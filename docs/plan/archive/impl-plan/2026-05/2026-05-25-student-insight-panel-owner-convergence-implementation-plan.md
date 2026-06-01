> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`StudentInsightPanel.vue` 当前结构、`StudentAnalysisPage.vue` 当前单点挂载方式
> 替代：无

# Student Insight Panel Owner Convergence Implementation Plan

## 目标

- 让 `StudentInsightPanel.vue` 从“同时拥有多个内容区块与局部展示 owner 的大面板”收口成组合面板。
- 把当前仍留在父面板内的 overview / recommendations 区块抽到 `student-insight` 子组件目录，与已存在的 writeups / manual review / attack sessions section 形成一致分层。
- 保持 `StudentAnalysisPage.vue` 继续拥有 tab、页面级 summary、query 同步和数据加载；保持 `useStudentAnalysisPage.ts` 继续拥有远端请求与工作流动作。

## 非目标

- 本轮不改教师学生分析页的用户可见流程、tab 结构、query 参数或 API 契约。
- 本轮不处理 `StudentAnalysisPage.vue` 本身的 tab / summary owner。
- 本轮不引入新的 feature、store 或额外页面入口。

## 输入依据

- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `docs/reviews/frontend/README.md`
- `code/frontend/src/components/teacher/StudentInsightPanel.vue`
- `code/frontend/src/components/teacher/student-insight/*`
- `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/views/teacher/__tests__/studentAnalysisPanelExtraction.test.ts`
- `code/frontend/src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`

## 当前结论

- `StudentInsightPanel.vue` 目前约 `393` 行，虽然已经不属于最重型的前端债面，但仍同时承接：
  - 学员画像 overview 展示
  - 推荐任务目录展示
  - 作为 writeups / manual review / review workspace / timeline 的组合入口
  - 对应区块的局部样式 owner
- 同目录下已经存在 `StudentInsightWriteupsSection.vue`、`StudentInsightManualReviewSection.vue`、`StudentInsightAttackSessionsSection.vue`，说明当前最小收口路径是继续沿 section 拆分，而不是再造并行 owner。
- `StudentAnalysisPage.vue` 已经通过单个 `StudentInsightPanel` 挂载服务所有 tab，当前不能把区块重新复制回页面层，否则会破坏已有 extraction 约束。

## 任务切片

### Slice 1：抽出 overview section

- 目标：
  - 新建 `StudentInsightOverviewSection.vue`，承接六维雷达图和维度得分占比展示。
  - `StudentInsightPanel.vue` 只保留 `activeSection` 控制和区块组合，不再直接渲染 overview 模板。
- 预期改动：
  - `code/frontend/src/components/teacher/StudentInsightPanel.vue`
  - `code/frontend/src/components/teacher/student-insight/StudentInsightOverviewSection.vue`
- 验证：
  - `git diff --check -- <touched files>`
  - `npm run test:run -- src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- Review focus：
  - overview 的样式和展示 owner 是否真正下沉到 section，而不是只做模板搬运后仍把关键样式留在父面板。
  - 父面板是否继续只保留 section 可见性和公共事件桥接。

### Slice 2：抽出 recommendations section 并让父面板成为纯组合层

- 目标：
  - 新建 `StudentInsightRecommendationsSection.vue`，承接推荐任务目录与点击题目事件。
  - `StudentInsightPanel.vue` 最终只组合 section 组件、timeline 组件和局部可见性逻辑。
- 预期改动：
  - `code/frontend/src/components/teacher/StudentInsightPanel.vue`
  - `code/frontend/src/components/teacher/student-insight/StudentInsightRecommendationsSection.vue`
  - `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
  - `code/frontend/src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts`
- 依赖：
  - 继续复用 `ChallengeCategoryDifficultyPills`、`SectionCard`、`AppEmpty`、现有 workspace directory 样式类。
- 验证：
  - `git diff --check -- <touched files>`
  - `npm run test:run -- src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts src/views/teacher/__tests__/studentAnalysisPanelExtraction.test.ts`
- Review focus：
  - `StudentInsightPanel.vue` 是否收口成组合面板，不再继续直接拥有 recommendations 目录模板。
  - `StudentAnalysisPage.vue` 的单点挂载约束是否保持。

## 风险

- `teacherDetailSurfaceAlignment.test.ts` 目前直接读取 `StudentInsightPanel.vue` 源码做样式断言；section 抽取后，这些断言可能需要改成组合源码断言，否则会把已经下沉的 owner 错判为缺失。
- 推荐任务行样式复用的是 workspace directory 语义类，若 section 抽取时误把这些类名改丢，会造成视觉与源码断言双回归。
- `Timeline` 当前仍直接由父面板挂载；如果本轮抽取过程中顺手继续扩大范围，容易把“先让父面板退成组合层”变成无边界重构。

## 回退方式

- 如 section 抽取引入结构或样式回归，可回退新增的 `StudentInsightOverviewSection.vue` / `StudentInsightRecommendationsSection.vue` 并恢复 `StudentInsightPanel.vue` 内联模板。
- 因本轮不改 API、route、query 与页面 owner，回退只需恢复前端组件层，不涉及数据迁移。
