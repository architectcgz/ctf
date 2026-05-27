> 状态：Current
> 事实源：`components/teacher/reports`、teacher / platform class & student analysis route views、frontend backlog
> 替代：无

# Class Report Export Dialog Public Owner Neutralization Implementation Plan

## 目标

- 为 `ClassReportExportDialog` 增加中立 public owner `@/components/reports`。
- 让 teacher / platform 的班级管理、班级学员页和学员分析页都通过中立入口引用共享导出对话框。

## 非目标

- 本轮不移动 `ClassReportExportDialog.vue` 物理文件位置。
- 本轮不改 `useClassReportExport`、导出参数、导出按钮行为或对话框内容。
- 本轮不处理 review archive、AWD review 等其他 teacher 组件 public owner。

## 输入依据

- `code/frontend/src/components/teacher/reports/index.ts`
- `code/frontend/src/components/teacher/reports/ClassReportExportDialog.vue`
- `code/frontend/src/views/teacher/ClassManagement.vue`
- `code/frontend/src/views/teacher/TeacherStudentManagement.vue`
- `code/frontend/src/views/teacher/TeacherClassStudents.vue`
- `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
- `code/frontend/src/views/platform/PlatformClassStudents.vue`
- `code/frontend/src/views/platform/PlatformStudentAnalysis.vue`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `ClassReportExportDialog` 已经是 teacher / platform 共用的共享对话框，但 public import 还停留在 `components/teacher/reports`，会继续放大 admin / teacher 结构耦合的表象。
- 这条债不需要动 feature owner 或实际组件位置，只要新增中立 public owner 并切 import，即可让共享语义和目录 owner 对齐。

## 任务切片

### Slice 1：提供中立 reports public owner

- 目标：
  - 新增 `components/reports/index.ts`，统一对外暴露 `ClassReportExportDialog`。
- 预期改动：
  - `code/frontend/src/components/reports/index.ts`
- review focus：
  - 不引入新的实际组件副本或重复实现

### Slice 2：迁移 teacher / platform 路由页 import

- 目标：
  - 让 teacher / platform 相关 route view 都改从 `@/components/reports` 引入共享导出对话框。
- 预期改动：
  - `code/frontend/src/views/teacher/ClassManagement.vue`
  - `code/frontend/src/views/teacher/TeacherStudentManagement.vue`
  - `code/frontend/src/views/teacher/TeacherClassStudents.vue`
  - `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
  - `code/frontend/src/views/platform/PlatformClassStudents.vue`
  - `code/frontend/src/views/platform/PlatformStudentAnalysis.vue`
- review focus：
  - route view 结构、对话框 props 和 `v-model` 行为不变

### Slice 3：同步源码断言测试与 backlog 记录

- 目标：
  - 更新相关源码断言测试，并在 backlog 里记录这条 public owner 收口进展。
- 预期改动：
  - `code/frontend/src/views/teacher/__tests__/ClassManagement.test.ts`
  - `code/frontend/src/views/teacher/__tests__/TeacherStudentManagement.test.ts`
  - `code/frontend/src/views/teacher/__tests__/TeacherClassStudents.test.ts`
  - `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
  - `code/frontend/src/views/platform/__tests__/PlatformClassStudents.test.ts`
  - `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-27-class-report-export-dialog-public-owner-neutralization-review.md`

## 验证

- `npm run test:run -- src/views/teacher/__tests__/ClassManagement.test.ts src/views/teacher/__tests__/TeacherStudentManagement.test.ts src/views/teacher/__tests__/TeacherClassStudents.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/platform/__tests__/PlatformClassStudents.test.ts src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `npm run typecheck`
- `bash scripts/check-workflow-complete.sh`

## 回退 / 恢复说明

- 本轮回退只需恢复 route view import 到 `@/components/teacher/reports`，并移除 `components/reports/index.ts`。

## 残余风险

- 这刀只收共享班级报告导出对话框的 public owner；review archive、AWD review 等 shared widget 仍可能继续依赖 teacher 组件入口，需要后续独立切片。
