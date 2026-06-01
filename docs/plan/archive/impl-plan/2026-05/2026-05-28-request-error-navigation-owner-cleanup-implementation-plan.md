> 状态：Current
> 事实源：`api/request.ts`、`main.ts`、`router/guards.ts`、`useWebSocket.ts`、相关测试与前端架构 review
> 替代：无

# Request Error Navigation Owner Cleanup Implementation Plan

## 目标

- 把请求层中的全局错误导航 owner 从 `api/request.ts` 收回到单一 runtime owner
- 一次性结束当前“429/500 已本地化、401 仍留在 request.ts” 的中间迁移状态
- 统一 HTTP、WebSocket、Vue runtime、router runtime 的全局错误跳转入口

## 非目标

- 本轮不让所有页面都去接入 429/500 专属错误页
- 本轮不重做 `errorStatusPage.ts` 的 route map 或错误页 UI
- 本轮不调整 router `beforeEach` 的登录/权限拦截语义

## 输入依据

- `code/frontend/src/api/request.ts`
- `code/frontend/src/api/__tests__/request.test.ts`
- `code/frontend/src/utils/errorStatusPage.ts`
- `code/frontend/src/utils/__tests__/errorStatusPage.test.ts`
- `code/frontend/src/main.ts`
- `code/frontend/src/router/guards.ts`
- `code/frontend/src/router/__tests__/guards.test.ts`
- `code/frontend/src/composables/useWebSocket.ts`
- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `docs/architecture/frontend/04-api-layer.md`
- `docs/architecture/frontend/08-build-deploy.md`

## 当前结论

- `request.ts` 现在已经不再为 429/500 跳状态页，但仍在 401 时直接 `logout + redirect`，这让 transport 层仍然持有一段导航 owner。
- 与此同时，全局错误跳转还分散在 `main.ts`、`router/guards.ts`、`useWebSocket.ts`，说明真正的问题不是单个 401 分支，而是 cross-cutting runtime error owner 没有独立落点。
- 最终 owner 应该是单一 runtime 模块：负责“哪些错误必须全局跳页”，而不是由请求层、websocket 和 bootstrap 各自判断。

## 设计边界

### `api/request.ts` 本轮负责

- `ApiError` 归一化
- `requestUrl` / `status` / `requestId` 等 transport 元数据保留
- 不再直接调用 store 或错误页导航 helper

### `runtime/globalErrorRuntime.ts` 本轮负责

- 安装全局 HTTP 401 处理
- 处理 WebSocket auth close
- 处理 Vue runtime error 到 `/500`
- 处理 router runtime error 到 `/500`

### 页面 / feature owner 本轮继续负责

- 429 / 5xx / 业务错误 / 网络错误的本地 toast、inline fallback、retry、draft 保留
- 当前已有 `ApiError` catch 分支的页面逻辑不变

## 任务切片

### Slice 1：抽 runtime global error owner

- 目标：
  - 新增 `runtime/globalErrorRuntime.ts`
  - 把 Vue runtime / router runtime / WebSocket auth close / HTTP 401 会话失效统一接到 runtime owner
- 验证：
  - `cd code/frontend && npm run test:run -- src/runtime/__tests__/globalErrorRuntime.test.ts src/router/__tests__/guards.test.ts`
- Review focus：
  - runtime owner 是否真的成为唯一全局导航入口
  - 401 是否仍是少数保留的 truly global auth/session failure

### Slice 2：清 request transport 层导航

- 目标：
  - `request.ts` 删除 store / redirect helper 依赖
  - `ApiError` 补足 runtime owner 需要的元数据
  - `request.test.ts` 改为断言“请求层只返回错误，不直接跳页”
- 验证：
  - `cd code/frontend && npm run test:run -- src/api/__tests__/request.test.ts src/utils/__tests__/errorStatusPage.test.ts`
- Review focus：
  - 是否彻底消除 request transport 层的导航 owner
  - 429/500/网络错误是否保持页面本地恢复能力

### Slice 3：事实源与 backlog 收尾

- 目标：
  - 更新前端 API / bootstrap 架构文档
  - 更新 backlog 进展
  - 归档 review
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 验证计划

- `cd code/frontend && npm run test:run -- src/api/__tests__/request.test.ts src/runtime/__tests__/globalErrorRuntime.test.ts src/utils/__tests__/errorStatusPage.test.ts src/router/__tests__/guards.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 401 全局失效策略仍然会跳 `/401`，这是当前明确保留的 truly global session failure；如果后续产品要改成无刷新回登录框，那是另一个专题。
- 页面 owner 目前并没有广泛接入 429/500 专属状态页，本轮只负责把 transport owner 清干净，不强推新的页面级 UX 策略。
