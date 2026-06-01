# route / widget typecheck 尾差收口计划

## Objective

- 收掉当前前端 `typecheck` 中 3 个由 route/widget 迁移遗留的类型尾差。
- 不改页面结构和用户行为，只把 contract 对齐到当前 owner。

## Non-goals

- 不继续做新的 route page -> widget 拆分。
- 不改 contest / notification / scoreboard 的运行时逻辑。
- 不处理与这 3 个错误无关的其它结构债。

## Source Inputs

- `code/frontend/src/features/contest-detail/model/contestListRoutes.ts`
- `code/frontend/src/features/notifications/ui/NotificationCategoryFilter.vue`
- `code/frontend/src/features/scoreboard/model/useScoreboardDetailPage.ts`
- `code/frontend/src/widgets/contest-list-workspace/ContestListWorkspace.vue`
- `code/frontend/src/widgets/notification-list-workspace/NotificationListWorkspace.vue`
- `code/frontend/src/widgets/scoreboard-detail-workspace/ScoreboardDetailWorkspace.vue`

## Task Slices

### Slice 1: contest list route target 对齐共享导航 contract

- 目标：让 contest list 明细跳转直接复用 `AppRouteTarget`，不再保留本地近似类型。
- 变更面：
  - `code/frontend/src/features/contest-detail/model/contestListRoutes.ts`
- 风险：
  - 如果 target 类型改错，会把 `AppRouteLink` 的 route object 推回字符串拼接。

### Slice 2: notification category options 收口 readonly contract

- 目标：让 `NotificationCategoryFilter.vue` 接受只读 options，与 widget / page owner 一致。
- 变更面：
  - `code/frontend/src/features/notifications/ui/NotificationCategoryFilter.vue`
- 风险：
  - 如果把 readonly 改回 mutable，会继续允许下游误把 options 当本地可变状态。

### Slice 3: scoreboard detail 空态 formatter 签名对齐

- 目标：让 `formatContestWindow()` 显式接受 `null | undefined`，与 widget 已声明的 `contest` 空态 contract 一致。
- 变更面：
  - `code/frontend/src/features/scoreboard/model/useScoreboardDetailPage.ts`
- 风险：
  - 如果只在 widget 侧做类型绕过，会继续留下 page owner 和 widget contract 的漂移。

## Validation Plan

- `npm run test:run -- src/pages/contests/__tests__/ContestList.test.ts src/pages/notifications/__tests__/NotificationList.test.ts src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- `npm run typecheck`
- `bash scripts/check-consistency.sh`

## Review Focus

- 是否复用了已有共享 contract，而不是新建一次性类型别名。
- readonly / nullable contract 是否回到真正 owner，而不是在调用点加断言糊过去。
- 本轮是否保持“只修类型尾差，不扩结构面”。

## Rollback / Recovery

- 如果共享 contract 对齐导致现有测试或调用方出错，可先恢复单文件旧类型，再按更小切片分别收口 route target、readonly options 和 scoreboard formatter。
