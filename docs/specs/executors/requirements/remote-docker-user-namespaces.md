---
status: draft
system: executors
created: 2026-11-20
owners:
  - kandev
---

# Remote Docker User Namespace Requirements

## Overview

Operators can enable the existing Docker user namespace support for remote Docker profiles, allowing compatible sandboxed tools to create user namespaces inside newly created task containers.

## Requirements

### REQ-EXECUTORS-REMOTE-USERNS-001: Remote Docker user namespaces

#### Acceptance criteria

- **AC-EXECUTORS-REMOTE-USERNS-001.1:** An operator shall be able to enable and disable user namespace support when creating or editing a remote Docker profile, and the saved value shall survive reload.
- **AC-EXECUTORS-REMOTE-USERNS-001.2:** The setting shall default to disabled. When enabled, a newly created remote Docker container shall receive the same namespace security options as a local Docker container with the setting enabled. When disabled, it shall use Docker's default security options.
- **AC-EXECUTORS-REMOTE-USERNS-001.3:** A change shall apply to newly created containers; existing environments shall retain their configuration until reset. The interface shall communicate this and the existing security tradeoff on desktop and mobile.
- **AC-EXECUTORS-REMOTE-USERNS-001.4:** Task metadata shall not override the profile setting, and agent-facing profile management shall not enable the setting.
- **AC-EXECUTORS-REMOTE-USERNS-001.5:** On a compatible remote Docker host, an enabled container shall run a user namespace sandbox; failures caused by host policy shall remain visible.

## Exclusions

Configuring the remote daemon or host kernel, daemon-wide user ID remapping, and changing the existing security profile are outside this capability.

## System design

See [Remote Docker user namespaces](../system-design/remote-docker-user-namespaces.md).
