# Reuse Decision

## Change type
frontend refactor / entity presentation owner strengthening

## Existing code searched
- `code/frontend/src/pages/contests/ContestListRoutePage.vue`
- `code/frontend/src/features/contest-detail/model/useContestListPage.ts`
- `code/frontend/src/widgets/contest-list-workspace/ContestListWorkspace.vue`
- `code/frontend/src/pages/contests/ContestDetailRoutePage.vue`
- `code/frontend/src/features/contest-detail/model/useContestDetailRoutePage.ts`
- `code/frontend/src/widgets/contest-detail-workspace/ContestDetailWorkspace.vue`
- `code/frontend/src/features/platform/contests/ui/PlatformContestTable.vue`
- `code/frontend/src/pages/contests/__tests__/ContestList.test.ts`
- `code/frontend/src/pages/contests/__tests__/ContestDetail.test.ts`
- `code/frontend/src/features/platform/contests/ui/PlatformContestTable.test.ts`
- `code/frontend/src/entities/contest/model/presentation.ts`
- `code/frontend/src/entities/contest/index.ts`
- 前端 feature-sliced architecture 迁移台账

## Similar implementations found
- `entities/notification` 已经承担通知对象的稳定展示规则，并让 list/detail workspace 直接消费实体层 owner。
- `entities/contest/model/presentation.ts` 已经有状态、模式、动作和 accent 基础映射，但 contest list/detail route page 与 platform table 仍保留展示透传和本地 class owner，说明 contest 展示 owner 还没有完全收口。

## Decision
refactor_existing

## Reason
当前竞赛链路里最适合继续收口的不是 workflow，而是对象展示 owner：

- student contest list route page 仍在透传 `getStatusLabel / getModeLabel / getContestActionLabel / contestAccentStyle`
- student contest detail route page 仍在透传 `contestAccentStyle`
- platform contest table 仍在本地维护 `status -> pill class` 映射

本轮最小正确改动是：

- 强化 `entities/contest` 的 presentation owner，让状态 badge class 与 accent css var style 也落在实体层
- 让 student contest list/detail workspace 直接从 `entities/contest` 取稳定展示规则
- 让 platform contest table 改为消费实体层的状态 pill class owner
- 用测试锁住展示 owner 不再回流到 route page / feature model / platform table 本地函数

本轮不做：

- 不改 contest 列表查询、筛选、分页、倒计时、路由同步、队伍/提交 flag workflow
- 不改 contest overview / team / announcements / awd workspace 的业务流程
- 不把 route、toast、mutation 或 query sync 搬进实体层

## Files to modify
- `.harness/reuse-decisions/contest-entity-presentation-owner.md`
- `docs/plan/impl-plan/2026-05-31-contest-entity-presentation-owner-plan.md`
- `code/frontend/src/entities/contest/model/presentation.ts`
- `code/frontend/src/entities/contest/model/index.ts`
- `code/frontend/src/entities/contest/index.ts`
- `code/frontend/src/entities/contest/model/presentation.test.ts`
- `code/frontend/src/widgets/contest-list-workspace/ContestListWorkspace.vue`
- `code/frontend/src/pages/contests/ContestListRoutePage.vue`
- `code/frontend/src/features/contest-detail/model/useContestListPage.ts`
- `code/frontend/src/widgets/contest-detail-workspace/ContestDetailWorkspace.vue`
- `code/frontend/src/pages/contests/ContestDetailRoutePage.vue`
- `code/frontend/src/features/contest-detail/model/useContestDetailRoutePage.ts`
- `code/frontend/src/features/platform/contests/ui/PlatformContestTable.vue`
- `code/frontend/src/pages/contests/__tests__/ContestList.test.ts`
- `code/frontend/src/pages/contests/__tests__/ContestDetail.test.ts`
- `code/frontend/src/features/platform/contests/ui/PlatformContestTable.test.ts`
- 前端 feature-sliced architecture 迁移台账

## After implementation
- `entities/contest` 会成为 contest 状态 / 模式 / CTA / accent / status badge class 的稳定展示 owner。
- student contest list/detail route page 不再为 widget 透传 contest 展示规则。
- platform contest table 不再本地维护状态 pill class 映射。
