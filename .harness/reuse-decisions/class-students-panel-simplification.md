# Reuse Decision

## Change type
component / page / feature

## Existing code searched
- code/frontend/src/components/teacher/class-management/ClassStudentsPage.vue
- code/frontend/src/components/teacher/student-management/StudentManagementPage.vue
- code/frontend/src/assets/styles/teacher-surface.css
- code/frontend/src/views/teacher/TeacherClassStudents.vue
- code/frontend/src/features/teacher-class-students/model/useTeacherClassStudentsPage.ts
- code/frontend/src/views/teacher/__tests__/TeacherClassStudents.test.ts

## Similar implementations found
- `StudentManagementPage.vue` 已经把学生目录列头和数据行都收口到统一的列表结构里，说明学生目录不需要在班级详情页再维护另一套头部控制区。
- `teacher-surface.css` 已经为 `.teacher-directory-head` 提供了基于 `--teacher-directory-columns` 的统一列模板，班级学生页应该沿用这层对齐，而不是只给数据行单独写 `grid-template-columns`。
- `TeacherClassStudents.vue` 和 `useTeacherClassStudentsPage.ts` 当前只是为了班级切换把 `classes` / `selectClass` 往下传，这次去掉切换入口后应同步收缩父子契约。

## Decision
refactor_existing

## Reason
这次不是新增页面能力，而是把具体班级页里的重复导航和无效班级切换去掉，并让学生目录重新对齐既有列表列模板。沿用现有教师目录样式和父子边界比继续保留一套只在当前页面存在的局部结构更稳。

## Files to modify
- .harness/reuse-decisions/class-students-panel-simplification.md
- code/frontend/src/components/teacher/class-management/ClassStudentsPage.vue
- code/frontend/src/views/teacher/TeacherClassStudents.vue
- code/frontend/src/features/teacher-class-students/model/useTeacherClassStudentsPage.ts
- code/frontend/src/views/teacher/__tests__/TeacherClassStudents.test.ts
