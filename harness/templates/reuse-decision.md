# Reuse Decision

This file is only for the current task and may be overwritten by the next protected change.
Durable reuse knowledge belongs in:

- `.harness/reuse-index.yaml` for searchable pattern entries.
- `.harness/reuse-history.md` for append-only decision history.

## Change type
page / component / hook / service / store / api / form / table / modal / layout / schema

## Existing code searched
- code/frontend/src/views/...
- code/frontend/src/components/...
- code/frontend/src/features/...
- code/frontend/src/composables/...
- code/frontend/src/api/...

## Similar implementations found
- code/frontend/src/views/example/ExampleList.vue
- code/frontend/src/components/common/WorkspaceDataTable.vue
- code/frontend/src/features/example/model/useExampleListQuery.ts

## Decision
reuse_existing / extend_existing / refactor_existing / create_new_with_reason

## Reason
Explain why the existing implementation can be reused, extended, refactored, or why a new implementation is unavoidable.

## Files to modify
- code/frontend/src/views/example/ExampleList.vue
- code/frontend/src/components/common/WorkspaceDataTable.vue

## After implementation
- If this decision is reusable, append a short entry to `.harness/reuse-history.md`.
- If future tasks should find this pattern without rereading old decisions, add or update an entry in `.harness/reuse-index.yaml`.
