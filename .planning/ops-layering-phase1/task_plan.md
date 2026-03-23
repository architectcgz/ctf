# Task Plan

## Goal

把 `ops` 从根包大平铺收敛成符合 CTF 规范的物理分层结构。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 `ops` 根包现状 | completed | 已确认根包同时混放 handler / service / repository |
| 2. 建立目标目录结构 | completed | 已建立 `api/http`、`application`、`infrastructure` |
| 3. 迁移审计/仪表盘/风控实现 | completed | 代码已按层拆分并更新引用 |
| 4. 收紧根包暴露面 | completed | 根包仅保留 `contracts.go`、`module.go` 与架构测试 |
| 5. 定向验证 | completed | `ops/...` 与 `app` 相关测试已通过 |

## Acceptance Checks

- `internal/module/ops` 根包不再保留 `*_handler.go`、`*_service.go`、`*_repository.go`
- HTTP handler 位于 `internal/module/ops/api/http`
- 应用服务位于 `internal/module/ops/application`
- 基础设施适配器位于 `internal/module/ops/infrastructure`
- 对外 contract 继续由根包 `contracts.go` 暴露

## Constraints

- 不新增兼容层
- 外部 API 路径不变
- 测试默认限核、最小充分范围
