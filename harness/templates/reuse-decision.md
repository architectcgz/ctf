# Reuse Decision

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
