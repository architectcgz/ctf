# Task Plan

## Goal

完成 `runtime-bridge-phase2`：把 practice runtime adapter 从 `practice` composition 挪回 `runtime` composition，并消除 `runtime` 对 `contest/infrastructure` 类型的依赖。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 runtime/practice/contest bridge 依赖 | completed | 已确认问题集中在 `practice_module.go` adapter 和 `runtime_module.go` 的 concrete cross-module type |
| 2. 以结构守卫暴露 red case | completed | 已补 practice glue 禁用与 runtime external ports deps 守卫 |
| 3. 将 practice runtime adapter 下沉到 runtime composition | completed | `practice_module.go` 已只保留装配，adapter 与拓扑映射迁回 `runtime_module.go` |
| 4. 消除 runtime 对 contest infrastructure 的反向依赖 | completed | 已通过 `contest/ports.AWDContainerFileWriter` 收口 `containerFiles` 类型 |
| 5. focused 验证 | completed | `internal/app`、`contest/...`、`practice/...` 相关测试已通过 |
