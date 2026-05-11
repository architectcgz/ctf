# Auth / Identity 边界阶段 1 Implementation Plan

## Objective

按 `docs/design/backend-module-boundary-target.md` 的阶段 1 先落一小段迁移：

- 去掉 `identity -> auth` 的反向依赖
- 保持 `auth` 继续作为 session、token、CAS、WebSocket ticket 的 owner
- 不改变现有 HTTP / WebSocket 行为和组合方式

## Non-goals

- 不在这一轮拆 `runtime -> instance + container_runtime`
- 不重命名模块目录
- 不改外部路由、cookie 行为、session 存储或 CAS 流程
- 不更新论文，只更新当前实现计划和必要的代码 / 架构测试

## Inputs

- `docs/design/backend-module-boundary-target.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/identity/contracts/auth.go`
- `code/backend/internal/module/identity/runtime/module.go`
- `code/backend/internal/app/composition/identity_module.go`
- `code/backend/internal/module/architecture_allowlist_test.go`

## Current Baseline

- `identity/contracts/auth.go` 里定义了 `identitycontracts.Authenticator`，本质只是包装 `authcontracts.TokenService`
- `identity/runtime/module.go` 通过 `identitycmd.NewAuthenticatorService` 把 `authcontracts.TokenService` 再包装一层
- `app/composition/identity_module.go` 继续把这层包装暴露给路由和其他组合模块
- `middleware.Auth`、`ops` 通知 WS、`contest` 实时 WS 已经直接依赖 `authcontracts.TokenService`
- `allowedModuleDependencies` 仍显式允许 `identity -> auth`

## Chosen Direction

这次不新增共享认证抽象，直接让 `authcontracts.TokenService` 成为唯一 token contract，并把 token service 从 `identity` 模块完全推出去：

1. `identity/contracts` 删除 `Authenticator` 包装接口
2. `identity/runtime` 和 `app/composition/identity_module.go` 不再持有 token service
3. `app/router.go` 统一创建 token service，并传给 `auth` 组合、认证中间件、通知 WS 和竞赛实时 WS
4. 移除 `identity/application/commands/authenticator_service.go` 里的历史包装实现，但不删除文件
5. 更新 allowlist 和架构测试，明确新的依赖方向

这样可以先把反向依赖真正收掉，同时不动认证行为 owner 和外部接口。

## Ownership Boundary

- `auth`
  - 负责：session、token、CAS、WS ticket 以及相关 contract
  - 不负责：用户资料 owner、用户仓储实现、管理端用户 CRUD
- `identity`
  - 负责：用户查询、资料修改、密码修改、管理端用户能力
  - 不负责：定义 token contract、包装 session/token service
- `composition`
  - 负责：在 `app/router.go` 把 `authcontracts.TokenService` 传给 `middleware`、`ops`、`contest` 实时 handler 和 `auth` runtime
  - 不负责：在 `identity` 模块内部重新创造另一套 auth abstraction

## Change Surface

- Modify: `code/backend/internal/app/router.go`
- Modify: `code/backend/internal/app/full_router_integration_test.go`
- Modify: `code/backend/internal/app/composition/auth_module.go`
- Modify: `code/backend/internal/module/identity/contracts/auth.go`
- Modify: `code/backend/internal/module/identity/runtime/module.go`
- Modify: `code/backend/internal/module/identity/application/commands/authenticator_service.go`
- Modify: `code/backend/internal/app/composition/identity_module.go`
- Modify: `code/backend/internal/app/router_test.go`
- Modify: `code/backend/internal/module/identity/architecture_test.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`

## Task Slices

### Slice 1: 收口 token contract

目标：

- `identity` 不再定义 `Authenticator`
- `identity` runtime / composition 不再持有 token service
- token service 只在外层 app 组合点和 `auth` runtime 之间流动

Validation:

- `rg -n "identitycontracts\\.Authenticator|AuthenticatorService|identity -> auth" code/backend/internal -g '*.go'`
- `go test ./internal/module/identity/... ./internal/app/... -run 'Identity|Composition|Router'`

Review focus:

- 是否还保留 `identity -> auth` 的包装语义
- token service 是否真的离开了 `identity` 模块

### Slice 2: 更新架构 guardrail

目标：

- 删掉 `allowedModuleDependencies` 里的 `identity -> auth`
- 更新 runtime / composition 的 typed deps 测试

Validation:

- `go test ./internal/module/... -run 'Architecture|Allowlist'`

Review focus:

- 架构测试是否真的阻止旧方向回流
- 新 typed deps 断言是否和目标边界一致

### Slice 3: 定向回归验证

目标：

- 确认 router / composition / identity runtime 仍可编译和通过测试

Validation:

- `go test ./internal/module/identity/... ./internal/module/auth/... ./internal/middleware/... ./internal/app/...`

Review focus:

- 是否有 compile-time contract 断言遗漏
- 是否有路由装配或 WS handler 因类型变化失配

## Risks

- `identity/application/commands/authenticator_service.go` 如果直接删除，会扩大本次切片并触发高风险删除确认；因此本轮只做去语义化保留文件
- `composition.IdentityModule` 被多个路由 / 模块消费，字段类型变化如果漏改测试，会在 `router_test` 里暴露
- `buildAuthModule` 签名变更会联动 router build mock；如果漏改，集成测试会直接编译失败
- allowlist 改掉后，任何残留 `identity -> auth` import 都会直接炸架构测试

## Verification Plan

1. `cd code/backend && rg -n "identitycontracts\\.Authenticator|AuthenticatorService|identity -> auth" internal -g '*.go'`
2. `cd code/backend && go test ./internal/module/identity/... ./internal/module/auth/... ./internal/middleware/... ./internal/app/...`
3. `cd code/backend && go test ./internal/module/... -run 'Architecture|Allowlist'`
4. `bash scripts/check-consistency.sh`

## Architecture-Fit Evaluation

- 目标 owner 明确：token contract 归 `auth`，用户 owner 归 `identity`
- 共享能力 landing zone 明确：直接使用 `auth/contracts/token_service.go`
- 本轮不是“让代码还能跑就行”，而是把 `identity -> auth` 反向依赖从 contract、runtime、composition 和架构测试一起收口
- 本切片完成后，不应再需要第二轮“把 identity 里的 auth 包装去掉”的重复改造
