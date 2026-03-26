# Task Plan

## Goal

完成 `assessment` Phase 2 收口：让 composition 依赖收口到 ports/contracts，并拆分局部 builder，避免继续把 concrete repo 和全部装配堆在单个函数里。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 assessment composition 遗留耦合点 | completed | 主要问题是 concrete repo 直连与单函数装配 |
| 2. 以结构守卫暴露 red case | completed | 已补 typed deps 与 sub-builder 守卫 |
| 3. 切换 composition 到 typed deps | completed | `assessmentModuleDeps` 已改为 ports/contracts |
| 4. 拆分局部 builder | completed | 已拆 profile/recommendation/report 局部装配 |
| 5. focused 验证 | completed | `internal/app` 与 `assessment/...` 相关测试已通过 |
