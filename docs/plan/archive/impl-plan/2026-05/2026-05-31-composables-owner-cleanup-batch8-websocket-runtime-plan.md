# 2026-05-31 composables owner cleanup batch8 websocket runtime plan

> 状态：Draft
> 关联 reuse decision：`.harness/reuse-decisions/composables-owner-cleanup-batch8-websocket-runtime.md`

## 目标

把 `useWebSocket` 从历史 `code/frontend/src/composables/` 收口到 `shared/model/realtime/`，并同步修正 realtime 消费方、类型引用、测试 mock 与架构文档。

## 非目标

- 不改 `getWsTicket()`、心跳间隔、pong timeout、重连次数、close code 或 session 过期回退语义
- 不改通知、公告、排行榜、AWD 预览各自的业务 payload 解析逻辑
- 不处理后端 realtime 端点

## 输入事实源

- `docs/architecture/frontend/01-architecture-overview.md`
- `docs/architecture/frontend/03-state-management.md`
- `docs/architecture/frontend/05-websocket-composables.md`
- `docs/architecture/frontend/08-build-deploy.md`
- `code/frontend/src/composables/useWebSocket.ts`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`

## 目标归属

- `useWebSocket` -> `shared/model/realtime/useWebSocket.ts`

理由：

- 它是共享 realtime runtime owner，依赖 `@/api/auth` 和 `@/runtime/globalErrorRuntime`
- `shared/model/common` / `shared/lib` 都不适合承接这类 API/runtime 依赖

## 任务切片

### Slice 1

迁移 owner：

- 新建 `shared/model/realtime/useWebSocket.ts`
- 删除 `composables/useWebSocket.ts`
- 修正运行时代码与类型引用

验证：

- `cd code/frontend && timeout 180s npm run typecheck`

### Slice 2

修正测试与文档：

- 修正所有 `vi.mock('@/composables/useWebSocket')`
- 更新 `01-architecture-overview.md`
- 更新 `03-state-management.md`
- 更新 `05-websocket-composables.md`
- 更新 `08-build-deploy.md`

验证：

- `cd code/frontend && timeout 180s npm run test:run -- src/features/contest-announcements/model/useContestAnnouncementRealtime.test.ts src/features/notifications/model/useNotificationRealtime.test.ts src/features/scoreboard/model/useContestScoreboardRealtime.test.ts src/pages/contests/__tests__/ContestDetail.test.ts src/pages/scoreboard/__tests__/ScoreboardView.test.ts`
- `python3 scripts/check-docs-consistency.py`
- `git diff --check`

## 风险点

- `WebSocketStatus` 同时被 feature model、shared model 和 shared ui 作为类型引用，路径漏改会直接卡 typecheck
- `?raw` 断言虽然这批不多，但测试里的 `vi.mock` 分布广，容易漏
- `shared/model/realtime` 是新目录，文档需要同步说明 owner 理由

## Review focus

- `useWebSocket` 是否确实落在 `shared/model/realtime`，而不是被塞进 `common` / `lib`
- 是否只发生 owner 迁移，没有混入行为变化
- 是否清干净了旧 `src/composables/useWebSocket` 路径与 mock
