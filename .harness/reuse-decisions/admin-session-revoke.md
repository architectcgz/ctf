# Reuse Decision

## Change type
handler / api / modal / service

## Existing code searched
- code/backend/internal/module/auth/contracts/token_service.go
- code/backend/internal/module/auth/infrastructure/token_service.go
- code/backend/internal/module/identity/api/http/handler.go
- code/backend/internal/module/identity/api/http/admin_user_types.go
- code/backend/internal/app/router_routes.go
- code/backend/internal/app/router.go
- code/frontend/src/api/admin/users.ts
- code/frontend/src/api/contracts.ts
- code/frontend/src/features/platform/user-management/ui/UserGovernanceDetailModal.vue
- code/frontend/src/features/platform/user-management/model/usePlatformUsers.ts

## Similar implementations found
- TokenService.RevokeAllUserSessions 已在 auth api http handler（密码变更场景）中调用，本次复用同一实现
- TokenService.DeleteSession 已存在，用于登出场景
- admin 路由注册模式：现有 admin/user CRUD 路由直接依赖 identityHandler，本次新增的 session 路由在 adminRouteDeps 上增加 tokenService 字段，沿用同一审计中间件、响应包装和错误处理路径
- AdminSurfaceModal 用于用户详情弹窗，遵循既有 modal 模板规范
- api/admin/users.ts 现有 CRUD 函数（getUsers / createUser / updateUser / deleteUser）使用统一的 request() 封装和 normalize 模式，本次新增的 session API 函数沿用同一模式

## Decision
extend_existing

## Reason
- 后端会话撤销能力（TokenService.DeleteSession / RevokeAllUserSessions）已存在，本次仅新增 ListUserSessions 查询方法和管理端 API 暴露入口
- 路由注册、审计、RBAC 保护均复用既有 admin routes 模式
- 前端 UI 在已有 UserGovernanceDetailModal 内新增 section，不新建独立页面或组件
- 不符合创建全新模块的条件：未改变模块边界，未引入新的仓储或外部依赖

## Files to modify
- code/backend/internal/module/auth/contracts/token_service.go
- code/backend/internal/module/auth/infrastructure/token_service.go
- code/backend/internal/module/auth/infrastructure/token_service_test.go
- code/backend/internal/module/auth/api/http/http_integration_test.go
- code/backend/internal/module/auth/application/commands/service_test.go
- code/backend/internal/module/identity/api/http/admin_user_types.go
- code/backend/internal/app/router_routes.go
- code/backend/internal/app/router.go
- code/frontend/src/api/contracts.ts
- code/frontend/src/api/admin/users.ts
- code/frontend/src/features/platform/user-management/model/usePlatformUserSessions.ts
- code/frontend/src/features/platform/user-management/ui/UserGovernanceDetailModal.vue
- code/frontend/src/pages/platform/__tests__/UserManage.test.ts
- code/backend/internal/app/router_session_routes_test.go

## After implementation
- 若后续 session 管理能力在其他页面复用（如审计页展示在线用户），在 `.harness/reuse-index/` 中添加 session API 模块级入口
