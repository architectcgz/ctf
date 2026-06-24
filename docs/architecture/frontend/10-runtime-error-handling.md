# 前端运行时错误处理

> 状态：Current
> 事实源：`code/frontend/src/main.ts`、`code/frontend/src/runtime/globalErrorRuntime.ts`、`code/frontend/src/api/request.ts`、`code/frontend/src/shared/model/realtime/useWebSocket.ts`
> 替代：无

## 定位

本文档说明前端运行时全局错误 owner、HTTP 401 会话失效、Vue / Router 崩溃跳转和 WebSocket 鉴权关闭的处理边界。

- 覆盖：`setupGlobalErrorRuntime()`、Axios 响应拦截器、Vue `app.config.errorHandler`、`router.onError()`、WebSocket auth close fallback。
- 不覆盖：页面内表单错误、业务 toast、列表空态、导出轮询 retry 或 WebSocket 消息业务语义。

## 当前设计

- `code/frontend/src/main.ts`
  - 负责：在 `app.use(pinia)`、`app.use(router)` 后调用 `setupGlobalErrorRuntime(app, router, pinia)`，保证全局错误处理能拿到 router 与 auth store。
  - 不负责：在入口文件中处理页面级错误展示或业务重试。

- `code/frontend/src/runtime/globalErrorRuntime.ts`
  - 负责：安装 Axios response interceptor；当错误是 `ApiError` 且 `status === 401` 时，调用 `handleGlobalSessionExpired()`，执行 `authStore.logout()` 并按 `shouldRedirectToErrorStatusPage(401, requestUrl)` 判断是否跳转 `/401`。
  - 不负责：处理普通 `429 / 5xx / 网络错误` 的页面提示或自动重试；这些错误继续由 feature / page owner 处理。

- `createGlobalVueErrorHandler()`
  - 负责：记录 Vue runtime error；如果错误不是 `ApiError`，跳转 `/500`。
  - 不负责：把 API 请求错误升级为全局崩溃页。

- `createGlobalRouterErrorHandler()`
  - 负责：记录 router runtime error 并跳转 `/500`。
  - 不负责：替 router guard 做权限判断或登录恢复；这些仍在 `router/guards.ts`。

- `code/frontend/src/shared/model/realtime/useWebSocket.ts`
  - 负责：当 WebSocket close code 是 `4001` 或 `4401` 时调用 `handleGlobalSessionExpired()`，与 HTTP 401 使用同一会话失效入口。
  - 不负责：把普通断线、心跳超时和重连失败都变成全局状态页；连接状态仍由 composable 暴露给使用方。

## 边界

- 全局错误导航只允许通过 `runtime/globalErrorRuntime.ts` 进入。
- `api/request.ts` 只负责构造标准化 `ApiError`，不直接跳状态页、不弹 toast、不做 refresh token 重试。
- 页面 / feature 仍拥有可恢复错误的 UX，例如 inline error、toast、retry、draft 保留和导出轮询。
- 当前没有全局 `window.onunhandledrejection` / `unhandledrejection` owner；未被 Vue、router 或 Axios 捕获的 Promise rejection 需要在调用方自己处理。

## 接口或数据影响

- HTTP 401 会清空前端 auth store，并在 `shouldRedirectToErrorStatusPage()` 允许时跳转错误页。
- `/500` 当前主要代表 Vue runtime error 或 router runtime error，不代表所有 HTTP 5xx。
- 运行时错误处理不改变 OpenAPI 契约；错误对象字段仍来自 `ApiError.message / code / requestId / status / errors / requestUrl`。
- 相关环境变量仍由请求层和 realtime 层读取：`VITE_API_BASE_URL`、`VITE_API_TIMEOUT`、`VITE_WS_BASE_URL`。

## Guardrail

- `code/frontend/src/runtime/__tests__/globalErrorRuntime.test.ts`：覆盖 HTTP 401、Vue error、router error 和安装接线。
- `code/frontend/src/api/__tests__/request.test.ts`：覆盖 `ApiError` 标准化和请求层不直接接管全局导航的边界。
- `code/frontend/src/router/__tests__/guards.test.ts`：覆盖 router guard 的登录态恢复与默认跳转。
- `code/frontend/src/shared/model/realtime/useWebSocket.ts` 由调用方测试覆盖实时连接行为；新增全局鉴权关闭行为时应补 runtime 或 realtime owner 测试。

## 历史迁移

- 当前已从“请求层遇到错误直接跳状态页”收口为 runtime owner；请求层只返回标准化错误。
- 当前认证模式依赖 HttpOnly session cookie，没有 refresh token 自动重试链路。
