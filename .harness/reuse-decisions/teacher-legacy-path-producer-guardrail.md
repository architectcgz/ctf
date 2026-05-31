# Reuse Decision

## Change type
frontend refactor / route compatibility guardrail

## Existing code searched
- `code/frontend/src/features/teacher/**`
- `code/frontend/src/shared/**`
- `code/frontend/src/pages/teacher/**`
- `code/frontend/src/utils/teacherLegacyRedirect.ts`
- `code/frontend/src/router/routes/teacherRoutes.ts`
- `code/frontend/src/router/__tests__/sharedRoutes.test.ts`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`
- `code/frontend/src/config/backofficeNavigation.ts`
- `code/frontend/src/utils/roleRoutes.ts`
- `code/frontend/src/shared/model/layout/useWorkspaceShellNavigation.ts`
- frontend sliced architecture migration ledger
- `docs/architecture/frontend/02-routing.md`

## Similar implementations found
- `sharedRoutes.test.ts` 已经把 `/teacher/*` runtime redirect 收口成 exact allowlist，说明 legacy 教师端页面入口已经有单点 route owner。
- `architectureBoundaries.test.ts` 已经承担源码级 guardrail，适合继续机械化“仓库内部不再产出旧前端路径”这条结构约束，而不是再散落到页面或 feature 层手工排查。

## Decision
refactor_existing

## Reason
当前扫描结果显示，活跃前端页面路径里已经统一转向 `/academy/*`；剩余 `/teacher/*` 页面路径语义基本只应存在于 `teacherLegacyRedirect.ts` 这个 runtime 兼容 owner。

在直接删除 runtime redirect 之前，最小正确改动是：

- 先用源码级 guardrail 锁住新的 `/teacher/*` 页面路径 producer
- 明确只有 legacy redirect owner 可以保留这些旧路径
- 把当前事实同步回迁移台账和路由事实文档

本轮不做：

- 不删除现有 `/teacher/*` redirect
- 不修改教师端后端 API `/teacher/*`
- 不调整 `/academy/*` canonical route

## Files to modify
- `.harness/reuse-decisions/teacher-legacy-path-producer-guardrail.md`
- `docs/plan/impl-plan/2026-05-31-teacher-legacy-path-producer-guardrail-plan.md`
- `code/frontend/src/__tests__/architectureBoundaries.test.ts`
- frontend sliced architecture migration ledger (`frontend-sliced-architecture.md`)
- `docs/architecture/frontend/02-routing.md`

## After implementation
- 前端仓库内部新增 `/teacher/*` 页面路径 producer 会被源码级测试直接拦住。
- legacy 教师端前端路径只剩 runtime 兼容 owner 保留。
- 后续删除 runtime redirect 时，不需要先再做一次内部入口清点。
