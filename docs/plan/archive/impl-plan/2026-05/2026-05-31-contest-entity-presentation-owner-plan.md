# 竞赛实体展示 owner 收口计划

## Objective

- 强化 `entities/contest` 的稳定展示 owner，让 contest 状态 / 模式 / CTA / accent / status badge class 由实体层统一提供。
- 让 student contest list/detail workspace 与 platform contest table 直接消费实体层 owner，而不是保留 route page 透传或本地展示映射。

## Non-goals

- 不修改 contest 列表分页、筛选、倒计时、详情加载、队伍操作和 flag 提交流程。
- 不改 contest announcements、awd workspace、platform contest manage 的业务 workflow。
- 不把 route、toast、mutation 或 query sync 搬进实体层。

## Source Inputs

- `code/frontend/src/pages/contests/ContestListRoutePage.vue`
- `code/frontend/src/features/contest-detail/model/useContestListPage.ts`
- `code/frontend/src/widgets/contest-list-workspace/ContestListWorkspace.vue`
- `code/frontend/src/pages/contests/ContestDetailRoutePage.vue`
- `code/frontend/src/features/contest-detail/model/useContestDetailRoutePage.ts`
- `code/frontend/src/widgets/contest-detail-workspace/ContestDetailWorkspace.vue`
- `code/frontend/src/features/platform/contests/ui/PlatformContestTable.vue`
- `code/frontend/src/entities/contest/model/presentation.ts`
- `TODO/frontend-sliced-architecture.md`

## Task Slices

### Slice 1: 强化 contest entity presentation owner

- 目标：让 `entities/contest` 明确导出状态 badge class 与 accent css var style，而不是只保留文案映射。
- 变更面：
  - `code/frontend/src/entities/contest/model/presentation.ts`
  - `code/frontend/src/entities/contest/model/index.ts`
  - `code/frontend/src/entities/contest/index.ts`
  - `code/frontend/src/entities/contest/model/presentation.test.ts`
- 风险：
  - 如果把 route target、pagination 或 workflow 判定混进实体层，会重新污染 owner 边界。

### Slice 2: 改 student contest list/detail 消费面

- 目标：让 contest list/detail workspace 直接消费实体层展示规则，并移除 route page / feature model 的展示透传。
- 变更面：
  - `code/frontend/src/widgets/contest-list-workspace/ContestListWorkspace.vue`
  - `code/frontend/src/pages/contests/ContestListRoutePage.vue`
  - `code/frontend/src/features/contest-detail/model/useContestListPage.ts`
  - `code/frontend/src/widgets/contest-detail-workspace/ContestDetailWorkspace.vue`
  - `code/frontend/src/pages/contests/ContestDetailRoutePage.vue`
  - `code/frontend/src/features/contest-detail/model/useContestDetailRoutePage.ts`
- 风险：
  - 如果 feature model 仍继续返回展示 helper，就会留下双 owner。

### Slice 3: 改 platform contest table 状态展示 owner

- 目标：让平台竞赛表复用实体层状态 pill class owner，不再本地维护 `status -> class` 映射。
- 变更面：
  - `code/frontend/src/features/platform/contests/ui/PlatformContestTable.vue`
- 风险：
  - 如果平台表继续保留本地 class owner，contest 展示 owner 仍然不是唯一。

### Slice 4: 锁住边界与迁移台账

- 目标：让测试和台账明确 contest 展示规则已经进一步收口到实体层。
- 变更面：
  - `code/frontend/src/pages/contests/__tests__/ContestList.test.ts`
  - `code/frontend/src/pages/contests/__tests__/ContestDetail.test.ts`
  - `code/frontend/src/features/platform/contests/ui/PlatformContestTable.test.ts`
  - `TODO/frontend-sliced-architecture.md`
- 风险：
  - 如果测试只验证 UI，不验证 owner，后续回流难以及时暴露。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision contest-entity-presentation-owner`
- `npm run test:run -- src/pages/contests/__tests__/ContestList.test.ts src/pages/contests/__tests__/ContestDetail.test.ts src/features/platform/contests/ui/PlatformContestTable.test.ts src/entities/contest/model/presentation.test.ts`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-reuse-first.sh`
- `bash scripts/check-workflow-complete.sh`

## Review Focus

- `entities/contest` 是否只承接稳定对象展示规则，而没有吸入 workflow / route owner。
- contest list/detail route page 与 feature model 是否已停止透传展示规则。
- platform contest table 是否已经移除本地状态 pill class owner。
- 竞赛列表、详情和平台目录现有行为是否保持不变。

## Rollback / Recovery

- 如果实体层抽取后让 widget contract 变得更难读，可保留实体 model 函数但回退局部 UI 抽取，前提是展示 owner 仍然唯一且 route page / feature model 不再透传展示规则。
