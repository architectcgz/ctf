# Task Plan

## Goal

继续收口 `contest` AWD 主流程，把 `awd_attack_commands.go` 从单文件拆成更清晰的 attack log 与 submit attack 命令结构，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 AWD attack command 职责 | completed | 已确认 `awd_attack_commands.go` 同时承载手工 attack log 与 flag submit 两条命令流程 |
| 2. 拆分文件结构 | completed | 已拆为 `awd_attack_commands.go`、`awd_attack_log_commands.go`、`awd_attack_submit_commands.go` |
| 3. focused 验证 | completed | 已运行 `./internal/module/contest/...` 定向测试 |

## Acceptance Checks

- `awd_attack_commands.go` 不再承载混杂命令流程
- attack log 与 submit attack 命令拆到独立文件
- `contest/...` 定向测试通过

## Result

- 不改 AWD commands 对外接口
- 仅改善 AWD attack 主流程文件边界，为后续继续收 AWD 写侧留下更清晰结构
