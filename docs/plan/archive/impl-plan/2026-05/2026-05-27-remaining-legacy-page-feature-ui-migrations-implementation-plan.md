> 状态：Current
> 事实源：剩余 legacy `components/*Page.vue` 到 feature `ui/` 的迁移边界
> 替代：无

# Remaining Legacy Page Feature UI Migrations Implementation Plan

## 目标

- 把剩余 4 个 legacy page shell 从 `components/` 迁到对应 feature 的 `ui/`：
  - `ClassManagementPage.vue`
  - `AWDChallengeLibraryPage.vue`
  - `ClassStudentsPage.vue`
  - `StudentAnalysisPage.vue`
- 让对应 route view 统一通过 feature public API 组合 page model 与 page shell。
- 收掉这些页面对应的 `legacyComponentPageAllowlist`，并移除只为旧 page path 存在的 `components/class-management/index.ts`。

## 非目标

- 本轮不改 page model 的 router / API / lifecycle / dialog owner。
- 本轮不重排 page shell 内已经拆好的子组件层级。
- 本轮不改用户可见文案、筛选规则、分页行为、review / writeup / 导入提交流程。

## 输入依据

- `docs/architecture/frontend/06-components.md`
- `docs/architecture/frontend/07-pages-dataflow.md`
- `code/frontend/src/components/teacher/class-management/ClassManagementPage.vue`
- `code/frontend/src/components/teacher/class-management/ClassStudentsPage.vue`
- `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- `code/frontend/src/components/platform/awd-service/AWDChallengeLibraryPage.vue`
- `code/frontend/src/views/teacher/ClassManagement.vue`
- `code/frontend/src/views/teacher/TeacherClassStudents.vue`
- `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
- `code/frontend/src/views/platform/AWDChallengeLibrary.vue`
- `code/frontend/src/views/platform/AWDChallengeImport.vue`
- `code/frontend/src/views/platform/PlatformClassStudents.vue`
- `code/frontend/src/views/platform/PlatformStudentAnalysis.vue`
- `code/frontend/src/features/**/model/*`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `ClassManagement.vue` 已经是标准薄 route 壳，最适合先迁。
- `AWDChallengeLibrary.vue` / `AWDChallengeImport.vue` 共用同一 page shell，但 shell 本体已经很薄，迁移主要是 public API 切换。
- `TeacherClassStudents.vue` / `PlatformClassStudents.vue` 与 `TeacherStudentAnalysis.vue` / `PlatformStudentAnalysis.vue` 已经共享 feature model；当前残留结构债只在 page shell 仍停留在 `components/` 目录。
- `components/class-management/index.ts` 当前只作为 `ClassStudentsPage` / `StudentAnalysisPage` 的 legacy barrel，迁移完成后应一并退场。

## 设计边界

### route view 继续负责

- 组合 feature public API 导出的 page shell 与 page model
- 组合 `ClassReportExportDialog`
- 不直接持有 router / API / lifecycle owner

### feature model 继续负责

- `teacher-class-management/model`：班级目录加载、分页、跳转、导出弹窗状态
- `platform-awd-challenges/model`：题库列表、导入队列、编辑对话框、删除、跳转、上传
- `class-students-workspace/model`：路由参数、insight window query、班级详情和学生列表
- `student-analysis-workspace/model`：路由参数、review workspace、writeup / manual review、导出归档、轮询

### feature ui 本轮负责

- page-sized shell 落位
- 消费上层派生状态和事件 handler
- 不直接持有 API、router 或对话框 owner

## 任务切片

### Slice 1：迁移 `ClassManagementPage` 到 `features/teacher-class-management/ui`

- 目标：
  - 新增 `features/teacher-class-management/ui/ClassManagementPage.vue`
  - route view 改从 `@/features/teacher/class-management` 取 page shell
- 验证：
  - `npm run test:run -- src/views/teacher/__tests__/ClassManagement.test.ts`
- Review focus：
  - route view 是否继续只保留组合壳

### Slice 2：迁移 `AWDChallengeLibraryPage` 到 `features/platform-awd-challenges/ui`

- 目标：
  - 新增 `features/platform-awd-challenges/ui/AWDChallengeLibraryPage.vue`
  - `AWDChallengeLibrary.vue` / `AWDChallengeImport.vue` 均改从 feature public API 取 page shell
  - 收掉 `componentFeatureImportAllowlist` 中对应条目
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/AWDChallengeLibrary.test.ts`
- Review focus：
  - 同一 page shell 在两个 route view 下都改走 feature public API

### Slice 3：迁移 `ClassStudentsPage` 到 `features/class-students-workspace/ui`

- 目标：
  - 新增 `features/class-students-workspace/ui/ClassStudentsPage.vue`
  - `TeacherClassStudents.vue` / `PlatformClassStudents.vue` 改从 feature public API 取 page shell
  - 清掉 `components/class-management/index.ts` 对它的 legacy 依赖
- 验证：
  - `npm run test:run -- src/views/teacher/__tests__/TeacherClassStudents.test.ts src/views/platform/__tests__/PlatformClassStudents.test.ts`
- Review focus：
  - teacher / platform 双 route 共享 page shell 是否仍只经过 feature public API

### Slice 4：迁移 `StudentAnalysisPage` 到 `features/student-analysis-workspace/ui`

- 目标：
  - 新增 `features/student-analysis-workspace/ui/StudentAnalysisPage.vue`
  - `TeacherStudentAnalysis.vue` / `PlatformStudentAnalysis.vue` 改从 feature public API 取 page shell
  - 清掉 `components/class-management/index.ts` 剩余用途并退场
- 验证：
  - `npm run test:run -- src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
- Review focus：
  - 宽 page contract 是否只是 UI 落位迁移，没有把 model owner 反向吸回 page shell

### Slice 5：统一护栏、原始源码测试和 backlog

- 目标：
  - 更新 allowlist、`components.d.ts`、raw-source 测试路径
  - 删除 4 个旧 page 文件和失效 barrel
  - 记录 backlog 进展
- 验证：
  - `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/ClassManagement.test.ts src/views/platform/__tests__/AWDChallengeLibrary.test.ts src/views/teacher/__tests__/TeacherClassStudents.test.ts src/views/platform/__tests__/PlatformClassStudents.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/platform/__tests__/PlatformStudentAnalysis.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/sharedPaginationControls.test.ts src/views/__tests__/studentDirectoryTypographyBoundary.test.ts src/views/__tests__/journalEyebrowStyles.test.ts src/views/__tests__/pageTabsStyles.test.ts src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts src/views/teacher/__tests__/teacherWorkspaceSubpanelAdoption.test.ts src/views/teacher/__tests__/classManagementTabsAdoption.test.ts src/views/teacher/__tests__/classStudentsPanelExtraction.test.ts src/views/teacher/__tests__/studentAnalysisPanelExtraction.test.ts`
- Review focus：
  - allowlist 是否真实下降到只剩 student dashboard 页

## 结构收口检查

- 4 个 legacy page shell 都不再作为 `components/*Page.vue` 存在。
- route view 统一只保留 feature public API 组合壳。
- `components/class-management/index.ts` 不再作为跨层共享入口存在。
- `legacyComponentPageAllowlist` 只剩 student dashboard 5 个页面。

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/teacher/__tests__/ClassManagement.test.ts src/views/platform/__tests__/AWDChallengeLibrary.test.ts src/views/teacher/__tests__/TeacherClassStudents.test.ts src/views/platform/__tests__/PlatformClassStudents.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/platform/__tests__/PlatformStudentAnalysis.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/sharedPaginationControls.test.ts src/views/__tests__/studentDirectoryTypographyBoundary.test.ts src/views/__tests__/journalEyebrowStyles.test.ts src/views/__tests__/pageTabsStyles.test.ts src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts src/views/teacher/__tests__/teacherWorkspaceSubpanelAdoption.test.ts src/views/teacher/__tests__/classManagementTabsAdoption.test.ts src/views/teacher/__tests__/classStudentsPanelExtraction.test.ts src/views/teacher/__tests__/studentAnalysisPanelExtraction.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `StudentAnalysisPage` 的测试覆盖面最大，这一 slice 最容易因为 raw-source 断言遗漏而返工。
- `StudentManagementPage` 的未提交迁移改动与本轮并存；提交时需要继续按任务边界分清文件集合。
