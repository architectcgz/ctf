# Task Plan

## Goal

继续收口 `contest` AWD jobs，把 `application/jobs/awd_probe.go` 从单文件拆成更清晰的 probe 主流程与 helper 结构，保持行为不变。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 AWD probe 职责 | completed | 已确认单文件同时承载 probe 执行、URL 构造、error normalize 与 timeout/path helper |
| 2. 拆分文件结构 | completed | 已拆为 `awd_probe.go` 与 `awd_probe_support.go` |
| 3. focused 验证 | completed | 已运行 `contest/...` 定向测试 |

## Acceptance Checks

- `awd_probe.go` 不再承载混杂 AWD probe helper
- probe 执行 与 URL/error/timeout helper 拆到独立文件
- `contest/...` 定向测试通过

## Result

- 不改 `AWDRoundUpdater` 对外接口
- 仅改善 AWD probe 内部文件边界，为后续继续收口 contest jobs 留下更清晰结构
