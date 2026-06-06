---
name: ctf-go-backend-architecture-reference
description: Use when applying onion-clean-architecture or go-backend inside this CTF repository and you need the current backend layout, local architecture documents, and repo-specific reference paths.
---

# CTF Go Backend Architecture Reference

Use this skill only as the CTF supplement to generic backend architecture skills.

## Read Alongside

- `onion-clean-architecture`
- `go-backend`

## Current Local References

- Backend root: `code/backend`
- Architecture note: `docs/architecture/01-backend-architecture-style-decision.md`
- Backend refactor note: `docs/architecture/backend/07-modular-monolith-refactor.md`
- Implementation plans: `docs/plan/impl-plan/`
- Review evidence: `docs/reviews/backend/` and `docs/reviews/general/`

## Current CTF Reading Focus

- Read the current module ownership before moving files around.
- Prefer local modular-monolith boundaries over textbook Clean Architecture folder theater.
- Treat `internal/module/*` as the primary bounded-context surface.
- When adding or splitting ports, keep handler -> application -> ports -> infrastructure ownership explicit and align with this repository's review and reuse-first gates.
