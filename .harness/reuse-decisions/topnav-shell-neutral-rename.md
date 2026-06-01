# Reuse Decision

## Change type
layout naming cleanup / shared top navigation shell

## Existing code searched
- `code/frontend/src/views`
- `code/frontend/src/components`
- `code/frontend/src/features`
- `code/frontend/src/widgets`
- `code/frontend/src/composables`
- `code/frontend/src/api`
- `code/frontend/src/shared/ui/layout/TopNav.vue`
- `code/frontend/src/shared/ui/layout/topnav/topNavShell.css`
- `code/frontend/src/shared/ui/layout/__tests__/TopNav.test.ts`
- `code/frontend/src/shared/ui/layout/AppLayout.vue`

## Similar implementations found
- `AppLayout.vue` 会在学生页、教师页、平台页统一挂载共享 `TopNav`。
- `TopNav.vue` 当前把 `topnav-shell--admin` 和 `topnav-tool-cluster--admin` 写死在共享顶栏上。
- `topNavShell.css` 与 `TopNav.test.ts` 也都围绕这个历史命名建立了样式和断言，但实际并没有学生专属的另一套壳样式。

## Decision
refactor_existing

## Reason
这次不是新增学生变体，而是把误导性的历史类名改成中性共享命名，避免学生页 DOM 上继续出现 `--admin`。

最小正确改动是：

- 把共享顶栏壳类名从 `topnav-shell--admin` 重命名为 `topnav-shell--workspace`。
- 把工具胶囊类名从 `topnav-tool-cluster--admin` 重命名为 `topnav-tool-cluster--workspace`。
- 同步更新共享样式选择器和测试断言。

本轮不做：

- 不新增 `topnav-shell--student` 与 `topnav-shell--backoffice` 两套样式分支。
- 不改路由判定、权限逻辑或面包屑行为。
- 不改顶栏视觉表现，只改共享命名。

## Files to modify
- `.harness/reuse-decisions/topnav-shell-neutral-rename.md`
- `code/frontend/src/shared/ui/layout/TopNav.vue`
- `code/frontend/src/shared/ui/layout/topnav/TopNavUserCard.vue`
- `code/frontend/src/shared/ui/layout/topnav/topNavShell.css`
- `code/frontend/src/shared/ui/layout/__tests__/TopNav.test.ts`

## After implementation
- 学生页 DOM 不会再挂载 `topnav-shell--admin` 这类误导性类名。
- 共享顶栏继续复用同一套样式，但命名会准确反映它的真实归属。
- 测试会改为围绕共享 workspace 顶栏命名做断言。
