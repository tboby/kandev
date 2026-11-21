---
status: draft
system: executors
requirements:
  - REQ-EXECUTORS-REMOTE-USERNS-001
created: 2026-11-20
owners:
  - kandev
---

# Remote Docker User Namespace System Design

## Boundary and mapping

The existing [security decision](../../../decisions/2026-08-18-executor-userns-security-options.md) defines the security options and authority boundary. Extend its per-profile option to remote Docker without changing its meaning.

| Criteria | Design |
| --- | --- |
| .1, .3 | Profile editor and create form |
| .2, .4 | Profile propagation and container creation |
| .5 | Real remote daemon evidence |

## Profile editor and create form

Use the existing localized user namespace card for both Docker types in create and edit flows. Serialize `allow_user_namespaces` for either Docker type, remove it when disabled, and preserve the existing reset and warning copy. Reuse the existing responsive card and switch semantics.

## Profile propagation and container creation

Trace the saved profile through `profileConfigAuthoritativeKeys`, `docker_launch.go` and the remote executor's shared container manager. Ensure only the profile selects the option, including missing-profile and stale-profile paths. Preserve the MCP operator-only key guard. `container.go` owns the existing tailored seccomp and AppArmor options. Existing containers cannot change their security options; reset creates a new one. Do not silently retry with weaker or different options if Docker rejects them.

## Real remote daemon evidence

Exercise the remote Docker path using the existing containers fixture and a sandbox command, alongside disabled and profile persistence checks. Host kernel and daemon policy can independently prevent namespaces; surface the failure and document that prerequisite rather than modifying the host.
