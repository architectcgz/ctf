# awd review detail owner 收口计划

## Objective

- 在 `awd-review-detail-workspace` 内新增 `useAwdReviewDetailData`，承接 AWD 复盘详情的数据 owner。
- 让 `useAwdReviewDetailPage` 只保留 route query、breadcrumb、导出和页面级 orchestration。

## Non-goals

- 不改 AWD detail route page 模板和 widget 结构。
- 不调整 `useAwdReviewExportFlow` 的导出逻辑。
- 不修改 AWD detail 的 API contract、query 参数或页面文案。

## Source Inputs

- `code/frontend/src/features/awd-review-detail-workspace/model/useAwdReviewDetailPage.ts`
- `code/frontend/src/pages/awd-review/__tests__/PlatformAwdReviewDetail.test.ts`
- `code/frontend/src/pages/awd-review/__tests__/TeacherAWDReviewDetail.test.ts`
- `code/frontend/src/features/contest-projector/model/useContestProjectorPage.ts`
- `code/frontend/src/features/teaching/student-analysis-workspace/model/useStudentAnalysisDataState.ts`

## Plan Review Result

- AWD detail 最合适的第一刀是 `page + data` 拆分，而不是继续把 route、导出和 detail 数据混在一起。
- route query 和返回列表跳转仍然留在 page owner，更符合页面级 orchestration。
- `review / selectedTeam / summaryStats / selectedTeamServices` 这些细节状态应当和详情读取一起落在 data owner。

## Task Slices

### Slice 1: 新建 useAwdReviewDetailData

- 目标：收口 AWD 详情读取、team 选择和详情派生数据。
- 风险：
  - 如果把 breadcrumb 或 export flow 一起搬走，会重新模糊 page owner。

### Slice 2: useAwdReviewDetailPage 改为消费 detail data owner

- 目标：保留 route query、返回列表、breadcrumb 和导出编排。
- 风险：
  - 如果 page 继续直接依赖 API，就没有真正收口 detail 数据 owner。

### Slice 3: 更新源码级和行为测试

- 目标：给 detail data owner 补直测，并更新 platform/teacher 详情页源码断言。
- 风险：
  - 不验证 team 选择回收和失败态，后面还会回流进 page owner。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision awd-review-detail-owner`
- `npm run test:run -- src/features/awd-review-detail-workspace/model/useAwdReviewDetailData.test.ts src/pages/awd-review/__tests__/PlatformAwdReviewDetail.test.ts src/pages/awd-review/__tests__/TeacherAWDReviewDetail.test.ts src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- `npm run typecheck`
- `git diff --check`
- `bash scripts/check-reuse-first.sh`

## Review Focus

- `useAwdReviewDetailData` 是否只承接 AWD detail 数据 owner。
- `useAwdReviewDetailPage` 是否只剩页面级路由和 workflow 编排。
- 平台端 / 教师端 detail route page 是否继续只负责组合 widget。

## Rollback / Recovery

- 如果 `useAwdReviewDetailData` 的返回接口不顺手，可以调整命名或结构，但 API 读取与 team 选择 owner 仍必须留在 data composable。
