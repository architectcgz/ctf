> 状态：Current
> 事实源：platform challenge detail page model、route transports、前端架构 allowlist
> 替代：无

# Platform Challenge Detail Route Owner Cleanup Plan

## 目标

- 收掉 `features/platform-challenge-detail/model/usePlatformChallengeDetailPage.ts -> vue-router`

## 非目标

- 不改 `usePlatformChallengeDetailRoutePage.ts` 的 query-tab owner。
- 不重构题目详情页的 Flag 草稿、附件下载和详情加载流程。
- 不把 challenge detail 的失败跳转策略下沉到 shared transport。

## 输入依据

- `usePlatformChallengeDetailPage.ts`
- `usePlatformChallengeDetailRoutePage.ts`
- `routeNavigationTransport.ts`
- `routeQueryTransport.ts`
- `ChallengeDetail.test.ts`
- `routeQueryTabsAdoption.test.ts`
- `architectureAllowlist.ts`

## 当前结论

- challenge detail 这条只剩 params 读取和几条薄导航，复杂度明显低于 `contest-detail` 或 `awd-review-detail`，适合在 platform-challenges 之后紧接着收。
- 由于 `platform-challenge-detail` 和 `platform-challenges` 是并列 feature，本轮不把 detail 页硬绑到前者的 route helper；而是在 detail feature 内保持自己的本地 route target contract。

## 设计边界

### platform challenge detail page model 本轮负责

- 保留挑战详情加载、Flag draft、附件下载和失败后的延迟跳转 owner
- 不再直接 import `vue-router`

### shared transports 本轮负责

- 提供 route `params` 读取
- 提供 `push()` 导航 transport
- 不承接 challenge detail 特有的失败跳转、Flag 保存或附件下载语义

### platform challenge detail routes 本轮负责

- 统一描述题库列表、拓扑、题解查看和题解编辑目标路由
- 不承接加载、计时器或错误处理

## 任务切片

- [ ] Slice 1：抽 detail feature 本地 route targets
  - 目标：
    - 新增 `platformChallengeDetailRoutes.ts`
    - 统一描述返回题库、拓扑、题解查看/编辑的 route target
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeDetail.test.ts`

- [ ] Slice 2：page model 改用 shared route transports
  - 目标：
    - `usePlatformChallengeDetailPage.ts` 去掉 `vue-router`
    - `challengeId` 改由 `routeQueryTransport` 读取
    - 导航与失败重定向改走 `routeNavigationTransport`
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeDetail.test.ts src/__tests__/architectureBoundaries.test.ts`

- [ ] Slice 3：allowlist / review / backlog 收尾
  - 目标：
    - 更新 allowlist、测试护栏、todo、review
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeDetail.test.ts src/__tests__/architectureBoundaries.test.ts`
    - `cd code/frontend && npm run typecheck`

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeDetail.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/platform-challenge-detail-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-platform-challenge-detail-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-platform-challenge-detail-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-challenge-detail/model/index.ts code/frontend/src/features/platform-challenge-detail/model/platformChallengeDetailRoutes.ts code/frontend/src/features/platform-challenge-detail/model/usePlatformChallengeDetailPage.ts code/frontend/src/views/platform/__tests__/ChallengeDetail.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `platform-challenge-detail` 和 `platform-challenges` 目前都会持有本地 route target helper；后续如果 challenge authoring 路由继续增长，再评估是否需要一个更中立的 challenge authoring route owner，而不是提前为了复用合并两个 feature。
- 这轮 review 默认仍是同上下文 self-review；独立 reviewer gate 仍需单独说明。
