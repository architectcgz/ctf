# auth 模块设计

> 状态：Current
> 事实源：`code/backend/internal/module/auth/`、`code/backend/internal/app/router.go`
> 替代：无

## 定位

`auth` 是认证、会话、登录入口和 OAuth Authorization Server owner，负责把账号凭证、CAS 票据、Session Cookie、WebSocket ticket、OAuth client / consent / code / token 等认证能力收口到一个模块。

## 本文档范围

| 本文档负责 | 本文档不负责（见其他文档） |
|-----------|-------------------------|
| auth 模块的职责边界、HTTP 入口和用例组织 | 用户资料和角色的真相源 → `identity` 模块文档 |
| 注册、登录、登出、改密、CAS 集成、MCP 浏览器 OAuth 授权的 owner | Session / OAuth token 存储的 Redis 键规范 → `docs/architecture/backend/design/redis-key-conventions.md` |
| 认证日志和审计要求 | 日志级别和审计策略 → `docs/architecture/backend/design/logging-and-audit.md` |
| 模块内部组件协作和数据流 | 跨模块事件发布策略 → `docs/architecture/features/事件发布与降级策略.md` |

## 事实来源

- HTTP 入口：`code/backend/internal/module/auth/api/http/handler.go`
- 用例编排：`code/backend/internal/module/auth/application/commands/`、`code/backend/internal/module/auth/application/queries/`
- 对外契约：`code/backend/internal/module/auth/contracts/`
- 外部适配：`code/backend/internal/module/auth/infrastructure/`
- 装配入口：`code/backend/internal/module/auth/runtime/module.go`、`code/backend/internal/app/composition/auth_module.go`

## 当前设计

- `auth/api/http`
  - 负责：注册、登录、登出、改密、资料读取、CAS status/login/callback、WebSocket ticket、OAuth metadata / client registration / authorize / token 的 HTTP 适配与请求/响应映射。
  - 不负责：直接访问用户仓储、CAS HTTP 客户端、Redis token store、OAuth 持久化表或审计持久化。
- `auth/application/commands`
  - 负责：注册、登录、登出、改密、CAS callback、Session 签发、OAuth client registration、authorization code、token exchange、refresh rotation、scope 校验和认证日志上下文编排。
  - 不负责：用户资料真相源、角色真相源、CAS validate HTTP 细节、Redis key 细节或 MCP JSON-RPC 协议适配。
- `auth/application/queries`
  - 负责：CAS 登录状态、OAuth protected resource metadata 和 authorization server metadata 等只读认证查询。
  - 不负责：用户管理查询、教师/管理员聚合查询或 MCP 工具列表生成。
- `auth/contracts`
  - 负责：`TokenService`、公开错误等跨模块稳定契约，供 middleware、通知 WebSocket、竞赛实时 WebSocket 等调用。
  - 不负责：暴露 auth 内部 DTO、Gin 类型、GORM 类型或 Redis concrete 类型。
- `auth/infrastructure`
  - 负责：CAS ticket validate request / XML principal 解析、Session / WebSocket ticket 的 Redis-backed 实现，以及 OAuth client / consent 的 PostgreSQL 存储和 OAuth code / access token / refresh token / consent nonce 的 Redis-backed 存储。
  - 不负责：决定用户是否存在、用户角色、审计语义或业务错误对外文案。

## API 入口设计

| 路由 | Handler | Service / 用例 |
| --- | --- | --- |
| `POST /api/v1/auth/register` | `auth/api/http.Handler.Register` | `commands.Service.Register`，注册用户并委托 `identity` 写用户事实 |
| `POST /api/v1/auth/login` | `Handler.Login` | `commands.Service.Login`，校验登录主体、密码和状态后签发 session |
| `GET /api/v1/auth/cas/status` | `Handler.CASStatus` | `queries.CASService`，读取 CAS 是否启用和登录入口信息 |
| `GET /api/v1/auth/cas/login` | `Handler.CASLogin` | `commands.CASService`，构造 CAS 登录跳转 |
| `GET /api/v1/auth/cas/callback` | `Handler.CASCallback` | `commands.CASService`，校验 ticket、同步用户、签发 session |
| `POST /api/v1/auth/logout` | `Handler.Logout` | `commands.Service.Logout`，吊销当前会话 |
| `GET /api/v1/auth/profile` | `Handler.Profile` | `identitycontracts.ProfileQueryService`，读取当前用户资料 |
| `PUT /api/v1/auth/password` | `Handler.ChangePassword` | `commands.Service.ChangePassword`，校验旧密码并更新密码 |
| `POST /api/v1/auth/ws-ticket` | `Handler.IssueWSTicket` | `authcontracts.TokenService`，签发短期 WebSocket ticket |
| `GET /.well-known/oauth-protected-resource` | `Handler.OAuthProtectedResourceMetadata` | `queries.OAuthMetadataService`，暴露 `/mcp` protected resource metadata |
| `GET /.well-known/oauth-authorization-server` | `Handler.OAuthAuthorizationServerMetadata` | `queries.OAuthMetadataService`，暴露 OAuth AS metadata |
| `POST /api/v1/oauth/register` | `Handler.RegisterOAuthClient` | `commands.OAuthService.RegisterClient`，注册 public OAuth client |
| `GET /api/v1/oauth/authorize` | `Handler.OAuthAuthorize` | `commands.OAuthService.PrepareAuthorization`，校验授权请求、登录跳转或渲染 consent HTML |
| `POST /api/v1/oauth/authorize` | `Handler.OAuthAuthorizeDecision` | `commands.OAuthService.ApproveAuthorization` / `DenyAuthorization`，处理用户同意并签发 authorization code |
| `POST /api/v1/oauth/token` | `Handler.OAuthToken` | `commands.OAuthService.ExchangeAuthorizationCode` / `RefreshAccessToken`，签发或刷新 OAuth token |

## Application / Service 设计

| Service | 代码路径 | 负责 | 不负责 |
| --- | --- | --- | --- |
| 本地认证 command service | `auth/application/commands/service.go` | 注册、登录、登出、改密、session 签发和认证错误映射 | 用户资料真相源、角色管理、审计持久化 |
| CAS command service | `auth/application/commands/cas_service.go` | CAS callback 编排、ticket 校验、用户同步和 session 签发 | CAS HTTP/XML 细节、用户仓储实现 |
| CAS query service | `auth/application/queries/cas_service.go` | CAS 状态查询和入口响应 | 登录状态写入 |
| OAuth command service | `auth/application/commands/oauth_service.go` | OAuth client registration、authorize 参数校验、consent、authorization code、token exchange、refresh rotation、scope 和 session version 校验 | MCP JSON-RPC 协议、用户资料真相源、Redis / SQL concrete 细节 |
| OAuth metadata query service | `auth/application/queries/oauth_metadata_service.go` | 根据配置或 request origin 生成 protected resource / authorization server metadata | MCP tool schema 或外部 client 存储 |
| Token service | `auth/infrastructure/token_service.go` | Redis-backed session 与 WebSocket ticket 校验和签发，提供 session version 给 OAuth token 主链路校验 | HTTP 路由权限判断、用户状态事实、OAuth client / consent 持久化 |
| OAuth store | `auth/infrastructure/oauth_store.go` | PostgreSQL backed OAuth client / consent；Redis backed authorization code、access token、refresh token 与 consent nonce | OAuth 业务规则和 handler 响应映射 |
| CAS ticket validator | `auth/infrastructure/cas_ticket_validator.go` | validate URL 请求、XML principal 解析、invalid ticket sentinel | 用户同步和 session 签发 |

## 数据设计

`auth` 不拥有用户资料或角色事实；这些事实由 `identity` 的 `users` / `roles` / `user_roles` 拥有。`auth` 拥有 OAuth client 与 consent 表，用于 MCP 浏览器授权。

| 数据 / 副作用 | Owner | 说明 |
| --- | --- | --- |
| Redis session / token / WebSocket ticket | `auth/infrastructure/token_service.go` | 保存登录会话、刷新和短期 WS ticket；key 和 TTL 不泄漏到 application |
| OAuth client / consent | `auth/infrastructure/oauth_store.go`、`oauth_clients`、`oauth_consents` | 保存 public client 注册信息和用户对 `mcp:challenge:read` 的 consent |
| OAuth code / access token / refresh token / consent nonce | `auth/infrastructure/oauth_store.go` | Redis TTL 存储；token 明文不入库，Redis key 使用 token hash |
| CAS validate 响应 | `auth/infrastructure/cas_ticket_validator.go` | 外部响应只转换成端口结果，不进入数据库 |
| 审计日志 | `ops` | auth 只通过 `auditlog.Recorder` 记录 login/logout、CAS、oauth_client/register、oauth_consent/grant、oauth_token/exchange、oauth_token/refresh |
| Cookie / Header | `auth/api/http` | 协议层设置，实际有效性由 token service 校验 |

## 边界

- `auth` 负责认证过程，不拥有用户资料、角色或用户状态；这些事实由 `identity` 拥有。
- `auth` 可通过 `identitycontracts.UserRepository`、`ProfileCommandService`、`ProfileQueryService` 完成认证需要的用户查询与同步。
- `auth` 可通过 app composition 注入 `auditlog.Recorder` 记录登录、登出、CAS 等审计事件，但不拥有审计日志查询。
- `auth` 的 `TokenService` 是认证中间件和 WebSocket 认证的共享入口，不回收到 `identity`。

## 主要用例

- 匿名用户注册：`POST /api/v1/auth/register`
- 用户登录：`POST /api/v1/auth/login`
- CAS 登录状态、跳转和回调：`GET /api/v1/auth/cas/status`、`GET /api/v1/auth/cas/login`、`GET /api/v1/auth/cas/callback`
- 已认证用户登出、资料读取、改密、签发 WebSocket ticket：`/api/v1/auth/*`
- 外部 MCP client 浏览器授权：`/.well-known/oauth-*`、`/api/v1/oauth/register`、`/api/v1/oauth/authorize`、`/api/v1/oauth/token`
- 中间件会话校验：`code/backend/internal/middleware` 通过 `auth/contracts.TokenService` 与 `identityModule.Users` 完成。

## 数据与副作用

- PostgreSQL：OAuth client / consent 由 `auth/infrastructure/oauth_store.go` 读写。
- Redis：Session / WebSocket ticket 由 `auth/infrastructure/token_service.go` 读写；OAuth code / access token / refresh token / consent nonce 由 `auth/infrastructure/oauth_store.go` 读写。
- CAS：`auth/infrastructure/cas_ticket_validator.go` 访问 CAS validate URL，并把外部响应转换成模块端口结果。
- 审计：auth command service 和 OAuth handler 通过注入的 `auditlog.Recorder` 写入审计能力，实际 owner 在 `ops`；不记录 token、code 或 verifier 明文。
- Cookie / Header：HTTP handler 只做协议层设置，认证有效性由 token service 与 middleware 判断。

## 跨模块依赖

| 依赖 | 用途 | 接线 |
| --- | --- | --- |
| `identity` | 用户存在性、资料同步、角色/状态读取 | `code/backend/internal/app/composition/auth_module.go` |
| `ops` | 审计记录能力 | app composition 注入 `auditlog.Recorder` |
| `middleware` | 认证中间件消费 token service | `code/backend/internal/app/router.go` |

## Guardrail

- `code/backend/internal/module/auth/architecture_test.go`：禁止 api / commands / queries 依赖 infrastructure 或 Gin 泄漏到 application，约束 runtime typed deps。
- `code/backend/internal/app/router_session_routes_test.go`：覆盖 session 相关路由。
- `code/backend/internal/module/auth/ports/cas_ticket_validator_context_contract_test.go`：约束 CAS validator context 传播。

## 已知限制

- 当前认证仍服务于单体 API 进程；没有把 auth 拆成独立进程。
- CAS 外部服务可用性只影响 CAS 登录链路，不改变本地账号密码登录的 owner。
