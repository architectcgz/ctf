# Reuse Decision

## Change type
frontend refactor / teacher route page migration residue cleanup

## Existing code searched
- code/frontend/src/views/teacher/*.vue
- code/frontend/src/views/teacher/__tests__/*.test.ts
- code/frontend/src/pages/teacher/**/*.vue
- code/frontend/src/pages/awd-review/TeacherAwdReviewDetailRoutePage.vue
- code/frontend/src/router/routes/teacherRoutes.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md

## Similar implementations found
- `views/platform/*.vue` 已在上一轮完成最终清场：先切测试引用，再删除旧页面文件，只保留 `views/platform/__tests__/`。
- `teacherRoutes.ts` 当前已统一指向 `pages/teacher/**` 与 `pages/awd-review/**`，说明 teacher 侧运行时入口也已迁完，剩余是旧 `views/teacher/*.vue` 残片。

## Decision
refactor_existing

## Reason
这轮与 platform 清场同类，不新增能力，只删除已经退出运行时的旧 teacher page 残片：

- 把仍直接 import `../TeacherAWDReviewDetail.vue` 的测试切到 `pages/awd-review/TeacherAwdReviewDetailRoutePage.vue`
- 删除 `views/teacher/*.vue`
- 同步 backlog 当前事实，让 `views/` 只剩测试目录

不做：

- 不迁移 `views/teacher/__tests__` 目录本身
- 不改历史 plan / review 文档里的旧路径记录
- 不继续扩展到 `views/**/__tests__` 的统一搬迁

## Files to modify
- .harness/reuse-decisions/teacher-views-final-cleanup.md
- docs/plan/impl-plan/2026-05-30-teacher-views-final-cleanup-plan.md
- code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts
- docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md
- code/frontend/src/views/teacher/*.vue

## After implementation
- `src/views/teacher` 将只剩 `__tests__/`
- `src/views` 将不再保留任何运行时 `.vue` 页面文件
- 当前事实文档会明确 `views/` 仅保留邻近测试支撑
