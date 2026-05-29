> 状态：Current
> 事实源：audit log page model、route query transport、前端架构 allowlist
> 替代：无

# Audit Log Route Owner Cleanup Plan

## 目标

- 收掉 `features/audit-log/model/useAuditLogPage.ts -> vue-router`

## 非目标

- 不改 audit log 的筛选交互、列表渲染和详情弹窗结构。
- 不把 audit log 的 query sync 再抽成新的 feature route wrapper。
- 不重做 audit log 的 auto-apply 策略。

## 输入依据

- `useAuditLogPage.ts`
- `routeQueryTransport.ts`
- `AuditLog.vue`
- `AuditLog.test.ts`
- `auditLogPageStateExtraction.test.ts`
- `architectureAllowlist.ts`

## 当前结论

- audit log 这条的 route owner 核心是 query hydrate 和 replace，并不涉及 route param、角色跳转或复杂 route contract。
- `routeQueryTransport.ts` 已经能承接“读取 query + replace query”这层 transport，本轮直接复用比新增 wrapper 更小也更合理。
- audit log feature 自己仍然应该持有 query normalize、auto-apply 节奏、分页刷新和请求取消 owner。

## 设计边界

### audit log page model 本轮负责

- 保留 query normalize、hydrate、auto-apply、分页、请求取消和加载流程
- 不再直接 import `vue-router`

### shared transport 本轮负责

- 提供当前 query 读取与 `replaceQuery()`
- 不承接 audit-log-specific query schema、默认值或节流策略

## 任务切片

- [ ] Slice 1：page model 改用 query transport
  - 目标：
    - `useAuditLogPage.ts` 去掉 `vue-router`
    - query hydrate 与 replace 改成消费 `routeQueryTransport.ts`
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/AuditLog.test.ts`

- [ ] Slice 2：allowlist / review / backlog 收尾
  - 目标：
    - 更新 allowlist、测试护栏、todo、review
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/AuditLog.test.ts src/views/platform/__tests__/auditLogPageStateExtraction.test.ts src/__tests__/architectureBoundaries.test.ts`
    - `cd code/frontend && npm run typecheck`

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/AuditLog.test.ts src/views/platform/__tests__/auditLogPageStateExtraction.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/audit-log-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-audit-log-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-audit-log-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/audit-log/model/useAuditLogPage.ts code/frontend/src/views/platform/__tests__/AuditLog.test.ts code/frontend/src/views/platform/__tests__/auditLogPageStateExtraction.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `routeQueryTransport.ts` 仍然只是 transport；如果后续 audit log 继续增长 query schema，不应把 normalize/default 也继续堆进 shared composable。
- auto-apply 的节流和请求取消都还在 page model，本轮需要确认迁移 transport 后这些行为不变。
