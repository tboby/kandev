# ADR-2026-11-20-remote-docker-workspace-storage: Store Remote Docker workspaces independently of containers

**Status:** proposed
**Date:** 2026-11-20
**Area:** backend

## Context

A checkout in a Remote Docker container's writable layer disappears when the
container is removed. Kandev can preserve a container during agent recovery,
but the daemon or an operator can still remove it independently.

## Decision

Use a named, environment-owned Docker volume on the original daemon for
`/workspace` on newly provisioned Remote Docker environments. Keep the
current task environment generation as the cleanup authority. A replacement
container may reuse storage only after both confirmed container loss and
positive storage ownership validation.

## Consequences

Container replacement can retain local work. Environment teardown must track
storage separately and survive partial cleanup. The design retains
container-only compatibility for existing environments. Daemon or volume
loss still requires separate backup and restore facilities.

## Alternatives considered

- **Keep the writable layer and narrow cleanup authority.** This addresses
  Kandev-initiated removal and leaves the same risk from external removal.
- **Bind-mount a directory on the remote host.** This introduces host-path
  configuration and cleanup and conflicts with the Engine API provisioning model.
- **Checkpoint Git automatically.** Git checkpoints leave non-repository files
  and working tree state unprotected and change Git history or publication behavior.
