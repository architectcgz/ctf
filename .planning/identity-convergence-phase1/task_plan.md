# Task Plan

## Goal

启动 `auth + adminuser -> identity` 收敛，让 `identity` 真正成为用户主数据与认证状态 owner。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 `auth / adminuser / identity` 当前职责 | completed | 已确认 user owner 分散在 `auth + adminuser` |
| 2. 在 `identity` 下建立用户与鉴权 contract | completed | 新增 `UserRepository / AdminService / ProfileService / Authenticator` |
| 3. 收编 `adminuser` 能力 | completed | `adminuser` 已物理删除，`/api/v1/admin/users*` 接到 `identity/api/http` |
| 4. 缩减 `auth` 到登录/CAS/令牌交付 | completed | `auth.Service` 只保留注册/登录/密码校验，profile/password 下沉到 `identity` |
| 5. 更新 composition/router 与规则 | completed | 新增 `composition.IdentityModule`，router 不再直接装配 `adminuser` |
| 6. 定向验证 | completed | `identity/auth/app/system` 关键测试已限核通过 |

## Key Files

- `code/backend/internal/app/composition/auth_module.go`
- `code/backend/internal/app/router.go`
- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/module/auth/*`
- `code/backend/internal/module/adminuser/*`
- `code/backend/internal/module/identity/*`

## Acceptance Checks

- composition 不再直接 import `internal/module/adminuser`
- `identity` 对外统一暴露用户查询 / 用户写操作 / access token 解析能力
- `auth` 只保留认证流程与 provider 相关能力
- 定向测试通过：
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/identity/... -count=1`
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/auth -count=1`
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/app -run 'TestBuildRoot|TestIdentityModuleContractsCompile|TestCompositionModulesExposeContracts|TestNewRouterRegistersStudentChallengeRoutes|TestNewRouterUsesRuntimeHandlersForInstanceRoutes|TestRouterBuildUsesCompositionModules|TestArchitectureRulesRejectConcreteCrossModuleImports|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`
  - `GOMAXPROCS=2 go -C code/backend test -p 1 -parallel 1 ./internal/module/system -run 'TestHTTP_NotificationsSupportTicketListReadAndWebSocketPush' -count=1`

## Constraints

- 不改 `/api/v1/auth/*` 与 `/api/v1/admin/users*` 的外部路径
- 不保留新的兼容包装层
- 不让用户主数据继续分散在多个 root module
