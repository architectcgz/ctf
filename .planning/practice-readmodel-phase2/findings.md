# Findings

- Remaining phase-2 gap is query ownership, not concrete cross-module imports.
- `practice` currently owns cross-domain training read SQL for progress/timeline.
- `GET /api/v1/users/me/progress` and `GET /api/v1/users/me/timeline` are the cleanest first cut.
- `teacher` already proved the readmodel extraction pattern is viable in this repo.
