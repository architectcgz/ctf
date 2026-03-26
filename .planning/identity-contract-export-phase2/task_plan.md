# Task Plan

## Goal

消除 `auth` 对 `IdentityModule` 私有字段的访问：将用户仓储以公开 contract 形式暴露，并补守卫禁止回退。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 identity/auth 边界遗留项 | completed | 已确认唯一遗留是 `identity.users` 私有字段被跨模块读取 |
| 2. 补结构守卫暴露 red case | completed | 已在 `router_test.go` 增加 `Users` contract 与私有字段禁用守卫 |
| 3. 切换 composition 到公开 contract | completed | `IdentityModule` 与 `auth_module.go` 已切换到 `Users` contract |
| 4. focused 验证 | completed | `internal/app`、`identity/...`、`auth/...` 定向测试已通过 |
