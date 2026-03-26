# Task Plan

## Goal

标准化 `auth` composition：为 `BuildAuthModule` 引入 typed deps，避免直接读取其他 composition 模块的内部字段。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 auth composition 遗留项 | completed | 已确认主要问题是 `identity.users` 与 `ops.AuditService` 直连使用 |
| 2. 补结构守卫暴露 red case | completed | 已在 `router_test.go` 增加 auth typed deps 守卫 |
| 3. 切换 auth composition 到 typed deps | completed | `auth_module.go` 已收口 users/token/profile/audit 依赖 |
| 4. focused 验证 | completed | `internal/app` 与 `internal/module/auth/...` 定向测试已通过 |
