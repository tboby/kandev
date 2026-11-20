---
status: draft
system: executors
created: 2026-11-20
owners:
  - kandev
---

# Remote Docker workspace retention requirements

## Overview

Remote Docker tasks can contain local commits, staged changes, and untracked
files. The executor owns retaining those files independently of the container
that runs the agent. This extends Remote Docker's existing stop and resume
contract to cover container loss.

## Requirements

### REQ-EXECUTORS-REMOTE-DOCKER-WORKSPACE-001: Preserve the checkout across runtime failures

**Intent:** Keep the task's local files available for recovery after a failed
resume or the loss of its container.

**User story:** As a Remote Docker task operator, I want to reopen the same
checkout after a runtime failure so that I can continue work I have yet to push.

#### Acceptance criteria

- **AC-EXECUTORS-REMOTE-DOCKER-WORKSPACE-001.1:** A prompt-readiness failure
  during recovery shall retain an established Remote Docker container and
  workspace. A failed or unconfirmed agent stop shall leave recovery retryable.
- **AC-EXECUTORS-REMOTE-DOCKER-WORKSPACE-001.2:** If the container for a new
  volume-backed environment is removed, a later resume shall attach the same
  verified workspace on the same daemon, preserving Git state and local files.
  Recovery shall honor existing prompt admission and delivery safeguards.
- **AC-EXECUTORS-REMOTE-DOCKER-WORKSPACE-001.3:** If storage identity or
  availability cannot be proven, recovery shall surface an actionable failure,
  leave the session outside STARTING once the attempt settles, and avoid
  provisioning an empty checkout in place of the original.
- **AC-EXECUTORS-REMOTE-DOCKER-WORKSPACE-001.4:** Reset Environment, archive,
  and delete shall remove owned workspace storage under their existing cleanup
  authority, including if the container is already absent. A stale cleanup
  operation shall leave successor resources intact.

## Out of scope

- Data restoration after both the container and its storage are gone.
- In-place migration of existing writable-layer checkouts.
- Cross-daemon replication, automatic commits or pushes, and changes to
  provider-native conversation restoration guarantees.

## System design

- [Remote Docker workspace retention](../system-design/remote-docker-workspace-retention.md)
