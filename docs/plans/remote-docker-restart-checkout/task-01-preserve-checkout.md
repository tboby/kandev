---
id: "01-preserve-checkout"
title: "Preserve the checkout on container restart"
status: in_progress
wave: 1
depends_on: []
plan: "plan.md"
requirements:
  - REQ-EXECUTORS-REMOTE-DOCKER-001
acceptance_criteria:
  - AC-EXECUTORS-REMOTE-DOCKER-001.7
  - AC-EXECUTORS-REMOTE-DOCKER-001.12
system_design:
  - ../../specs/executors/system-design/remote-docker-executor.md
---

# Preserve the checkout on container restart

## Scope

Resolve the built-in Remote Docker preparation in
`apps/backend/internal/agent/runtime/lifecycle/docker_launch.go`, using the
existing template in `default_scripts.go` and checking any built-in checkout
transform in `repository_checkout_options.go`. Keep custom scripts and Local
Docker behavior unchanged. Existing containers with frozen startup commands
are outside the script upgrade. Preserve their existing reuse and cleanup
semantics.

## Acceptance

1. Fresh preparation materializes the configured repository and branch.
2. Repeating built-in preparation against the same retained checkout succeeds
   without another clone and preserves the active task branch, unpublished
   commit and working-tree edits. Exercise both blank and exact stored default
   profiles. An invalid or unrelated occupied workspace stays intact and
   preparation fails.
3. A browser Resume of the same stopped Remote Docker container accepts a
   follow-up turn, preserves the container and checkout, and produces no
   duplicate clone attempt or new prepare failure in restart logs.

## Verification and lifecycle evidence

Implement `TestRemoteDockerPrepareScriptRetainedCheckout` alongside the
lifecycle source. Execute the normally resolved script twice using a temporary
Git origin and a temporary workspace. Capture the expected nonempty destination
error in the red run. In the green run assert successful preparation and use a
clone counter or Git trace to prove the clone runs once. Use a fixture setup hook that appends to a log outside the
checkout and assert it runs once at launch and again on replay. This proves
the guard skips cloning while preserving setup-hook behavior. Add fresh,
retained and invalid workspace subtests. Cover custom-script and Local
Docker resolution where the changed branch could affect them.

Extend `apps/web/e2e/tests/remote-docker/remote-docker-recovery.spec.ts` with
the test named "Resume restarts the same container without cloning again".
Stop the session, record a checkpoint for container logs, stop the exact Docker
container, use the existing browser Resume flow and verify the new turn and
retained checkout. Existing running-container reattachment coverage is useful
alongside this scenario because it does not exercise startup replay.

| Scenario | Criterion and observable invariant | Evidence and result |
| --- | --- | --- |
| Pause and resume | .12: same container and unpublished changes survive intentional stop and Resume | Playwright test below. Pending. |
| Restart and recovery | .12: Docker startup replay succeeds against retained data | Playwright and Go commands below. Pending. Backend restart is covered by existing recovery paths, whose ownership state is outside this change. |
| Duplicate and stale delivery | .12: repeated preparation invokes clone once | Go command below. Passed. Journal delivery and stale owner rules are outside the prepare step. |
| Entry paths | .7 and .12: normal launch resolves the guarded script and browser Resume reaches stopped-container startup | Go and Playwright commands below. Pending. |
| Ownership changes | Excluded. The guard touches checkout preparation only. Session claims and container ownership are resolved before startup. | None. |
| Concurrent or merged work | Excluded. Docker replays one entrypoint per start of the identified container. The change introduces no concurrent prepare runner or work merge. | None. |

Run from the repository root:

```sh
(cd apps/backend && go test ./internal/agent/runtime/lifecycle -run '^TestRemoteDockerPrepareScriptRetainedCheckout$' -count=1)
(cd apps/web && KANDEV_E2E_CONTAINERS=1 pnpm e2e:run --host --shards 1 --project containers tests/remote-docker/remote-docker-recovery.spec.ts)
```

Run the Go command before and after the fix. The Playwright check requires
Docker, the existing SSH fixture and the web workspace dependencies.

## Results

- The focused Go test failed against the original sources with exit 128 on
  the second clone for both profile cases. The same test passed with the fix.
- The adjacent Go run hit the Kubernetes fixture's root check because the
  test temp directory was inside the repository. A retry with an external
  test temp directory is pending.
- Specification and public documentation validators passed. The browser spec
  passed Prettier. Playwright stopped at the missing `make` executable
  before test execution. The runner also has no Docker daemon. Browser
  evidence for pause, restart and the Resume entry path remains pending.
