---
created: 2026-11-20
status: draft
requirements:
  - REQ-EXECUTORS-REMOTE-DOCKER-WORKSPACE-001
system_design:
  - ../../specs/executors/system-design/remote-docker-workspace-retention.md
legacy_specs: []
---

# Implementation plan: Remote Docker workspace retention

## Overview

First close the prompt-readiness cleanup gap. Then make the checkout survive
container removal through environment-owned storage and prove recovery through
the normal Resume flow on a disposable SSH-to-Docker daemon.

The [requirements](../../specs/executors/requirements/remote-docker-workspace-retention.md)
define the workspace guarantee. The [system design](../../specs/executors/system-design/remote-docker-workspace-retention.md)
defines storage ownership, recovery, and teardown.

## Evidence

The report describes Docker kill/remove events and a stale container binding.
No incident backend log or context bundle path was supplied. The precise
historical trigger remains unknown. In this checkout,
`stopPromptUnreadyExecution` passes force with the reason
`prompt readiness recovery`. The Remote Docker retention predicate recognizes
recoverable agent failure and retained bootstrap failure. The executor's
force path removes the container. Its required-workspace reconnect path blocks
replacement when the recorded container is absent. The earlier
[recoverable resume fix](../remote-docker-recoverable-resume/plan.md) covers
different stop reasons and currently relies on retention of the container.

## Scope

- Exact-execution cleanup after prompt-readiness failure and concurrent cancel.
- New environment volume provisioning, verified replacement, and fenced cleanup.
- Existing Resume feedback for missing or ambiguous evidence.
- Disposable real-daemon regression and user-facing executor documentation.

Existing writable-layer environments keep their container-only behavior.
Recovery of already deleted data and stronger provider conversation guarantees
are outside scope. The proposed storage scope is the checkout at `/workspace`.

## Technical approach

Keep retention policy in runtime lifecycle and assignment authority in task
environment recovery. Extend the Engine API wrapper for named volumes.
Use the task environment inventory and existing generation-fenced operations,
keeping container and volume cleanup separately retryable. Keep the existing
Resume interface and use its existing feedback surfaces.

## Work orders

- [ ] [Task 01: Retain Remote Docker on prompt-readiness failure](task-01-retain-prompt-unready-workspace.md)
- [ ] [Task 02: Persist and recover the Remote Docker workspace](task-02-recover-volume-backed-workspace.md)

Execute Task 01 before Task 02. Both require RED evidence before permanent code
and exact targeted results in their work orders.

## Risks

- An unavailable daemon can resemble container absence if errors are flattened.
- A replacement can keep workspace files while losing its prior delivery
  journal. Pending prompt admission must remain fenced.
- Docker can create a missing volume during container creation. Ensure the
  mounted resource is the verified volume and abort on disappearance or mismatch
  before repository access or publication.
- Older checkouts require their original containers until explicitly replaced.

## Verification results

- Catalog validation passed for 459 decisions and 2072 specifications.
- Specification lint and `git diff --check` passed.
- Implementation and regression evidence are pending.
