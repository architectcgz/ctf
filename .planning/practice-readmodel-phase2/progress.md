# Progress

## 2026-03-22

- Created isolated worktree `codex/ctf-practice-readmodel-phase2`
- Narrowed phase-2 scope to student progress and timeline reads
- Wrote phase-2 spec and implementation plan
- Implemented `practice_readmodel` and hard-switched `/api/v1/users/me/progress` plus `/api/v1/users/me/timeline`
- Removed migrated progress/timeline read handlers, service methods, and repository queries from `practice`
- Refactored `practice_readmodel` to follow backend architecture rules:
  - `api/http` depends on module contract
  - `application` depends on minimal repository contract
  - `infrastructure` split by progress/timeline concern
- Added router/composition and integration coverage for the new readmodel owner
- Added backend architecture/code-style docs and linked them from `CODEX.md`
- Verification completed with:
  - `go test ./internal/module/practice ./internal/module/practice_readmodel/... ./internal/app`
  - `go test ./...`
