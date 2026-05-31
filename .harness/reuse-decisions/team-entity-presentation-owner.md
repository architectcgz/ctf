# Reuse Decision

## Change type
frontend refactor / entity presentation owner strengthening

## Existing code searched
- `code/frontend/src/features/contest-detail/ui/ContestTeamPanel.vue`
- `code/frontend/src/features/contest-detail/ui/ContestTeamWorkspaceSection.vue`
- `code/frontend/src/features/contest-detail/model/useContestDetailPage.ts`
- `code/frontend/src/features/contest-detail/model/useContestDetailRoutePage.ts`
- `code/frontend/src/widgets/contest-detail-workspace/ContestDetailWorkspace.vue`
- `code/frontend/src/pages/contests/ContestDetailRoutePage.vue`
- `code/frontend/src/pages/contests/__tests__/ContestDetail.test.ts`
- `code/frontend/src/entities/notification/*`
- `code/frontend/src/entities/contest/*`
- 前端 feature-sliced architecture 迁移台账文档

## Similar implementations found
- `entities/contest` 已经承接 contest 状态、模式、CTA、accent 和 badge class 这类稳定展示规则。
- `entities/notification` 已经承接通知类型、已读状态与 accent 展示规则，并通过公共入口给 workspace/detail 消费。
- 当前 `team` 还没有实体层入口，`captain_user_id` 判定、成员数派生、空态文案和邀请码展示仍散在 feature model 与 feature ui 中。

## Decision
refactor_existing

## Reason
当前最小正确切片不是迁移队伍 workflow，而是先把 `team` 对象的稳定展示 owner 收口：

- `useContestDetailPage.ts` 仍在本地判断当前用户是否为队长
- `useContestDetailRoutePage.ts` 仍在本地派生成员数
- `ContestTeamPanel.vue` 仍直接用 `captain_user_id` 渲染队长标签和空态文案
- `ContestTeamWorkspaceSection.vue` 仍直接拼接邀请码文案

本轮最小正确改动是：

- 新建 `entities/team`，承接成员数、队长判定、队长标签、未入队空态、邀请码展示和成员展示项构建
- 让 `ContestTeamPanel.vue` 与 `ContestTeamWorkspaceSection.vue` 通过实体层公共入口消费 team 展示规则
- 让 `useContestDetailPage.ts` 与 `useContestDetailRoutePage.ts` 停止本地持有 team 展示派生
- 用测试锁住 `captain_user_id` / `members.length` / 邀请码展示 owner 不再回流到 feature route model 或 feature ui 私有实现

本轮不做：

- 不改创建队伍、加入队伍、踢出成员、确认弹窗、toast、route query 和 dialog workflow
- 不改 AWD 战场、题目提交、公告和详情加载流程
- 不把 mutation、confirm 或 route 状态搬进实体层

## Files to modify
- `.harness/reuse-decisions/team-entity-presentation-owner.md`
- `docs/plan/impl-plan/2026-05-31-team-entity-presentation-owner-plan.md`
- `code/frontend/src/entities/team/index.ts`
- `code/frontend/src/entities/team/model/index.ts`
- `code/frontend/src/entities/team/model/presentation.ts`
- `code/frontend/src/entities/team/model/presentation.test.ts`
- `code/frontend/src/features/contest-detail/ui/ContestTeamPanel.vue`
- `code/frontend/src/features/contest-detail/ui/ContestTeamWorkspaceSection.vue`
- `code/frontend/src/features/contest-detail/model/useContestDetailPage.ts`
- `code/frontend/src/features/contest-detail/model/useContestDetailRoutePage.ts`
- `code/frontend/src/pages/contests/__tests__/ContestDetail.test.ts`

## After implementation
- `entities/team` 会成为 team 成员数、队长关系、队长标签、邀请码文案和成员展示项的稳定展示 owner。
- `useContestDetailPage.ts` 与 `useContestDetailRoutePage.ts` 不再直接持有 `team` 的展示派生。
- `ContestTeamPanel.vue` 与 `ContestTeamWorkspaceSection.vue` 不再本地解释 `captain_user_id` 和邀请码展示格式。
