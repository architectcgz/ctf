# Reuse Decision

## Change type
route / test

## Existing code searched
- `code/frontend/src/router/routes/teacherRoutes.ts`
- `code/frontend/src/router/__tests__/sharedRoutes.test.ts`
- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardPage.vue`
- `code/frontend/src/features/teacher/student-management/ui/StudentManagementPage.vue`
- `frontend sliced architecture migration ledger`

## Similar implementations found
- `code/frontend/src/router/routes/platformRoutes.ts`
- `code/frontend/src/router/__tests__/sharedRoutes.test.ts`

## Decision
refactor_existing

## Reason
教师端 runtime 页面已经统一到 `/academy/*`，但 `teacherRoutes.ts` 里仍保留多条 `/teacher/* -> /academy/*` 兼容 redirect。当前更合适的最小切片不是直接删行为，而是先把遗留兼容入口收口成明确 allowlist，并补源码级 guardrail，避免后续继续新增旧命名空间。

## Files to modify
- `code/frontend/src/router/routes/teacherRoutes.ts`
- `code/frontend/src/router/__tests__/sharedRoutes.test.ts`
- `frontend sliced architecture migration ledger`
- `docs/plan/impl-plan/2026-05-31-teacher-legacy-route-allowlist-guardrail-plan.md`
- `.harness/reuse-decisions/teacher-legacy-route-allowlist-guardrail.md`

## After implementation
- 教师端遗留 `/teacher/*` redirect 有单点 allowlist owner。
- 路由测试会限制只允许当前白名单里的 legacy redirect 存在。
- 台账会把“先整理 allowlist”从口头建议收成已执行事实。
