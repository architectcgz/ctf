# 安全审查待修复项

> 来源：2026-06-02 项目网络安全审查
> 审查范围：`code/backend/`、`code/frontend/`、`docker/`、`configs/`

## P1 — 高风险

- [ ] **Docker socket 挂载权限收窄**
  `docker/ctf/docker-compose.dev.yml:66` — `/var/run/docker.sock` 直接挂载到 API 容器，容器被攻破后攻击者可控制宿主机 Docker daemon。考虑：Docker TLS 远程访问 + 受限权限、只读 rootfs、非 root 运行 API 容器。
  关联：`docker/ctf/docker-compose.dev.yml:63-66`

- [ ] **iptables 命令参数校验加固**
  `code/backend/internal/module/runtime/infrastructure/acl.go:45-72` — `buildACLCommand` 将 `InstanceRuntimeACLRule` 的 `SourceIP`、`TargetIP`、`Comment` 等字段直接拼入 iptables 参数。需确认数据来源是否用户可控，并对所有字段做白名单校验（IP 格式、协议名、action 名、Comment 字符集限制）。

## P2 — 中等风险

- [ ] **开发环境凭据去硬编码**
  `docker/ctf/docker-compose.dev.yml` 和 `code/backend/scripts/dev-run.sh` 中硬编码了 `postgres123456`、`redis123456`、`ctf-container-flag-secret-0123456789abcdef`。建议改为随机生成（`.env` 或启动脚本自动生成），并在 README 标注不可用于生产。

- [ ] **WebSocket 端点认证校验确认**
  `code/backend/internal/app/router.go:135-136,191-193` — `/ws/notifications`、`/ws/contests/:id/announcements` 等 WebSocket 端点绕过 Auth 中间件，直接挂在 `engine` 上。需逐一确认 handler 内部是否有独立的 ticket/query param 认证，防止未授权连接。

- [ ] **生产配置中去掉 `change_me` 占位**
  `code/backend/configs/config.prod.yaml:11,18` — PostgreSQL 和 Redis 密码的 `change_me` 占位符虽然启动校验会拦截，但仍应替换为空字符串，强制通过环境变量注入，降低校验被绕过的风险面。

- [ ] **添加 Content-Security-Policy 响应头**
  当前缺少 CSP 头。建议在中间件或反向代理层添加合理的 CSP 策略（至少 `default-src 'self'; script-src 'self'`），作为 XSS 纵深防御。

## P3 — 低风险

- [ ] **报表 SQL 拼接模式收敛**
  `code/backend/internal/module/assessment/infrastructure/report_repository.go:548-614` — `GetStudentTimeline` 使用 `fmt.Sprintf` 拼接 SQL 片段。当前参数均为硬编码常量无实际风险，但模式容易被复制。建议改为参数化或加注释标明参数必须为常量。

## 已确认无需处理

- CSRF：session cookie 已设置 `SameSite=Lax`，结合 `HttpOnly`，保护充分
- XSS：所有 `v-html` 使用点均经过 DOMPurify 白名单净化
- 密码存储：bcrypt + DefaultCost，符合要求
- CORS：白名单 + 启动校验，无通配符风险
- 会话撤销：session version 机制已覆盖改密踢出场景
