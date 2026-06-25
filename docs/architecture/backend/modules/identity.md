# identity 模块设计

> 状态：Current
> 事实源：`code/backend/internal/module/identity/`、`code/backend/internal/app/composition/identity_module.go`
> 替代：无

## 定位

`identity` 是用户、角色、资料和基础权限事实 owner，给认证、教师聚合、管理员用户管理和当前用户解析提供稳定用户能力。

## 本文档范围

| 本文档负责 | 本文档不负责（见其他文档） |
|-----------|-------------------------|
| identity 模块的职责边界、HTTP 入口和用例组织 | 认证流程和会话管理 → `auth` 模块文档 |
| 用户、角色、资料的真相源和基础权限模型 | 权限判定和访问控制实现 → `docs/architecture/backend/design/权限与访问控制.md` |
| 用户管理命令和查询的实现细节 | 用户画像和评估数据 → `assessment` 模块文档 |
| 模块内部组件协作和数据流 | 跨模块事件发布策略 → `docs/architecture/features/事件发布与降级策略.md` |

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

## 权限模型

### 用户、角色、组织三层结构

`identity` 模块采用 **User - Role - Organization** 三层结构：

1. **User（用户层）**：用户账号、资料、状态和登录失败计数
   - Entity：`identity/entity.User`
   - 核心字段：`id`、`username`、`role`（单一角色字段）、`student_no`、`teacher_no`、`class_name`、`status`
   
2. **Role（角色层）**：系统角色定义
   - Entity：`identity/entity.Role`
   - 角色类型：`student`、`teacher`、`admin`（常量定义在 `entity/role.go`）
   - 当前没有独立角色 CRUD API，角色参考数据由 migration / seed 维护

3. **Organization（组织层）**：班级归属
   - 当前通过 `users.class_name` 字符串字段表达
   - 班级不是独立实体，只作为用户资料的一部分

### 角色类型与权限范围

| 角色常量 | 代码值 | 权限范围 | 典型路由前缀 |
| --- | --- | --- | --- |
| `RoleStudent` | `"student"` | 学生侧题目、训练、提交、排行榜、社区题解 | `/api/v1/challenges`、`/api/v1/practice`、`/api/v1/contests` |
| `RoleTeacher` | `"teacher"` | 教师侧班级管理、复盘分析、题解审核、AWD 赛事管理 | `/api/v1/teacher`、`/api/v1/academy` |
| `RoleAdmin` | `"admin"` | 管理员侧用户管理、题目出题、镜像构建、系统配置 | `/api/v1/admin`、`/api/v1/authoring` |

### 权限检查层级

权限检查分为两层：

1. **API Handler 层**：粗粒度角色检查
   - 通过 `middleware.Auth()` 中间件解析当前用户，将 `authctx.CurrentUser` 存入 `gin.Context`
   - Handler 通过 `middleware.MustCurrentUser(c)` 获取当前用户和角色
   - 例：管理员路由 `/api/v1/admin/*` 统一通过路由分组 + 中间件检查角色

2. **Application Service 层**：细粒度业务权限检查
   - 服务层通过 `ctx context.Context` 参数接收当前用户信息
   - 业务逻辑内检查资源归属、班级访问权限、跨模块权限
   - 例：`teaching_analysis` 模块中教师只能访问自己班级的学生复盘

### Context 传递机制

当前用户信息通过 **Gin Context** 传递，不使用 Go 标准库 `context.WithValue`：

- **存储**：`middleware.Auth()` 调用 `authctx.SetCurrentUser(c, authctx.CurrentUser{...})`，将当前用户存入 `gin.Context`
- **读取**：Handler 或 Service 通过 `middleware.MustCurrentUser(c)` 或 `authctx.MustCurrentUser(c)` 获取
- **传递路径**：`middleware.Auth()` → Gin Context → Handler → Service（如需细粒度检查）

**代码位置**：
- `code/backend/internal/authctx/authctx.go`：`CurrentUser` 结构体和 Context 存取
- `code/backend/internal/middleware/auth.go`：认证中间件和当前用户解析
- `code/backend/internal/module/identity/contracts/auth.go`：角色常量和用户 repository 契约
- `code/backend/internal/module/identity/entity/user.go`：User Entity 和角色字段
- `code/backend/internal/module/identity/entity/role.go`：Role Entity 和角色常量

**相关专题**：
- 认证与会话管理 → `docs/architecture/backend/modules/auth.md`
- 教师班级访问权限 → `docs/architecture/backend/modules/teaching_analysis.md`

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
- `code/backend/internal/middleware/auth.go`：认证中间件覆盖当前用户解析和角色提取。
- `code/backend/internal/authctx/authctx.go`：Context 存取当前用户信息的标准方式。

## 已知限制

- 当前没有独立权限策略服务；角色判断由 identity contract、middleware 和各 handler/application 的权限检查共同消费。
