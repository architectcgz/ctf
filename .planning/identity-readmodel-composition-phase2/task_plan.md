# Task Plan

## Goal

完成 `identity + readmodel` 轻量 composition 收口：为 `identity`、`practice_readmodel`、`teaching_readmodel` 引入 typed deps，消除 inline concrete 装配。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点轻量 composition 遗留项 | completed | 已确认主要集中在 identity 与两个 readmodel 模块 |
| 2. 以结构守卫暴露 red case | completed | 已补 identity/readmodel typed deps 守卫 |
| 3. 切换 composition 到 typed deps | completed | 三个模块已引入局部 deps builder |
| 4. focused 验证 | completed | `internal/app`、`identity/...`、两个 readmodel 模块测试已通过 |
