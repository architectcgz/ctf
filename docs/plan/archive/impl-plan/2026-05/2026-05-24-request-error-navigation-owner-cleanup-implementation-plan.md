> 状态：Current
> 事实源：2026-05-24 前端架构 review、当前 `request.ts` 错误拦截策略、现有页面级错误 owner
> 替代：无

# Request Error Navigation Owner Cleanup Implementation Plan

## 目标

- 把 recoverable HTTP 错误的导航决策从 `code/frontend/src/api/request.ts` 收回到页面或 feature owner。
- 保留请求层的 transport owner：错误标准化、取消请求处理、session 失效场景下的全局 auth 跳转。

## 非目标

- 本轮不重写全站页面的错误展示方式。
- 本轮不改 `errorStatusPage.ts` 的状态页路由映射。
- 本轮不调整 `ApiError` 结构或 API envelope 解析策略。

## 输入依据

- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `code/frontend/src/api/request.ts`
- `code/frontend/src/api/__tests__/request.test.ts`
- `code/frontend/src/utils/errorStatusPage.ts`
- `code/frontend/src/features/platform-overview/model/usePlatformOverviewPage.ts`
- `code/frontend/src/features/student-directory/model/useStudentDirectoryQuery.ts`

## 当前结论

- 现在 `request.ts` 在响应拦截器里直接处理 `429`、`401` 和通用 HTTP 错误页跳转。
- 页面 owner 同时又在本地 catch 后维护 `error`、toast 或重试入口，导致错误 UX owner 分裂。
- 这次最小收口方式不是全站补错误页逻辑，而是先让除 session 失效以外的错误回到调用方决定。

## 任务切片

### Slice 1：收缩请求层全局错误跳转 owner

- 目标：
  - `429` 与普通 `4xx/5xx` 失败不再由请求层强制跳错误页。
  - `401` 的全局 session 失效跳转保持不变。
- 预期改动：
  - `docs/plan/impl-plan/2026-05-24-request-error-navigation-owner-cleanup-implementation-plan.md`
  - `.harness/reuse-decisions/frontend-architecture-review-prompt-and-platform-feature-split.md`
  - `code/frontend/src/api/request.ts`
  - `code/frontend/src/api/__tests__/request.test.ts`
- 依赖：
  - 继续复用现有 `ApiError`、`mapErrorCode`、`errorStatusPage` 和 `authStore.logout()` 逻辑。
  - 不新增并行请求层或新的全局错误处理抽象。
- 验证：
  - `git diff --check -- docs/plan/impl-plan/2026-05-24-request-error-navigation-owner-cleanup-implementation-plan.md .harness/reuse-decisions/frontend-architecture-review-prompt-and-platform-feature-split.md code/frontend/src/api/request.ts code/frontend/src/api/__tests__/request.test.ts`
  - `npm run test:run -- src/api/__tests__/request.test.ts src/utils/__tests__/errorStatusPage.test.ts`
  - `npm run typecheck`
  - `bash scripts/check-consistency.sh`
- Review focus：
  - 请求层是否只保留 transport / session owner，而不是继续决定 recoverable failure 的页面导航。
  - `401` 的全局 logout + redirect 是否仍只在明确需要的路径触发。
  - 新测试是否覆盖 `429`、通用 HTTP 错误和 `401` 的 owner 边界。

## 风险

- 某些页面如果历史上完全依赖全局错误页，而本地没有错误展示 owner，收口后可能只剩 reject 路径未消费。
- `401` 行为依赖 `shouldRedirectToErrorStatusPage`，若调用方 URL 命名继续扩展，auth entry flow 的排除条件仍要单独维护。

## 回退方式

- 如本轮收口引出页面级回归，可回退 `request.ts` 与对应测试，恢复此前的 `429/普通 HTTP` 全局跳转策略。
- 计划和 reuse decision 文档可保留，作为该方向曾评估过的证据。
