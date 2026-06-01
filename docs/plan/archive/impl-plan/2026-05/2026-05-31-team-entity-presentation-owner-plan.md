# 队伍实体展示 owner 收口计划

## Objective

- 新建 `entities/team`，统一承接 team 对象的稳定展示规则，包括成员数、队长关系、成员展示项、空态文案和邀请码文案。
- 让 contest detail 的 team 区块与 route 派生直接消费 `entities/team`，而不是继续在 feature model / feature ui 中散落 team 展示 owner。

## Non-goals

- 不修改创建队伍、加入队伍、踢出成员、toast、confirm 和 dialog workflow。
- 不改 contest detail 的加载、公告、题目、AWD 战场和 flag 提交流程。
- 不把 route、mutation、权限工作流或异步状态搬进实体层。

## Source Inputs

- `code/frontend/src/features/contest-detail/ui/ContestTeamPanel.vue`
- `code/frontend/src/features/contest-detail/ui/ContestTeamWorkspaceSection.vue`
- `code/frontend/src/features/contest-detail/model/useContestDetailPage.ts`
- `code/frontend/src/features/contest-detail/model/useContestDetailRoutePage.ts`
- `code/frontend/src/pages/contests/__tests__/ContestDetail.test.ts`
- `code/frontend/src/entities/contest/model/presentation.ts`
- `code/frontend/src/entities/notification/model/presentation.ts`
- `TODO/frontend-sliced-architecture.md`

## Brainstorming Conclusion

- 推荐方向：先用 `entities/team/model/presentation.ts` 收 team 的稳定展示派生，不急着抽 team 专属 UI 组件。
- 原因：当前消费面只有 contest detail 一条主链路，先把 owner 从 route model / feature ui 挪干净，保持最小可审阅切片；如果后续出现第二个消费面，再决定是否补 `entities/team/ui`。
- TDD：需要。这里会改动派生展示逻辑和 raw source owner guardrail，应该先用实体层测试和源码级断言锁住边界。

## Plan Review Result

- `team` 的展示 owner 与 workflow owner 边界明确：
  - `entities/team` 负责“队伍对象如何稳定展示”
  - `features/contest-detail` 继续负责“用户如何创建 / 加入 / 踢人”
- 当前切片不会留下“先把展示 helper 抽出来，下一刀再清 route owner”的半收口状态；本轮直接一起收掉 route 派生和 feature ui 内联展示解释。

## Task Slices

### Slice 1: 建立 team entity presentation owner

- 目标：新增 `entities/team` 公共入口，提供 team 展示派生函数和测试。
- 变更面：
  - `code/frontend/src/entities/team/index.ts`
  - `code/frontend/src/entities/team/model/index.ts`
  - `code/frontend/src/entities/team/model/presentation.ts`
  - `code/frontend/src/entities/team/model/presentation.test.ts`
- 风险：
  - 如果把 workflow 文案或按钮行为一起搬进实体层，会污染 owner。

### Slice 2: 迁移 team feature ui 与 route 派生

- 目标：让 `ContestTeamPanel.vue`、`ContestTeamWorkspaceSection.vue`、`useContestDetailPage.ts`、`useContestDetailRoutePage.ts` 改为消费实体层展示规则。
- 变更面：
  - `code/frontend/src/features/contest-detail/ui/ContestTeamPanel.vue`
  - `code/frontend/src/features/contest-detail/ui/ContestTeamWorkspaceSection.vue`
  - `code/frontend/src/features/contest-detail/model/useContestDetailPage.ts`
  - `code/frontend/src/features/contest-detail/model/useContestDetailRoutePage.ts`
- 风险：
  - 如果 feature model 还保留本地 `members.length` 或 `captain_user_id` 派生，会形成双 owner。

### Slice 3: 锁住边界并更新迁移台账

- 目标：补测试和迁移记录，明确 `team` 已经收口为实体展示 owner。
- 变更面：
  - `code/frontend/src/pages/contests/__tests__/ContestDetail.test.ts`
  - `TODO/frontend-sliced-architecture.md`
- 风险：
  - 如果只测渲染结果，不测 owner 回流，后续很容易退化。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision team-entity-presentation-owner`
- `npm run test:run -- src/pages/contests/__tests__/ContestDetail.test.ts src/entities/team/model/presentation.test.ts`
- `git diff --check -- .harness/reuse-decisions/team-entity-presentation-owner.md docs/plan/impl-plan/2026-05-31-team-entity-presentation-owner-plan.md code/frontend/src/entities/team/index.ts code/frontend/src/entities/team/model/index.ts code/frontend/src/entities/team/model/presentation.ts code/frontend/src/entities/team/model/presentation.test.ts code/frontend/src/features/contest-detail/ui/ContestTeamPanel.vue code/frontend/src/features/contest-detail/ui/ContestTeamWorkspaceSection.vue code/frontend/src/features/contest-detail/model/useContestDetailPage.ts code/frontend/src/features/contest-detail/model/useContestDetailRoutePage.ts code/frontend/src/pages/contests/__tests__/ContestDetail.test.ts TODO/frontend-sliced-architecture.md`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-reuse-first.sh`
- `bash scripts/check-workflow-complete.sh`

## Review Focus

- `entities/team` 是否只承接稳定 team 展示规则，没有吸入 workflow / async owner。
- contest detail 的 team feature ui 和 route model 是否已停止本地解释 `captain_user_id`、`members.length` 和邀请码展示。
- 用户当前的创建队伍、加入队伍、踢人行为是否保持不变。

## Rollback / Recovery

- 如果实体层抽取后的展示项对象让 feature 代码更难读，可以回退具体 helper 形态，但不能回退 owner 边界；`team` 展示规则仍必须停留在 `entities/team` 公共入口。
