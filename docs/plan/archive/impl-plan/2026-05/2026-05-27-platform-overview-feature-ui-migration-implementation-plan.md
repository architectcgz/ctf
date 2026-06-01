> 状态：Current
> 事实源：`PlatformOverviewPage.vue` 当前 owner、`feature-owned UI` 既有规则、前端架构 allowlist 与平台总览 route page 组合边界
> 替代：无

# Platform Overview Feature UI Migration Implementation Plan

## 目标

- 把 `PlatformOverviewPage.vue` 从 `components/platform/dashboard/` 迁到 `features/platform-overview/ui/`。
- 让 `views/platform/PlatformOverview.vue` 直接通过 `features/platform-overview` public API 组合 page model 与 page-sized UI。
- 收掉 `PlatformOverviewPage.vue` 对应的 `componentFeatureImportAllowlist` 和 `legacyComponentPageAllowlist` 例外。

## 非目标

- 本轮不改 `usePlatformOverviewPage()` 的加载、失败处理和导航 owner。
- 本轮不改 `PlatformOverviewHeroPanel.vue`、`PlatformOverviewAlertsSection.vue`、`PlatformOverviewHotspotsSection.vue` 的职责边界，只保留它们作为平台总览的稳定分区。
- 本轮不顺手迁移 `TeacherDashboardPage.vue`、`ContestOrchestrationPage.vue` 或其它 dashboard/contest 页面。

## 输入依据

- `docs/architecture/frontend/06-components.md`
- `docs/architecture/frontend/07-pages-dataflow.md`
- `code/frontend/src/components/platform/dashboard/PlatformOverviewPage.vue`
- `code/frontend/src/components/platform/dashboard/PlatformOverviewHeroPanel.vue`
- `code/frontend/src/components/platform/dashboard/PlatformOverviewAlertsSection.vue`
- `code/frontend/src/components/platform/dashboard/PlatformOverviewHotspotsSection.vue`
- `code/frontend/src/features/platform-overview/model/usePlatformOverviewPage.ts`
- `code/frontend/src/features/platform-overview/model/usePlatformOverviewWorkspace.ts`
- `code/frontend/src/views/platform/PlatformOverview.vue`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `PlatformOverviewPage.vue` 已经不是中立组件，而是平台总览 feature 的 page-sized UI：它直接消费 `usePlatformOverviewWorkspace()`，并且仅服务 `PlatformOverview` route view。
- 当前继续把它留在 `components/platform/dashboard/`，只会让 `components/* -> features/*` 与 legacy component page 两条 allowlist 一直保留。
- 因为平台总览内部已经拆出 hero / alerts / hotspots 三个稳定区块，这次迁移不需要再混入额外的模板拆分。

## 设计边界

### route view 继续负责

- 路由入口组合 `usePlatformOverviewPage()` 与 `PlatformOverviewPage`
- 不直接落平台总览的模板细节

### `features/platform-overview/model` 继续负责

- 平台总览请求、失败态、重试和导航 owner
- `dashboard` 原始数据到 page shell 所需展示数据的组合

### `features/platform-overview/ui` 本轮负责

- 平台总览 page-sized UI surface
- 消费 feature model 派生后的只读数据与父层 emit
- 继续组合 `PlatformOverviewHeroPanel`、`PlatformOverviewAlertsSection`、`PlatformOverviewHotspotsSection`

### `components/platform/dashboard/*` 继续保留

- 平台总览的稳定展示分区与样式原语
- 不直接引入 `@/features/platform-overview`

## 任务切片

### Slice 1：迁移 feature-owned page shell

- 目标：
  - 新增 `features/platform-overview/ui/PlatformOverviewPage.vue`
  - `views/platform/PlatformOverview.vue` 改从 feature public API 引用
- 预期改动：
  - `code/frontend/src/features/platform-overview/index.ts`
  - `code/frontend/src/features/platform-overview/ui/*`
  - `code/frontend/src/views/platform/PlatformOverview.vue`
  - `code/frontend/src/components.d.ts`
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/PlatformOverview.test.ts`
- Review focus：
  - route view 是否仍然保持薄壳
  - feature ui 是否没有吸入 router / API owner

### Slice 2：清理 guardrail 与 backlog

- 目标：
  - 清理 `PlatformOverviewPage.vue` 对应的 allowlist 例外
  - 更新 raw-source 测试路径和前端债务 backlog
- 预期改动：
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/views/platform/__tests__/PlatformOverview.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformOverviewSurfaceAlignment.test.ts`
  - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
  - `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
  - `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
  - `code/frontend/src/views/__tests__/adminRootHeroLayout.test.ts`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-27-platform-overview-feature-ui-migration-review.md`
- 验证：
  - `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/PlatformOverview.test.ts src/views/platform/__tests__/platformOverviewSurfaceAlignment.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/adminRootHeroLayout.test.ts`
- Review focus：
  - allowlist 是否真实下降
  - raw-source 测试是否已经指向新的 feature ui owner

## 结构收口检查

- `PlatformOverviewPage.vue` 不再作为 `components/*Page.vue` 遗留页存在。
- `views/platform/PlatformOverview.vue` 只组合 `usePlatformOverviewPage()` 与 feature public API。
- touched surface 上至少移除一条 component->feature allowlist 与一条 legacy component page allowlist。

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/PlatformOverview.test.ts src/views/platform/__tests__/platformOverviewSurfaceAlignment.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/adminRootHeroLayout.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/platform-overview-feature-ui-migration.md docs/plan/impl-plan/2026-05-27-platform-overview-feature-ui-migration-implementation-plan.md docs/reviews/frontend/2026-05-27-platform-overview-feature-ui-migration-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/features/platform-overview code/frontend/src/views/platform/PlatformOverview.vue code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/components.d.ts code/frontend/src/views/platform/__tests__/PlatformOverview.test.ts code/frontend/src/views/platform/__tests__/platformOverviewSurfaceAlignment.test.ts code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts code/frontend/src/views/__tests__/workspaceShellStyles.test.ts code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts code/frontend/src/views/__tests__/adminRootHeroLayout.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Review 关注点

- `features/platform-overview/ui` 是否成为平台总览 page shell 的唯一 owner。
- 迁移后是否没有把 `components/platform/dashboard/*` 继续变成 feature model 的中间桥。
- 测试与 allowlist 是否同步反映新边界，而不是继续引用旧路径。

## 回退 / 恢复说明

- 如迁移后出现问题，可把 `PlatformOverviewPage.vue` 移回 `components/platform/dashboard/` 并恢复 route view import。
- 本轮不触碰 API、DTO、路由名和平台总览文案，因此回退只涉及目录与 import 关系。

## 残余风险

- 平台总览的三个子分区仍留在 `components/platform/dashboard/`，后续如果要进一步下沉到 feature ui，需要另开切片判断是否值得。
- 其它同模式的 legacy component page 仍然存在，本轮只处理平台总览这一组。
