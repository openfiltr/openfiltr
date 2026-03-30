# ADR-0004: Keep Changelog Updates Inside the Implementation PR

**Date:** 2026-03-30
**Status:** Accepted

## Context

ADR-0002 introduced a post-merge GitHub Actions workflow that opened a follow-up GitHub issue for every merged pull request so an agent could append a changelog entry later.

That workflow created unnecessary noise:

- every merged PR produced another issue with no product value of its own
- maintainers had to triage or ignore repetitive changelog-only issues
- the repository already requires implementation PRs to update `CHANGELOG.md` before merge through contributor guidance and `.github/workflows/changelog-guard.yml`

The extra automation duplicates an existing gate and shifts changelog maintenance away from the pull request that actually made the change.

## Decision

OpenFiltr will keep changelog updates inside the implementation PR that introduces the change.

The repository should:

- remove `.github/workflows/changelog-update.yml`
- stop creating post-merge changelog issues
- continue to require changelog updates in non-trivial implementation PRs through the existing guard workflow and repository guidance

This decision supersedes only the post-merge changelog issue automation described in ADR-0002. The pre-task protocol in ADR-0002 remains in force.

## Consequences

**Positive:**
- merged PRs stop generating follow-up changelog housekeeping issues
- the changelog stays coupled to the code and docs change that required it
- reviewers can assess the changelog text before merge instead of after the fact

**Negative:**
- contributors must keep the changelog entry accurate before merge rather than relying on a later cleanup pass
- maintainers lose the fallback automation that previously reminded agents to backfill missing changelog text after merge

**Neutral:**
- `CHANGELOG.md` remains required reading before edits
- `.github/workflows/changelog-guard.yml` remains the enforcement mechanism for non-trivial pull requests

## References

- [ADR-0002](0002-agent-pre-task-protocol.md)
- `.github/workflows/changelog-guard.yml`
- `docs/codex-agents/README.md`
- `CHANGELOG.md`
