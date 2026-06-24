# Contest Manage Page Shell Content Split 计划

## Objective

- 把 `ContestOrchestrationPage.vue` 收口成真正的 contest manage shell owner。
- 将 overview 与 create 两个大区块拆成明确的内部 panel owner，并把页面局部样式移到独立 css 文件。

## Non-goals

- 不改 `useContestManagePage.ts` 的 query 同步、公告抽屉、创建/编辑保存、AWD readiness override 或刷新编排。
- 不修改 `ContestManageRoutePage.vue` 的 route-level 组合方式。
- 不顺手继续拆 `PlatformContestTable.vue`、`PlatformContestFormPanel.vue` 或 create/edit dialog。

## Source Inputs

- `code/frontend/src/features/platform/contests/ui/ContestOrchestrationPage.vue`
- `code/frontend/src/features/platform/contests/ui/ContestEditWorkspacePanel.vue`
- `code/frontend/src/features/platform/contests/ui/AWDChallengeConfigPanel.vue`
- `code/frontend/src/features/platform/contests/model/useContestManagePage.ts`
- `code/frontend/src/pages/platform/contests/ContestManageRoutePage.vue`
- `code/frontend/src/pages/platform/contests/__tests__/ContestManage.test.ts`
- `code/frontend/src/features/platform/contests/ui/contestAdminUiStrategy.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Plan Review Result

- 当前最小且能真实压缩 `platform/contests` 过宽 owner 面的切口，不是回去再做 route page owner cleanup，而是直接拆 `ContestOrchestrationPage.vue` 本身的 content assembly。
- 保持 route page 与 page model 不动，可以把回归面收敛在 feature 内部 UI，而不是同时打开 route-level owner 与 panel owner 两条线。

## Task Slices

### Slice 1: 抽离 overview panel owner

- 目标：新增 `ContestManageOverviewPanel.vue`，承接目录 header、summary、toolbar、status filter、loading/empty/table 装配。
- 风险：筛选 reset、创建入口、列表分页和公告动作的 props/emits 契约必须保持不变。

### Slice 2: 抽离 create panel owner

- 目标：新增 `ContestManageCreatePanel.vue`，承接创建竞赛工作区 header 和 `PlatformContestFormPanel` 桥接。
- 风险：返回 overview 与创建保存事件仍要保持当前 switch panel 语义，不引入新的本地状态。

### Slice 3: 收口 shell、样式与测试护栏

- 目标：让 `ContestOrchestrationPage.vue` 只保留 shell、active panel 切换与事件桥接；新增 `contestOrchestrationPage.css` 承接这页的局部样式，并同步 raw-source 护栏与 backlog 进展。
- 风险：如果只拆模板不补 source-boundary 断言，后续容易把 overview/create 逻辑重新写回 page shell。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision contest-manage-page-shell-content-split`
- `cd code/frontend && npm run test:run -- src/pages/platform/contests/__tests__/ContestManage.test.ts src/features/platform/contests/ui/contestAdminUiStrategy.test.ts`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/contest-manage-page-shell-content-split.md docs/plan/archive/impl-plan/2026-06/2026-06-01-contest-manage-page-shell-content-split-plan.md docs/reviews/frontend/2026-06-01-contest-manage-page-shell-content-split-review.md code/frontend/src/features/platform/contests/ui/ContestOrchestrationPage.vue code/frontend/src/features/platform/contests/ui/ContestManageOverviewPanel.vue code/frontend/src/features/platform/contests/ui/ContestManageCreatePanel.vue code/frontend/src/features/platform/contests/ui/contestOrchestrationPage.css code/frontend/src/features/platform/contests/ui/contestOrchestrationPage.types.ts code/frontend/src/features/platform/contests/ui/index.ts code/frontend/src/features/platform/contests/ui/contestAdminUiStrategy.test.ts code/frontend/src/pages/platform/contests/__tests__/ContestManage.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Review Focus

- `ContestOrchestrationPage.vue` 是否真正退回 shell owner，而不是继续混放 overview / create 内容装配。
- 新增 panel 是否只承接本地展示组合，不把 route-aware 状态或异步 workflow 下沉进去。
- raw-source 护栏是否已经转向新 panel owner，避免后续继续只盯父页 SFC 做表面断言。
