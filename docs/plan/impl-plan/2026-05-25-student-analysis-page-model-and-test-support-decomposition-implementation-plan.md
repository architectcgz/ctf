> 状态：Current
> 事实源：`useStudentAnalysisPage.ts` 当前 owner、teacher/platform 学员分析测试现状、2026-05-24/2026-05-25 已落地 student-analysis 相关实现计划与 review
> 替代：无

# Student Analysis Page Model And Test Support Decomposition Implementation Plan

## 目标

- 收口 `useStudentAnalysisPage.ts` 的 page-model 过宽问题，把最稳定、边界最清楚的职责拆成独立 model helper。
- 收口 teacher / platform 学员分析测试中的重复 mock、route state、fixture 和 dialog stub，降低双份测试面继续膨胀的风险。
- 保持用户可见行为、route owner、共享 page contract 和现有护栏语义不变。

## 非目标

- 本轮不重写 `StudentAnalysisPage.vue` 或 `StudentInsightPanel.vue` 的 UI owner。
- 本轮不改变 `useReviewWorkspace`、`useSubmissionReviewFlows`、`useReviewArchiveExportFlow` 这些既有 workflow 的对外 contract。
- 本轮不把学员分析整条页面模型拆成更多 feature 或 store。

## 输入依据

- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `docs/plan/impl-plan/2026-05-24-platform-route-owner-decoupling-implementation-plan.md`
- `docs/reviews/frontend/2026-05-25-student-analysis-page-prop-contract-convergence-review.md`
- `docs/plan/impl-plan/2026-05-25-platform-student-analysis-runtime-guardrail-tests-implementation-plan.md`

## 当前结论

- `useStudentAnalysisPage.ts` 当前同时拥有 route 解析、主数据加载、review workspace query 同步、writeup / manual review / report export workflow 编排、breadcrumb 和导航桥接，已经超出“页面装配 owner”的最小边界。
- 更早计划已经明确把这块过宽 owner 作为已知风险保留，而不是已完成收口；本轮需要在 touched surface 内继续拿掉这段结构债。
- `TeacherStudentAnalysis.test.ts` 与 `PlatformStudentAnalysis.test.ts` 目前各自维护一套几乎相同的 router mock、API mock、基础 fixture 和 `ClassReportExportDialog` stub；继续沿着这个结构迭代会放大维护成本。
- 最小安全方案不是重做测试体系，而是抽一个共享 test support，并把 route/page 独有断言继续留在各自 test file。

## 任务切片

### Slice 1：拆分学员分析 page model 的稳定职责

- 目标：
  - 从 `useStudentAnalysisPage.ts` 提取“主数据状态与加载”和“review workspace query 同步”两块独立 helper。
  - 让 `useStudentAnalysisPage.ts` 回到页面装配 owner：组装 workflow、绑定 watch、暴露页面 contract。
- 预期改动：
  - `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts`
  - `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisDataState.ts`
  - `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisReviewQuerySync.ts`
  - `code/frontend/src/features/student-analysis-workspace/model/index.ts`
- 验证：
  - `npm run test:run -- src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
  - 如新增 unit test，则一并跑新增 test file
- Review focus：
  - route/query 同步、error policy、loading owner 是否保持原语义。
  - 页面装配 owner 是否明显收窄，而不是只把长函数平移成另一个 catch-all helper。

### Slice 2：抽共享学员分析 route test support

- 目标：
  - 抽出 teacher / platform 学员分析测试共同使用的 route mock、API mock、基础 fixture、auth 初始化和 dialog stub。
  - 保留 teacher/platform 各自独有的 route owner 断言、交互路径和源码护栏。
- 预期改动：
  - `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
  - `code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
  - `code/frontend/src/views/__tests__/studentAnalysisRouteTestSupport.ts`
- 验证：
  - `npm run test:run -- src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
  - `git diff --check -- code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisDataState.ts code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisReviewQuerySync.ts code/frontend/src/features/student-analysis-workspace/model/index.ts code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts code/frontend/src/views/__tests__/studentAnalysisRouteTestSupport.ts docs/plan/impl-plan/2026-05-25-student-analysis-page-model-and-test-support-decomposition-implementation-plan.md .harness/reuse-decisions/student-analysis-page-model-and-test-support-decomposition.md`
- Review focus：
  - 共享 support 是否只承接重复 fixture，不侵入各自测试意图。
  - platform runtime guardrail 与 teacher 交互测试是否仍各自覆盖独有风险。

## 风险

- 如果把 query sync 和数据加载同时拆错 owner，容易把 `route.params` / `route.query` watch 责任弄散，产生隐藏回归。
- 如果共享 test support 过度抽象，会把 teacher/platform 独有断言语义压平，反而削弱测试可读性。
- `useStudentAnalysisPage.ts` 的 async flow 同时触达 review workspace、writeup、manual review、report export；本轮只能拆已经边界清楚的部分，不能为降行数硬拆剩余流程。

## 回退方式

- 如拆分引入回归，可逐文件回退新 helper 与 test support，并恢复 `useStudentAnalysisPage.ts` 和两份测试的原结构。
