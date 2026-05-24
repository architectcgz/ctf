# Reuse Decision

## Change type
hook / page / layout

## Existing code searched
- `code/frontend/src/views/platform`
- `code/frontend/src/features/platform-users`
- `code/frontend/src/features/teacher-class-management`
- `code/frontend/src/features/teacher-class-students`
- `code/frontend/src/features/teacher-instances`
- `code/frontend/src/features/student-directory`
- `code/frontend/src/components/platform`
- `code/frontend/src/composables`
- `harness/prompts`

## Similar implementations found
- `code/frontend/src/features/platform-users/model/usePlatformUsers.ts`
- `code/frontend/src/features/platform-users/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/features/platform-users/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/features/platform-users/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/teacher-class-management/model/useTeacherClassManagementPage.ts`
- `code/frontend/src/features/teacher-class-students/model/useTeacherClassStudentsPage.ts`
- `code/frontend/src/features/teacher-instances/model/useTeacherInstanceManagementPage.ts`
- `code/frontend/src/features/teacher-student-management/model/useTeacherStudentManagementPage.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisPage.ts`
- `code/frontend/src/features/student-directory/model/useStudentDirectoryQuery.ts`
- `code/frontend/src/features/student-directory/model/useStudentListQuery.ts`
- `harness/prompts/architecture-diagram-generation.md`
- `harness/prompts/coding-agent-system-prompt.md`

## Decision
refactor_existing

## Reason
这次不新增并行实现。`platform-users` 已经承担了用户治理之外的班级、学生目录和实例管理职责，适合把现有页面模型迁到更细的 feature owner 下，而不是继续往该桶里加代码。学生目录查询能力继续复用 `student-directory` 里的既有 query owner；平台侧的 `usePlatformStudentManagementPage` 只保留页面级筛选、分页和路由跳转编排，不再新造一条与 `useTeacherStudentManagementPage`、`useTeacherStudentAnalysisPage`、`useStudentListQuery` 平行的学生查询栈。review 提示词也放到现有 `harness/prompts/` 目录，复用已有项目 prompt 资产形态，不新建额外 prompt 目录。

## Files to modify
- `harness/prompts/AGENTS.md`
- `harness/prompts/frontend-architecture-review.md`
- `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- `code/frontend/src/views/platform/ClassManage.vue`
- `code/frontend/src/views/platform/StudentManage.vue`
- `code/frontend/src/views/platform/InstanceManage.vue`
- `code/frontend/src/views/platform/__tests__/ClassManage.test.ts`
- `code/frontend/src/views/platform/__tests__/StudentManage.test.ts`
- `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/features/platform-users/index.ts`
- `code/frontend/src/features/platform-users/model/index.ts`
- `code/frontend/src/features/platform-users/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/features/platform-users/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/features/platform-users/model/usePlatformInstanceManagementPage.ts`
- `code/frontend/src/features/platform-class-management/index.ts`
- `code/frontend/src/features/platform-class-management/model/index.ts`
- `code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts`
- `code/frontend/src/features/platform-student-management/index.ts`
- `code/frontend/src/features/platform-student-management/model/index.ts`
- `code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts`
- `code/frontend/src/features/platform-instance-management/index.ts`
- `code/frontend/src/features/platform-instance-management/model/index.ts`
- `code/frontend/src/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts`

## After implementation
- 如果后续 review 流程继续稳定复用，再考虑补 `feedback/` 记录 prompt 资产化经验；本次先以 `harness/prompts/` 作为唯一事实源。
