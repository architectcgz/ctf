# Task Plan

## Goal

标准化 `practice` composition 的跨模块装配：拆分 persistence deps、external deps、handler builder，统一到 phase2 风格。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 practice 剩余耦合点 | completed | 已确认主要集中在 `practice_module.go` 的跨模块装配混写 |
| 2. 补结构守卫暴露 red case | completed | 已在 `router_test.go` 增加 practice cross-deps 与 sub-builder 守卫 |
| 3. 切换 practice composition 到标准 builder | completed | `practice_module.go` 已拆分 persistence/external/handler builder |
| 4. focused 验证 | completed | `internal/app` 与 `internal/module/practice/...` 定向测试已通过 |
