---
id: "01-remote-userns"
title: "Enable remote Docker user namespaces"
status: in_progress
wave: 1
depends_on: []
plan: "plan.md"
requirements:
  - REQ-EXECUTORS-REMOTE-USERNS-001
acceptance_criteria:
  - AC-EXECUTORS-REMOTE-USERNS-001.1
  - AC-EXECUTORS-REMOTE-USERNS-001.2
  - AC-EXECUTORS-REMOTE-USERNS-001.3
  - AC-EXECUTORS-REMOTE-USERNS-001.4
  - AC-EXECUTORS-REMOTE-USERNS-001.5
system_design:
  - ../../specs/executors/system-design/remote-docker-user-namespaces.md
---

# Task 01: Enable Remote Docker User Namespaces

## Summary

Expose the existing option on remote Docker profiles and carry its operator-controlled value to newly created remote containers. Verify the actual sandbox behavior and document host prerequisites.

## In scope

- Create and edit UI, serialization, authoritative metadata, remote launch and public documentation.
- Targeted unit and real-daemon browser coverage, including mobile settings.

## Out of scope

- Changing host configuration, daemon user ID remapping, or the existing seccomp profile.

## Acceptance

- Remote profiles persist the setting and display existing security and reset guidance.
- New remote containers receive the correct security options; task and agent inputs cannot override operator choice.
- A compatible remote host runs the sandbox probe and failures remain visible.

## ASCII UI preview

UI-01: See the [plan preview](plan.md#ascii-ui-preview); the same card is full width on phone. Covers .1 and .3.

```text
User namespace support                 [off/on]
Allow compatible sandbox tools to run
Changes apply to new environments
Security warning
```

## Verification

Run from the repository root after adding focused cases to these existing suites and a remote sandbox case to the existing remote Docker spec:

```bash
(cd apps/backend && go test ./internal/orchestrator/executor ./internal/agent/runtime/lifecycle ./internal/mcp/handlers)
(cd apps/web && pnpm exec vitest run components/settings/profile-edit/serialize-executor-config.test.ts components/settings/profile-edit/use-user-namespaces-form-state.test.ts)
(cd apps/web && pnpm e2e:run tests/settings/docker-profile-persistence.spec.ts tests/settings/mobile-user-namespaces.spec.ts)
(cd apps/web && KANDEV_E2E_CONTAINERS=1 pnpm e2e:raw --project=containers tests/remote-docker/remote-docker-task.spec.ts)
python3 scripts/list-docs.py validate
python3 scripts/lint-spec-files.py --all
git diff --check
```

## Files likely touched

- `apps/web/components/settings/profile-edit/`
- `apps/web/app/settings/executors/`
- `apps/backend/internal/orchestrator/executor/`
- `apps/backend/internal/agent/runtime/lifecycle/`
- `apps/web/e2e/tests/remote-docker/`
- `docs/public/executors.md` and `docs/public/security.md`

## Dependencies

None.

## Risks

The real daemon test requires a host allowing nested namespaces. A skipped or blocked probe is not evidence of acceptance.

## Parallelism

Sequential.

## Inputs

The linked requirement, design, security decision and existing Docker namespace tests.

## Results

- Host compatibility probe on Mellifera: `unshare --user --map-root-user true`
  succeeded in a disposable, network-isolated `kandev/multi-agent:latest`
  container with the generated Kandev seccomp profile and
  `apparmor=unconfined`. The same command with Docker defaults failed with
  `Operation not permitted`. Both containers used `--rm`. This is host
  compatibility evidence, not proof of the Kandev remote launch path.
- Specification validation passed: `python3 scripts/list-docs.py validate` and
  `python3 scripts/lint-spec-files.py --all`.
- The two specified frontend unit suites passed (21 tests).
