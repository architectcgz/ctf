> 状态：Current
> 事实源：contest awd config page model、route transports、前端架构 allowlist
> 替代：无

# Contest AWD Config Route Owner Cleanup Plan

## 目标

- 收掉 `features/contest-awd-config/model/useContestAwdConfigPage.ts -> vue-router`

## 非目标

- 不改 `useAwdChallengeSelection.ts` 的 service fallback / selected service owner。
- 不重构 checker draft、preview 或 save workflow。
- 不改 `ContestEdit.vue` 的 panel 路由语义。

## 输入依据

- `useContestAwdConfigPage.ts`
- `useAwdChallengeSelection.ts`
- `routeNavigationTransport.ts`
- `routeQueryTransport.ts`
- `ContestAwdConfig.test.ts`
- `architectureAllowlist.ts`

## 当前结论

- 这条只有 params/query 与单条 back navigation，复杂度明显低于 `challenge-detail` 或 `awd-review-detail-workspace`。
- `service` query 写回可以直接复用 `replaceQuery()`，不需要新造一层 query helper。

## 设计边界

### contest awd config page model 本轮负责

- 保留 mounted 初始化、breadcrumb、service 选择初始化、checker draft / preview / save owner
- 不再直接 import `vue-router`

### shared transports 本轮负责

- 提供 route `params / query` 读取
- 提供 query 写回和 `push()` 导航
- 不承接 AWD 配置页自己的 service fallback / checker workflow

### contest awd config routes 本轮负责

- 描述返回 `ContestEdit?panel=awd-config` 的 route target
- 不承接 query 同步和 mounted 初始化

## 任务切片

- [ ] Slice 1：抽本地 back route target
  - 目标：
    - 新增 `contestAwdConfigRoutes.ts`
    - 返回工作台导航不再在 page model 内直接拼 route object
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts`

- [ ] Slice 2：page model 改用 shared route transports
  - 目标：
    - `contestId` / `service` query 改用 `routeQueryTransport`
    - `replaceServiceQuery()` 改用共享 `replaceQuery()`
    - `goBackToStudio()` 改用 `routeNavigationTransport`
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts src/__tests__/architectureBoundaries.test.ts`

- [ ] Slice 3：allowlist / review / backlog 收尾
  - 目标：
    - 更新 allowlist、测试护栏、todo、review
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts src/__tests__/architectureBoundaries.test.ts`
    - `cd code/frontend && npm run typecheck`

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestAwdConfig.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/contest-awd-config-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-contest-awd-config-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-contest-awd-config-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/contest-awd-config/model/contestAwdConfigRoutes.ts code/frontend/src/features/contest-awd-config/model/index.ts code/frontend/src/features/contest-awd-config/model/useContestAwdConfigPage.ts code/frontend/src/views/platform/__tests__/ContestAwdConfig.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `replaceQuery()` 当前只写 query；如果后续 AWD 配置页开始依赖 hash 或更复杂的 route normalize，再评估是否需要更强的 query-route contract，而不是现在提前抽象。
- 这轮 review 默认仍是同上下文 self-review；独立 reviewer gate 仍需单独说明。
