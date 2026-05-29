> 状态：Current
> 事实源：platform challenge route wrappers、platform route views、前端架构 allowlist
> 替代：无

# Platform Challenge Route Page Convergence Plan

## 目标

- 把 `useChallengeTopologyStudioRoutePage.ts`、`useChallengeWriteupPage.ts`、`useChallengeWriteupViewPage.ts` 三份重复 route wrapper 合并成一个显式 `usePlatformChallengeRoutePage(mode)`。
- 在不改 route view 调用面的前提下，把这组 allowlist 从 3 条收敛到 1 条。

## 非目标

- 不改 platform challenge route view 的模板结构。
- 不处理 `usePlatformChallengeDetailRoutePage.ts`。
- 不改 challenge topology / writeup 的业务状态 owner。

## 输入依据

- `code/frontend/src/features/platform-challenges/model/useChallengeTopologyStudioRoutePage.ts`
- `code/frontend/src/features/platform-challenges/model/useChallengeWriteupPage.ts`
- `code/frontend/src/features/platform-challenges/model/useChallengeWriteupViewPage.ts`
- `code/frontend/src/views/platform/ChallengeTopologyStudio.vue`
- `code/frontend/src/views/platform/ChallengeWriteup.vue`
- `code/frontend/src/views/platform/ChallengeWriteupView.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- `code/frontend/src/views/platform/__tests__/ChallengeWriteup.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`

## 当前结论

- 三个 wrapper 本质上是同一个 route page owner，只在返回 query 和附加动作上有微小差异。
- 继续保留三份独立 `useRoute/useRouter` 会让 allowlist 和重复逻辑一起冻结。
- 保持旧 public API 名字，对外委托到统一 wrapper，是最小可审阅改动。

## 设计边界

### `usePlatformChallengeRoutePage(mode)` 本轮负责

- 读取 `challengeId`
- 返回题目详情导航
- 在 `writeup-view` 模式下提供“去题解编辑页”导航

### 旧三个 wrapper 本轮负责

- 仅做 mode-specific 委托和向后兼容 public API
- 不再直接持有 `vue-router`

## 任务切片

### Slice 1：新增统一 route-aware page wrapper

- 目标：
  - 新增 `usePlatformChallengeRoutePage(mode)`
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - 只有一个文件继续 import `vue-router`

### Slice 2：旧 wrapper 收口为纯委托层

- 目标：
  - 三个旧 wrapper 不再直接拿 router
  - route view 调用面不变
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeTopologyStudio.test.ts src/views/platform/__tests__/ChallengeWriteup.test.ts`
- Review focus：
  - raw-source 护栏是否仍符合“view 不直接 useRoute/useRouter”

### Slice 3：allowlist / review / backlog 收尾

- 目标：
  - 更新 allowlist、review 和 backlog
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts src/views/platform/__tests__/ChallengeWriteup.test.ts`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - 这组三条 allowlist 是否已收敛成一条

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts src/views/platform/__tests__/ChallengeWriteup.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/platform-challenge-route-page-convergence.md docs/plan/impl-plan/2026-05-29-platform-challenge-route-page-convergence-plan.md docs/reviews/frontend/2026-05-29-platform-challenge-route-page-convergence-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/platform-challenges/model/usePlatformChallengeRoutePage.ts code/frontend/src/features/platform-challenges/model/useChallengeTopologyStudioRoutePage.ts code/frontend/src/features/platform-challenges/model/useChallengeWriteupPage.ts code/frontend/src/features/platform-challenges/model/useChallengeWriteupViewPage.ts code/frontend/src/features/platform-challenges/model/index.ts code/frontend/src/features/platform-challenges/index.ts code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts code/frontend/src/views/platform/__tests__/ChallengeWriteup.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮只收 platform challenge route wrapper 的重复 owner，不继续并动 `usePlatformChallengeDetailRoutePage.ts`。
