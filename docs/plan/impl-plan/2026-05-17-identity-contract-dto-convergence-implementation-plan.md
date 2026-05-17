# identity contract DTO 收口实现方案

## Objective

把 `identity/contracts` 对 `internal/dto` 的直接依赖收回到模块内 contract 类型，先完成 `admin` 与 `profile` 两条对外服务面的收口；HTTP 层继续维持现有请求/响应 DTO 与外部接口不变。

## Non-goals

- 不迁移整个仓库的 `internal/dto`
- 不调整外部 HTTP 路由、JSON 字段、分页结构和状态码
- 不顺手重做 `auth` 登录响应、数据库模型或仓储接口

## Inputs

- `docs/architecture/backend/01-system-architecture.md`
- `docs/architecture/backend/04-api-design.md`
- `docs/architecture/backend/07-modular-monolith-refactor.md`
- `code/backend/internal/module/identity/contracts/*.go`
- `code/backend/internal/module/identity/api/http/*.go`
- `code/backend/internal/module/auth/api/http/*.go`

## Task Slices

1. 定义 `identity/contracts` 自有 query / response 类型
   - 在 `identity/contracts/admin.go` 声明 `AdminUserListQuery`、`AdminUser`、`ImportUsersResult`
   - 在 `identity/contracts/profile.go` 声明 `ProfileUser`
   - 验证：contract 文件不再 import `internal/dto`

2. 收口 identity application 输出
   - `application/commands` 与 `application/queries` 改为返回 `identity/contracts` 类型
   - 更新 response mapper 和对应单测
   - 验证：`go test ./internal/module/identity/...`

3. 把 HTTP DTO 映射压回 API 边界
   - `identity/api/http` 负责 `dto <-> contracts` 映射
   - `auth/api/http` 负责把 `identitycontracts.ProfileUser` 转成 `dto.AuthUser`
   - 验证：`go test ./internal/module/identity/... ./internal/module/auth/...`

4. 守住新边界
   - 更新 `identity/architecture_test.go`，禁止 `contracts` 继续依赖 `internal/dto`
   - 验证：`go test ./internal/module/identity/...`

## Expected Changes

- `code/backend/internal/module/identity/contracts/`
- `code/backend/internal/module/identity/application/commands/`
- `code/backend/internal/module/identity/application/queries/`
- `code/backend/internal/module/identity/api/http/`
- `code/backend/internal/module/auth/api/http/`
- `code/backend/internal/module/identity/architecture_test.go`

## Compatibility

- 外部 HTTP 请求/响应仍复用现有 `internal/dto`，前端和集成测试不需要跟着改协议
- 这次只改变模块内 contract 落点，不改变数据存储、权限和分页语义

## Validation

- `go test ./internal/module/identity/...`
- `go test ./internal/module/auth/...`

## Review Focus

- `identity/contracts` 是否已经不再依赖 `internal/dto`
- API 层是否成为唯一 DTO 映射 owner，而不是把 DTO 继续漏进 application / contracts
- `auth` 消费 `identity` profile 时是否仍通过 contract，而不是重新耦合 HTTP DTO

## Rollback

- 如果收口引起编译或映射回归，可先回退本次 contract 类型替换，恢复 `identity/contracts -> internal/dto` 依赖，再按更小切片重做
