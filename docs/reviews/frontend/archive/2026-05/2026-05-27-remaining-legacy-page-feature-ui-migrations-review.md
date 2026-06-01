# Remaining Legacy Page Feature UI Migrations 收口复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-27-remaining-legacy-page-feature-ui-migrations-implementation-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/remaining-legacy-page-feature-ui-migrations.md`
    - `docs/plan/impl-plan/2026-05-27-remaining-legacy-page-feature-ui-migrations-implementation-plan.md`
    - `docs/reviews/frontend/2026-05-27-remaining-legacy-page-feature-ui-migrations-review.md`
    - `code/frontend/src/features/teacher-class-management/**/*`
    - `code/frontend/src/features/platform-awd-challenges/**/*`
    - `code/frontend/src/features/class-students-workspace/**/*`
    - `code/frontend/src/features/student-analysis-workspace/**/*`
    - `code/frontend/src/views/teacher/ClassManagement.vue`
    - `code/frontend/src/views/platform/AWDChallengeLibrary.vue`
    - `code/frontend/src/views/platform/AWDChallengeImport.vue`
    - `code/frontend/src/views/teacher/TeacherClassStudents.vue`
    - `code/frontend/src/views/platform/PlatformClassStudents.vue`
    - `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
    - `code/frontend/src/views/platform/PlatformStudentAnalysis.vue`
    - `code/frontend/src/__tests__/architectureAllowlist.ts`
    - `code/frontend/src/components.d.ts`
    - `code/frontend/src/views/**/__tests__/*`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：预期属于 allowlist 驱动的前端结构债收口，这次只处理剩余 legacy page shell 落位，不引入新的 route / API owner。
- Gate verdict：Self-check pass，独立 reviewer gate 待补

## Findings

- 未发现阻塞性问题。

## Material findings

- None.

## Senior implementation assessment

- `ClassManagementPage.vue`、`AWDChallengeLibraryPage.vue`、`ClassStudentsPage.vue`、`StudentAnalysisPage.vue` 已分别迁入各自 feature 的 `ui/` 目录，对应 route view 均改为通过 feature public API 组合 page shell 与 page model。
- `TeacherClassStudents.vue` / `PlatformClassStudents.vue` 与 `TeacherStudentAnalysis.vue` / `PlatformStudentAnalysis.vue` 不再经过 `components/class-management/index.ts` 这个 legacy barrel，teacher / platform 双 route 共享 page shell 的边界已经收口到 feature 公共入口。
- `AWDChallengeLibrary.vue` / `AWDChallengeImport.vue` 现在都通过 `features/platform-awd-challenges` 获取 page shell，`AWDChallengeLibraryPage.vue` 对应的一条 `componentFeatureImportAllowlist` 与一条 `legacyComponentPageAllowlist` 已同时收掉。
- `legacyComponentPageAllowlist` 已从 teacher / platform 这批 page shell 中清空，当前只剩 student dashboard 的 5 个页面。
- 当前记录只覆盖实现者 self-check；如果要完全满足 development-pipeline 的独立 review gate，还需要在用户显式授权 delegation 后补一轮 reviewer agent 复核。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/ClassManagement.test.ts src/views/platform/__tests__/AWDChallengeLibrary.test.ts src/views/teacher/__tests__/TeacherClassStudents.test.ts src/views/platform/__tests__/PlatformClassStudents.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/platform/__tests__/PlatformStudentAnalysis.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/sharedPaginationControls.test.ts src/views/__tests__/studentDirectoryTypographyBoundary.test.ts src/views/__tests__/journalEyebrowStyles.test.ts src/views/__tests__/pageTabsStyles.test.ts src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts src/views/teacher/__tests__/teacherWorkspaceSubpanelAdoption.test.ts src/views/teacher/__tests__/classManagementTabsAdoption.test.ts src/views/teacher/__tests__/classStudentsPanelExtraction.test.ts src/views/teacher/__tests__/studentAnalysisPanelExtraction.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- `StudentAnalysisPage` 的原始源码断言面最大，最容易出现漏改路径或断言文案。

## Touched known-debt status

- 这轮 touched surface 上的 4 个 legacy page shell 与 1 个 legacy barrel 已完成收口。
