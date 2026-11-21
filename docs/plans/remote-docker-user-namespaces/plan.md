---
created: 2026-11-20
status: draft
requirements:
  - REQ-EXECUTORS-REMOTE-USERNS-001
system_design:
  - ../../specs/executors/system-design/remote-docker-user-namespaces.md
legacy_specs: []
---

# Implementation Plan: Remote Docker User Namespaces

## Overview

Extend the existing per-profile Docker setting to remote Docker in one vertical work order, then verify persistence and execution on a compatible remote daemon.

## Scope

Include profile creation and editing, profile authority, remote container security options, documentation, and desktop/mobile evidence. Host configuration and daemon-wide user ID remapping are excluded.

## Technical approach

Reuse the existing card, config key, launch metadata and shared container security profile. Update the local-only UI predicates and verify the remote path carries the authoritative value. Preserve the MCP restriction and default-off behavior. Update public executor and security documentation to name remote Docker.

## ASCII UI preview

UI-01: Remote Docker profile create/edit, desktop and phone. The existing card flows in the settings column; phone uses the same full-width control and wrapping copy. The warning and reset guidance remain visible. Structure is required; spacing is illustrative. Maps to .1 and .3.

```text
User namespace support                 [off/on]
Allow compatible sandbox tools to run
Changes apply to new environments
Security warning
```

## Tests

Extend `serialize-executor-config.test.ts` and create-profile config tests for enabled, disabled and non-Docker cases (.1, .2). Extend authoritative metadata and remote container tests for default, enabled, task override and missing profile (.2, .4). Verify security options and error propagation (.2, .5).

## E2E tests

Extend settings persistence and mobile user namespace Playwright coverage for remote Docker (.1, .3). Add a remote Docker containers-project scenario that runs the existing namespace sandbox probe on a compatible host (.5), including reset behavior (.3).

## Work orders

- [ ] [Task 01: Enable remote Docker user namespaces](task-01-remote-userns.md)

## Verification results

Pending.

## Risks

- Host kernel policy may block namespace creation even with the Docker security options; real-daemon evidence must report the host conditions.
- Applying task metadata or dropping an existing profile value on save would weaken the operator authority boundary.
