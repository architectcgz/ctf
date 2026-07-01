# MCP OAuth 登录操作说明

> 状态：Current
> 事实源：`code/backend/internal/interfaces/mcp/handler.go`、`code/backend/internal/module/auth/api/http/oauth_handler.go`、`docs/contracts/api-contract-v1.md`
> 替代：旧 MCP token 手工签发流程

## 定位

本文档说明外部 AI 客户端如何通过浏览器 OAuth 授权连接 CTF 平台 `/mcp`。

- 负责：说明 MCP client 的登录入口、平台需要暴露的 URL、授权后访问 `/mcp` 的行为，以及旧 token 配置的移除。
- 不负责：定义 OAuth handler 的字段契约、实现本地开发环境启动步骤或替代 OpenAPI；字段契约见 `docs/contracts/api-contract-v1.md`。

## 前置条件

- 平台 HTTPS 地址可访问 `/mcp`、`/.well-known/oauth-protected-resource`、`/.well-known/oauth-authorization-server`。
- OAuth issuer 与生产域名一致；生产环境必须配置 `auth.oauth.issuer_url` 为 HTTPS origin。
- 外部客户端的 MCP server 配置只需要指向平台 `/mcp` URL，不再配置静态 Bearer token。

## 客户端登录

Codex：

```bash
codex mcp login ctf
```

Claude：

```bash
claude mcp login ctf
```

客户端会读取 `/mcp` 的 protected resource metadata，打开浏览器进入 `/api/v1/oauth/authorize`。用户在浏览器登录 CTF 平台并同意 `mcp:challenge:read` 后，客户端获得 OAuth access token 和 refresh token。

## MCP 调用行为

- `initialize`、`notifications/initialized` 和 `tools/list` 可无 token 调用。
- `tools/call` 必须携带 OAuth access token，且 token 必须包含 `mcp:challenge:read`。
- 未授权时 `/mcp` 返回 HTTP `401`，`WWW-Authenticate` 指向 `/.well-known/oauth-protected-resource`。
- 成功调用 `get_current_challenge` 后，响应 `structuredContent` 包含 `has_current_challenge`、`instance` 和 `challenge`；没有活动实例时 `has_current_challenge=false`。

## 已移除配置

不要再使用或保留以下静态 token 配置：

```toml
bearer_token_env_var = "CTF_MCP_TOKEN"
```

也不再需要用户手工调用 `POST /api/v1/auth/mcp-token`。旧 endpoint 已删除，旧 token 不再是 `/mcp` 的认证路径。
