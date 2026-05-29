> 状态：Current
> 事实源：challenge detail page model、route transports、前端架构 allowlist
> 替代：无

# Challenge Detail Route Owner Cleanup Plan

## 目标

- 收掉 `features/challenge-detail/model/useChallengeDetailPage.ts -> vue-router`

## 非目标

- 不重做学生侧 challenge detail 的 page shell、workspace tabs 或题解 / 实例交互结构。
- 不调整 `useUrlSyncedTabs()` 当前基于 URL query 的 page-level tab 同步方式。
- 不处理 `auth` 剩余的 route / guard allowlist。

## 输入依据

- `useChallengeDetailPage.ts`
- `ChallengeDetail.vue`
- `ChallengeDetail.test.ts`
- `routeNavigationTransport.ts`
- `routeQueryTransport.ts`
- `challengeListRoutes.ts`
- `architectureAllowlist.ts`

## 当前结论

- 这条 route owner 比 `auth` 更单纯，只有 params 读取和单条 back navigation。
- 当前 feature 已经稳定持有 challenge 加载、tab 预取、实例 workflow 和错误态 owner；本轮只收 transport，不改这些业务归属。

## 设计边界

### challenge detail page model 本轮负责

- 保留 challenge 加载、错误态、题解 / 提交记录 / writeup 预取、实例 workflow、solution tab keyboard 和 page reset owner
- 不再直接 import `vue-router`

### shared transports 本轮负责

- 提供 route params 读取
- 提供 `push()` 导航
- 不承接 challengeId normalize、错误态判定或 challenge detail workflow

### challenge detail routes 本轮负责

- 描述返回题目列表的 route target
- 不承接 panel query sync、题目加载或实例动作

## 任务切片

- [ ] Slice 1：抽本地 back route target
  - 目标：
    - 新增 `challengeDetailRoutes.ts`
    - 返回题目列表不再在 page model 里直接写路径字符串
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/challenges/__tests__/ChallengeDetail.test.ts`

- [ ] Slice 2：page model 改用 shared route transports
  - 目标：
    - `challengeId` 改用 `routeQueryTransport`
    - `goBackToChallengeList()` 改用 `routeNavigationTransport`
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/challenges/__tests__/ChallengeDetail.test.ts src/__tests__/architectureBoundaries.test.ts`

- [ ] Slice 3：allowlist / review / backlog 收尾
  - 目标：
    - 更新 allowlist、测试护栏、todo、review
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/challenges/__tests__/ChallengeDetail.test.ts src/__tests__/architectureBoundaries.test.ts`
    - `cd code/frontend && npm run typecheck`

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/challenges/__tests__/ChallengeDetail.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/challenge-detail-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-challenge-detail-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-challenge-detail-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/challenge-detail/model/challengeDetailRoutes.ts code/frontend/src/features/challenge-detail/model/index.ts code/frontend/src/features/challenge-detail/model/useChallengeDetailPage.ts code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `useUrlSyncedTabs()` 仍通过 `window.location` 同步 page query；本轮不把这层一并改到共享 transport。
- 这轮 review 默认仍是同上下文 self-review；独立 reviewer gate 仍需单独说明。
