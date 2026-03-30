# ADR-0002: Agent Pre-Task Protocol — Sync, Read ADRs and Changelog, Run Tests

**Date:** 2026-03-19
**Status:** Accepted

## Context

This repository is developed with heavy AI coding agent involvement (GitHub Copilot, Claude, etc.). Without a standardised starting protocol, agents risk:

- Working from a stale branch and creating conflicts.
- Repeating or reversing decisions that were already deliberately made and recorded (for example the storage direction captured in ADR-0001 and ADR-0003).
- Breaking the existing test baseline before they have made any changes, making it impossible to distinguish pre-existing failures from regressions they introduced.
- Writing changelog entries inconsistent with the established format.

A consistent pre-task protocol solves all four risks.

## Decision

Every AI coding agent — and human contributor — **must** complete the following steps before making any code change:

1. **Sync with `main`**: fetch and rebase (or merge) the latest `origin/main` onto the working branch.
2. **Read `CHANGELOG.md`**: understand what has changed recently and in which direction the project is moving.
3. **Read all ADRs in `docs/adr/`**: understand why significant decisions were made so they are not accidentally reversed.
4. **Run the test suite** (`go test ./...`) to record the current passing baseline.

Additionally, upon every PR merge to `main`, a GitHub Actions workflow creates a structured issue for the Copilot coding agent to append a new changelog entry that records what changed, why, and when. This keeps the changelog useful as a machine-readable history for future agents.

## Consequences

**Positive:**
- Agents start every task with an accurate picture of the codebase state.
- Architectural decisions recorded in ADRs are not inadvertently reversed.
- The changelog grows as a reliable, append-only audit trail that agents can query to understand the current state of the project.
- Human reviewers gain confidence that agents are working from a consistent, known baseline.

**Negative:**
- Agents must spend a few extra steps reading before coding; however, these steps prevent far costlier rework.
- The automated changelog workflow creates one GitHub issue per merged PR. Repositories with a very high merge rate should monitor issue volume.

**Neutral:**
- The protocol applies equally to human contributors — it is good hygiene regardless of whether the author is human or AI.

## References

- `AGENTS.md` — root agent guidance file (enforces this protocol)
- `.github/copilot-instructions/project.md` — Copilot-specific agent instructions
- `.github/workflows/changelog-update.yml` — automated changelog update workflow
- `CHANGELOG.md` — the append-only project changelog
