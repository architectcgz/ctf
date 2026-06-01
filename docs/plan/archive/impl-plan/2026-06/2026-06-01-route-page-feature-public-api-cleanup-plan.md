# Route Page Feature Public API Cleanup 计划

## Objective

- 把 `TeacherDashboardRoutePage.vue` 与 `PlatformOverviewRoutePage.vue` 的 feature internal import 收回各自 feature 根入口。
- 为 route page 层补一条自动护栏，禁止直接 import `@/features/*/(model|ui|lib|api|types)/*`。

## Non-goals

- 不改 teacher dashboard 与 platform overview 的数据加载、错误处理和 retry owner。
- 不调整 feature 根出口已经存在的导出结构。
- 不扩展到 widgets、entities 或其它更宽的公共 API 清理。

## Source Inputs

- `code/frontend/src/pages/teacher/TeacherDashboardRoutePage.vue`
- `code/frontend/src/pages/platform/PlatformOverviewRoutePage.vue`
- `code/frontend/src/features/teacher/dashboard/index.ts`
- `code/frontend/src/features/platform/overview/index.ts`
- `code/frontend/src/__tests__/routePageArchitectureBoundary.test.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherDashboard.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformOverview.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Plan Review Result

- 这刀的 owner 很清楚：feature 根出口已经是事实源，route page 只需要停用内部路径。
- 护栏应保持足够窄，只防 route page 深导入 feature internal module，不借机重写整套 import policy。

## Task Slices

### Slice 1: 收口两个 route page 的 import 路径

- 目标：`TeacherDashboardRoutePage.vue` 与 `PlatformOverviewRoutePage.vue` 都只从 feature 根入口导入 page component 和 page model。
- 风险：如果 feature 根出口没有稳定转发 page model，会把改动扩散到 feature API 组织。

### Slice 2: 补 route page 公共出口护栏

- 目标：在 `routePageArchitectureBoundary.test.ts` 新增 route page 不得深导入 feature internal module 的断言。
- 风险：规则过宽会误伤合法的 feature 根入口导入；规则过窄则继续放过 `model/use*Page.ts` 这类残片。

### Slice 3: 更新 raw-source 测试与 backlog 事实

- 目标：让 `TeacherDashboard.test.ts`、`PlatformOverview.test.ts` 和 backlog 都显式体现“route page 只吃 feature public API”。
- 风险：如果只改源码不补护栏，后续很容易再次回流到内部路径。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision route-page-feature-public-api-cleanup`
- `cd code/frontend && npm run test:run -- src/__tests__/routePageArchitectureBoundary.test.ts src/pages/teacher/__tests__/TeacherDashboard.test.ts src/pages/platform/__tests__/PlatformOverview.test.ts`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/route-page-feature-public-api-cleanup.md docs/plan/impl-plan/2026-06-01-route-page-feature-public-api-cleanup-plan.md docs/reviews/frontend/2026-06-01-route-page-feature-public-api-cleanup-review.md code/frontend/src/pages/teacher/TeacherDashboardRoutePage.vue code/frontend/src/pages/platform/PlatformOverviewRoutePage.vue code/frontend/src/__tests__/routePageArchitectureBoundary.test.ts code/frontend/src/pages/teacher/__tests__/TeacherDashboard.test.ts code/frontend/src/pages/platform/__tests__/PlatformOverview.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Review Focus

- 两个 route page 是否都已经不再直接引用 `model/use*Page.ts`。
- 新护栏是否只限制 feature internal import，而没有误伤 feature 根入口。
- raw-source 测试是否真正锁住 route page 对 feature public API 的依赖方向。

## Rollback / Recovery

- 如果护栏误伤现有合法路径，优先收窄正则，不回退 route page 的公共出口改动。
- 如果发现还有其它 route page 同类残片，可以在同一规则下继续补，不再新增一套平行约束。
