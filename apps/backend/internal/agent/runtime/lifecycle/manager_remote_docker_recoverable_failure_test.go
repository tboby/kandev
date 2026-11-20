package lifecycle

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kandev/kandev/internal/agent/executor"
	agentctl "github.com/kandev/kandev/internal/agent/runtime/agentctl"
	"github.com/kandev/kandev/internal/agentruntime"
	"github.com/kandev/kandev/internal/task/models"
)

// @covers AC-EXECUTORS-REMOTE-DOCKER-WORKSPACE-001.1
func TestRemoteDockerPromptReadinessRecovery(t *testing.T) {
	for _, stopFails := range []bool{false, true} {
		name := "agent stops"
		if stopFails {
			name = "unconfirmed stop retains ownership for retry"
		}
		t.Run(name, func(t *testing.T) {
			engine, dockerServer := newRecordingDockerEngine(t)
			log := newTestRegistryLogger()
			registry := NewExecutorRegistry(log)
			registry.Register(newRemoteDockerTestExecutor(t, dockerServer))
			mgr := NewManager(newTestRegistry(), &MockEventBus{}, registry, nil, nil, nil, ExecutorFallbackWarn, "", log)
			cleanupManagerStopCh(t, mgr)
			var failStop atomic.Bool
			failStop.Store(stopFails)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if failStop.Load() {
					w.WriteHeader(http.StatusInternalServerError)
				}
				_, _ = w.Write([]byte(`{"success":true}`))
			}))
			t.Cleanup(server.Close)
			host, port := parseHTTPURL(t, server.URL)
			execution := &AgentExecution{
				ID: "instance-1", TaskID: "task-1", SessionID: "session-1",
				RuntimeName: executor.NameRemoteDocker,
				agentctl:    agentctl.NewClient(host, port, log),
				ContainerID: "container-1",
			}
			require.NoError(t, mgr.executionStore.Add(execution))

			const reason = StopReasonPromptReadinessRecovery
			err := mgr.StopAgentWithReason(t.Context(), execution.ID, reason, true)
			require.Empty(t, engine.snapshot(), "prompt-readiness cleanup must retain the remote container")
			if stopFails {
				require.ErrorIs(t, err, errRecoverableStopUnconfirmed)
				current, exists := mgr.executionStore.Get(execution.ID)
				require.True(t, exists)
				require.Same(t, execution, current)
				failStop.Store(false)
				err = mgr.StopAgentWithReason(t.Context(), execution.ID, reason, true)
			}
			require.NoError(t, err)
			require.Empty(t, engine.snapshot(), "retry must also retain the container")
			_, exists := mgr.executionStore.Get(execution.ID)
			require.False(t, exists, "confirmed stop releases the execution slot")
		})
	}
}

// @covers AC-EXECUTORS-REMOTE-DOCKER-WORKSPACE-001.1
func TestRemoteDockerPromptReadinessRecoveryWithPendingOwner(t *testing.T) {
	row := &models.ExecutorRunning{
		ID: "session-pending", SessionID: "session-pending", TaskID: "task-pending",
		AgentExecutionID: "execution-pending", Runtime: agentruntime.RuntimeRemoteDocker,
	}
	backend := &task08DetailedRecoveryBackend{
		MockExecutor: &MockExecutor{name: executor.NameRemoteDocker},
		recoveryResult: func([]*models.ExecutorRunning) ([]*ExecutorInstance, map[string]RecoveryCandidateOutcome, error) {
			return nil, map[string]RecoveryCandidateOutcome{row.SessionID: RecoveryOutcomeNoMatchingInstance}, nil
		},
	}
	mgr, _ := task08ManagerWithBackend(t, backend)
	mgr.SetExecutorRunningWriter(&task08ExecutorInventoryWriter{rows: []*models.ExecutorRunning{row}})
	task08SetSnapshots(mgr, []*models.ExecutorRunning{row})
	mgr.queueRemoteRecovery(row, task08RemoteSnapshot(row))

	require.NoError(t, mgr.StopAgentWithReason(t.Context(), row.AgentExecutionID, StopReasonPromptReadinessRecovery, true))
	require.Empty(t, backend.stopInstanceCalls, "proven agent absence must not force cleanup of the remote container")
	require.NoError(t, mgr.recoveryGuard.CheckLaunchAllowed(row.SessionID))
}

func newAgentctlStatusServer(t *testing.T, status int) (host string, port int) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	t.Cleanup(server.Close)
	return parseHTTPURL(t, server.URL)
}

// @covers AC-EXECUTORS-REMOTE-DOCKER-001.12
func TestRemoteDockerRecoverableFailurePreservesRuntime(t *testing.T) {
	for _, tc := range []struct {
		name     string
		reason   string
		resumed  bool
		attached bool
		force    bool
		want     bool
	}{
		{"established recoverable failure", StopReasonRecoverableAgentFailure, false, false, true, false},
		{"resumed recoverable failure", StopReasonRecoverableAgentFailure, true, true, true, false},
		{"reused bootstrap failure", StopReasonAgentBootstrapFailed, false, true, true, false},
		{"fresh bootstrap rollback", StopReasonAgentBootstrapFailed, false, false, true, true},
		{"conversation resume on fresh compute", StopReasonAgentBootstrapFailed, true, false, true, true},
		{"archive remains destructive", StopReasonTaskArchived, true, true, true, true},
		{"delete remains destructive", StopReasonTaskDeleted, true, true, true, true},
		{"explicit force remains destructive", "", true, true, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			log := newTestRegistryLogger()
			registry := NewExecutorRegistry(log)
			backend := &stopTrackingExecutor{MockExecutor: MockExecutor{name: executor.NameRemoteDocker}}
			registry.Register(backend)
			mgr := NewManager(newTestRegistry(), &MockEventBus{}, registry, nil, nil, nil, ExecutorFallbackWarn, "", log)
			cleanupManagerStopCh(t, mgr)
			host, port := newAgentctlStatusServer(t, http.StatusOK)
			execution := &AgentExecution{
				ID: "execution-1", TaskID: "task-1", SessionID: "session-1",
				RuntimeName: executor.NameRemoteDocker, isResumedSession: tc.resumed,
				attachedExistingCompute: tc.attached,
				agentctl:                agentctl.NewClient(host, port, log),
				ContainerID:             "container-1",
			}
			require.NoError(t, mgr.executionStore.Add(execution))

			require.NoError(t, mgr.StopAgentWithReason(context.Background(), execution.ID, tc.reason, tc.force))
			require.Equal(t, []bool{tc.want}, backend.forces)
			_, stillTracked := mgr.executionStore.Get(execution.ID)
			require.False(t, stillTracked, "a successful stop releases the in-memory execution slot")
		})
	}
}

// @covers AC-EXECUTORS-REMOTE-DOCKER-001.12
func TestRemoteDockerFailedResumePreservesRuntime(t *testing.T) {
	log := newTestRegistryLogger()
	registry := NewExecutorRegistry(log)
	backend := &stopTrackingExecutor{MockExecutor: MockExecutor{name: executor.NameRemoteDocker}}
	registry.Register(backend)
	mgr := NewManager(newTestRegistry(), &MockEventBus{}, registry, nil, nil, nil, ExecutorFallbackWarn, "", log)
	cleanupManagerStopCh(t, mgr)
	host, port := newAgentctlStatusServer(t, http.StatusOK)
	execution := &AgentExecution{
		ID: "execution-1", TaskID: "task-1", SessionID: "session-1",
		RuntimeName: executor.NameRemoteDocker, isResumedSession: true,
		attachedExistingCompute: true,
		agentctl:                agentctl.NewClient(host, port, log),
		ContainerID:             "container-1",
		metadata:                map[string]interface{}{"retained": "evidence"},
	}
	require.NoError(t, mgr.executionStore.Add(execution))

	require.NoError(t, mgr.StopAgentWithReason(context.Background(), execution.ID, StopReasonAgentBootstrapFailed, true))
	require.Equal(t, []bool{false}, backend.forces,
		"a resumed attempt that attached to retained compute must not force-remove it")
	require.Equal(t, map[string]interface{}{"retained": "evidence"}, execution.MetadataSnapshot())
}

// @covers AC-EXECUTORS-REMOTE-DOCKER-001.12
func TestRemoteDockerRecoverableStopFailureRetainsOwner(t *testing.T) {
	log := newTestRegistryLogger()
	registry := NewExecutorRegistry(log)
	backend := &stopTrackingExecutor{MockExecutor: MockExecutor{name: executor.NameRemoteDocker}}
	registry.Register(backend)
	mgr := NewManager(newTestRegistry(), &MockEventBus{}, registry, nil, nil, nil, ExecutorFallbackWarn, "", log)
	cleanupManagerStopCh(t, mgr)
	host, port := newAgentctlStatusServer(t, http.StatusInternalServerError)
	execution := &AgentExecution{
		ID: "execution-1", TaskID: "task-1", SessionID: "session-1",
		RuntimeName: executor.NameRemoteDocker,
		agentctl:    agentctl.NewClient(host, port, log),
		ContainerID: "container-1",
	}
	require.NoError(t, mgr.executionStore.Add(execution))

	err := mgr.StopAgentWithReason(context.Background(), execution.ID, StopReasonRecoverableAgentFailure, true)
	require.ErrorIs(t, err, errRecoverableStopUnconfirmed)
	require.Empty(t, backend.forces, "an unconfirmed stop must not reach destructive runtime cleanup")
	current, exists := mgr.executionStore.Get(execution.ID)
	require.True(t, exists, "the failed stop keeps its exact cleanup ownership for retry")
	require.Same(t, execution, current)
}

// @covers AC-EXECUTORS-REMOTE-DOCKER-001.12
func TestRemoteDockerResumeRollbackPreservesRuntime(t *testing.T) {
	for _, tc := range []struct {
		name    string
		resumed bool
	}{
		{"resumed session with previous execution", true},
		{"workspace-only attach without previous execution", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			log := newTestRegistryLogger()
			registry := NewExecutorRegistry(log)
			backend := &stopTrackingExecutor{MockExecutor: MockExecutor{name: executor.NameRemoteDocker}}
			registry.Register(backend)
			mgr := NewManager(newTestRegistry(), &MockEventBus{}, registry, nil, nil, nil, ExecutorFallbackWarn, "", log)
			cleanupManagerStopCh(t, mgr)
			execution := &AgentExecution{
				ID: "execution-1", TaskID: "task-1", SessionID: "session-1",
				RuntimeName: executor.NameRemoteDocker, isResumedSession: tc.resumed,
				attachedExistingCompute: true,
				ContainerID:             "container-1",
			}
			require.NoError(t, mgr.executionStore.Add(execution))

			execInstance := &ExecutorInstance{InstanceID: execution.ID, ContainerID: "container-1"}
			require.NoError(t, mgr.stopRegisteredLaunchRuntime(backend, execInstance, execution))
			require.Equal(t, []bool{false}, backend.forces)
			require.Equal(t, StopReasonAgentBootstrapFailed, execInstance.StopReason,
				"a retained rollback must keep a preserving stop reason")
			mgr.finishRegisteredLaunchRollback(execution, false, true)
			_, exists := mgr.executionStore.Get(execution.ID)
			require.False(t, exists, "the failed attempt releases its in-memory slot")
		})
	}
}
