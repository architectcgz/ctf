# Task Plan

## Goal

继续收口 `contest` AWD infrastructure，把 `infrastructure/awd_flag_injector.go` 从单文件拆成更清晰的 factory/noop、docker injector、container support 三段，保持 `AWDFlagInjector` 行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 awd flag injector 职责 | completed | 已确认单文件同时承载 noop factory、docker 注入实现、container id 解析 support |
| 2. 拆分文件结构 | completed | 已保留 `awd_flag_injector.go` 为入口，并新增 `awd_docker_flag_injector.go`、`awd_flag_injector_support.go` |
| 3. focused 验证 | completed | `contest/...` 与相关 `internal/app` 定向测试已通过 |

## Acceptance Checks

- `awd_flag_injector.go` 不再混载 factory/noop、docker injector、support 三类职责
- docker 注入流程与 container id 解析 support 拆到独立文件
- `contest/...` 与相关 `internal/app` 定向测试通过

## Result

- 不改 `AWDFlagInjector` 对外行为与 composition 装配方式
- 仅改善 `contest` AWD injector infrastructure 文件边界，为后续继续收口 AWD infrastructure 链路留下更清晰结构
