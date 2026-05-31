# awd review directory owner 收口计划

## Objective

- 在 `awd-review-workspace` 内新增 `useAwdReviewDirectory`，承接 AWD 复盘目录的请求时序 owner。
- 让 `useAwdReviewIndex` 只保留 AWD 复盘自己的业务派生和展示语义。

## Non-goals

- 不改 AWD detail、export flow、route target builder。
- 不把 AWD 目录 owner 抽成新的 shared 通用目录层。
- 不调整 teacher/platform AWD 页面 UI。

## Source Inputs

- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndex.ts`
- `code/frontend/src/features/awd-review-workspace/model/useAwdReviewIndexPage.ts`
- `code/frontend/src/pages/awd-review/__tests__/PlatformAWDReviewIndex.test.ts`
- `code/frontend/src/pages/awd-review/__tests__/TeacherAWDReviewIndex.test.ts`
- `code/frontend/src/features/managed-instance-directory/model/useManagedInstanceDirectory.ts`

## Plan Review Result

- AWD 目录 owner 先在 feature 内收口最合理，因为它仍然带着 AWD 自己的 query contract 和 summary 语义。
- 本轮只拆目录状态，不碰 route page 和 widgets。
- stale request / debounce / cleanup 需要直接单测。

## Task Slices

### Slice 1: 新建 useAwdReviewDirectory

- 目标：收口 `listAwdReviewsByRole` 的加载、分页、debounce 和 cleanup。
- 风险：
  - 如果直接复用太泛的 shared owner，会模糊 AWD 复盘自己的 query contract。

### Slice 2: useAwdReviewIndex 改为消费目录 owner

- 目标：保留 rows / summary / status label / status options 的 AWD 语义。
- 风险：
  - 如果把这些展示派生也抽走，会伤到 AWD feature 的业务边界。

### Slice 3: 更新源码级和行为测试

- 目标：给目录 owner 补直测，并更新 platform/teacher 源码断言。
- 风险：
  - 不测 stale request 和 debounce，后面还会回流。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision awd-review-directory-owner`
- `npm run test:run -- src/features/awd-review-workspace/model/useAwdReviewDirectory.test.ts src/pages/awd-review/__tests__/PlatformAWDReviewIndex.test.ts src/pages/awd-review/__tests__/TeacherAWDReviewIndex.test.ts src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- `npm run typecheck`
- `git diff --check`
- `bash scripts/check-reuse-first.sh`

## Review Focus

- `useAwdReviewDirectory` 是否只承接目录状态 owner。
- `useAwdReviewIndex` 是否只剩 AWD 复盘派生语义。
- teacher/platform route page 是否继续只负责组合和 route target。

## Rollback / Recovery

- 如果目录 owner 接口形态不顺手，可以调整返回值，但 `load / debounce / stale request` owner 仍必须留在 `useAwdReviewDirectory`。
