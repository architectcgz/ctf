# Task Plan

## Goal

标准化 `assessment` composition 的跨模块装配：把 `challenge` 依赖拆到独立 external deps builder。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 assessment 剩余耦合点 | completed | 已确认只剩 `challenge` contract 混在主 deps builder |
| 2. 补结构守卫暴露 red case | completed | 已在 `router_test.go` 增加 assessment cross-deps 守卫 |
| 3. 切换 assessment composition 到标准 builder | completed | `assessment_module.go` 已拆分 main deps 与 external deps |
| 4. focused 验证 | completed | `internal/app` 与 `internal/module/assessment/...` 定向测试已通过 |
