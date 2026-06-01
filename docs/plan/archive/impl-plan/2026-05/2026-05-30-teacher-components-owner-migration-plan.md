# teacher 组件 owner 迁移计划

> 状态：Current
> 事实源：当前前端 feature/page 引用与 `components/teacher/*` consumer 扫描
> 替代：无

## 目标

把当前 teacher 历史业务组件目录拆回最终 feature owner：

- 教师独占能力进入 `features/teacher/**`
- teacher / platform 共用能力进入 `features/teaching/**`

结束 teacher 页面层继续跨目录依赖旧业务组件目录、以及共享能力继续挂 `teacher-*` 命名的状态。

## 非目标

- 本轮不处理 `components/teacher/awd-review/*`
- 本轮不处理 `components/teacher/review-archive/*`
- 本轮不顺手迁 `components/contests/*`、`components/platform/*`

## 方案依据

- `code/frontend/src/features/teacher/dashboard/ui/TeacherDashboardPage.vue`
- `code/frontend/src/features/teacher/instances/ui/TeacherInstanceManagementPage.vue`
- `code/frontend/src/features/teacher/class-management/ui/ClassManagementPage.vue`
- `code/frontend/src/features/teaching/class-students-workspace/ui/ClassStudentsPage.vue`
- `code/frontend/src/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue`
- `code/frontend/src/features/teaching/class-report-export/ui/ClassReportExportPreviewSection.vue`
- `code/frontend/src/pages/teacher/__tests__/*`
- `code/frontend/src/pages/__tests__/workspacePageHeaderStyles.test.ts`

## 当前边界

- `features/teacher/dashboard` 拥有 dashboard 分页、画像、趋势、复盘、介入建议面板，但这些面板文件仍放在 `components/teacher/dashboard/*`
- `features/teacher/instances` 拥有教师实例目录页，但 hero / directory section 仍放在 `components/teacher/instance-management/*`
- `features/teacher/class-management`、`features/teaching/class-students-workspace`、`features/teaching/student-analysis-workspace`、`features/teaching/class-report-export` 共同消费 `components/teacher/class-management/*`、`ClassInsightsPanel.vue`、`ClassReviewPanel.vue`、`ClassTrendPanel.vue`、`StudentInsightPanel.vue` 与 `components/teacher/student-insight/*`

## 任务切片

### Slice 1：教师独占能力收进 `features/teacher/**`

- 迁移：
  - `components/teacher/dashboard/*` -> `features/teacher/dashboard/ui/*`
  - `components/teacher/instance-management/*` -> `features/teacher/instances/ui/*`
- 更新：
  - feature imports
  - 相关 `pages/**/__tests__` raw-source imports

### Slice 2：共享教学能力收进 `features/teaching/**`

- 迁移：
  - `components/teacher/class-management/*` -> `features/teacher/class-management/ui/*`、`features/teaching/class-students-workspace/ui/*`、`features/teaching/student-analysis-workspace/ui/*`
  - `components/teacher/ClassInsightsPanel.vue`、`ClassReviewPanel.vue`、`ClassTrendPanel.vue` -> `features/teaching/class-students-workspace/ui/*`
  - `components/teacher/StudentInsightPanel.vue`
  - `components/teacher/student-insight/*` -> `features/teaching/student-analysis-workspace/ui/*`
- 更新：
  - `teaching/class-report-export` 对班级洞察面板的引用
  - teacher / platform / shared 页面测试 raw-source imports

## 预期改动文件

- `code/frontend/src/features/teacher/**`
- `code/frontend/src/features/teaching/**`
- `code/frontend/src/pages/**/__tests__/*`
- `code/frontend/src/components.d.ts`

## 验证

- `cd code/frontend && npm run test:run -- src/pages/teacher/__tests__/TeacherDashboard.test.ts src/pages/teacher/__tests__/InstanceManagement.test.ts src/pages/teacher/__tests__/ClassManagement.test.ts src/pages/teacher/__tests__/TeacherClassStudents.test.ts src/pages/teacher/__tests__/TeacherStudentAnalysis.test.ts src/pages/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts src/pages/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts src/pages/teacher/__tests__/teacherEyebrowSharedStyles.test.ts src/pages/teacher/__tests__/teacherSharedDirectoryStyles.test.ts src/pages/teacher/__tests__/teacherSurface.test.ts`
- `cd code/frontend && npm run test:run -- src/pages/platform/__tests__/PlatformClassStudents.test.ts src/pages/platform/__tests__/PlatformStudentAnalysis.test.ts src/pages/__tests__/workspacePageHeaderStyles.test.ts src/pages/__tests__/workspaceShellStyles.test.ts src/pages/__tests__/studentDirectoryTypographyBoundary.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-frontend-architecture.sh --quick`

## Review 关注点

- feature owner 是否清晰，是否还有 feature/page 直接依赖旧 `components/teacher/*`
- `ClassInsightsPanel` / `ClassReviewPanel` / `ClassTrendPanel` / `StudentInsightPanel` 的落点是否仍保持 teacher-neutral 的共享边界
- `platform` route page 是否不再引用 `teacher-*` feature 命名
- 迁移后测试是否只验证当前 owner，而不是守旧路径名称

## 回退

- 如果某一组 teacher 组件迁移导致 raw-source 守卫与 feature owner 冲突，回退该组 imports 到上一提交状态，再重新拆分 owner 范围，不在同一批里保留 bridge 壳
