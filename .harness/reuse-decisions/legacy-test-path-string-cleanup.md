# Reuse Decision

## Change type
test / cleanup

## Existing code searched
- `code/frontend/src/pages/dashboard/__tests__/DashboardView.test.ts`
- `code/frontend/src/pages/teacher/__tests__/*`
- `code/frontend/src/pages/platform/__tests__/*`
- `code/frontend/src/pages/platform/contests/__tests__/*`
- `code/frontend/src/pages/platform/awd-challenges/__tests__/*`
- `code/frontend/src/features/challenge-topology-studio/**/*.test.ts`
- `code/frontend/src/features/contest-projector/**/*.test.ts`

## Similar implementations found
- 当前前端架构边界已经由 `src/__tests__/architectureBoundaries.test.ts`、`src/__tests__/routePageArchitectureBoundary.test.ts` 和 `scripts/check-frontend-architecture.sh` 承担。
- 这批命中的断言多数只是在检查历史 `components/*` / `pages/*` 路径字符串没有再次出现，不再提供新的运行时或架构语义。

## Decision
refactor_existing

## Reason
- 本次不新增新的迁移态测试，也不保留“旧路径字符串不能出现”的历史断言。
- 保留仍然表达 router owner、API owner、feature owner、页面职责边界的测试；删除纯迁移残痕，避免测试继续绑死历史目录结构。

## Files to modify
- `.harness/reuse-decisions/legacy-test-path-string-cleanup.md`
- `code/frontend/src/pages/dashboard/__tests__/DashboardView.test.ts`
- `code/frontend/src/pages/teacher/__tests__/ClassManagement.test.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentManagement.test.ts`
- `code/frontend/src/pages/teacher/__tests__/InstanceManagement.test.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherClassStudents.test.ts`
- `code/frontend/src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/pages/platform/contests/__tests__/ContestAwdConfig.test.ts`
- `code/frontend/src/pages/platform/contests/__tests__/ContestProjector.test.ts`
- `code/frontend/src/pages/platform/awd-challenges/__tests__/AWDChallengeLibrary.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `code/frontend/src/pages/platform/__tests__/PlatformClassStudents.test.ts`
- `code/frontend/src/features/challenge-topology-studio/ui/ChallengeTopologyStudioPage.test.ts`
- `code/frontend/src/features/challenge-topology-studio/model/useChallengeTopologyStudioBoundary.test.ts`
- `code/frontend/src/features/contest-projector/model/useContestProjectorBoundary.test.ts`

## After implementation
- 页面 / feature 测试只保留当前仍有业务或架构语义的断言。
- 历史目录迁移是否回退，交给统一架构守卫检查，不再由零散字符串断言重复承担。
