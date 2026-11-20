---
id: "02-recover-volume-backed-workspace"
title: "Persist and recover the Remote Docker workspace"
status: pending
wave: 2
depends_on:
  - "01-retain-prompt-unready-workspace"
plan: "plan.md"
requirements:
  - REQ-EXECUTORS-REMOTE-DOCKER-WORKSPACE-001
acceptance_criteria:
  - AC-EXECUTORS-REMOTE-DOCKER-WORKSPACE-001.2
  - AC-EXECUTORS-REMOTE-DOCKER-WORKSPACE-001.3
  - AC-EXECUTORS-REMOTE-DOCKER-WORKSPACE-001.4
system_design:
  - ../../specs/executors/system-design/remote-docker-workspace-retention.md
---

# Task 02: Persist and recover the Remote Docker workspace

## Scope

Implement environment-owned Engine volume provisioning, verified replacement,
generation-fenced cleanup, and existing recovery feedback. Update executor
documentation. Legacy migration and container-external backup are outside scope.

## Acceptance

1. On a disposable remote daemon, remove the task container and restart the
   backend. Resume from the UI. Verify a new container ID and the same volume,
   unpublished commit, staged change, and untracked marker.
2. Missing, mismatched, or ambiguous storage settles as an actionable recovery
   failure with no replacement checkout or replay of uncertain delivery.
3. Explicit teardown removes owned storage when the container is absent.
   Retry partial cleanup and prove stale or concurrent owners cannot delete or
   publish over the successor.

## Verification

Extend the existing Remote Docker Playwright fixture and target the actual
Resume control. Cover both prompt-readiness retention after a backend restart
and replacement after Engine removal. Keep a separate blocked case for missing
storage. Use current desktop and phone feedback specs for parity. Run commands
from the root, sequentially. Record expected RED assertions and actual results.

```bash
(cd apps/backend && go test -trimpath -race ./internal/agent/docker ./internal/agent/runtime/lifecycle ./internal/orchestrator/executor ./internal/task/repository/sqlite ./internal/task/service -run 'TestRemoteDockerWorkspaceVolume' -count=1)
(cd apps/web && KANDEV_E2E_CONTAINERS=1 pnpm e2e:run --host --shards 1 --project containers tests/remote-docker/remote-docker-task.spec.ts tests/remote-docker/remote-docker-recovery.spec.ts)
(cd apps/web && pnpm e2e:run --host --shards 1 --project chromium tests/session/session-error-recovery-ui.spec.ts)
(cd apps/web && pnpm e2e:run --host --shards 1 --project mobile-chrome tests/session/mobile-session-error-recovery-ui.spec.ts)
node scripts/validate-public-docs.mjs
python3 scripts/list-docs.py validate
python3 scripts/lint-spec-files.py --all
git diff --check
```

| Lifecycle scenario | Criterion and invariant | Evidence and result |
| --- | --- | --- |
| Pause/resume and restart | .2, replacement retains exactly the former checkout after reattachment | Real-daemon Resume scenario, second command, pending |
| Duplicate delivery | .2/.3, missing journal proof cannot produce an automatic replay | Admission and blocked-state cases, first and second commands, pending |
| Ownership changes and stale work | .4, old generation and repeated cleanup cannot delete current storage | Cleanup and reset/archive/delete cases, first and second commands, pending |
| Concurrent work | .2/.4, competing recoveries yield one committed binding and rollback leaves volume intact | Concurrent recovery and partial teardown tests, first command, pending |
| UI, API, and runtime entry | .2/.3, Resume and the executor agree on verified attachment and blocked settlement | First four commands, pending |

No scheduling or workflow-trigger changes are planned. Test fixture readiness
and installation failures are blocked evidence, not passing scenarios.

## Files likely touched

- `apps/backend/internal/agent/docker/` and `apps/backend/internal/agent/runtime/lifecycle/executor_remote_docker.go`
- `apps/backend/internal/orchestrator/executor/` and task environment model, repository, and cleanup code
- `apps/web/e2e/tests/remote-docker/remote-docker-recovery.spec.ts`
- `docs/public/executors.md`

## Dependencies and risks

Depends on Task 01. Engine absence must be distinguished from transport or
permission failure. Verify actual mounted storage after creation and before
workspace access. A partial teardown must keep enough saved identity to retry
without a live execution. Record exact commands and fixture teardown.

## Results

Pending implementation.
