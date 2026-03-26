# Task Plan

## Goal

标准化 `runtime` composition：为 `BuildRuntimeModule` 引入 typed deps 与局部 builder，拆分后台任务注册和各子能力装配。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 runtime composition 遗留项 | completed | 已确认主要集中在 `BuildRuntimeModule` 单函数承担过多装配职责 |
| 2. 补结构守卫暴露 red case | completed | 已在 `router_test.go` 增加 runtime typed deps 与 sub-builder 守卫 |
| 3. 切换 runtime composition 到标准 builder | completed | `runtime_module.go` 已拆分 deps、jobs、http/practice/challenge/ops/contest builders |
| 4. focused 验证 | completed | `internal/app` 与 `internal/module/runtime/...` 定向测试已通过 |
