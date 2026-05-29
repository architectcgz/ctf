> 状态：Current
> 事实源：`useWebSocket.ts`、前端架构 allowlist、各 realtime feature model 调用面
> 替代：无

# WebSocket Composable Boundary Cleanup Plan

## 目标

- 去掉 `useWebSocket.ts` 里不必要的 `store` 依赖。
- 清空 `composableMultiBoundaryAllowlist`。
- 保持现有 WebSocket ticket、会话失效、心跳和重连语义不变。

## 非目标

- 不重写 WebSocket 生命周期逻辑。
- 不改 realtime feature model 的业务 payload 处理。
- 不顺手清理 `featureRouterImportAllowlist`。

## 输入依据

- `code/frontend/src/composables/useWebSocket.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/features/notifications/model/useNotificationRealtime.ts`
- `code/frontend/src/features/scoreboard/model/useContestScoreboardRealtime.ts`
- `code/frontend/src/features/contest-announcements/model/useContestAnnouncementRealtime.ts`
- `code/frontend/src/features/awd-inspector/model/useContestAwdPreviewRealtime.ts`

## 当前结论

- `useWebSocket.ts` 的 `store` 边界来自未使用的 `useAuthStore` import，不是实际 owner 需求。
- `useWebSocket.ts` 继续保留 `api/auth` 的 ticket 获取和 `runtime/globalErrorRuntime` 的 session 失效处理是合理的 shared transport / runtime 边界。

## 设计边界

### `useWebSocket.ts` 本轮负责

- WebSocket 连接、关闭、发送
- 心跳、pong timeout、自动重连
- 通过 `getWsTicket()` 获取 ticket
- 通过 `handleGlobalSessionExpired()` 处理鉴权关闭

### `useWebSocket.ts` 本轮不负责

- 直接读取 Pinia store
- 承接任一业务 feature 的 payload 映射或 UI 提示

## 任务切片

### Slice 1：去掉无效 store 依赖

- 目标：
  - 删除 `useWebSocket.ts` 中未使用的 `useAuthStore` import
- 验证：
  - `cd code/frontend && npm run typecheck`
- Review focus：
  - 行为是否零变化
  - composable 边界是否从 `api+store` 收回到 `api`

### Slice 2：收掉 allowlist 与文档

- 目标：
  - 清空 `composableMultiBoundaryAllowlist`
  - 更新 backlog / review
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - allowlist 是否真实下降
  - 文档是否准确说明“误挂 store 边界被移除”，而不是泛化成新的结构规则

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/websocket-composable-boundary-cleanup.md docs/plan/impl-plan/2026-05-29-websocket-composable-boundary-cleanup-plan.md docs/reviews/frontend/2026-05-29-websocket-composable-boundary-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/composables/useWebSocket.ts code/frontend/src/__tests__/architectureAllowlist.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮没有新增 `useWebSocket.ts` 的行为测试，主要依赖现有架构护栏与 typecheck；如果仓库后续想更强约束 WebSocket shared owner，应该再补专门的 composable 测试。
- 这轮 review 默认仍是同上下文 self-review；独立 reviewer gate 仍需单独说明。
