> 状态：Current
> 事实源：`ContestOrchestrationPage.vue` 当前 owner、`feature-owned UI` 规则、竞赛目录 route page 与导航 owner 边界
> 替代：无

# Contest Orchestration Feature UI Migration Implementation Plan

## 目标

- 把 `ContestOrchestrationPage.vue` 从 `components/platform/contest/` 迁到 `features/platform-contests/ui/`。
- 把竞赛编辑页 / 运维页跳转 owner 从 page shell 收回 `useContestManagePage()`。
- 让 `views/platform/ContestManage.vue` 直接通过 `features/platform-contests` public API 组合 page model 与 page-sized UI。
- 收掉 `ContestOrchestrationPage.vue` 对应的 `componentFeatureImportAllowlist` 和 `legacyComponentPageAllowlist` 例外。

## 非目标

- 本轮不改 `usePlatformContests()` 的列表请求、保存、AWD readiness gate、公告抽屉等业务 owner。
- 本轮不拆 `PlatformContestTable.vue`、`PlatformContestFormPanel.vue`、`ContestAnnouncementManageDrawer.vue`。
- 本轮不顺手处理 `ContestAwdConfigWorkspaceShell.vue` 的大组件拆分。

## 输入依据

- `docs/architecture/frontend/06-components.md`
- `docs/architecture/frontend/07-pages-dataflow.md`
- `code/frontend/src/components/platform/contest/ContestOrchestrationPage.vue`
- `code/frontend/src/views/platform/ContestManage.vue`
- `code/frontend/src/features/platform-contests/model/useContestManagePage.ts`
- `code/frontend/src/features/platform-contests/model/usePlatformContests.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/platform/__tests__/ContestManage.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `ContestOrchestrationPage.vue` 已经是竞赛目录 feature 的 page-sized UI：它只服务 `ContestManage` route，并直接依赖 `features/platform-contests` contract。
- 这页当前的额外问题是自己持有 router；如果直接搬目录，只会把错误 owner 原样迁到 feature ui。
- 竞赛编辑 / 运维跳转应由 route-aware feature model `useContestManagePage()` 持有，page shell 只发出事件。

## 设计边界

### route view 继续负责

- 组合 `useContestManagePage()`、`ContestOrchestrationPage`、公告抽屉、表单对话框和 AWD 覆盖确认对话框
- 不直接持有 router / API owner

### `features/platform-contests/model` 继续负责

- 列表请求、筛选、创建 / 保存、AWD readiness gate、公告抽屉状态
- 竞赛编辑页 / 运维页跳转 owner

### `features/platform-contests/ui` 本轮负责

- 竞赛目录 page-sized shell
- 消费上层派生的数据与事件 handler
- 不直接持有 `vue-router`

### `components/platform/contest/*` 继续保留

- 稳定的竞赛表格、表单、抽屉和其它可复用展示分区
- 不再承担竞赛目录整页 shell owner

## 任务切片

### Slice 1：收回导航 owner 并迁移 page shell

- 目标：
  - `useContestManagePage()` 接管竞赛编辑页 / 运维页跳转
  - 新增 `features/platform-contests/ui/ContestOrchestrationPage.vue`
  - `ContestManage.vue` 改从 feature public API 引用
- 预期改动：
  - `code/frontend/src/features/platform-contests/model/useContestManagePage.ts`
  - `code/frontend/src/features/platform-contests/index.ts`
  - `code/frontend/src/features/platform-contests/ui/*`
  - `code/frontend/src/views/platform/ContestManage.vue`
  - `code/frontend/src/components.d.ts`
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/ContestManage.test.ts`
- Review focus：
  - page shell 是否已经不再持有 router
  - route view 是否仍然保持薄壳

### Slice 2：清理 guardrail 与 backlog

- 目标：
  - 清理 `ContestOrchestrationPage.vue` 对应 allowlist 例外
  - 更新 raw-source 测试路径和 backlog 进展
- 预期改动：
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - 相关 raw-source 测试
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-27-contest-orchestration-feature-ui-migration-review.md`
- 验证：
  - `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/ContestManage.test.ts src/views/platform/__tests__/contestManageExportFlowExtraction.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase21.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase22.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase27.test.ts src/views/platform/__tests__/contestOrchestrationTabsAdoption.test.ts src/views/platform/__tests__/journalPlatformShellStyles.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/views/__tests__/adminRootHeroLayout.test.ts src/views/__tests__/journalNoteStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/workspaceShellStyles.test.ts`
- Review focus：
  - allowlist 是否真实下降
  - raw-source 测试是否已经指向新 owner

## 结构收口检查

- `ContestOrchestrationPage.vue` 不再作为 `components/*Page.vue` 遗留页存在。
- `ContestManage.vue` 只组合 `useContestManagePage()` 与 feature public API。
- `ContestOrchestrationPage` 不再直接持有 `vue-router`。
- touched surface 上至少移除一条 component->feature allowlist、一条 legacy component page allowlist，并把竞赛目录跳转 owner 收口到 feature model。

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/ContestManage.test.ts src/views/platform/__tests__/contestManageExportFlowExtraction.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase21.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase22.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase27.test.ts src/views/platform/__tests__/contestOrchestrationTabsAdoption.test.ts src/views/platform/__tests__/journalPlatformShellStyles.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/views/__tests__/adminRootHeroLayout.test.ts src/views/__tests__/journalNoteStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/workspaceShellStyles.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/contest-orchestration-feature-ui-migration.md docs/plan/impl-plan/2026-05-27-contest-orchestration-feature-ui-migration-implementation-plan.md docs/reviews/frontend/2026-05-27-contest-orchestration-feature-ui-migration-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/features/platform-contests code/frontend/src/views/platform/ContestManage.vue code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/components.d.ts code/frontend/src/views/platform/__tests__/ContestManage.test.ts code/frontend/src/views/platform/__tests__/contestManageExportFlowExtraction.test.ts code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase21.test.ts code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase22.test.ts code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase27.test.ts code/frontend/src/views/platform/__tests__/contestOrchestrationTabsAdoption.test.ts code/frontend/src/views/platform/__tests__/journalPlatformShellStyles.test.ts code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts code/frontend/src/views/__tests__/adminRootHeroLayout.test.ts code/frontend/src/views/__tests__/journalNoteStyles.test.ts code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Review 关注点

- `features/platform-contests/ui` 是否成为竞赛目录 page shell 的唯一 owner。
- `useContestManagePage()` 是否明确接管竞赛编辑页 / 运维页跳转 owner。
- 测试与 allowlist 是否同步反映新边界，而不是继续绑定旧路径。

## 回退 / 恢复说明

- 如迁移后出现问题，可把 `ContestOrchestrationPage.vue` 移回 `components/platform/contest/` 并恢复 route view import。
- 如导航 owner 迁移有问题，可把 `useContestManagePage()` 中新增的 router 行为回退到 page shell。

## 残余风险

- `ContestOrchestrationPage` 的模板和样式仍然偏大，本轮只处理 owner 与目录归位，不做进一步拆分。
- `ContestAwdConfigWorkspaceShell.vue` 的大组件债仍在另一条 backlog 下继续跟踪。
