# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for OpenFiltr.

An ADR captures a significant architectural or technical decision made during the project's lifetime, including the context, the decision taken, and its consequences. They are **append-only**: once recorded, an ADR is never deleted. If a decision is reversed, a new ADR is written to supersede it.

## Why ADRs matter for AI agents

When an AI coding agent starts work on this repository, it must read all ADRs in this directory before making any changes. ADRs explain *why* the codebase is the way it is, so that agents do not accidentally reverse deliberate decisions or miss important constraints.

## Format

Each ADR file is named `NNNN-short-title.md` (zero-padded to four digits) and follows this template:

```markdown
# ADR-NNNN: Title

**Date:** YYYY-MM-DD
**Status:** Proposed | Accepted | Deprecated | Superseded by ADR-XXXX

## Context

What situation or problem prompted this decision?

## Decision

What was decided?

## Consequences

What are the results of this decision — positive, negative, and neutral?

## References

- Links to relevant issues, PRs, or external resources
```

## Index

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [0001](0001-postgresql-only-storage.md) | PostgreSQL as the only supported database backend | Superseded by ADR-0003 | 2026-03-18 |
| [0002](0002-agent-pre-task-protocol.md) | Agent pre-task protocol: sync, read ADRs and changelog, run tests | Accepted | 2026-03-19 |
| [0003](0003-bbolt-default-storage.md) | bbolt as the default embedded backend with PostgreSQL compatibility | Accepted | 2026-03-30 |
| [0004](0004-manual-changelog-updates.md) | Keep changelog updates inside the implementation PR | Accepted | 2026-03-30 |
