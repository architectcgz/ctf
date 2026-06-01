> 状态：Current
> 事实源：AWD config page / service selection owner、前端架构 allowlist、ContestAwdConfig 测试
> 替代：无

# AWD Challenge Selection Router Owner Cleanup Plan

## 目标

- 把 `useAwdChallengeSelection.ts` 从 route-aware helper 收口回纯 service selection owner。
- 让 `useContestAwdConfigPage.ts` 保留唯一 router/query owner。
- 删除对应 `featureRouterImportAllowlist` 条目。

## 非目标

- 不重构 AWD checker 草稿、预览或保存流程。
- 不改 AWD 配置页的 UI 结构和服务列表展示方式。
- 不处理 `featureRouterImportAllowlist` 其它 feature 条目。

## 输入依据

- `code/frontend/src/features/contest-awd-config/model/useAwdChallengeSelection.ts`
- `code/frontend/src/features/contest-awd-config/model/useContestAwdConfigPage.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts`

## 当前结论

- `useAwdChallengeSelection.ts` 当前职责是 service selection / fallback / sorting；route query 写回混在这里，会把 router owner 漂进 helper 层。
- `useContestAwdConfigPage.ts` 已经持有 `useRoute()`、`useRouter()` 和 `contestId`，继续作为唯一 query/router owner 更合理。

## 设计边界

### `useContestAwdConfigPage.ts` 本轮负责

- `useRoute()` / `useRouter()` 获取
- 读取 `service` query
- 把服务选择结果写回路由 query
- 返回 AWD 工作台的导航动作

### `useAwdChallengeSelection.ts` 本轮负责

- 当前服务选择 state
- query value 与本地选择值的 reconcile
- checker 类型与排序派生
- 调用外部注入的 query read / replace callback

## 任务切片

### Slice 1：selection helper 去掉 router 依赖

- 目标：
  - 从 `useAwdChallengeSelection.ts` 移除 `vue-router` 类型依赖
  - 改成 callback 注入 query read / replace
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts`
- Review focus：
  - helper 是否已经不再 import `vue-router`
  - fallback service 与手动切换 service 的 query 同步是否保持不变

### Slice 2：allowlist / 护栏 / backlog 收尾

- 目标：
  - 删除 allowlist 条目
  - 更新 raw-source 护栏与 backlog 记录
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/ContestAwdConfig.test.ts`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - router owner 是否只回到 page，而没有漂到别的 helper
  - allowlist 是否真实下降

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/ContestAwdConfig.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/awd-challenge-selection-router-owner-cleanup.md docs/plan/impl-plan/2026-05-29-awd-challenge-selection-router-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-awd-challenge-selection-router-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/contest-awd-config/model/useAwdChallengeSelection.ts code/frontend/src/features/contest-awd-config/model/useContestAwdConfigPage.ts code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮只收口一条 `featureRouterImportAllowlist`，不代表剩余条目都不合理；仍需逐条判定。
- 这轮 review 默认仍是同上下文 self-review；独立 reviewer gate 仍需单独说明。
