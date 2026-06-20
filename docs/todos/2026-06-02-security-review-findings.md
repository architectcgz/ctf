# 安全审查待修复项

> 来源：2026-06-02 项目网络安全审查
> 审查范围：`code/backend/`、`code/frontend/`、`docker/`、`configs/`

## P1 — 高风险

- [x] **Docker socket 挂载权限收窄**（2026-06-20 已修复）
  `docker/docker-compose.dev.yml` 已将宿主 `/var/run/docker.sock` 从 `ctf-api` 和 `ctf-awd-defense-ssh-gateway` 移除，默认改为 API / gateway 通过 mTLS 调用 `ctf-runtime-agent`，只有 runtime-agent 服务继续持有 Docker socket 作为本地 dev execution node。
  关联：`docker/docker-compose.dev.yml`

- [x] **iptables 命令参数校验加固**（2026-06-02 已修复）
  `code/backend/internal/module/runtime/infrastructure/acl.go` — 已在 `validateAndCanonicalizeACLRule()` 中对 SourceIP/TargetIP（`net/netip` 单 IPv4）、Action（`allow`/`deny` 白名单）、Protocol（`any`/`tcp`/`udp` 白名单）、Ports（1-65535 去重排序、multiport 上限 15、`protocol=any` 禁止端口）、Comment（系统重建，不信任持久化值）做执行前白名单校验。同时 ACL cleanup authority 已从 `acl_rules` 数据库字段收口到实例级 iptables chain handle（`runtime_details.acl`），`acl_rules` 降级为调试快照。

## P2 — 中等风险

- [ ] **开发环境凭据去硬编码**
  `docker/docker-compose.dev.yml` 和 `code/backend/scripts/dev-run.sh` 中硬编码了 `postgres123456`、`redis123456`、`ctf-container-flag-secret-0123456789abcdef`。建议改为随机生成（`.env` 或启动脚本自动生成），并在 README 标注不可用于生产。

- [ ] **WebSocket 端点认证校验确认**
  `code/backend/internal/app/router.go:135-136,191-193` — `/ws/notifications`、`/ws/contests/:id/announcements` 等 WebSocket 端点绕过 Auth 中间件，直接挂在 `engine` 上。需逐一确认 handler 内部是否有独立的 ticket/query param 认证，防止未授权连接。

- [ ] **生产配置中去掉 `change_me` 占位**
  `code/backend/configs/config.prod.yaml:11,18` — PostgreSQL 和 Redis 密码的 `change_me` 占位符虽然启动校验会拦截，但仍应替换为空字符串，强制通过环境变量注入，降低校验被绕过的风险面。

- [x] **添加 Content-Security-Policy 响应头**（2026-06-02 已修复）
  `code/frontend/nginx/default.conf` 已增加文档响应 CSP，并将 allowlist 收紧为当前前端真实依赖：脚本仅 `self`，样式 / 字体仅放行 Google Fonts，`img-src` 收口为 `self + data:`，`connect-src` 收口为 `self`。`code/backend/internal/middleware/security_headers.go` 同时为 API 响应补充最小防御头；原先把 `img-src https:` 放宽的 DiceBear 外链头像已从 `code/frontend/src/pages/utility/UILabRoutePage.vue` 移除。

## P3 — 低风险

- [ ] **报表 SQL 拼接模式收敛**
  `code/backend/internal/module/assessment/infrastructure/report_repository.go:548-614` — `GetStudentTimeline` 使用 `fmt.Sprintf` 拼接 SQL 片段。当前参数均为硬编码常量无实际风险，但模式容易被复制。建议改为参数化或加注释标明参数必须为常量。

## 已确认无需处理

- CSRF：session cookie 已设置 `SameSite=Lax`，结合 `HttpOnly`，保护充分
- XSS：所有 `v-html` 使用点均经过 DOMPurify 白名单净化
- 密码存储：bcrypt + DefaultCost，符合要求
- CORS：白名单 + 启动校验，无通配符风险
- 会话撤销：session version 机制已覆盖改密踢出场景
