# Task Plan

## Goal

消除其他模块对 `RuntimeModule` 私有嵌套字段的读取：将 runtime 跨模块 contract 正式公开，并补守卫禁止回退。

## Phases

| Phase | Status | Notes |
|---|---|---|
| 1. 盘点 runtime 对外 contract 遗留项 | completed | 已确认遗留集中在 `challenge / ops / practice / contest` 对私有字段的读取 |
| 2. 补结构守卫暴露 red case | completed | 已在 `router_test.go` 增加 runtime contract 暴露与私有字段禁用守卫 |
| 3. 切换 consumers 到公开 contract | completed | `runtime` 与相关 composition 模块已切到公开 contract 字段 |
| 4. focused 验证 | completed | `internal/app` 与受影响模块定向测试已通过 |
