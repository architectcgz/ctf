> 状态：Current
> 事实源：`TeacherDashboardPage.vue` 当前 owner、`feature-owned UI` 规则、教师总览 route page 组合边界
> 替代：无

# Teacher Dashboard Feature UI Migration Implementation Plan

## 目标

- 把 `TeacherDashboardPage.vue` 从 `components/teacher/dashboard/` 迁到 `features/teacher-dashboard/ui/`。
- 让 `views/teacher/TeacherDashboard.vue` 直接通过 `features/teacher-dashboard` public API 组合 page model 与 page-sized UI。
- 收掉 `TeacherDashboardPage.vue` 对应的 `componentFeatureImportAllowlist` 和 `legacyComponentPageAllowlist` 例外。

## 非目标

- 本轮不改 `useDashboardPage()` 的请求、失败处理、URL tab 和导航 owner。
- 本轮不改 `TeacherDashboardPortraitPanel.vue`、`TeacherDashboardStudentInsightPanel.vue`、`TeacherDashboardTrendPanel.vue`、`TeacherDashboardReviewPanel.vue`、`TeacherDashboardInterventionPanel.vue` 的职责边界。
- 本轮不顺手处理 `ClassManagementPage.vue`、`StudentAnalysisPage.vue` 或其它 teacher legacy page。

## 输入依据

- `docs/architecture/frontend/06-components.md`
- `docs/architecture/frontend/07-pages-dataflow.md`
- `code/frontend/src/components/teacher/dashboard/TeacherDashboardPage.vue`
- `code/frontend/src/features/teacher-dashboard/model/useDashboardPage.ts`
- `code/frontend/src/features/teacher-dashboard/model/useDashboardMetrics.ts`
- `code/frontend/src/views/teacher/TeacherDashboard.vue`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherDashboard.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `TeacherDashboardPage.vue` 已经不是中立组件，而是教师总览 feature 的 page-sized UI：它直接消费 `useDashboardMetrics()`，并且只服务 `TeacherDashboard` route view。
- 继续把它留在 `components/teacher/dashboard/`，只会让 component->feature allowlist 与 legacy component page allowlist 保持增长。
- 由于教师总览已经拆成稳定子分区，本轮迁移不需要再混入模板拆分或 teacher 结构耦合收口。

## 设计边界

### route view 继续负责

- 路由入口组合 `useDashboardPage()` 与 `TeacherDashboardPage`
- 不直接承担教师总览的模板细节

### `features/teacher-dashboard/model` 继续负责

- 教师概览请求、失败态、重试与班级管理跳转 owner
- `overview` 原始数据到 page shell 展示数据的组合

### `features/teacher-dashboard/ui` 本轮负责

- 教师总览 page-sized UI surface
- 消费 feature model 派生后的只读数据与父层 emit
- 继续组合现有 portrait / insight / trend / review / intervention 子面板

### `components/teacher/dashboard/*` 继续保留

- 教师总览的稳定展示分区
- 不直接引入 `@/features/teacher/dashboard`

## 任务切片

### Slice 1：迁移 feature-owned page shell

- 目标：
  - 新增 `features/teacher-dashboard/ui/TeacherDashboardPage.vue`
  - `views/teacher/TeacherDashboard.vue` 改从 feature public API 引用
- 预期改动：
  - `code/frontend/src/features/teacher-dashboard/index.ts`
  - `code/frontend/src/features/teacher-dashboard/ui/*`
  - `code/frontend/src/views/teacher/TeacherDashboard.vue`
  - `code/frontend/src/components.d.ts`
- 验证：
  - `npm run test:run -- src/views/teacher/__tests__/TeacherDashboard.test.ts`
- Review focus：
  - route view 是否仍然保持薄壳
  - feature ui 是否没有吸入 router / API owner

### Slice 2：清理 guardrail 与 backlog

- 目标：
  - 清理 `TeacherDashboardPage.vue` 对应的 allowlist 例外
  - 更新 raw-source 测试路径与 backlog 进展
- 预期改动：
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
  - `code/frontend/src/views/teacher/__tests__/TeacherDashboard.test.ts`
  - `code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts`
  - `code/frontend/src/views/teacher/__tests__/teacherWorkspaceSubpanelAdoption.test.ts`
  - `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
  - `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-27-teacher-dashboard-feature-ui-migration-review.md`
- 验证：
  - `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/TeacherDashboard.test.ts src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts src/views/teacher/__tests__/teacherWorkspaceSubpanelAdoption.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts`
- Review focus：
  - allowlist 是否真实下降
  - raw-source 测试是否已经对齐新的 feature ui owner

## 结构收口检查

- `TeacherDashboardPage.vue` 不再作为 `components/*Page.vue` 遗留页存在。
- `views/teacher/TeacherDashboard.vue` 只组合 `useDashboardPage()` 与 feature public API。
- touched surface 上至少移除一条 component->feature allowlist 与一条 legacy component page allowlist。

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/TeacherDashboard.test.ts src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts src/views/teacher/__tests__/teacherWorkspaceSubpanelAdoption.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/teacher-dashboard-feature-ui-migration.md docs/plan/impl-plan/2026-05-27-teacher-dashboard-feature-ui-migration-implementation-plan.md docs/reviews/frontend/2026-05-27-teacher-dashboard-feature-ui-migration-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/features/teacher-dashboard code/frontend/src/views/teacher/TeacherDashboard.vue code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/components.d.ts code/frontend/src/views/teacher/__tests__/TeacherDashboard.test.ts code/frontend/src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts code/frontend/src/views/teacher/__tests__/teacherWorkspaceSubpanelAdoption.test.ts code/frontend/src/views/__tests__/workspaceShellStyles.test.ts code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Review 关注点

- `features/teacher-dashboard/ui` 是否成为教师总览 page shell 的唯一 owner。
- 迁移后是否没有让 `components/teacher/dashboard/*` 继续承担 feature model 的中间桥职责。
- 测试与 allowlist 是否同步反映新边界，而不是继续引用旧路径。

## 回退 / 恢复说明

- 如迁移后出现问题，可把 `TeacherDashboardPage.vue` 移回 `components/teacher/dashboard/` 并恢复 route view import。
- 本轮不触碰 API、DTO、路由名和教师总览文案，因此回退只涉及目录与 import 关系。

## 残余风险

- 教师总览的子分区仍保留在 `components/teacher/dashboard/`，是否继续迁到 feature ui 需要另开切片判断。
- teacher 其它 legacy component page 仍然存在，本轮只处理教师总览这一组。
