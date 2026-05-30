# Reuse Decision

## Change type
frontend refactor / route page migration residue cleanup

## Existing code searched
- code/frontend/src/views/platform/*.vue
- code/frontend/src/views/__tests__/adminRootHeroLayout.test.ts
- code/frontend/src/views/platform/__tests__/*.test.ts
- code/frontend/src/pages/platform/**/*.vue
- code/frontend/src/pages/awd-review/PlatformAwdReviewDetailRoutePage.vue
- code/frontend/src/router/routes/platformRoutes.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- 学生侧、`auth`、`errors`、`utility` 的运行时入口已经统一迁到 `pages/**`，旧 `views/*.vue` 已删除，只保留 `views/**/__tests__`。
- 当前 `platformRoutes.ts` 已全部指向 `@/pages/platform/**` 或 `@/pages/awd-review/**`，不再以 `@/views/platform/**` 作为运行时入口。

## Decision
refactor_existing

## Reason
这轮不是新增 page owner，也不是继续拆 feature，而是把 `views/platform` 里的迁移残片做最终清场：

- 保留 `views/platform/__tests__` 作为邻近测试目录
- 把唯一还直接 raw import 旧 `ChallengeDetail.vue` 的测试切到 `ChallengeDetailRoutePage.vue`
- 把仍以 platform 旧 page 命名落位的 `ChallengePackageFormat` 邻近测试切到 `pages/platform/challenges/__tests__/ChallengePackageFormat.test.ts`
- 删除已经失去运行时引用的旧 `views/platform/*.vue`
- 同步更新 backlog 中关于 `views/` 入口层收口的当前结论

不做：

- 不重写历史 plan / review 文档里的旧路径记录
- 不迁移 `views/platform/__tests__` 目录本身
- 不继续扩大到 teacher / contests / 其他 `views/**` 测试目录重命名

## Files to modify
- .harness/reuse-decisions/platform-views-final-cleanup.md
- docs/plan/impl-plan/2026-05-30-platform-views-final-cleanup-plan.md
- code/frontend/src/views/__tests__/adminRootHeroLayout.test.ts
- code/frontend/src/pages/platform/challenges/__tests__/ChallengePackageFormat.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/views/platform/*.vue

## After implementation
- `src/views/platform` 将只剩 `__tests__/`
- platform 运行时入口与当前测试 raw source 将统一以 `pages/**` 或 feature / widget owner 为准
- backlog 中关于 `views/` 入口层的结论会与代码现状一致
