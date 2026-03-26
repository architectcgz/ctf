# Task Plan

## Goal

收紧 `contest` composition 的跨模块依赖：`contestModuleDeps` 不再保存整块 `ChallengeModule` / `RuntimeModule`，改为只持有 typed contracts。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 contest 剩余耦合点 | completed | 已确认遗留集中在 `contest_module.go` 的跨模块依赖字段 |
| 2. 补结构守卫暴露 red case | completed | 已在 `router_test.go` 增加 typed cross-module deps 守卫 |
| 3. 切换 contest composition 到 typed deps | completed | 已切到 `challengeCatalog / flagValidator / containerFiles` 三类 contract 字段 |
| 4. focused 验证 | completed | `internal/app` 与 `internal/module/contest/...` 定向测试已通过 |
