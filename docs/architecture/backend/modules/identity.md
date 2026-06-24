# identity 模块设计

> 状态：Current
> 事实源：`code/backend/internal/module/identity/`、`code/backend/internal/app/composition/identity_module.go`
> 替代：无

## 定位

`identity` 是用户、角色、资料和基础权限事实 owner，给认证、教师聚合、管理员用户管理和当前用户解析提供稳定用户能力。

## 事实来源

- HTTP 入口：`code/backend/internal/module/identity/api/http/handler.go`
- 用例编排：`code/backend/internal/module/identity/application/commands/`、`code/backend/internal/module/identity/application/queries/`
- 对外契约：`code/backend/internal/module/identity/contracts/`
- 持久化：`code/backend/internal/module/identity/infrastructure/repository.go`
- 装配入口：`code/backend/internal/module/identity/runtime/module.go`

## 当前设计

- `identity/api/http`
  - 负责：管理员用户管理 HTTP 入口、请求响应映射。
  - 不负责：会话签发、CAS 认证、Token 存储或教师复盘聚合。
- `identity/application/commands`
  - 负责：管理员用户变更、profile command、认证过程需要的用户同步 / authenticator 能力。
  - 不负责：登录凭证校验和 token 颁发，这些由 `auth` 编排。
- `identity/application/queries`
  - 负责：管理员用户查询、profile query、基础用户 lookup。
  - 不负责：教学复盘、班级洞察、排行榜或训练进度这类跨 owner 聚合。
- `identity/contracts`
  - 负责：`UserRepository`、profile service、role 常量和公开错误等稳定接口。
  - 不负责：暴露 GORM model 或 HTTP DTO。
- `identity/infrastructure`
  - 负责：用户、角色、资料的 PostgreSQL 读写。
  - 不负责：Redis token、审计日志、通知或业务报表输出。

## API 入口设计

| 路由 | Handler | Service / 用例 |
| --- | --- | --- |
| `GET /api/v1/admin/users` | `identity/api/http.Handler.ListUsers` | `queries.AdminService.ListUsers`，管理员用户列表 |
| `POST /api/v1/admin/users` | `Handler.CreateUser` | `commands.AdminService.CreateUser`，创建用户和角色信息 |
| `POST /api/v1/admin/users/import` | `Handler.ImportUsers` | `commands.AdminService.ImportUsers`，批量导入学生/教师账号 |
| `PUT /api/v1/admin/users/:id` | `Handler.UpdateUser` | `commands.AdminService.UpdateUser`，更新用户资料、角色和状态 |
| `DELETE /api/v1/admin/users/:id` | `Handler.DeleteUser` | `commands.AdminService.DeleteUser`，软删除或禁用用户 |
| `GET /api/v1/admin/users/:id/sessions` | app router + auth token service | 管理端查看用户会话，token owner 仍是 `auth` |
| `DELETE /api/v1/admin/users/:id/sessions/:sid` / `DELETE /api/v1/admin/users/:id/sessions` | app router + auth token service | 管理端吊销用户会话，identity 只提供用户事实 |

## Application / Service 设计

| Service | 代码路径 | 负责 | 不负责 |
| --- | --- | --- | --- |
| Admin command service | `identity/application/commands/admin_service.go` | 管理端创建、导入、更新、删除用户 | token 吊销、审计持久化 |
| Profile command service | `identity/application/commands/profile_service.go` | 当前用户资料变更和 profile command contract | 登录、CAS、session 签发 |
| Authenticator service | `identity/application/commands/authenticator_service.go` | 为 auth 提供用户查找、密码校验前置和 CAS 同步能力 | 认证协议细节 |
| Admin query service | `identity/application/queries/admin_service.go` | 管理端用户列表、详情和筛选 | 业务面板聚合 |
| Profile query service | `identity/application/queries/profile_service.go` | 当前用户资料和基础用户 lookup | 教师复盘聚合 |

## 数据设计

| 表 | Entity | Owner 语义 | 主要写入方 |
| --- | --- | --- | --- |
| `users` | `identity/entity.User` | 用户账号、资料、角色字段、状态、登录失败计数和软删除 | identity command |
| `roles` | `identity/entity.Role` | 系统角色参考数据，当前角色包括 `student`、`teacher`、`admin`；当前没有独立角色 CRUD API | migration / seed 数据 |
| `user_roles` | `identity/entity.UserRole` | 用户与角色的关联事实 | identity command |

缓存和会话不属于 `identity`：Redis token / session 由 `auth` 拥有，教学聚合缓存由各查询 owner 或 `ops` / `assessment` 拥有。

## 边界

- `identity` 拥有用户主数据和角色事实。
- `auth` 只消费 identity contract，不反向拥有用户资料。
- `teaching_analysis` 可消费 identity 基础用户 lookup 拼装教师视角，但不写 identity 状态。
- `ops` 审计里出现的用户 ID / 角色只是记录，不成为用户事实源。

## 主要用例

- 管理端用户创建、启禁、用户角色字段维护和列表查询；独立角色定义不由管理端 API 维护。
- 当前用户资料读取和资料修改所需的 profile command / query。
- 认证链路使用的用户查找、密码校验前置资料和 CAS 用户同步。
- 教师聚合使用的基础用户 lookup。

## 数据与副作用

- PostgreSQL：`identity/entity/user.go`、`identity/entity/role.go` 是模块持久化实体。
- 密码/资料变更：由 identity command 维护用户事实，token 失效或会话处理由 `auth` 负责。
- 审计：用户管理动作通过 app 层注入的审计能力记录，不在 identity 仓储中复制审计表。

## 跨模块依赖

| 依赖 | 用途 | 接线 |
| --- | --- | --- |
| `auth` | auth 消费 identity 用户与 profile 能力 | `composition.BuildAuthModule` |
| `teaching_analysis` | 教师聚合需要基础用户信息 | `composition.BuildTeachingAnalysisModule` |
| `middleware` | 认证后角色和当前用户解析 | `code/backend/internal/app/router.go` |

## Guardrail

- `code/backend/internal/module/identity/architecture_test.go`：禁止 concrete Go 文件停留在根包，约束 API / application / contracts 分层。
- `code/backend/internal/app/router_user_self_domain_routes_test.go`：覆盖当前用户域路由归属。
- `code/backend/internal/app/router_admin_identity_routes.go`：管理员 identity 路由集中注册。

## 已知限制

- 当前没有独立权限策略服务；角色判断由 identity contract、middleware 和各 handler/application 的权限检查共同消费。
