# teacher views 最终清场计划

> 状态：In Progress
> 事实输入：`code/frontend/src/router/routes/teacherRoutes.ts`、`code/frontend/src/pages/teacher/**`、`code/frontend/src/pages/awd-review/**`、`code/frontend/src/views/teacher/**`

## Plan Summary

- Objective
  - 删除 `code/frontend/src/views/teacher/*.vue` 的旧运行时页面残片，让 `src/views/teacher` 只保留 `__tests__/`。
- Non-goals
  - 不迁移 `views/teacher/__tests__` 目录。
  - 不修改历史 review / plan 中的旧路径记录。
  - 不顺带搬迁 `views/__tests__` 到其他目录。
- Source architecture or design docs
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `code/frontend/src/router/routes/teacherRoutes.ts`
- Dependency order
  - 先改唯一活测试引用，再删旧页面，再更新 backlog，最后验证。
- Expected specialist skills
  - `frontend-engineer`
  - `development-pipeline`

## Task 1

- Goal
  - 确认 `views/teacher/*.vue` 是否已经退出运行时与源码直接引用。
- Touched modules or boundaries
  - `router/routes/teacherRoutes.ts`
  - `views/teacher/*.vue`
  - `views/teacher/__tests__`
- Dependencies
  - 无
- Validation
  - `rg` 检查 `@/views/teacher/`、`src/views/teacher/`、`../*.vue` 相对导入命中。
- Review focus
  - `TeacherStudentManagement` 与 `StudentManagementRoutePage`、`TeacherAWDReviewDetail` 与 `pages/awd-review/**` 的非同名映射是否已经稳定。
- Risk notes
  - 如果遗漏动态 import 或相对 import，删除后会直接打断 Vitest。

## Task 2

- Goal
  - 把仍直接 import 旧 teacher view 的测试切到当前 route page。
- Touched modules or boundaries
  - `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
  - `code/frontend/src/pages/awd-review/TeacherAwdReviewDetailRoutePage.vue`
- Dependencies
  - Task 1
- Validation
  - 相关测试通过，旧 `TeacherAWDReviewDetail.vue` 不再被源码引用。
- Review focus
  - 测试断言语义不变，只替换当前事实入口。
- Risk notes
  - route page 与旧 view 如存在轻微壳层差异，要确认测试仍对准最终行为。

## Task 3

- Goal
  - 删除 `views/teacher/*.vue`，并同步 backlog 当前结论。
- Touched modules or boundaries
  - `code/frontend/src/views/teacher/*.vue`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Dependencies
  - Task 2
- Validation
  - `find code/frontend/src/views/teacher -maxdepth 1 -name '*.vue'` 为空。
  - `rg` 确认源码不再引用这批旧页面。
- Review focus
  - 删除只覆盖旧运行时页，不误删测试支撑。
- Risk notes
  - 删除操作统一走回收站，避免误删无法恢复。

## Integration Checks

- `teacherRoutes.ts` 仍全部指向 `pages/**`
- `views/teacher` 只剩 `__tests__/`
- `src/views` 不再保留任何运行时 `.vue` 页面文件

## Rollback / Recovery Notes

- 若发现误删，可从 Git 恢复单个旧 view 文件。
- 删除统一使用回收站方式执行。

## Residual Risks

- 历史 plan / review 文档仍会出现旧 `views/teacher/*.vue` 路径，这是保留的时间点记录，不在本轮收口范围内。
