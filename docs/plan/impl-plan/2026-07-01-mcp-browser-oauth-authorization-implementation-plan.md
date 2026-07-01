# MCP 浏览器 OAuth 授权实施计划

> **给 agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development 或 superpowers:executing-plans 按任务逐项执行本计划。所有步骤使用 checkbox（`- [ ]`）追踪，完成一个可验证步骤后立即勾选。

**Goal:** 将外部 agent 调用 CTF `/mcp` 的认证方式从用户手工签发 MCP token 改为标准浏览器 OAuth 2.1 Authorization Code + PKCE 授权，并删除旧 MCP token 方案。

**目标：** 将外部 agent 调用 CTF `/mcp` 的认证方式从用户手工签发 MCP token 改为标准浏览器 OAuth 2.1 Authorization Code + PKCE 授权，并删除旧 MCP token 方案。

**Architecture:** `auth` 模块作为 OAuth Authorization Server owner，负责客户端注册、授权码、access token、refresh token、scope、同意记录和审计；`interfaces/mcp` 只作为 MCP / JSON-RPC 协议适配层，验证 OAuth Bearer access token 后调用既有 `instance` / `challenge` query service。OAuth 授权端点放在 `/api/v1/oauth/authorize`，以复用当前 `ctf_session` cookie 的 `/api/v1` path；标准 discovery metadata 仍暴露在 `/.well-known/*`。

**架构：** `auth` 模块作为 OAuth Authorization Server owner，负责客户端注册、授权码、access token、refresh token、scope、同意记录和审计；`interfaces/mcp` 只作为 MCP / JSON-RPC 协议适配层，验证 OAuth Bearer access token 后调用既有 `instance` / `challenge` query service。OAuth 授权端点放在 `/api/v1/oauth/authorize`，以复用当前 `ctf_session` cookie 的 `/api/v1` path；标准 discovery metadata 仍暴露在 `/.well-known/*`。

**Tech Stack:** Go、Gin、Redis、PostgreSQL migration、OAuth 2.1、PKCE S256、MCP Streamable HTTP、JSON-RPC 2.0、OpenAPI、Vue 登录跳转微调、`go test`、`pnpm test:run`

**技术栈：** Go、Gin、Redis、PostgreSQL migration、OAuth 2.1、PKCE S256、MCP Streamable HTTP、JSON-RPC 2.0、OpenAPI、Vue 登录跳转微调、`go test`、`pnpm test:run`

---

## Task Metadata

- Task Slug: `2026-07-01-mcp-browser-oauth-authorization`
- Started At: `2026-07-01T06:41:00Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-07-01-mcp-browser-oauth-authorization`
- Branch: `task/2026-07-01-mcp-browser-oauth-authorization`
- Plan Type: `formal implementation`
- 当前分支: `multi-instance`
- 计划位置: `docs/plan/impl-plan/2026-07-01-mcp-browser-oauth-authorization-implementation-plan.md`
- 关联旧计划: `docs/plan/impl-plan/2026-07-01-agent-current-challenge-mcp-implementation-plan.md`

## Task Classification

- Classification: `非琐碎任务`
- Why: 本任务新增 OAuth Authorization Server、PostgreSQL migration、Redis token owner、HTTP / MCP contract、前端登录跳转和 OpenAPI / 架构文档，触达认证安全路径和多个受保护实现面。
- Required workflow: `code-workflow` + `executing-plans` + `test-driven-development`，每个可验证步骤完成后同步更新本计划 checkbox。

## 计划状态

- Status: `active`
- 进入编码前必须完成：
  - [x] 绑定 task slug 和 startup gate。
  - [x] 完成本计划 review / architecture-fit check。
  - [x] 确认不保留旧 `POST /api/v1/auth/mcp-token` 兼容路径。

### 编码前 Review 记录

- 2026-07-01：已在 task worktree `/home/azhi/workspace/projects/.worktrees/ctf/2026-07-01-mcp-browser-oauth-authorization` 绑定 `task/2026-07-01-mcp-browser-oauth-authorization` 和 `.harness/session-gates/2026-07-01-mcp-browser-oauth-authorization.json`。
- 2026-07-01：计划边界与现有 `auth` / `interfaces/mcp` 分层一致；`auth` 承担 OAuth Authorization Server，`interfaces/mcp` 仅消费 Bearer access token 和 scope。
- 2026-07-01：旧 MCP token 路径不做兼容保留；`POST /api/v1/auth/mcp-token`、`auth.mcp_token_ttl`、`MCPToken`、`IssueMCPToken`、`ResolveMCPToken` 和 `token_url` 都属于删除范围。
- 2026-07-01：`TokenService` 方法“不可存在”不能用普通 Go 编译测试直接表达，执行时用路由/handler/config/source guard 组合锁定旧 surface 删除。

## Intake Analysis Gate

- Relevant superpowers analysis pass: `brainstorming`、`grill-with-docs`、`systematic-debugging`、`test-driven-development`。
- Why this pass fits: 本任务是认证能力替换和外部 MCP 授权流程改造，需要先确认 owner 边界、旧兼容删除范围、安全路径和验证粒度，再进入测试先行实现。
- grill-with-docs findings: 现有 `docs/architecture/backend/modules/auth.md` 把 auth 定位为认证、会话和登录 owner；`docs/architecture/backend/04-api-design.md` 把 `/mcp` 定位为协议适配入口。计划将 OAuth AS 收口到 auth、MCP 只消费 OAuth token，与当前事实边界一致。
- Plan adjustments after challenge: 保持不兼容删除旧 `POST /api/v1/auth/mcp-token`；任务 1 中“接口不存在”的断言采用 HTTP / route / source guard 组合表达，避免用无法编译的负接口断言。

## 目标与非目标

- 目标：外部 agent 首次访问 `/mcp` 时可根据 OAuth metadata 发起浏览器授权；用户在浏览器登录 CTF 平台并同意 `mcp:challenge:read` scope 后，agent 使用 OAuth access token 调用 `get_current_challenge`。
- 目标：删除当前 MCP token 签发、解析、配置、契约和文档，不提供 fallback，不保留 `CTF_MCP_TOKEN` 操作路径。
- 目标：支持 public OAuth client + PKCE；动态客户端注册仅允许受限 public client，不引入 client secret。
- 目标：保留当前 `/mcp` 用户级限流和成功工具调用审计，但审计资源语义从 `mcp_token` 转为 `oauth_client` / `oauth_consent` / `mcp_tool`。
- 非目标：不实现 OIDC 登录，不签发 `id_token`，不开放 `password` grant / `client_credentials` grant，不允许 agent 获得提交 flag、下载附件、解锁 hint 等写能力。
- 非目标：不把 OAuth access token 存入前端 localStorage；agent token 存储由外部 MCP client 自己负责。

## 问题陈述

- 当前行为：外部 agent 必须拿到用户手工签发的 `Authorization: Bearer <mcp_token>`，签发入口是 `POST /api/v1/auth/mcp-token`，配置项是 `auth.mcp_token_ttl`，未认证 MCP 错误会返回 `token_url`。
- 目标行为：外部 agent 只需要知道 `/mcp` URL；未授权时通过 MCP/OAuth metadata 得到授权服务器信息，执行浏览器授权流程。用户在浏览器完成 CTF 登录和授权同意后，agent 获得 OAuth access token，后续调用 `/mcp`。
- 当前 token 方案的问题：它依赖用户复制或环境变量注入 token，不是标准 MCP 客户端预期的浏览器授权；token 签发接口本身也会成为被脚本化调用的爬取入口。
- 这次任务的结构性变化：认证 owner 从“auth 模块临时 MCP token 方法”升级为“auth 模块内 OAuth Authorization Server 子能力”，MCP 层只消费标准 OAuth token 和 scope。

## 当前实现必须删除的旧 surface

- 后端路由：`POST /api/v1/auth/mcp-token`
- 后端契约：`authcontracts.MCPToken`、`TokenService.IssueMCPToken`、`TokenService.ResolveMCPToken`
- Redis payload：`mcpTokenPayload`、`mcpTokenKey`
- 配置：`auth.mcp_token_ttl`
- MCP handler 字段：`TokenURL`、`tokenResolver`、`auth_method=bearer_token`
- 审计：`mcp_token/create`
- 文档和契约：`docs/contracts/api-contract-v1.md`、`docs/contracts/openapi-v1/paths/auth.yaml`、`docs/contracts/openapi-v1.yaml`、`docs/architecture/backend/04-api-design.md`、`docs/architecture/backend/modules/auth.md` 中所有 MCP token 说明
- 本地 agent 配置说明：不再使用 `bearer_token_env_var = "CTF_MCP_TOKEN"`

## 授权协议目标形态

### 标准端点

- `GET /.well-known/oauth-protected-resource`
  - 返回 MCP protected resource metadata。
  - `resource` 指向当前 origin 的 `/mcp`。
  - `authorization_servers` 指向当前 origin 的 OAuth issuer。
- `GET /.well-known/oauth-authorization-server`
  - 返回 Authorization Server metadata。
  - `issuer` 使用 `auth.oauth.issuer_url`，生产必须是 HTTPS origin。
  - `authorization_endpoint` 为 `${issuer}/api/v1/oauth/authorize`。
  - `token_endpoint` 为 `${issuer}/api/v1/oauth/token`。
  - `registration_endpoint` 为 `${issuer}/api/v1/oauth/register`。
  - `response_types_supported=["code"]`。
  - `grant_types_supported=["authorization_code","refresh_token"]`。
  - `code_challenge_methods_supported=["S256"]`。
  - `token_endpoint_auth_methods_supported=["none"]`。
  - `scopes_supported=["mcp:challenge:read"]`。
- `POST /api/v1/oauth/register`
  - Dynamic Client Registration，public clients only。
  - 只接受 `token_endpoint_auth_method=none`。
  - 只接受 `authorization_code` 和可选 `refresh_token`。
  - `redirect_uris` 默认只允许 loopback：`http://127.0.0.1:<port>/*`、`http://localhost:<port>/*`、`http://[::1]:<port>/*`；生产若要支持固定 HTTPS agent callback，必须通过 `auth.oauth.allowed_redirect_uri_prefixes` 显式配置。
- `GET /api/v1/oauth/authorize`
  - 标准 authorization endpoint。
  - 未登录：302 到前端 `/login?redirect=/api/v1/oauth/authorize?...`。
  - 已登录未同意：渲染 server-side consent HTML。
  - 已登录且已有有效同意：直接签发 authorization code 并 302 回 `redirect_uri?code=...&state=...`。
- `POST /api/v1/oauth/authorize`
  - consent decision endpoint。
  - `approve=true` 后签发 authorization code。
  - `approve=false` 后 302 回 `redirect_uri?error=access_denied&state=...`。
- `POST /api/v1/oauth/token`
  - `grant_type=authorization_code`：校验 code、client、redirect_uri、PKCE verifier、scope 后签发 access token 和 refresh token。
  - `grant_type=refresh_token`：校验 refresh token、client、session version、scope 后旋转 refresh token 并签发新 access token。
- `POST /mcp`
  - 只接受 OAuth Bearer access token。
  - 缺失、过期、scope 不足或 client 不匹配时返回 HTTP 401，并设置 `WWW-Authenticate` 指向 protected resource metadata。

### Scope 与 token 规则

- 第一版只定义 `mcp:challenge:read`。
- `get_current_challenge` 必须要求 `mcp:challenge:read`。
- access token 默认 TTL：15 分钟。
- refresh token 默认 TTL：30 天，旋转使用；旧 refresh token 交换成功后立即失效。
- authorization code 默认 TTL：5 分钟，单次使用。
- 所有 OAuth token 必须绑定：
  - user id
  - username
  - role
  - client id
  - scope
  - session version
  - issued_at / expires_at
- 用户改密、禁用或撤销会话后，refresh token 和 access token 在解析主链路失效；短期 access token 可依赖解析时校验 session version，不只依赖 TTL。

### MCP 未授权响应

`/mcp` 在缺少 OAuth access token 时应返回：

```http
HTTP/1.1 401 Unauthorized
WWW-Authenticate: Bearer realm="ctf-mcp", resource_metadata="https://ctf.example.edu/.well-known/oauth-protected-resource"
Content-Type: application/json
```

响应 body 可以继续使用 JSON-RPC error，便于不完整 MCP client 展示错误，但不再包含 `token_url`：

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32001,
    "message": "请通过浏览器授权 CTF MCP",
    "data": {
      "auth_method": "oauth",
      "resource_metadata": "/.well-known/oauth-protected-resource",
      "required_scope": "mcp:challenge:read"
    }
  }
}
```

## Files

### 新增

- `code/backend/migrations/000002_oauth_browser_authorization.up.sql`
- `code/backend/migrations/000002_oauth_browser_authorization.down.sql`
- `code/backend/internal/module/auth/contracts/oauth.go`
- `code/backend/internal/module/auth/application/commands/oauth_service.go`
- `code/backend/internal/module/auth/application/commands/oauth_service_test.go`
- `code/backend/internal/module/auth/application/queries/oauth_metadata_service.go`
- `code/backend/internal/module/auth/infrastructure/oauth_store.go`
- `code/backend/internal/module/auth/infrastructure/oauth_store_test.go`
- `code/backend/internal/module/auth/api/http/oauth_handler.go`
- `code/backend/internal/module/auth/api/http/oauth_types.go`
- `code/backend/internal/module/auth/api/http/oauth_handler_test.go`
- `code/backend/internal/interfaces/mcp/oauth_auth_test.go`
- `docs/operations/mcp-oauth-login.md`
- `docs/contracts/openapi-v1/paths/oauth.yaml`
- `docs/contracts/openapi-v1/components/schemas/oauth.yaml`

### 修改

- `code/backend/internal/interfaces/mcp/handler.go`
- `code/backend/internal/interfaces/mcp/handler_test.go`
- `code/backend/internal/module/auth/contracts/token_service.go`
- `code/backend/internal/module/auth/contracts/public_errors.go`
- `code/backend/internal/module/auth/infrastructure/token_service.go`
- `code/backend/internal/module/auth/infrastructure/token_service_test.go`
- `code/backend/internal/module/auth/api/http/handler.go`
- `code/backend/internal/module/auth/api/http/auth_types.go`
- `code/backend/internal/module/auth/api/http/http_integration_test.go`
- `code/backend/internal/module/auth/runtime/module.go`
- `code/backend/internal/app/composition/auth_module.go`
- `code/backend/internal/app/router.go`
- `code/backend/internal/app/router_route_wiring_test.go`
- `code/backend/internal/config/types.go`
- `code/backend/internal/config/defaults.go`
- `code/backend/internal/config/validate.go`
- `code/backend/internal/config/config_test.go`
- `code/backend/configs/config.yaml`
- `code/backend/configs/config.dev.yaml`
- `code/backend/configs/config.prod.yaml`
- `code/frontend/src/features/auth/model/useLoginPage.ts`
- `code/frontend/src/features/auth/model/useLoginPage.test.ts`
- `docs/contracts/api-contract-v1.md`
- `docs/contracts/openapi-v1/index.yaml`
- `docs/contracts/openapi-v1.yaml`
- `docs/contracts/openapi-v1/paths/auth.yaml`
- `docs/contracts/openapi-v1/components/schemas/auth.yaml`
- `docs/architecture/backend/04-api-design.md`
- `docs/architecture/backend/modules/auth.md`
- `docs/design/ctf-tutor-agent-and-mcp.md`

### 评审

- `code/backend/internal/middleware/auth.go`
- `code/backend/internal/authctx/authctx.go`
- `code/backend/internal/module/identity/contracts/auth.go`
- `code/backend/tests/README.md`
- `docs/plan/impl-plan/2026-07-01-agent-current-challenge-mcp-implementation-plan.md`

## 复用与 Owner 决策

- `auth` owner：OAuth client、authorization code、token exchange、refresh rotation、session version 校验、同意记录和审计。
- `interfaces/mcp` owner：MCP JSON-RPC / tools 协议、`WWW-Authenticate` 响应、scope 到工具的映射。
- `identity` owner：仍只负责用户状态和角色事实，不新增 OAuth 知识。
- `ops` owner：复用 `auditlog.Recorder` 记录授权同意、token exchange 失败和 MCP 工具读取，不新增独立审计系统。
- `frontend` owner：只修正登录后的 redirect 处理，保证 `/api/v1/oauth/authorize?...` 这类后端授权端点能通过 `window.location.assign` 完成跳转；不承载 OAuth consent 页面。
- server-side consent HTML：第一版由后端渲染，原因是授权端点本身必须能被外部 agent 浏览器直接打开和重定向，且当前 `ctf_session` cookie path 是 `/api/v1`。这避免同时迁移 cookie path 和前端路由 owner。

## 数据设计

### PostgreSQL 表

`000002_oauth_browser_authorization.up.sql` 创建：

```sql
CREATE TABLE oauth_clients (
  id BIGSERIAL PRIMARY KEY,
  client_id TEXT NOT NULL UNIQUE,
  client_name TEXT NOT NULL,
  client_uri TEXT,
  redirect_uris JSONB NOT NULL,
  grant_types JSONB NOT NULL,
  response_types JSONB NOT NULL,
  scope TEXT NOT NULL,
  token_endpoint_auth_method TEXT NOT NULL DEFAULT 'none',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE oauth_consents (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  client_id TEXT NOT NULL REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
  scope TEXT NOT NULL,
  granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ,
  revoked_at TIMESTAMPTZ,
  UNIQUE (user_id, client_id, scope)
);

CREATE INDEX idx_oauth_consents_user_client ON oauth_consents(user_id, client_id, revoked_at);
```

不把 authorization code、access token、refresh token 做成长期表；它们是短生命周期凭证，使用 Redis TTL 保存 token hash 与 payload，减少数据库清理任务和敏感 token 残留。

### Redis key

- `ctf:auth:oauth:code:<sha256(code)>`，TTL 5m，`GETDEL` 消费。
- `ctf:auth:oauth:access:<sha256(token)>`，TTL 15m。
- `ctf:auth:oauth:refresh:<sha256(token)>`，TTL 30d，refresh 成功后删除旧 key。
- payload 不保存 token 明文，只保存 user/client/scope/session_version/issued_at/expires_at。

### 配置

新增 `AuthOAuthConfig`：

```go
type AuthOAuthConfig struct {
    IssuerURL                  string        `mapstructure:"issuer_url"`
    AuthorizationCodeTTL       time.Duration `mapstructure:"authorization_code_ttl"`
    AccessTokenTTL             time.Duration `mapstructure:"access_token_ttl"`
    RefreshTokenTTL            time.Duration `mapstructure:"refresh_token_ttl"`
    ClientRegistrationEnabled  bool          `mapstructure:"client_registration_enabled"`
    AllowedRedirectURIPrefixes []string      `mapstructure:"allowed_redirect_uri_prefixes"`
    RedisKeyPrefix             string        `mapstructure:"redis_key_prefix"`
}
```

删除：

```go
MCPTokenTTL time.Duration `mapstructure:"mcp_token_ttl"`
```

默认值：

- `auth.oauth.issuer_url=""`：dev 环境允许从 request origin 推导；prod 必须显式配置 HTTPS。
- `auth.oauth.authorization_code_ttl=5m`
- `auth.oauth.access_token_ttl=15m`
- `auth.oauth.refresh_token_ttl=720h`
- `auth.oauth.client_registration_enabled=true`
- `auth.oauth.redis_key_prefix=ctf:auth:oauth`

## 执行切片

### 任务 1：删除 MCP token surface 的红测

**目标：** 先用测试锁定“旧 token 方案不存在”，避免后续实现时把它当兼容路径保留下来。

**文件：**

- 修改：`code/backend/internal/module/auth/api/http/http_integration_test.go`
- 修改：`code/backend/internal/module/auth/infrastructure/token_service_test.go`
- 修改：`code/backend/internal/interfaces/mcp/handler_test.go`
- 修改：`code/backend/internal/config/config_test.go`
- 修改：`code/backend/internal/app/router_route_wiring_test.go`

- [x] 步骤 1：新增 HTTP 红测：`POST /api/v1/auth/mcp-token` 返回 404 或路由不存在。

```go
func TestHTTP_MCPTokenEndpointRemoved(t *testing.T) {
    env := newHTTPTestEnv(t)
    resp := performJSONRequest(t, env.router, http.MethodPost, "/api/v1/auth/mcp-token", nil, nil, nil)
    if resp.Code != http.StatusNotFound {
        t.Fatalf("expected removed MCP token endpoint to return 404, got %d body=%s", resp.Code, resp.Body.String())
    }
}
```

- [x] 步骤 2：新增 token service 编译/行为红测：`TokenService` 不再提供 `IssueMCPToken` 和 `ResolveMCPToken`，删除现有 MCP token 测试后用 OAuth token 测试替代。
- [x] 步骤 3：新增 MCP handler 红测：无 token 时 `WWW-Authenticate` 包含 `resource_metadata`，错误 data 不包含 `token_url`。
- [x] 步骤 4：新增 config 红测：配置中出现 `auth.mcp_token_ttl` 不再参与 `Validate()`；默认配置不再设置该 key。
- [x] 步骤 5：运行红测。

运行：

```bash
cd code/backend && go test ./internal/module/auth/api/http ./internal/module/auth/infrastructure ./internal/interfaces/mcp ./internal/config ./internal/app -run 'MCPToken|OAuth|Route' -count=1
```

预期：至少新增测试失败，失败原因指向旧 endpoint / 旧 handler / 旧 config 尚未删除。

- [x] 步骤 6：按仓库提交规则提交本切片测试，提交前读取 `committing-changes` skill。

### 任务 2：添加 OAuth 数据层和配置

**目标：** 建立 OAuth client / consent 的持久 owner，以及 authorization code / access token / refresh token 的 Redis TTL owner。

**文件：**

- 新增：`code/backend/migrations/000002_oauth_browser_authorization.up.sql`
- 新增：`code/backend/migrations/000002_oauth_browser_authorization.down.sql`
- 新增：`code/backend/internal/module/auth/contracts/oauth.go`
- 新增：`code/backend/internal/module/auth/infrastructure/oauth_store.go`
- 新增：`code/backend/internal/module/auth/infrastructure/oauth_store_test.go`
- 修改：`code/backend/internal/config/types.go`
- 修改：`code/backend/internal/config/defaults.go`
- 修改：`code/backend/internal/config/validate.go`
- 修改：`code/backend/internal/config/config_test.go`
- 修改：`code/backend/configs/config.yaml`
- 修改：`code/backend/configs/config.dev.yaml`
- 修改：`code/backend/configs/config.prod.yaml`

- [x] 步骤 1：写 migration 结构测试，要求 `000002_oauth_browser_authorization.up.sql` 创建 `oauth_clients` 和 `oauth_consents`，down 文件反向删除。
- [x] 步骤 2：写 config 红测，覆盖默认 TTL、prod issuer 必须 HTTPS、redirect prefix 不能含空字符串、registration disabled 时不允许 DCR。
- [x] 步骤 3：写 `oauth_store_test.go` 红测，覆盖 client 注册、redirect URI 精确匹配、consent upsert/revoke、authorization code `GETDEL` 单次消费、access token 解析、refresh token 旋转。
- [x] 步骤 4：实现 `AuthOAuthConfig`、默认值和校验。
- [x] 步骤 5：实现 `OAuthStore`，Redis 中只存 hash key，不存 token 明文。
- [x] 步骤 6：确保所有 token payload 使用 `time.Now().UTC()`，输出 RFC3339 UTC。
- [x] 步骤 7：运行数据层验证。

运行：

```bash
cd code/backend && go test ./internal/config ./internal/module/auth/infrastructure -run 'OAuth|Migration' -count=1
```

预期：PASS。

- [x] 步骤 8：运行 migration 文件验证。

运行：

```bash
cd code/backend && go test ./internal/app -run 'Migration|TestMigrationFiles' -count=1
```

预期：PASS 或只出现既有 fixture 问题；若失败涉及新增 migration，先修复。

- [x] 步骤 9：提交本切片。

### 任务 3：实现 OAuth service、metadata 和动态客户端注册

**目标：** 让 MCP client 能通过标准 metadata 发现授权服务器，并注册 public client。

**文件：**

- 新增：`code/backend/internal/module/auth/application/queries/oauth_metadata_service.go`
- 新增：`code/backend/internal/module/auth/application/commands/oauth_service.go`
- 新增：`code/backend/internal/module/auth/application/commands/oauth_service_test.go`
- 新增：`code/backend/internal/module/auth/api/http/oauth_handler.go`
- 新增：`code/backend/internal/module/auth/api/http/oauth_types.go`
- 新增：`code/backend/internal/module/auth/api/http/oauth_handler_test.go`
- 修改：`code/backend/internal/module/auth/runtime/module.go`
- 修改：`code/backend/internal/app/composition/auth_module.go`
- 修改：`code/backend/internal/app/router.go`

- [x] 步骤 1：写 metadata HTTP 红测。

关键断言：

```go
if got := body["code_challenge_methods_supported"]; !contains(got, "S256") { t.Fatal(...) }
if got := body["token_endpoint_auth_methods_supported"]; !contains(got, "none") { t.Fatal(...) }
```

- [x] 步骤 2：写 DCR 红测：合法 loopback redirect URI 返回 `client_id`；非 loopback 且未配置 allow prefix 的 redirect URI 返回 400。
- [x] 步骤 3：写 service 红测：client id 必须不可预测、注册 scope 只能是 `mcp:challenge:read` 子集、`client_secret` 不返回。
- [x] 步骤 4：实现 metadata query service，issuer 生成规则为：
  - prod：必须使用 `auth.oauth.issuer_url`
  - dev：配置为空时从 `X-Forwarded-Proto` / `Host` 或 request origin 推导
- [x] 步骤 5：实现 dynamic client registration command。
- [x] 步骤 6：接线路由：
  - `engine.GET("/.well-known/oauth-protected-resource", ...)`
  - `engine.GET("/.well-known/oauth-authorization-server", ...)`
  - `apiV1.POST("/oauth/register", ...)`
- [x] 步骤 7：对 `/api/v1/oauth/register` 添加 IP 级 rate limit，例如 `rate_limit.auth_anonymous` 或新增 `rate_limit.oauth_client_registration`。
- [x] 步骤 8：运行验证。

运行：

```bash
cd code/backend && go test ./internal/module/auth/application/commands ./internal/module/auth/api/http ./internal/app -run 'OAuth|WellKnown|ClientRegistration|Route' -count=1
```

预期：PASS。

- [x] 步骤 9：提交本切片。

### 任务 4：实现 authorization endpoint、登录跳转和 consent

**目标：** 外部 agent 打开浏览器后，未登录用户会被引导到 CTF 登录页；登录完成后回到授权端点并完成同意。

**文件：**

- 修改：`code/backend/internal/module/auth/application/commands/oauth_service.go`
- 修改：`code/backend/internal/module/auth/application/commands/oauth_service_test.go`
- 修改：`code/backend/internal/module/auth/api/http/oauth_handler.go`
- 修改：`code/backend/internal/module/auth/api/http/oauth_types.go`
- 修改：`code/backend/internal/module/auth/api/http/oauth_handler_test.go`
- 修改：`code/backend/internal/app/router.go`
- 修改：`code/frontend/src/features/auth/model/useLoginPage.ts`
- 修改：`code/frontend/src/features/auth/model/useLoginPage.test.ts`

- [x] 步骤 1：写 authorize 参数校验红测：
  - `response_type` 只能是 `code`
  - 必须有 `client_id`
  - `redirect_uri` 必须和注册记录精确匹配
  - 必须有 `code_challenge`
  - `code_challenge_method` 必须是 `S256`
  - `scope` 必须包含且只包含允许 scope
  - `state` 原样透传，不强制服务端理解
- [x] 步骤 2：写未登录重定向红测：`GET /api/v1/oauth/authorize?...` 无 session 时 302 到 `/login?redirect=<escaped authorize path>`。
- [x] 步骤 3：写前端登录跳转红测：当 redirect target 是 `/api/v1/oauth/authorize?...` 时，登录成功后使用 `window.location.assign()`，而不是 Vue Router `push()`。
- [x] 步骤 4：写 consent 红测：已登录但未同意时返回 HTML 表单，表单 action 为 `/api/v1/oauth/authorize`，隐藏字段保留原始 authorize 参数和 CSRF nonce。
- [x] 步骤 5：写 approve 红测：`POST /api/v1/oauth/authorize` 同意后创建 `oauth_consents`，签发 authorization code，302 到 `redirect_uri?code=...&state=...`。
- [x] 步骤 6：写 deny 红测：拒绝后 302 到 `redirect_uri?error=access_denied&state=...`，不创建 code。
- [x] 步骤 7：实现 authorize request parser 和 service 校验。
- [x] 步骤 8：实现 server-side consent HTML，页面只显示客户端名、scope 和当前用户；禁止输出 access token / refresh token。
- [x] 步骤 9：实现 consent CSRF nonce，nonce 存 Redis 或 signed form value，TTL 不超过 authorization code TTL。
- [x] 步骤 10：实现前端 `useLoginPage` 的 hard navigation 分支。
- [x] 步骤 11：运行后端和前端验证。

运行：

```bash
cd code/backend && go test ./internal/module/auth/application/commands ./internal/module/auth/api/http -run 'Authorize|Consent|OAuth' -count=1
cd code/frontend && pnpm test:run src/features/auth/model/useLoginPage.test.ts
```

预期：PASS。

- [x] 步骤 12：提交本切片。

### 任务 5：实现 token endpoint 和 refresh token 旋转

**目标：** OAuth client 能用 authorization code + PKCE 换取 access token，并用 refresh token 续期。

**文件：**

- 修改：`code/backend/internal/module/auth/contracts/oauth.go`
- 修改：`code/backend/internal/module/auth/application/commands/oauth_service.go`
- 修改：`code/backend/internal/module/auth/application/commands/oauth_service_test.go`
- 修改：`code/backend/internal/module/auth/api/http/oauth_handler.go`
- 修改：`code/backend/internal/module/auth/api/http/oauth_types.go`
- 修改：`code/backend/internal/module/auth/api/http/oauth_handler_test.go`
- 修改：`code/backend/internal/module/auth/runtime/module.go`

- [x] 步骤 1：写 token exchange 红测：合法 code + `code_verifier` 返回 `access_token`、`refresh_token`、`token_type=Bearer`、`expires_in`、`scope`。
- [x] 步骤 2：写 PKCE 失败红测：错误 verifier、plain method、缺 verifier、重复 code 消费均返回 OAuth error JSON。
- [x] 步骤 3：写 redirect URI 绑定红测：token request 的 `redirect_uri` 必须和 authorize 时完全一致。
- [x] 步骤 4：写 refresh 红测：旧 refresh token 只能使用一次，新 refresh token 可用，旧 token 复用返回 `invalid_grant` 并记录审计。
- [x] 步骤 5：写 session version 红测：用户 session version 变化后，refresh token 和 access token 解析失败。
- [x] 步骤 6：实现 `ExchangeAuthorizationCode`。
- [x] 步骤 7：实现 `RefreshAccessToken` 和 refresh rotation。
- [x] 步骤 8：实现 OAuth error response：

```json
{
  "error": "invalid_grant",
  "error_description": "authorization code is invalid or expired"
}
```

- [x] 步骤 9：记录审计：
  - `oauth_client/register`
  - `oauth_consent/grant`
  - `oauth_token/exchange`
  - `oauth_token/refresh`
  - token 失败只记录 client id、user id、错误类型，不记录 token/code 明文。
- [x] 步骤 10：运行验证。

运行：

```bash
cd code/backend && go test ./internal/module/auth/application/commands ./internal/module/auth/api/http ./internal/module/auth/infrastructure -run 'OAuth|Refresh|PKCE' -count=1
cd code/backend && go test ./internal/app -run '^TestNewRouterRegistersStudentChallengeRoutes$' -count=1
```

预期：PASS。

- [x] 步骤 11：提交本切片。

### 任务 6：将 `/mcp` 切到 OAuth access token + scope

**目标：** `/mcp` 不再解析 MCP token，只解析 OAuth access token，并按工具要求检查 scope。

**文件：**

- 修改：`code/backend/internal/interfaces/mcp/handler.go`
- 修改：`code/backend/internal/interfaces/mcp/handler_test.go`
- 新增：`code/backend/internal/interfaces/mcp/oauth_auth_test.go`
- 修改：`code/backend/internal/app/router.go`
- 修改：`code/backend/internal/app/router_route_wiring_test.go`
- 修改：`code/backend/internal/module/auth/contracts/oauth.go`

- [x] 步骤 1：写 MCP OAuth 红测：无 `Authorization` 返回 HTTP 401 + `WWW-Authenticate`。
- [x] 步骤 2：写 MCP OAuth 红测：无效 access token、过期 token、scope 不包含 `mcp:challenge:read` 都返回 401，且不会调用 instance/challenge service。
- [x] 步骤 3：写 MCP OAuth 绿测：合法 access token + scope 可调用 `tools/call get_current_challenge`。
- [x] 步骤 4：写 `tools/list` 行为确认：可以无需 token 返回工具列表，或按 MCP 客户端兼容性决定仍返回 401；实现前在测试名中固定选择。建议第一版允许 `initialize` / `tools/list` 无 token，`tools/call` 必须有 token。
- [x] 步骤 5：删除 `tokenResolver`、`TokenURL`、`defaultTokenURL`、`ResolveMCPToken` 调用和 `auth_method=bearer_token`。
- [x] 步骤 6：新增 `oauthTokenResolver` 接口：

```go
type oauthTokenResolver interface {
    ResolveOAuthAccessToken(ctx context.Context, token string, requiredScope string) (*authctx.CurrentUser, error)
}
```

- [x] 步骤 7：实现 `WWW-Authenticate` header builder，`resource_metadata` 根据 issuer / request origin 生成。
- [x] 步骤 8：保留 `rate_limit.mcp` 和 `mcp_tool/read` 审计，但 detail 增加 `client_id` 和 `scope`，不记录 token。
- [x] 步骤 9：运行验证。

运行：

```bash
cd code/backend && go test ./internal/interfaces/mcp -run 'MCP|OAuth' -count=1
cd code/backend && go test ./internal/app -run '^TestNewRouterRegistersStudentChallengeRoutes$' -count=1
cd code/backend && go test ./internal/module/auth/... ./internal/interfaces/mcp ./internal/app -run '^$' -count=1
```

预期：PASS。

- [x] 步骤 10：提交本切片。

### 任务 7：删除旧 token 实现、配置和测试

**目标：** 彻底移除旧 MCP token 方案，避免留下死代码或隐藏兼容路径。

**文件：**

- 修改：`code/backend/internal/module/auth/contracts/token_service.go`
- 修改：`code/backend/internal/module/auth/infrastructure/token_service.go`
- 修改：`code/backend/internal/module/auth/infrastructure/token_service_test.go`
- 修改：`code/backend/internal/module/auth/api/http/handler.go`
- 修改：`code/backend/internal/module/auth/api/http/auth_types.go`
- 修改：`code/backend/internal/module/auth/api/http/http_integration_test.go`
- 修改：`code/backend/internal/app/router.go`
- 修改：`code/backend/internal/config/types.go`
- 修改：`code/backend/internal/config/defaults.go`
- 修改：`code/backend/internal/config/validate.go`
- 修改：`code/backend/configs/config.yaml`
- 修改：`code/backend/configs/config.dev.yaml`
- 修改：`code/backend/configs/config.prod.yaml`

- [ ] 步骤 1：删除 `MCPToken` struct、`IssueMCPToken`、`ResolveMCPToken`、`mcpTokenPayload`、`mcpTokenKey`。
- [ ] 步骤 2：删除 `Handler.IssueMCPToken` 和 `MCPTokenResp`。
- [ ] 步骤 3：删除 router 中 `protected.POST("/auth/mcp-token", ...)`。
- [ ] 步骤 4：删除 `auth.mcp_token_ttl` 默认值、配置项、校验和 YAML 示例。
- [ ] 步骤 5：删除旧 MCP token 相关测试，保留“endpoint removed”测试作为迁移 guard。
- [ ] 步骤 6：全仓搜索旧关键字。

运行：

```bash
rg -n "mcp-token|MCPToken|IssueMCPToken|ResolveMCPToken|mcp_token_ttl|token_url|CTF_MCP_TOKEN|bearer_token_env_var" .
```

预期：只允许本计划、历史旧计划、归档说明或明确标注为 superseded 的文档命中；生产代码和当前契约不得命中。

- [ ] 步骤 7：运行后端验证。

运行：

```bash
cd code/backend && go test ./internal/module/auth/... ./internal/interfaces/mcp ./internal/config ./internal/app -run 'OAuth|MCP|Token|Route|Config' -count=1
```

预期：PASS。

- [ ] 步骤 8：提交本切片。

### 任务 8：更新 OpenAPI、API contract、架构事实和操作文档

**目标：** 当前事实文档不再描述 MCP token，改为 OAuth 浏览器授权。

**文件：**

- 修改：`docs/contracts/api-contract-v1.md`
- 新增：`docs/contracts/openapi-v1/paths/oauth.yaml`
- 新增：`docs/contracts/openapi-v1/components/schemas/oauth.yaml`
- 修改：`docs/contracts/openapi-v1/paths/auth.yaml`
- 修改：`docs/contracts/openapi-v1/components/schemas/auth.yaml`
- 修改：`docs/contracts/openapi-v1/index.yaml`
- 修改：`docs/contracts/openapi-v1.yaml`
- 修改：`docs/architecture/backend/04-api-design.md`
- 修改：`docs/architecture/backend/modules/auth.md`
- 修改：`docs/design/ctf-tutor-agent-and-mcp.md`
- 新增：`docs/operations/mcp-oauth-login.md`

- [ ] 步骤 1：删除 API contract 中 `POST /api/v1/auth/mcp-token`。
- [ ] 步骤 2：新增 OAuth endpoint 契约，至少覆盖 metadata、register、authorize、token 的 request/response/error。
- [ ] 步骤 3：更新 `/mcp` 契约，说明 OAuth Bearer token、`WWW-Authenticate`、`mcp:challenge:read` scope。
- [ ] 步骤 4：更新后端 API 架构文档：`/mcp` 不依赖 Cookie、不依赖 MCP token；`auth` 模块是 OAuth AS owner。
- [ ] 步骤 5：更新 auth 模块文档：新增 OAuth client/code/token/consent owner，删除 MCP token owner。
- [ ] 步骤 6：更新设计稿：把旧 token 方式标记为 `Superseded by OAuth browser authorization`。
- [ ] 步骤 7：新增操作文档，写明外部 agent 使用方式：

```bash
codex mcp login ctf
claude mcp login ctf
```

并写明 Codex 配置不再包含：

```toml
bearer_token_env_var = "CTF_MCP_TOKEN"
```

- [ ] 步骤 8：同步 OpenAPI bundle。

运行：

```bash
python3 tools/sync_openapi_from_contract.py
```

预期：PASS，`docs/contracts/openapi-v1.yaml` 更新。

- [ ] 步骤 9：运行文档一致性检查。

运行：

```bash
python3 scripts/check-docs-consistency.py
```

预期：PASS；若命中既有无关缺失引用，记录具体路径并继续跑 workflow gate。

- [ ] 步骤 10：提交本切片。

### 任务 9：端到端本地验证

**目标：** 用真实 HTTP 请求验证 OAuth + MCP 链路可被外部 agent 调用。

**文件：**

- 修改或新增：`tools/` 下的本地验证脚本，只有在手工 curl 过长且可复用时才新增。
- 修改：`docs/operations/mcp-oauth-login.md`，补充本地验证输出。

- [ ] 步骤 1：启动 Docker 开发环境，确认 API 和前端端口。
- [ ] 步骤 2：请求 protected resource metadata。

运行：

```bash
curl -i http://127.0.0.1:8080/.well-known/oauth-protected-resource
```

预期：200，返回 `/mcp` resource 和 authorization server。

- [ ] 步骤 3：请求 authorization server metadata。

运行：

```bash
curl -i http://127.0.0.1:8080/.well-known/oauth-authorization-server
```

预期：200，包含 `authorization_endpoint`、`token_endpoint`、`registration_endpoint`、`S256`。

- [ ] 步骤 4：执行 dynamic client registration。

运行：

```bash
curl -sS -X POST http://127.0.0.1:8080/api/v1/oauth/register \
  -H 'Content-Type: application/json' \
  -d '{"client_name":"local-codex","redirect_uris":["http://127.0.0.1:14567/callback"],"grant_types":["authorization_code","refresh_token"],"response_types":["code"],"scope":"mcp:challenge:read","token_endpoint_auth_method":"none"}'
```

预期：返回 `client_id`，不返回 `client_secret`。

- [ ] 步骤 5：无 access token 调用 `/mcp`。

运行：

```bash
curl -i -X POST http://127.0.0.1:8080/mcp \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_current_challenge","arguments":{}}}'
```

预期：401，`WWW-Authenticate` 指向 `/.well-known/oauth-protected-resource`。

- [ ] 步骤 6：用浏览器打开 authorize URL，确认未登录会跳到 `/login`，登录后回到授权页。
- [ ] 步骤 7：同意授权后，本地 callback 收到 `code` 和原始 `state`。
- [ ] 步骤 8：用 `code_verifier` 换 token，确认返回 access token 和 refresh token。
- [ ] 步骤 9：用 access token 调 `/mcp`，确认 `get_current_challenge` 返回当前用户题目信息或 `has_current_challenge=false`。
- [ ] 步骤 10：运行 workflow gate。

运行：

```bash
bash scripts/run-workflow-stage.sh pre-commit-quick
bash scripts/run-workflow-stage.sh completion-full
git diff --check
```

预期：PASS。

- [ ] 步骤 11：提交最终文档/验证补充。

## 风险与缓解

- 风险：OAuth DCR 被滥用注册大量 client。
  - 缓解：默认只允许 public loopback redirect URI；注册接口走 IP rate limit；client id 不赋予任何权限，必须用户授权 scope 才能访问 MCP。
- 风险：redirect URI 校验不严导致 authorization code 泄露。
  - 缓解：注册和 authorize/token 三处使用同一个 canonical validation；禁止 fragment；token exchange 要求 `redirect_uri` 与 authorize 时完全一致。
- 风险：PKCE 实现错误导致 code 被截获后可用。
  - 缓解：只支持 `S256`；authorization code payload 存储 `code_challenge` 和 method；token endpoint 使用 constant-time compare。
- 风险：旧 MCP token 代码残留形成第二条认证路径。
  - 缓解：任务 1 和任务 7 均有 `rg` guard；生产代码和当前契约不得再命中旧关键字。
- 风险：现有前端登录用 Vue Router `push()`，不能跳到 `/api/v1/oauth/authorize`。
  - 缓解：只对 `/api/v1/oauth/authorize` 这类后端授权 path 使用 `window.location.assign()`，普通登录跳转保持原逻辑。
- 风险：`ctf_session` cookie path 是 `/api/v1`，如果授权端点放到 `/oauth/authorize` 会拿不到登录态。
  - 缓解：authorization endpoint 放到 `/api/v1/oauth/authorize`；metadata 指向该标准端点 URL。
- 风险：refresh token 泄露后长期可用。
  - 缓解：refresh token 旋转、session version 校验、TTL、审计复用检测；后续如需要可增加 consent 管理页和单 client revoke。

## 架构适配评估

- 文档语言：本计划正文使用中文，代码标识、路径、协议字段保持原文。
- 目标边界：`auth` 负责 OAuth Authorization Server；`interfaces/mcp` 负责 MCP 协议和 scope 检查；`instance` / `challenge` 继续负责题目事实。
- 复用点：复用 Redis token TTL 模式、现有 session version 撤销语义、`auditlog.Recorder`、`rate_limit.mcp`、现有登录页和登录后 redirect 参数。
- 结构收敛：本计划不只替换响应文案，而是删除旧 MCP token 签发/解析链路，新增 OAuth owner 和 discovery endpoint，避免完成后立刻二次重构。
- 有意延期：不做 consent 管理页面、不做 OIDC、不做写 scope；这些不是浏览器授权接入的必要条件，应作为后续独立任务。

## Validation

- 后端单元 / 集成：

```bash
cd code/backend && go test ./internal/module/auth/... ./internal/interfaces/mcp ./internal/config -count=1
cd code/backend && go test ./internal/app -run 'Migration|Route|OAuth|MCP' -count=1
```

- 前端登录跳转：

```bash
cd code/frontend && pnpm test:run src/features/auth/model/useLoginPage.test.ts
```

- 契约和文档：

```bash
python3 tools/sync_openapi_from_contract.py
python3 scripts/check-docs-consistency.py
```

- 完成门禁：

```bash
bash scripts/run-workflow-stage.sh pre-commit-quick
bash scripts/run-workflow-stage.sh completion-full
git diff --check
```

## 完成判定

- `POST /api/v1/auth/mcp-token` 已删除，生产代码无旧 MCP token 认证路径。
- `/.well-known/oauth-protected-resource` 和 `/.well-known/oauth-authorization-server` 可发现 OAuth 配置。
- 外部 public client 可通过 DCR 注册 loopback redirect URI。
- 用户可通过浏览器登录 CTF 并同意 `mcp:challenge:read`。
- `/api/v1/oauth/token` 支持 authorization code + PKCE 和 refresh token rotation。
- `/mcp` 只接受 OAuth Bearer access token，并检查 `mcp:challenge:read`。
- `codex mcp login ctf` / `claude mcp login ctf` 成为目标操作路径，`CTF_MCP_TOKEN` 不再需要。
- OpenAPI、API contract、后端架构文档和操作文档与实现一致。
