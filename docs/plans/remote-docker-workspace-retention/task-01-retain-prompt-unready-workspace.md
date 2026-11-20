---
id: "01-retain-prompt-unready-workspace"
title: "Retain Remote Docker on prompt-readiness failure"
status: in_progress
wave: 1
depends_on: []
plan: "plan.md"
requirements:
  - REQ-EXECUTORS-REMOTE-DOCKER-WORKSPACE-001
acceptance_criteria:
  - AC-EXECUTORS-REMOTE-DOCKER-WORKSPACE-001.1
system_design:
  - ../../specs/executors/system-design/remote-docker-workspace-retention.md
---

# Task 01: Retain Remote Docker on prompt-readiness failure

## Scope

Connect the production prompt-readiness cleanup reason to the existing remote
runtime retention policy. Preserve unconfirmed-stop retry and exact-execution
ownership. Volume provisioning and UI are outside this work order.

## Acceptance

1. Prompt-readiness cleanup stops the failed agent without a remote container
   removal. Failed stop keeps ownership and supports retry.
2. Concurrent cancellation and stale cleanup cannot resume or remove successor
   work. Explicit teardown and other executors retain their current policies.

## Verification

Add the focused lifecycle case before changing the policy. Record the expected
RED assertion and GREEN result. Use the existing orchestrator ownership cases
for the production entry and cancellation. Run from the repository root.

```bash
(cd apps/backend && go test -trimpath -race ./internal/agent/runtime/lifecycle -run 'TestRemoteDockerPromptReadinessRecovery|TestRemoteDockerRecoverableFailurePreservesRuntime|TestRemoteDockerFailedResumePreservesRuntime|TestRemoteDockerExecutorPreservesRetainedWorkspace|TestRemoteDockerExecutorRemovesFreshRollbackAndExplicitCleanup' -count=1 -v)
(cd apps/backend && go test -trimpath -race ./internal/orchestrator -run 'TestReapPromptUnreadyExecution' -count=1 -v)
python3 scripts/list-docs.py validate
python3 scripts/lint-spec-files.py --all
node scripts/validate-public-docs.mjs
git diff --check
```

| Lifecycle scenario | Criterion and invariant | Evidence and result |
| --- | --- | --- |
| Pause and resume, API entry | .1, readiness recovery retains container and failed stop is retryable | First command passed with race detection |
| Stale or duplicate cleanup, cancellation, concurrent work | .1, only the claimed execution can settle and concurrent cancel prevents automatic resume | Second command blocked: compiler killed by memory limit on three attempts |
| Restart and UI entry | .1, retained checkout survives backend reattachment through Resume | Task 02 real-daemon command, pending |

## Files likely touched

- `apps/backend/internal/orchestrator/task_operations.go`
- `apps/backend/internal/agent/runtime/lifecycle/recovery_stop.go`
- `apps/backend/internal/agent/runtime/lifecycle/remote_recovery_stop.go`
- `apps/backend/internal/agent/runtime/lifecycle/manager_remote_docker_recoverable_failure_test.go`
- `apps/backend/internal/agent/runtime/lifecycle/remote_docker_stop_preserves_test.go`
- `apps/backend/internal/orchestrator/execution_teardown_ownership_test.go`
- `docs/public/executors.md` and the Remote Docker executor system design.

## Dependencies and risks

No prerequisite work order. Changing force alone can leave the unconfirmed
stop fallback destructive. Verify both the exact force decision and the
absence of container removal.

## Results

Pre-fix regression command:

```bash
(cd apps/backend && go test -trimpath -race ./internal/agent/runtime/lifecycle -run 'TestRemoteDockerPromptReadinessRecovery|TestRemoteDockerExecutorPreservesRetainedWorkspace/unconfirmed_stop_during_prompt' -count=1)
```

RED exited 1 with four expected assertion failures. The live execution cases
recorded Docker kill and delete requests for the original container. The
persisted-owner case invoked runtime cleanup despite confirmed agent absence.
The executor fallback case recorded a Docker stop after an unconfirmed stop.

Catalog validation, specification lint, public documentation validation, and
whitespace checks passed. Real-daemon restart and browser Resume evidence
remain in Task 02. This host has no Docker socket.

Changed-file Go lint was started with the installed Go toolchain but did not
finish on this host and was interrupted.
