# 教师端 legacy route allowlist guardrail 计划

## Objective

- 把 `teacherRoutes.ts` 中仍保留的 `/teacher/*` 兼容 redirect 收成明确 allowlist。
- 补源码级测试，限制后续不能继续新增新的 `/teacher/*` 前端页面入口。

## Non-goals

- 本轮不直接删除现有 `/teacher/*` redirect。
- 不改教师端 API `/teacher/*` 后端接口命名。
- 不调整 `/academy/*` 现有 canonical route 行为。

## Source Inputs

- `TODO/frontend-sliced-architecture.md`
- `code/frontend/src/router/routes/teacherRoutes.ts`
- `code/frontend/src/router/__tests__/sharedRoutes.test.ts`

## Task Slices

### Slice 1: 收口 legacy redirect allowlist

- 目标：把遗留教师端 redirect 整理成单点 allowlist owner。
- 变更面：
  - `code/frontend/src/router/routes/teacherRoutes.ts`
- 风险：
  - 路径白名单不全会误伤现有兼容入口。

### Slice 2: 补 shared route guardrail

- 目标：用路由测试限制 `/teacher/*` legacy redirect 只能来自当前 allowlist。
- 变更面：
  - `code/frontend/src/router/__tests__/sharedRoutes.test.ts`
- 风险：
  - 断言如果只测存在、不测全集，后续仍可能悄悄长出新的 legacy route。

### Slice 3: 同步迁移台账

- 目标：把“教师端兼容入口先整理 allowlist”记成已完成的当前事实。
- 变更面：
  - `TODO/frontend-sliced-architecture.md`

## Validation Plan

- `npm run test:run -- src/router/__tests__/sharedRoutes.test.ts src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- `bash scripts/check-consistency.sh`

## Review Focus

- allowlist 是否覆盖当前仍保留的所有 `/teacher/*` redirect，且没有混入新的 runtime 页面入口。
- 测试是否检查“exact allowlist”，而不是宽松的存在性断言。

## Rollback / Recovery

- 如果 allowlist 抽取影响现有 redirect 行为，可先回退到原始逐条 route 定义，再单独保留 guardrail 测试。
