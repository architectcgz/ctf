# Class Directory Contract Naming Neutralization 收口复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-27-class-directory-contract-naming-neutralization-implementation-plan.md`
  - files reviewed：
    - `code/frontend/src/api/contracts.ts`
    - `code/frontend/src/api/teaching/classes.ts`
    - `code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts`
    - `code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts`
    - `code/frontend/src/features/teacher-class-management/model/useClassManagementPage.ts`
    - `code/frontend/src/features/teacher-student-management/model/useStudentManagementPage.ts`
    - `code/frontend/src/features/teacher-instances/model/useInstances.ts`
    - `code/frontend/src/features/teacher-workspace/model/useWorkspace.ts`
    - `code/frontend/src/features/admin-notification-publisher/model/useAdminNotificationPublisher.ts`
    - `code/frontend/src/components/teacher/class-management/ClassManagementPage.vue`
    - `code/frontend/src/components/teacher/student-management/StudentManagementPage.vue`
    - `code/frontend/src/components/teacher/instance-management/TeacherInstanceManagementPage.vue`
    - `code/frontend/src/components/teacher/instance-management/TeacherInstanceDirectorySection.vue`
    - `code/frontend/src/components/platform/student/StudentManageWorkspacePanel.vue`
    - `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
    - `docs/contracts/api-contract-v1.md`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：同意本轮属于前端 P1 级结构收口，范围限定在共享班级目录项 contract 的中性化命名。
- Gate verdict：Pass

## Findings

- 无 material finding。

## Material findings

- 无。

## Senior implementation assessment

- `TeacherClassItem` 已同时服务 teacher、platform 和 admin 通知发布等共享目录流，本轮把它收口成 `ClassDirectoryItem`，比继续保留角色前缀更符合实际 owner 语义。
- 这刀没有把 `TeacherInstanceItem`、`TeacherStudentItem` 或班级洞察 DTO 混进来，提交边界合理，blast radius 主要集中在类型引用而不是行为面。
- `getClasses()` 的 path 和 admin wrapper 保持不变，只收 contract 命名，不会把本轮误扩成 API owner 迁移。

## Required re-validation

- `npm run test:run -- src/api/__tests__/teacher.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `npm run typecheck`
- `bash scripts/check-workflow-complete.sh`
- `git diff --check -- code/frontend/src/api/contracts.ts code/frontend/src/api/teaching/classes.ts code/frontend/src/features/platform-class-management/model/usePlatformClassManagementPage.ts code/frontend/src/features/platform-student-management/model/usePlatformStudentManagementPage.ts code/frontend/src/features/teacher-class-management/model/useClassManagementPage.ts code/frontend/src/features/teacher-student-management/model/useStudentManagementPage.ts code/frontend/src/features/teacher-instances/model/useInstances.ts code/frontend/src/features/teacher-workspace/model/useWorkspace.ts code/frontend/src/features/admin-notification-publisher/model/useAdminNotificationPublisher.ts code/frontend/src/components/teacher/class-management/ClassManagementPage.vue code/frontend/src/components/teacher/student-management/StudentManagementPage.vue code/frontend/src/components/teacher/instance-management/TeacherInstanceManagementPage.vue code/frontend/src/components/teacher/instance-management/TeacherInstanceDirectorySection.vue code/frontend/src/components/platform/student/StudentManageWorkspacePanel.vue code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts docs/contracts/api-contract-v1.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md .harness/reuse-decisions/class-directory-contract-naming-neutralization.md docs/plan/impl-plan/2026-05-27-class-directory-contract-naming-neutralization-implementation-plan.md docs/reviews/frontend/2026-05-27-class-directory-contract-naming-neutralization-review.md`

## Residual risk

- `TeacherAWDReviewContestItemData`、`TeacherAttackSessionQuery` 等其他共享 contract 命名残余还在，需要后续独立切片。
- 本轮 review 在同一主会话里完成，没有独立子 agent 复核；因此这里只作为同上下文 gate review 证据，不表述为独立审查。

## Touched known-debt status

- 本轮 touched 的已知结构债是共享班级目录 contract 仍保留 `TeacherClassItem` 前缀。
- 在本轮 touched surface 上，这条债务已经完成当前阶段收口；剩余 teacher 前缀共享 contract 已进一步收敛到 AWD review 和 attack session query 两组。
