> 状态：Current
> 事实源：auth feature model、layout session bridge、登录/注册页测试、前端架构 allowlist
> 替代：无

# Auth Router Owner Cleanup Plan

## 目标

- 把 `useAuth.ts` 从 route-aware feature helper 收口为纯 auth workflow helper。
- 将登录、注册、退出登录的导航 owner 回收到各自实际 workflow owner。
- 删除 `features/auth/model/useAuth.ts -> vue-router` 这条非 page owner allowlist。

## 非目标

- 不调整 `useLoginViewPage.ts` 当前的 redirect query sanitize 逻辑。
- 不修改登录页、注册页和顶部导航的视觉结构。
- 不继续处理 `featureRouterImportAllowlist` 里的其它 auth / route 条目。

## 输入依据

- `code/frontend/src/features/auth/model/useAuth.ts`
- `code/frontend/src/features/auth/model/useLoginPage.ts`
- `code/frontend/src/features/auth/model/useRegisterPage.ts`
- `code/frontend/src/features/auth/model/useLoginViewPage.ts`
- `code/frontend/src/widgets/layout-shell/model/useLayoutSessionActionsBridge.ts`
- `code/frontend/src/views/auth/__tests__/LoginView.test.ts`
- `code/frontend/src/views/auth/__tests__/RegisterView.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`

## 当前结论

- `useAuth.ts` 是跨页面 auth helper，不是单一 page owner，也不是 layout bridge，本身不该继续拿 `vue-router`。
- 登录 / 注册跳转属于各自页面 workflow 的成功后续动作，应回到 `useLoginPage.ts`、`useRegisterPage.ts`。
- 退出登录后跳回登录页属于 layout session bridge 的跨 feature 会话动作，放在 `useLayoutSessionActionsBridge.ts` 更合理。

## 设计边界

### `useAuth.ts` 本轮负责

- 调用 `loginApi` / `registerApi` / `logoutApi`
- 更新 `authStore`
- 发成功 toast
- 返回调用方继续导航所需的数据

### `useLoginPage.ts` 本轮负责

- 继续持有 `redirectTo` 读取
- 登录成功后决定 push 到 `redirectTo` 或角色首页
- 继续持有提交中、错误态和重复提交防护

### `useRegisterPage.ts` 本轮负责

- 注册成功后决定 push 到角色首页
- 继续持有提交中、错误态和重复提交防护

### `useLayoutSessionActionsBridge.ts` 本轮负责

- 退出登录后的 `/login` 导航
- 继续作为 layout shell 暴露的会话动作 bridge

## 任务切片

### Slice 1：收口 auth helper 的 router 依赖

- 目标：
  - `useAuth.ts` 去掉 `vue-router`
  - login / register 返回用户，logout 仅做 API/store/toast side effect
- 验证：
  - `cd code/frontend && npm run test:run -- src/features/auth/model/useAuth.test.ts`
- Review focus：
  - `useAuth.ts` 是否不再直接导航
  - store / toast side effect 是否保持现有行为

### Slice 2：把导航回收到真实 workflow owner

- 目标：
  - `useLoginPage.ts`、`useRegisterPage.ts`、`useLayoutSessionActionsBridge.ts` 承接导航
  - 保持登录、注册、退出登录的用户可见行为不变
- 验证：
  - `cd code/frontend && npm run test:run -- src/features/auth/model/useLoginPage.test.ts src/features/auth/model/useRegisterPage.test.ts src/widgets/layout-shell/model/useLayoutSessionActionsBridge.test.ts`
- Review focus：
  - redirect owner 是否清楚
  - page / layout owner 是否没有重复 side effect

### Slice 3：allowlist / backlog / review 收尾

- 目标：
  - 更新 `featureRouterImportAllowlist`
  - 记录 backlog 进展与 review
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/auth/__tests__/LoginView.test.ts src/views/auth/__tests__/RegisterView.test.ts`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - `useAuth.ts` 是否已从 allowlist 移除
  - 新增 router owner 是否落在更合理的 page owner

## 验证计划

- `cd code/frontend && npm run test:run -- src/features/auth/model/useAuth.test.ts src/features/auth/model/useLoginPage.test.ts src/features/auth/model/useRegisterPage.test.ts src/widgets/layout-shell/model/useLayoutSessionActionsBridge.test.ts src/views/auth/__tests__/LoginView.test.ts src/views/auth/__tests__/RegisterView.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/auth-router-owner-cleanup.md docs/plan/impl-plan/2026-05-29-auth-router-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-auth-router-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/auth/model/useAuth.ts code/frontend/src/features/auth/model/useAuth.test.ts code/frontend/src/features/auth/model/useLoginPage.ts code/frontend/src/features/auth/model/useLoginPage.test.ts code/frontend/src/features/auth/model/useRegisterPage.ts code/frontend/src/features/auth/model/useRegisterPage.test.ts code/frontend/src/widgets/layout-shell/model/useLayoutSessionActionsBridge.ts code/frontend/src/widgets/layout-shell/model/useLayoutSessionActionsBridge.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮会把 `featureRouterImportAllowlist` 的 `useAuth.ts` 挪走，但可能新增 `useLoginPage.ts` / `useRegisterPage.ts` 这类更明确的 page owner 条目，allowlist 总量未必下降。
- `useLoginViewPage.ts` 仍保留对 `redirect` query 的读取和 sanitize；若后续 auth 页面再增长，可能还要继续把 query owner 收回更完整的 page owner。
