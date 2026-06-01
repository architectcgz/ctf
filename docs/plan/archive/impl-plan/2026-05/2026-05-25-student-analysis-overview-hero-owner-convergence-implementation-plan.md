> 状态：Current
> 事实源：`StudentAnalysisPage.vue` 当前结构、`TD-1` 超大组件专题拆分、教师端 workspace hero 既有抽层模式
> 替代：无

# Student Analysis Overview Hero Owner Convergence Implementation Plan

## 目标

- 让 `StudentAnalysisPage.vue` 把 overview tab 顶部 hero 壳层收口为独立组件，只保留 tab / query / 面板装配 owner。
- 把当前内联在页面里的标题、四个快捷动作、三张 summary 卡和底部分隔线抽到单独的 overview hero panel。
- 保持 `TeacherStudentAnalysis.vue`、`useStudentAnalysisPage.ts` 与 `StudentInsightPanel.vue` 的 owner 不变，不改路由、API、query 或交互流程。

## 非目标

- 本轮不调整 `StudentInsightPanel.vue`、review workspace、writeup / evidence / timeline tab 的内部 owner。
- 本轮不引入新的 composable、store 或 feature 目录。
- 本轮不改 overview hero 的用户可见文案和动作语义。

## 输入依据

- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- `code/frontend/src/components/teacher/class-management/ClassStudentsPage.vue`
- `code/frontend/src/components/platform/class/ClassManageHeroPanel.vue`
- `code/frontend/src/components/platform/student/StudentManageHeroPanel.vue`
- `code/frontend/src/components/teacher/review-archive/ReviewArchiveHero.vue`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`

## 当前结论

- `StudentAnalysisPage.vue` 当前约 `756` 行，主要可安全继续拆分的稳定区块是 overview tab 顶部 hero 壳层。
- 这块 hero 只消费现成的 `selectedStudent`、`progress`、`solvedRate`、`weakDimensions` 和四个页面动作事件，没有独立的远端请求、路由同步或跨 tab 状态。
- 教师 / 平台端已有 `*HeroPanel.vue` 模式，说明当前最小收口路径是把 overview hero 抽成展示组件，而不是把 owner 再上推到路由页或扩散到新的 page model。

## 任务切片

### Slice 1：抽出 overview hero panel

- 目标：
  - 新建 `StudentAnalysisOverviewHeroPanel.vue`，承接标题、快捷动作、summary 指标卡和分隔线。
  - `StudentAnalysisPage.vue` 只保留 overview tab 的条件渲染与事件桥接，不再内联 hero 模板。
- 预期改动：
  - `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
  - `code/frontend/src/components/teacher/class-management/StudentAnalysisOverviewHeroPanel.vue`
- 验证：
  - `git diff --check -- <touched files>`
  - `npm run test:run -- src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- Review focus：
  - 页面是否继续只拥有 tab / query / 数据装配和事件桥接。
  - hero 的 props / emits 边界是否足够窄，不把路由或请求 owner 带进子组件。

### Slice 2：补源码护栏并复验

- 目标：
  - 更新 `TeacherStudentAnalysis.test.ts` 的源码边界断言，确保 `StudentAnalysisPage.vue` 通过 hero panel 装配 overview 顶栏，而不是继续内联。
  - 保持交互测试仍覆盖“导出班级报告 / 完整复盘页 / 返回学生列表”这些真实动作链路。
- 预期改动：
  - `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- 验证：
  - `npm run test:run -- src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/teacher/__tests__/studentAnalysisPanelExtraction.test.ts`
  - 如样式 token 或壳层类名发生变化，再补 `npm run check:theme-tail`
- Review focus：
  - 测试是否继续覆盖真实用户路径，而不是只验证组件名字符串。
  - 是否保持 overview hero 的可见文案、动作按钮和 summary 指标不回退。

## 风险

- `TeacherStudentAnalysis.test.ts` 现在既有源码断言，也有真实交互测试；如果只更新源码断言而不补交互路径，容易再次出现“源码收口了，但运行时挂载或行为不对”的假绿。
- overview hero 的样式 token 目前在 `StudentAnalysisPage.vue` 本地定义；抽出时如果忘了让变量从页面根继续向子组件透传，会造成视觉回归。
- 如果顺手把 summary 计算、动作 owner 或 tab 逻辑一起搬走，会把本轮从“稳定展示区抽层”扩大成 page owner 重构。

## 回退方式

- 如 hero 抽取引入回归，可回退 `StudentAnalysisOverviewHeroPanel.vue` 并恢复 `StudentAnalysisPage.vue` 内联 hero 模板。
- 因本轮不改 API、route、query 或 page model，回退只涉及前端组件层与测试文件，不涉及数据迁移。
