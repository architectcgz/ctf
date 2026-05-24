# Reuse Decision

## Change type
api / hook / page

## Existing code searched
- `code/frontend/src/api/teacher.ts`
- `code/frontend/src/api/teacher`
- `code/frontend/src/api/teacher/students.ts`
- `code/frontend/src/features/platform-class-management`
- `code/frontend/src/features/platform-student-management`
- `code/frontend/src/features/platform-instance-management`
- `code/frontend/src/views/platform/__tests__`

## Similar implementations found
- `code/frontend/src/api/teacher/classes.ts`
- `code/frontend/src/api/teacher/students.ts`
- `code/frontend/src/api/teacher/instances.ts`
- `code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`

## Decision
refactor_existing

## Reason
这次不新增后台专用并行接口，也不改后端路由。现有 platform 班级、学生、实例管理已经在页面 owner 上拆开，但 query helper 仍然直接挂在 `@/api/teacher` 命名空间下。继续扩 `code/frontend/src/api/teacher/classes.ts` 或 `code/frontend/src/api/teacher/students.ts` 虽然能复用现有实现，但会把 platform 侧新的共享查询继续沉积在 teacher namespace 里，和这轮“去 teacher owner 耦合”的目标相反。更低风险的收口方式是把这些已经被 admin/teacher 共用的查询实现迁到中性的 `@/api/teaching` owner，再让 `@/api/teacher` 退成兼容 re-export。这样既能保留现有请求路径和 DTO 形状，也能让 platform feature 不再显式依赖 teacher namespace。

## Files to modify
- `code/frontend/src/api/teaching.ts`
- `code/frontend/src/api/teaching/index.ts`
- `code/frontend/src/api/teaching/classes.ts`
- `code/frontend/src/api/teaching/instances.ts`
- `code/frontend/src/api/teacher/classes.ts`
- `code/frontend/src/api/teacher/instances.ts`
- `code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/views/platform/__tests__/ClassManage.test.ts`
- `code/frontend/src/views/platform/__tests__/StudentManage.test.ts`
- `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
- `docs/plan/impl-plan/2026-05-24-admin-teaching-query-owner-decoupling-implementation-plan.md`

## After implementation
- 如果后续还要继续把 class/student review、AWD review detail 等共享 workflow 从 `@/api/teacher` 收口到中性 owner，再沿用 `@/api/teaching` 扩展，不再新增新的 role-specific wrapper。
