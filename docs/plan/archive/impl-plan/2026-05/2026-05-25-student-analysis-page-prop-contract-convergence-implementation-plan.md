> 状态：Current
> 事实源：`StudentAnalysisPage.vue` 当前模板、teacher/platform route view 调用面、`useStudentAnalysisPage.ts` 返回契约
> 替代：无

# Student Analysis Page Prop Contract Convergence Implementation Plan

## 目标

- 删除 `StudentAnalysisPage.vue` 中已经没有消费方的 dead props。
- 删除 teacher / platform 两个 route view 对这些 props 的冗余传参与对 dead emits 的无效监听。
- 让 `useStudentAnalysisPage.ts` 和 `useStudentAnalysisNavigation.ts` 只暴露当前页面真实使用的导航动作，保持 page owner 更窄更清楚。
- 去掉 `useStudentAnalysisPage.ts` 中已经无主的 class list 初始化依赖，避免无关接口继续阻断学员分析页加载。

## 非目标

- 本轮不改 `StudentInsightPanel.vue`、overview hero、review workspace、writeup / evidence / timeline tab 的内部 owner。
- 本轮不调整 `useStudentAnalysisPage.ts` 的数据加载流程、路由同步或错误处理。
- 本轮不继续做 feature public API 的更大范围清理。

## 输入依据

- `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
- `code/frontend/src/views/platform/PlatformStudentAnalysis.vue`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisNavigation.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`

## 当前结论

- `StudentAnalysisPage.vue` 当前已不再拥有 class switch、student directory 或返回班级管理的入口，因此不应继续声明对应 props / emits。
- route view 继续传递无效 props，本质上是在把 page owner 之外的历史上下文强塞给共享页面壳，增加 review 噪音和误导性契约。
- `useStudentAnalysisPage.ts` 当前只有 `openClassStudents`、`openChallenge`、`openReviewArchivePage` 仍被 route view 通过页面壳间接使用；其余导航桥接可以在本轮一起删掉。
- `getClasses()` 和 `loadingClasses` 继续留在 `initialize()` 里，会让已经退出页面 contract 的班级列表接口仍然决定整个学员分析页能否加载；这条依赖要在本轮一起拿掉。

## 任务切片

### Slice 1：收紧 StudentAnalysisPage 组件契约

- 目标：
  - 从 `StudentAnalysisPage.vue` 删除未消费 props：`classes`、`students`、`selectedClassName`、`selectedStudentId`、`loadingClasses`、`loadingStudents`。
  - 删除未发出的 emits：`openClassManagement`、`selectClass`、`selectStudent`。
- 预期改动：
  - `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- 验证：
  - `git diff --check -- <touched files>`
  - `npm run test:run -- src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
- Review focus：
  - 页面壳是否只保留当前真实存在的页面级动作和数据 contract。
  - 是否误删仍被 template 或子组件事件桥接消费的字段。

### Slice 2：同步 route view 和 page model

- 目标：
  - 从 `TeacherStudentAnalysis.vue`、`PlatformStudentAnalysis.vue` 删除冗余传参与 dead emits 监听。
  - 从 `useStudentAnalysisPage.ts` / `useStudentAnalysisNavigation.ts` 删除 route view 已不再需要的导航桥接。
  - 从 `useStudentAnalysisPage.ts` 的初始化链路里移除无主的 class list 依赖。
- 预期改动：
  - `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
  - `code/frontend/src/views/platform/PlatformStudentAnalysis.vue`
  - `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts`
  - `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisNavigation.ts`
- 验证：
  - 同上
- Review focus：
  - route view 是否仍只负责共享 page workflow 装配和导出弹窗 owner。
  - page model 是否没有误丢 `openClassStudents` / `openReviewArchivePage` 这些仍在用的动作。
  - 学员分析页是否不再被 `getClasses()` 失败牵连进全局错误态。

### Slice 3：更新源码护栏与行为断言

- 目标：
  - 更新 `TeacherStudentAnalysis.test.ts` / `PlatformStudentAnalysis.test.ts` 的源码断言，只保留当前真实 contract。
  - 保持真实交互测试仍覆盖导出、返回学生列表、复盘归档等仍可发生的用户路径，并移除对 dead emit 的伪契约测试。
  - 增加失败路径测试，证明班级列表接口不会再阻断学员分析页加载。
- 预期改动：
  - `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
  - `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
- 验证：
  - `npm run test:run -- src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts src/views/teacher/__tests__/studentAnalysisPanelExtraction.test.ts src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
  - `bash scripts/check-consistency.sh`
- Review focus：
  - 测试是否覆盖真实 owner 收口，而不是继续为页面不会发出的事件保留假护栏。
  - 平台 / 教师 route view 是否仍通过同一组中性 page workflow 消费共享页面。

## 风险

- 如果把仍被 `ClassReportExportDialog` 或 review archive 导航链路依赖的字段一并删掉，会造成运行时回归。
- `TeacherStudentAnalysis.test.ts` 当前混合了源码断言和交互断言，如果只改字符串不复验行为，容易留下“契约文本变窄了，但真实页面桥接断了”的假绿。
- `useStudentAnalysisNavigation.ts` 是 feature 公共导出的一部分，删符号时要确认没有其它消费者依赖。

## 回退方式

- 如 contract 收紧引入回归，可逐文件回退 `StudentAnalysisPage.vue`、两个 route view 和相关测试。
- 因本轮不改 API、路由定义或数据结构，回退只涉及前端组件 / composable / 测试层。
