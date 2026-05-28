# Reuse Decision

## Change type
frontend api owner / shared transport wrapper alignment

## Existing code searched
- `code/frontend/src/api/teaching/instances.ts`
- `code/frontend/src/api/teacher/instances.ts`
- `code/frontend/src/api/admin/teaching.ts`
- `code/frontend/src/api/instances.ts`
- `code/frontend/src/api/__tests__/teacher.test.ts`
- `code/frontend/src/api/__tests__/admin.test.ts`
- `code/frontend/src/api/__tests__/instances.test.ts`
- `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
- `code/frontend/src/views/teacher/__tests__/InstanceManagement.test.ts`
- `docs/contracts/api-contract-v1.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `.harness/reuse-decisions/awd-review-api-implementation-owner-neutralization.md`
- `docs/plan/impl-plan/2026-05-28-awd-review-api-implementation-owner-neutralization-plan.md`
- `docs/reviews/frontend/2026-05-28-awd-review-api-implementation-owner-neutralization-review.md`

## Similar implementations found
- AWD review 已经完成同型收口：`api/teaching/*` 作为共享实现 owner，`api/admin/*` / `api/teacher/*` 只在 public owner 层保留角色语义。
- 当前实例目录 feature 已有 `api/instances.ts` 作为 role-aware facade，说明共享 workflow owner 已收回单点；剩余 teacher 语义残片主要停在共享实现函数名和 admin alias teacher function 这一层。

## Decision
refactor_existing

## Reason
- 本轮不新增后端 `/admin/instances*` 路径，也不修改 `/api/v1/teacher/instances*` 的 HTTP contract。
- 最小正确改动是把 `api/teaching/instances.ts` 的共享实现符号收口成中性 `InstanceDirectory*` / `ManagedInstance*` owner，再让 `api/teacher` / `api/admin` 在 public owner 层分别保留 teacher / platform 命名。
- 这样可以停止 platform public API 继续 alias teacher 命名函数，同时不影响现有 feature、DTO、HTTP path 和权限语义。

## Files to modify
- `.harness/reuse-decisions/instance-api-implementation-owner-neutralization.md`
- `docs/plan/impl-plan/2026-05-28-instance-api-implementation-owner-neutralization-plan.md`
- `docs/reviews/frontend/2026-05-28-instance-api-implementation-owner-neutralization-review.md`
- `code/frontend/src/api/teaching/instances.ts`
- `code/frontend/src/api/teacher/instances.ts`
- `code/frontend/src/api/admin/teaching.ts`
- `code/frontend/src/api/__tests__/teacher.test.ts`
- `code/frontend/src/api/__tests__/admin.test.ts`
- `docs/contracts/api-contract-v1.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## After implementation
- `api/teaching/instances.ts` 不再把 teacher 语义写死在共享实现符号里。
- `api/teacher/instances.ts` 继续保留 `getTeacherInstances()` / `destroyTeacherInstance()` 作为 teacher public owner。
- `api/admin/teaching.ts` 继续保留 `getPlatformInstances()` / `destroyPlatformInstance()` 作为 platform public owner，但不再 alias teacher 命名函数。
- 实例目录这条 admin / teacher 结构耦合在前端本地 API owner 维度会进一步缩到后端既有 teacher HTTP path 与 teacher public wrapper 命名。
