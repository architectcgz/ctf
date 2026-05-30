# Reuse Decision

## Change type
frontend architecture / feature ui / migration

## Existing code searched
- code/frontend/src/components/teacher/class-management/ClassManagementPage.vue
- code/frontend/src/components/teacher/class-management/ClassStudentsPage.vue
- code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue
- code/frontend/src/components/platform/awd-service/AWDChallengeLibraryPage.vue
- code/frontend/src/components/class-management/index.ts
- code/frontend/src/views/teacher/ClassManagement.vue
- code/frontend/src/views/teacher/TeacherClassStudents.vue
- code/frontend/src/views/teacher/TeacherStudentAnalysis.vue
- code/frontend/src/views/platform/AWDChallengeLibrary.vue
- code/frontend/src/views/platform/AWDChallengeImport.vue
- code/frontend/src/views/platform/PlatformClassStudents.vue
- code/frontend/src/views/platform/PlatformStudentAnalysis.vue
- code/frontend/src/views/teacher/__tests__/ClassManagement.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherClassStudents.test.ts
- code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts
- code/frontend/src/views/platform/__tests__/AWDChallengeLibrary.test.ts
- code/frontend/src/views/platform/__tests__/PlatformClassStudents.test.ts
- code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts
- code/frontend/src/features/teacher-class-management/index.ts
- code/frontend/src/features/teacher-class-management/model/useClassManagementPage.ts
- code/frontend/src/features/platform-awd-challenges/index.ts
- code/frontend/src/features/platform-awd-challenges/model/useAwdChallengeLibraryPage.ts
- code/frontend/src/features/platform-awd-challenges/model/useAwdChallengeImportPage.ts
- code/frontend/src/features/class-students-workspace/index.ts
- code/frontend/src/features/class-students-workspace/model/useClassStudentsPage.ts
- code/frontend/src/features/student-analysis-workspace/index.ts
- code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/components.d.ts
- docs/architecture/frontend/06-components.md
- docs/architecture/frontend/07-pages-dataflow.md
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `TeacherDashboardPage.vue`、`TeacherInstanceManagementPage.vue`、`StudentManagementPage.vue` 已证明，page-sized UI 可以迁到 `features/*/ui`，route view 保留组合壳。
- `ClassManagement.vue` 与 `TeacherStudentManagement.vue` 的结构几乎一致，都是 `feature model + page shell + report dialog` 的 route-level composition。
- `TeacherClassStudents.vue` / `PlatformClassStudents.vue` 与 `TeacherStudentAnalysis.vue` / `PlatformStudentAnalysis.vue` 已经共享中立 feature model；这次只需要把共享 page shell 从 `components/*` 落到对应 feature 的 `ui/`，不再经过 legacy barrel。
- `AWDChallengeLibrary.vue` / `AWDChallengeImport.vue` 共享 `AWDChallengeLibraryPage.vue`，且 page shell 已经拆成 `AwdChallengeWorkspaceHeader`、`AwdChallengeLibrarySection`、`AwdChallengeImportSection`，迁移成本主要在 public API 和 source-boundary 更新。

## Decision
refactor_existing

## Reason
这次不是新增页面能力，而是继续收口 `components/*Page.vue -> @/features/*/ui` 的剩余遗留例外。最小正确改动是把 4 个仍挂在 `components/` 下的 page shell 迁到各自 feature 的 `ui/`，让 route view 统一通过 feature public API 组合 page model 与 page shell，同时移除对应的 `legacyComponentPageAllowlist`，并清掉只为旧 page path 存在的 `components/class-management/index.ts` barrel。

## Files to modify
- .harness/reuse-decisions/remaining-legacy-page-feature-ui-migrations.md
- docs/plan/impl-plan/2026-05-27-remaining-legacy-page-feature-ui-migrations-implementation-plan.md
- docs/reviews/frontend/2026-05-27-remaining-legacy-page-feature-ui-migrations-review.md
- code/frontend/src/features/teacher-class-management/index.ts
- code/frontend/src/features/teacher-class-management/ui/*
- code/frontend/src/features/platform-awd-challenges/index.ts
- code/frontend/src/features/platform-awd-challenges/ui/*
- code/frontend/src/features/class-students-workspace/index.ts
- code/frontend/src/features/class-students-workspace/ui/*
- code/frontend/src/features/student-analysis-workspace/index.ts
- code/frontend/src/features/student-analysis-workspace/ui/*
- code/frontend/src/views/teacher/ClassManagement.vue
- code/frontend/src/views/platform/AWDChallengeLibrary.vue
- code/frontend/src/views/platform/AWDChallengeImport.vue
- code/frontend/src/views/platform/ChallengePackageFormat.vue
- code/frontend/src/views/platform/ChallengeImportManage.vue
- code/frontend/src/views/platform/ImageManage.vue
- code/frontend/src/views/platform/CheatDetection.vue
- code/frontend/src/views/platform/PlatformAwdReviewDetail.vue
- code/frontend/src/views/platform/__tests__/challengePackageFormatGuideExtraction.test.ts
- code/frontend/src/views/teacher/TeacherClassStudents.vue
- code/frontend/src/views/platform/PlatformClassStudents.vue
- code/frontend/src/views/teacher/TeacherStudentAnalysis.vue
- code/frontend/src/views/platform/PlatformStudentAnalysis.vue
- code/frontend/src/views/teacher/TeacherAWDReviewDetail.vue
- code/frontend/src/pages/platform/challenges/ChallengePackageFormatRoutePage.vue
- code/frontend/src/pages/platform/challenges/ChallengeImportManageRoutePage.vue
- code/frontend/src/pages/platform/ImageManageRoutePage.vue
- code/frontend/src/pages/platform/CheatDetectionRoutePage.vue
- code/frontend/src/pages/awd-review/PlatformAwdReviewDetailRoutePage.vue
- code/frontend/src/pages/awd-review/TeacherAwdReviewDetailRoutePage.vue
- code/frontend/src/router/routes/platformRoutes.ts
- code/frontend/src/router/routes/teacherRoutes.ts
- code/frontend/src/__tests__/architectureAllowlist.ts
- code/frontend/src/components.d.ts
- code/frontend/src/views/**/__tests__/*
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- docs/architecture/frontend/06-components.md
- docs/architecture/frontend/07-pages-dataflow.md

## After implementation
- 如果迁移顺利，`legacyComponentPageAllowlist` 会只剩学生 dashboard 那 5 个 page，teacher / platform 这批 page shell 会全部切到各自 feature 的 `model + ui` public API。
- 本轮不改 `useClassManagementPage.ts`、`useClassStudentsPage.ts`、`useStudentAnalysisPage.ts`、`useAwdChallengeLibraryPage.ts`、`useAwdChallengeImportPage.ts` 的 router / API / async owner，只收口 page shell 落位和 route import 边界。
- 后续同一条链路继续把 `ChallengePackageFormat`、`ChallengeImportManage`、`ImageManage`、`CheatDetection`、`TeacherAWDReviewDetail`、`PlatformAwdReviewDetail` 的运行时路由入口从 `views/**` 迁到 `pages/**`，并把 `views/*.vue` 降级成测试桥接壳。
