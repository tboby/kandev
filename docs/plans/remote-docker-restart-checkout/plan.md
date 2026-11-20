---
created: 2026-11-20
status: draft
requirements:
  - REQ-EXECUTORS-REMOTE-DOCKER-001
system_design:
  - ../../specs/executors/system-design/remote-docker-executor.md
legacy_specs: []
---

# Remote Docker restart checkout

Resume of a stopped Remote Docker container goes through
`DockerExecutor.ensureContainerRunning`. Docker replays the creation-time
command and `buildContainerEntrypoint` evaluates the stored prepare script
again. The built-in template unconditionally clones into its existing
workspace. The prepare error is logged while agentctl can still start, so
session readiness alone cannot prove the fix.

Two consecutive clones of a local fixture, using the built-in clone flags and
the same workspace, reproduced Git's nonempty destination error. The
executor system owns the retained container and its checkout preparation.
This package implements AC-EXECUTORS-REMOTE-DOCKER-001.7 and
AC-EXECUTORS-REMOTE-DOCKER-001.12.

## Executor comparison

- Kubernetes can prepare a replacement pod against a retained PVC. Its built-in
  script confirms the checkout root and normalized origin, reuses the existing
  checkout, and fails on unexpected contents. Reuse those checks and GitHub URL
  normalization for Remote Docker. Kubernetes stages a fresh clone outside the
  mount and allows `lost+found` on a fresh PVC. Remote Docker can keep cloning
  directly into its container workspace.
- SSH bypasses preparation when an execution must reuse the existing task
  directory. Docker restart runs the original entrypoint again, so the guard
  belongs in the preparation script that the container starts with.
- Local Docker shares the built-in template and stopped-container restart
  path, so it can encounter the same duplicate clone. This work order keeps
  the change scoped to Remote Docker.

## Dependency order

1. [Preserve the checkout on container restart](task-01-preserve-checkout.md).

## Scope and risks

Keep this change within Remote Docker's built-in preparation. Exact stored
copies of the built-in profile template must also receive the runtime fix at
container creation. Docker retains the original command and environment on
existing containers, so already-created containers cannot acquire this change
on restart. Keep those workspaces intact and leave reclamation and replacement
policy as-is.

The work order specifies the shell-backed Go regression and container-backed
Playwright check. No rendered UI changes are needed.
