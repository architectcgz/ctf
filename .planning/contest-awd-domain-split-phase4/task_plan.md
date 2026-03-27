# Task Plan

## Goal

继续收口 `contest` AWD domain，把 `domain/awd.go` 从单文件拆成响应映射、check source/result support 与 flag/error support 三段，保持 AWD domain helper 行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd domain 职责 | completed | 已确认单文件同时承载 DTO 映射、check support、flag/error support |
| 2. 拆分文件结构 | completed | 已拆为 `awd_response.go`、`awd_check_support.go`、`awd_flag_support.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `domain/awd.go` 的混合职责被拆到独立语义文件
- AWD DTO 映射与 summary 排序拆到 `awd_response.go`
- check source / check result 规范化拆到 `awd_check_support.go`
- flag 生成与 unique error support 拆到 `awd_flag_support.go`
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 AWD domain helper 的对外函数签名与行为
- 仅改善 contest AWD domain 文件边界，为后续 AWD application / query 复用留下更清晰结构
