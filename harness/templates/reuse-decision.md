# Reuse Decision

Suggested path: `.harness/reuse-decisions/<task-slug>.md`

This file is task-scoped current-task evidence.
Keep one reuse decision file per protected task and do not overwrite another task's decision file.
Durable reuse knowledge belongs in:

- `.harness/reuse-index/index.yaml` for searchable local pattern entries.
- `.harness/reuse-index/<source-path>/README.md` for module-level or module-internal secondary indexes mirrored from real code paths.

## Change type
page / component / hook / service / handler / repository / port / job / mapper / readmodel / composition / store / api / form / table / modal / layout / schema / migration

## Existing code searched
- 只列本次实际搜索过的目录或文件；不要保留未实际搜索的占位项。
- 可以写配置过的搜索根，例如 `code/frontend/src/features`、`code/backend/internal/module`。
- 也可以直接写真实目录或文件，例如 `code/frontend/src/shared/ui/layout/TopNav.vue`、`harness/checks/common.py`。

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
- If future tasks should find this pattern without rereading old decisions, add or update a local entry in `.harness/reuse-index/index.yaml`.
- If the pattern belongs to a concrete module or subdirectory, also update the nearest mirrored `README.md` under `.harness/reuse-index/`.
