---
status: draft
system: executors
requirements:
  - REQ-EXECUTORS-REMOTE-DOCKER-WORKSPACE-001
---

# Remote Docker workspace retention system design

## Purpose and boundaries

The executor owns storage and compute attachment. Task environments own
physical workspace identity and cleanup authority. Durable delivery continues
to own prompt admission and uncertain submissions. The existing Resume action
and recovery feedback expose outcomes.

## Requirement mapping

| Criterion | Design section |
| --- | --- |
| AC-EXECUTORS-REMOTE-DOCKER-WORKSPACE-001.1 | [Prompt-readiness recovery](#prompt-readiness-recovery) |
| AC-EXECUTORS-REMOTE-DOCKER-WORKSPACE-001.2 | [Container replacement](#container-replacement) |
| AC-EXECUTORS-REMOTE-DOCKER-WORKSPACE-001.3 | [Failure and recovery](#failure-and-recovery) |
| AC-EXECUTORS-REMOTE-DOCKER-WORKSPACE-001.4 | [Storage ownership and cleanup](#storage-ownership-and-cleanup) |

## Prompt-readiness recovery

Route the orchestrator's prompt-readiness stop through a shared reason and
the Remote Docker recoverable-stop policy. Confirm the exact agent has stopped
before retiring its execution slot. Preserve remote compute, environment binding,
and cleanup ownership when the stop cannot be confirmed. Use existing
cancellation and execution ownership checks before resuming. Keep the existing
behavior for other executors.

## Storage ownership and cleanup

Create a named Docker-managed volume on the saved daemon and mount it at
`/workspace` for newly provisioned environments. Use an environment-scoped name
and immutable labels for installation and environment identity. Persist the
allocation identity before provisioning through the existing materialization
owner. Carry it through environment inventory and durable cleanup snapshots.
Verify the labels on the trusted SSH target before attaching or deleting.
Use the current database owner and generation to authorize operations.

A missing legacy volume binding selects existing container-only behavior.
Keep ordinary stop, retry, and replacement rollback free of volume deletion.
Initial allocation rollback may remove only storage the attempt created when
no committed workspace or successor depends on it. Explicit destructive
cleanup removes the owned container first, then the volume. Treat confirmed
absence as idempotent success. A partial teardown retains the inventory needed
to retry. Volume cleanup must also work when there is no surviving execution.

## Container replacement

Try the recorded container first. Require a confirmed Docker not-found for the
exact recorded container and positively verified volume ownership before
allocating replacement compute. Use the existing task environment recovery
claim and generation fence to serialize replacement and publication. Use an
attempt-specific container identity. On a failed or abandoned replacement,
remove only that attempt's compute.

Prepare helpers and connection credentials through the existing Engine API
delivery path. Mount the verified volume without image copy-up and signal
reused-workspace provenance so launch validates repositories in place. Skip
fresh repository preparation and any destructive workspace initialization.
Publish the replacement container, connection metadata, and ownership through
the existing fenced environment update. Prevent late operations from clearing
or destroying the replacement binding.

Workspace retention covers `/workspace`. Container-local journals and harness
state outside it may be lost. Use platform recovery rules for missing journal
evidence. Block uncertain delivery and prohibit automatic replay or unsafe
legacy downgrade. A recoverable session may proceed only when existing
delivery and conversation admission permits it. The preserved checkout remains
available for an explicitly started session under existing environment rules.

## Failure and recovery

Ambiguous Docker responses, identity mismatch, and inaccessible or missing
storage enter the existing recovery error path. Settle the owned attempt
without leaving a stale STARTING state. Preserve authoritative inventory and
permit a later retry when evidence becomes available. Legacy environments
with missing containers surface the loss through existing feedback. Existing
legacy containers remain resumable.

## Security and observability

Use the existing trusted SSH endpoint and Docker Engine API. Never accept a
caller-supplied mount source or an arbitrary host path. Record bounded
failure reasons and environment, container, and execution identifiers in
structured logs without credentials, repository contents, or prompts.

## Related decisions

- [Independent Remote Docker workspace storage](../../../decisions/2026-11-20-remote-docker-workspace-storage.md)
- [Generation-fenced environment ownership](../../../decisions/2026-09-04-generation-fenced-task-environment-ownership.md)
- [Remote Docker executor](remote-docker-executor.md)
- [Durable journal ownership](../../platform/system-design/durable-agent-journal-ownership.md)
