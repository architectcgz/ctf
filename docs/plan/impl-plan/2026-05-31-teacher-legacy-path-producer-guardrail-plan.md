# 教师端 legacy path producer guardrail 计划

## Objective

- 用源码级 guardrail 限制前端仓库内部不再新增 `/teacher/*` 页面路径 producer。
- 明确 `teacherLegacyRedirect.ts` 是当前唯一允许保留旧教师端页面路径的 runtime 兼容 owner。

## Non-goals

- 不删除现有 `/teacher/*` redirect。
- 不修改 `api/teacher/*` 或 `api/teaching/*` 的后端 HTTP path。
- 不在本轮改教师端页面结构或 feature owner。

## Source Inputs

- `TODO/frontend-sliced-architecture.md`
- `docs/architecture/frontend/02-routing.md`
- `code/frontend/src/utils/teacherLegacyRedirect.ts`
- `code/frontend/src/router/__tests__/sharedRoutes.test.ts`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`

## Task Slices

### Slice 1: 前端路径 producer guardrail

- 目标：在架构测试里扫描前端源码，禁止新的 `/teacher/*` 页面路径 literal 漏回活跃代码。
- 变更面：
  - `code/frontend/src/__tests__/architectureBoundaries.test.ts`
- 风险：
  - 如果匹配规则过宽，会误伤 `@/pages/teacher/*` 这类目录名；需要只盯 legacy route literal，而不是一般 `teacher` 词面。

### Slice 2: 同步当前事实文档

- 目标：把“仓库内部活跃入口已不再产出旧教师端前端路径，剩余只在 runtime redirect owner”写回迁移台账和路由文档。
- 变更面：
  - `TODO/frontend-sliced-architecture.md`
  - `docs/architecture/frontend/02-routing.md`
- 风险：
  - 文档若写成“已经删除 legacy redirect”会与当前实现不符。

## Validation Plan

- `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/router/__tests__/sharedRoutes.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- `bash scripts/check-consistency.sh`

## Review Focus

- guardrail 是否只针对旧前端页面路径 literal，而不误伤 teacher feature 目录或后端 API path。
- 当前事实是否准确表达“内部 producer 已清空，但 runtime redirect 仍存在”。
- 本轮是否保持为 runtime redirect 删除前的准备切片，而不是半删半留。

## Rollback / Recovery

- 如果源码扫描规则误伤现有正常路径，可先回退该 guardrail 测试，仅保留文档事实更新，再按更小范围重写匹配规则。
