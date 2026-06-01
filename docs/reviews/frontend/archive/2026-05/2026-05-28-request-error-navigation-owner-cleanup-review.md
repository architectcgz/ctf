> 状态：Current
> 事实源：本次工作树 diff、相关 request/runtime/router 测试结果
> 替代：无

# Request Error Navigation Owner Cleanup Review

## Review target

- Repository: `ctf`
- Branch: `main`
- Diff source: 当前工作树未提交改动
- Files reviewed:
  - `code/frontend/src/api/request.ts`
  - `code/frontend/src/api/__tests__/request.test.ts`
  - `code/frontend/src/runtime/globalErrorRuntime.ts`
  - `code/frontend/src/runtime/__tests__/globalErrorRuntime.test.ts`
  - `code/frontend/src/router/guards.ts`
  - `code/frontend/src/router/__tests__/guards.test.ts`
  - `code/frontend/src/main.ts`
  - `code/frontend/src/composables/useWebSocket.ts`
  - `docs/architecture/frontend/04-api-layer.md`
  - `docs/architecture/frontend/08-build-deploy.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Classification check

- 结论：同意本次改动按 non-trivial frontend architecture cleanup 处理。
- 原因：改动同时触达请求层、bootstrap/runtime、router、WebSocket、架构事实源与 backlog，不是局部语句修补。

## Gate verdict

- 结论：`pass with minor issues`
- 说明：本次同上下文自审未发现 material finding，相关测试与类型检查已覆盖；但按 pipeline 标准，独立 review gate 仍未满足。

## Findings

- 无 material finding。

## Material findings

- 无。

## Senior implementation assessment

- 当前方案把 `request.ts` 收回到 transport owner，只保留 `ApiError` 标准化与 `requestUrl` 等元数据，这比在请求层直接 `logout + redirect` 更符合分层。
- `runtime/globalErrorRuntime.ts` 现在承接 HTTP `401`、WebSocket auth close、Vue runtime error 与 router runtime error，global owner 比之前更集中，也更容易测试。
- `router.onError` 已从 `router/guards.ts` 移除，只保留 bootstrap/runtime 这一处注册点，避免继续停留在双 owner 的过渡状态。
- 当前仍保留 `401` 作为 truly global session failure，这个取舍是合理的；`429 / 5xx / 网络错误 / 业务错误` 继续交给页面 / feature owner，避免把可恢复错误升级成全局跳页。

## Required re-validation

- 已执行，无额外 required re-validation。
- 如果后续继续下钻全局错误 owner，优先重跑：
  - `cd code/frontend && npm run test:run -- src/api/__tests__/request.test.ts src/runtime/__tests__/globalErrorRuntime.test.ts src/router/__tests__/guards.test.ts`
  - `cd code/frontend && npm run typecheck`

## Residual risk

- `installGlobalHttpErrorHandling()` 目前仍用模块级单次安装标志；应用实际只会 bootstrap 一次，所以产品风险可接受，但测试或未来多应用装配场景需要继续谨慎。
- `errorStatusPage.ts` 里仍保留 `429 / 5xx` 的 route map；这次已经把调用 owner 从请求层拿掉，但映射本身是否继续保留，是下一个更小的整理点。
- 独立 review gate 未满足：当前工具策略下没有启用子 agent reviewer，本次只能归档同上下文自审结果。

## Touched known-debt status

- 已触达的已知债务：`docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md` 中“请求层错误导航 owner”这一条。
- 收口结果：`request.ts` 不再直接替页面或 feature 决定可恢复错误的跳转；唯一保留的全局导航 owner 已收口到 `runtime/globalErrorRuntime.ts`。
- 未在本轮收口的相邻债务：`errorStatusPage.ts` 的历史 route map 还偏宽，但当前已经不再由 transport 层直接使用这些可恢复错误状态。
