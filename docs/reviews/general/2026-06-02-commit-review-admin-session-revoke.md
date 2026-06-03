# 5766b556c review

## Review target

- Repository: `ctf`
- Branch: `main`
- Commit: `5766b556c064f9ce268a4b184aaa986a59b5ba6c`
- Scope: 管理员手动撤销用户会话（后端 API + 前端用户详情弹窗）
- Files reviewed:
  - `code/backend/internal/app/router.go`
  - `code/backend/internal/app/router_routes.go`
  - `code/backend/internal/module/auth/contracts/token_service.go`
  - `code/backend/internal/module/auth/infrastructure/token_service.go`
  - `code/backend/internal/module/auth/infrastructure/token_service_test.go`
  - `code/backend/internal/module/identity/api/http/admin_user_types.go`
  - `code/frontend/src/api/admin/users.ts`
  - `code/frontend/src/api/contracts.ts`
  - `code/frontend/src/features/platform/user-management/model/usePlatformUserSessions.ts`
  - `code/frontend/src/features/platform/user-management/ui/UserGovernanceDetailModal.vue`
  - `code/frontend/src/features/platform/user-management/ui/UserGovernancePage.vue`
  - `code/frontend/src/pages/platform/__tests__/UserManage.test.ts`

## Classification check

- Route chosen: `HARNESS`
- Task type: independent code review
- Complexity: non-trivial
- Specialist skill: `code-reviewer`

## Gate verdict

- `blocked`

## Findings

### 1. Blocker: 撤销全部确认态会跨用户泄漏，破坏二次确认语义

- Location:
  - `code/frontend/src/features/platform/user-management/ui/UserGovernanceDetailModal.vue:33-41`
  - `code/frontend/src/features/platform/user-management/ui/UserGovernanceDetailModal.vue:205-224`
  - `code/frontend/src/features/platform/user-management/ui/UserGovernancePage.vue:118-123`
- Risk:
  - `UserGovernanceDetailModal` 组件实例在页面里常驻，只有内部 `AdminSurfaceModal` 受 `v-if="user"` 控制。
  - `showRevokeAllConfirm` 只在点击“撤销全部”、取消或成功确认后修改，没有在关闭弹窗或切换用户时重置。
  - 结果是：管理员在用户 A 上打开“撤销全部”确认框后直接关闭弹窗，再打开用户 B，确认框会直接保留到用户 B 上。此时再次点击“确认撤销全部会话”会针对当前 `userId` 执行，确认动作已经脱离最初触发它的用户上下文。
- Fix direction:
  - 在用户切换和弹窗关闭时显式重置 `showRevokeAllConfirm`。
  - 更稳妥的做法是把确认态绑定到当前用户 ID，或把确认 UI 挪到一次性弹层中，避免跨会话残留。

### 2. Blocker: 提交说明里的“AbortController 防竞态”并没有真正落地

- Location:
  - `code/frontend/src/features/platform/user-management/model/usePlatformUserSessions.ts:20-33`
  - `code/frontend/src/features/platform/user-management/model/usePlatformUserSessions.ts:88-108`
  - `code/frontend/src/api/admin/users.ts:147-166`
- Risk:
  - `fetchController` 从未赋值，也没有把 `signal` 传入 `getUserSessions` / `revoke*` API。
  - `watch(userId)` 里虽然调用了 `fetchController?.abort()`，但当前实现实际上不会取消任何请求。
  - 这意味着快速切换用户详情时，旧请求仍可能晚于新请求返回并覆盖 `sessions`，让管理员看到错误用户的会话列表。这个提交把“防竞态”写进了 commit message，但代码没有实现对应保证。
- Fix direction:
  - 给 `getUserSessions` 增加 `options?: { signal?: AbortSignal }`，在 composable 里为每次拉取创建并保存 controller。
  - 对取消错误做专门分支，不要把被取消的请求当成失败去清空列表或打日志。
  - 如果不想手写控制器，按项目现有模式复用 `useAbortController` / `latestRequestId` 这类 owner 明确的实现。

### 3. Major: 新增会话管理路径没有被自动化测试真正覆盖，现有页面测试已经暴露了未 mock 的真实请求

- Location:
  - `code/backend/internal/app/router_routes.go:386-463`
  - `code/frontend/src/pages/platform/__tests__/UserManage.test.ts:110-130`
- Risk:
  - 后端只补了 `TokenService.ListUserSessions` 的基础单测，没有覆盖新增的三个管理员路由，也没有覆盖“会话不属于该用户”“已过期会话返回特定文案”等新分支。
  - 前端没有针对 `usePlatformUserSessions` 或详情弹窗新增交互测试。现有 `UserManage.test.ts` 在打开详情弹窗后会触发真实 `getUserSessions` 请求，测试仍然通过，但 stderr 已经出现 `ApiError: 网络连接失败` 警告。
  - 这说明当前测试既没有验证新功能，也没有阻止未处理的网络副作用进入测试环境，因此前两条问题很容易漏过去。
- Fix direction:
  - 后端至少补一组路由级测试，覆盖 list / revoke-one / revoke-all 的 happy path 与关键错误分支。
  - 前端补 `usePlatformUserSessions` 或详情弹窗测试，覆盖用户切换取消、确认态重置、单条撤销与全部撤销。
  - 页面测试里应显式 mock `getUserSessions`，并在需要时断言是否带上 `AbortSignal`。

## Material findings

- 重置“撤销全部”确认态，确保不会跨用户 / 跨开关周期残留。
- 真正实现会话列表请求的取消与防竞态。

## Senior implementation assessment

- 后端整体思路是顺着既有 `TokenService` 和 admin 路由模式扩展，方向是对的，变更面也不算大。
- 风险主要集中在前端状态 owner 没有真正收口：会话请求状态、取消逻辑、确认态生命周期都还留在一个半完成状态。更低风险的实现方式是直接复用项目里已经存在的“带 signal 的分页/请求 owner”模式，把“当前用户 + 当前请求 + 当前 destructive confirm” 三个状态绑定在同一个 composable owner 里，而不是只在 UI 层做布尔值拼装。

## Required re-validation

- 前端：
  - 打开用户 A 的详情，点“撤销全部”后直接关闭，再打开用户 B，确认框不应残留。
  - 快速在两个用户之间切换详情，最终展示的会话列表必须与最后打开的用户一致。
  - 新增 / 更新对应 Vitest 用例，确保不会再出现未 mock 的真实网络请求。
- 后端：
  - 补路由级测试后，复验 list / revoke-one / revoke-all 的成功与关键失败分支。

## Residual risk

- 本次没有跑前端 typecheck / 全量测试，只做了最小相关验证。
- 后端路由层目前缺少直接测试证据；我只能基于代码路径和已有单测判断。

## Touched known-debt status

- 本次 review 未发现提交直接触达当前 review index 里已登记、且必须在 touched surface 内同时收口的结构性债面。

## Validation evidence

- `cd code/backend && go test ./internal/module/auth/infrastructure -run 'TestTokenServiceListUserSessions' -count=1`
  - 结果：通过
- `cd code/frontend && npm run test:run -- src/pages/platform/__tests__/UserManage.test.ts`
  - 结果：测试通过，但 stderr 出现 `getUserSessions` 触发的真实网络失败日志
- `cd code/backend && go test ./internal/app -run 'TestFullRouter.*Admin.*User|TestFullRouterStateMatrix' -count=1`
  - 结果：`internal/app` 下无匹配测试，未提供新路由覆盖证据
