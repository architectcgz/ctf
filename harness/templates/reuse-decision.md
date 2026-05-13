# Reuse Decision

Suggested path: `.harness/reuse-decisions/<task-slug>.md`

This file is task-scoped current-task evidence.
Keep one reuse decision file per protected task and do not overwrite another task's decision file.
Durable reuse knowledge belongs in:

- `harness/reuse/index.yaml` for searchable pattern entries.
- `harness/reuse/history.md` for append-only decision history.

## Change type
page / component / hook / service / handler / repository / port / job / mapper / readmodel / composition / store / api / form / table / modal / layout / schema / migration

## Existing code searched
- code/frontend/src/views/...
- code/frontend/src/components/...
- code/frontend/src/features/...
- code/frontend/src/composables/...
- code/frontend/src/api/...
- code/backend/internal/module/...
- code/backend/internal/app/composition/...
- code/backend/internal/model/...
- code/backend/migrations/...

## Similar implementations found
- code/frontend/src/views/example/ExampleList.vue
- code/frontend/src/components/common/WorkspaceDataTable.vue
- code/frontend/src/features/example/model/useExampleListQuery.ts
- code/backend/internal/module/example/ports/...
- code/backend/internal/module/example/infrastructure/...
- code/backend/internal/module/example/api/...
- code/backend/internal/module/example/application/...

## Decision
reuse_existing / extend_existing / refactor_existing / create_new_with_reason

## Reason
Explain why the existing implementation can be reused, extended, refactored, or why a new implementation is unavoidable.

## Files to modify
- code/frontend/src/views/example/ExampleList.vue
- code/frontend/src/components/common/WorkspaceDataTable.vue

## After implementation
- If this decision is reusable, append a short entry to `harness/reuse/history.md`.
- If future tasks should find this pattern without rereading old decisions, add or update an entry in `harness/reuse/index.yaml`.
