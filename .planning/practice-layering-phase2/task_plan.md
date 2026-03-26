# Task Plan

## Goal

完成 `practice` Phase 2 第一刀：删除 legacy 宽 `PracticeRepository`，把应用层切到按用例划分的窄端口，并让 composition 依赖类型收口到 ports。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 practice 应用层真实依赖面 | completed | 已确认 command / score-write / score-read 三类依赖面清晰可拆 |
| 2. 以结构守卫暴露 red case | completed | 已补 ports 宽接口禁用与 composition typed deps 守卫 |
| 3. 在 `practice/ports` 中拆出窄接口 | completed | 已拆为 command / command-tx / score / ranking 四组 |
| 4. 切换应用层与测试桩 | completed | `service / score_service / ranking query` 与 stub 已收口 |
| 5. composition 收口 | completed | `BuildPracticeModule` 已增加 typed deps，避免直接持有 concrete practice repo 字段 |
| 6. focused 验证 | completed | `practice/...` 与 `internal/app` 相关测试已通过 |
