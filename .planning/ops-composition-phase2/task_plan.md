# Task Plan

## Goal

标准化 `ops` composition：收口 `BuildOpsModule` 与 `BuildNotificationHandler` 的装配依赖，统一为 typed deps + 局部 builder。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 ops composition 遗留项 | completed | 已确认遗留集中在 audit/dashboard/risk/notification 装配 |
| 2. 补结构守卫暴露 red case | completed | 已在 `router_test.go` 增加 ops typed deps 与 sub-builder 守卫 |
| 3. 切换 composition 到标准装配模式 | completed | `ops_module.go` 与 `runtime_module.go` 已收口 typed deps 与局部 builder |
| 4. focused 验证 | completed | `internal/app`、`internal/module/ops/...` 定向测试已通过 |
