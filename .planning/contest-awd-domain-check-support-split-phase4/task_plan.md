# Task Plan

## Goal

继续收口 `contest` AWD domain support，把 `domain/awd_check_support.go` 从单文件拆成 source 归一化与 check result JSON support 两段，保持 domain helper 对外行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd check support 职责 | completed | 已确认单文件同时承载 source 归一化与 check result parse/marshal |
| 2. 拆分文件结构 | completed | 已拆为 `awd_source_support.go` 与 `awd_check_result_support.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_check_support.go` 不再混载 source 与 check result 两类 helper
- source 归一化 helper 下沉到 `awd_source_support.go`
- check result JSON parse/marshal 与手工补全下沉到 `awd_check_result_support.go`
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `contest/domain` 对外 helper 函数签名与行为
- 仅改善 AWD domain support 文件边界，为后续继续收口 contest domain/support 留下更清晰结构
