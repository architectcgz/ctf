> 状态：Current
> 事实源：auth login page model、router guards、前端架构 allowlist
> 替代：无

# Auth Login Route Owner Cleanup Plan

## 目标

- 收掉 `features/auth/model/useLoginPage.ts -> vue-router`
- 收掉 `features/auth/model/useLoginPage.ts -> @/router/guards`

## 非目标

- 不重做登录页 UI、文案或提交交互。
- 不调整 `useAuth.ts`、注册页和 layout session bridge 的 owner。
- 不重写 router guards 的登录态恢复、权限校验或标题更新逻辑。

## 输入依据

- `useLoginPage.ts`
- `useLoginPage.test.ts`
- `useLoginViewPage.ts`
- `router/guards.ts`
- `router/__tests__/guards.test.ts`
- `routeNavigationTransport.ts`
- `routeQueryTransport.ts`
- `architectureAllowlist.ts`

## 当前结论

- `useLoginPage.ts` 当前真正需要的 route 能力只有 redirect query 读取和登录成功后跳转。
- `sanitizeRedirectPath()` 被放在 `router/guards.ts` 只是历史位置问题，本质上是可共享的 path sanitize helper，更适合放进中性 util。

## 设计边界

### login page model 本轮负责

- 保留表单态、提交 guard、probe 提示、默认 dashboard fallback 与错误文案 owner
- 不再直接 import `vue-router`
- 不再直接依赖 router guards 模块

### shared transports 本轮负责

- 提供 redirect query 读取
- 提供登录成功后的 `push()` 导航
- 不承接 redirect sanitize 或默认 dashboard 决策

### redirect path util 本轮负责

- 提供安全的 redirect path sanitize
- 供 router guards 与 login page 共用
- 不承接登录成功默认目标决策

## 任务切片

- [ ] Slice 1：抽 redirect sanitize util
  - 目标：
    - 新增中性 redirect path util
    - router guards 与 login page 共享同一实现
  - 验证：
    - `cd code/frontend && npm run test:run -- src/router/__tests__/guards.test.ts src/features/auth/model/useLoginPage.test.ts`

- [ ] Slice 2：login page 改用 shared route transports
  - 目标：
    - redirect query 改用 `routeQueryTransport`
    - 登录成功跳转改用 `routeNavigationTransport`
  - 验证：
    - `cd code/frontend && npm run test:run -- src/features/auth/model/useLoginPage.test.ts src/router/__tests__/guards.test.ts src/__tests__/architectureBoundaries.test.ts`

- [ ] Slice 3：allowlist / review / backlog 收尾
  - 目标：
    - 更新 allowlist、review、todo
  - 验证：
    - `cd code/frontend && npm run test:run -- src/features/auth/model/useLoginPage.test.ts src/router/__tests__/guards.test.ts src/__tests__/architectureBoundaries.test.ts`
    - `cd code/frontend && npm run typecheck`

## 验证计划

- `cd code/frontend && npm run test:run -- src/features/auth/model/useLoginPage.test.ts src/router/__tests__/guards.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/auth-login-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-auth-login-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-auth-login-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/auth/model/useLoginPage.ts code/frontend/src/features/auth/model/useLoginPage.test.ts code/frontend/src/router/guards.ts code/frontend/src/utils/redirectPath.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `resolveLoginRedirectTarget()` 仍留在 `useLoginViewPage.ts` 这个历史命名文件内；本轮不顺带处理命名债，只聚焦 allowlist 清空。
- 这轮 review 默认仍是同上下文 self-review；独立 reviewer gate 仍需单独说明。
