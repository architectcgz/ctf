> 状态：Current
> 事实源：platform challenges page model、route transports、前端架构 allowlist
> 替代：无

# Platform Challenges Route Owner Cleanup Plan

## 目标

- 收掉 `features/platform-challenges/model/useChallengeManagePage.ts -> vue-router`
- 收掉 `features/platform-challenges/model/usePlatformChallengeRoutePage.ts -> vue-router`

## 非目标

- 不重构 `usePlatformChallenges()` 的列表请求、筛选、排序或发布/删除流程。
- 不改题目详情页 `usePlatformChallengeDetailPage.ts` 的 route owner。
- 不把 platform challenge 的业务路由语义下沉到 shared composable。

## 输入依据

- `useChallengeManagePage.ts`
- `usePlatformChallengeRoutePage.ts`
- `useChallengeTopologyStudioRoutePage.ts`
- `useChallengeWriteupPage.ts`
- `useChallengeWriteupViewPage.ts`
- `routeNavigationTransport.ts`
- `routeQueryTransport.ts`
- `ChallengeManage.test.ts`
- `ChallengeTopologyStudio.test.ts`
- `ChallengeWriteup.test.ts`
- `architectureAllowlist.ts`

## 当前结论

- `platform-challenges` 这组剩余 allowlist 都落在同一个 feature，适合一组切片一起收口。
- 共享 transport 已经能承接 route `params` 读取和 `push()` 导航，本轮不需要再引入新的 bridge 或 feature 外 route wrapper。
- 题目导入预览、题目详情、拓扑、题解和题解编辑这些导航目标应保留在 `platform-challenges` 自己的 route target helper 中，避免业务路由语义漂到 shared。

## 设计边界

### platform challenges page model 本轮负责

- 保留题目目录列表、排序、筛选、发布/删除动作和空态文案 owner
- 保留 topology / writeup route page 的 mode 语义 owner
- 不再直接 import `vue-router`

### shared transports 本轮负责

- 提供 route `params / query` 读取
- 提供 `push()` 导航 transport
- 不承接 platform challenge 的业务 route 语义

### platform challenge routes 本轮负责

- 统一描述 import preview、detail、topology、writeup panel、writeup editor、import manage 的 route target
- 不承接列表加载、mode 判定或 UI workflow

## 任务切片

- [ ] Slice 1：抽 platform challenge route targets
  - 目标：
    - 新增 `platformChallengeRoutes.ts`
    - 管理页与 route page 的导航目标统一走本地 route target helper
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeManage.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts src/views/platform/__tests__/ChallengeWriteup.test.ts`

- [ ] Slice 2：page model 改用 shared route transports
  - 目标：
    - `useChallengeManagePage.ts` 改用 `useRouteNavigationTransport()`
    - `usePlatformChallengeRoutePage.ts` 改用 `useRouteQueryTransport()` + `useRouteNavigationTransport()`
    - 两个文件都不再 import `vue-router`
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeManage.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts src/views/platform/__tests__/ChallengeWriteup.test.ts`

- [ ] Slice 3：allowlist / review / backlog 收尾
  - 目标：
    - 更新 allowlist、raw-source 护栏、todo、review
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeManage.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts src/views/platform/__tests__/ChallengeWriteup.test.ts src/__tests__/architectureBoundaries.test.ts`
    - `cd code/frontend && npm run typecheck`

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeManage.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts src/views/platform/__tests__/ChallengeWriteup.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/platform-challenges-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-platform-challenges-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-platform-challenges-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-challenges/model/index.ts code/frontend/src/features/platform-challenges/model/platformChallengeRoutes.ts code/frontend/src/features/platform-challenges/model/useChallengeManagePage.ts code/frontend/src/features/platform-challenges/model/usePlatformChallengeRoutePage.ts code/frontend/src/views/platform/__tests__/ChallengeManage.test.ts code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts code/frontend/src/views/platform/__tests__/ChallengeWriteup.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `routeQueryTransport.ts` 目前仍同时依赖 `useRoute()` 与 `useRouter()`；这层 transport 应继续只保留 transport 语义，不能开始吸收平台题目特有逻辑。
- 本轮 review 默认仍是同上下文 self-review；独立 reviewer gate 仍需单独说明。
