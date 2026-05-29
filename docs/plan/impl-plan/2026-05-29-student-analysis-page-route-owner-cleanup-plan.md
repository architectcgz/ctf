> 状态：Current
> 事实源：student analysis page model、navigation helper、review query helper、前端架构 allowlist
> 替代：无

# Student Analysis Page Route Owner Cleanup Plan

## 目标

- 收掉 `features/student-analysis-workspace/model/useStudentAnalysisPage.ts -> vue-router`

## 非目标

- 不重构 student analysis 的 review workspace、写题解审核和报表导出 workflow。
- 不改 `useStudentAnalysisDataState.ts`、`useReviewWorkspace()` 或 `useSubmissionReviewFlows()` 的数据 owner。
- 不把 review query schema 或 breadcrumb policy 下沉到 shared transport。

## 输入依据

- `useStudentAnalysisPage.ts`
- `useStudentAnalysisNavigation.ts`
- `useStudentAnalysisReviewQuerySync.ts`
- `TeacherStudentAnalysis.test.ts`
- `PlatformStudentAnalysis.test.ts`
- `routeQueryTransport.ts`
- `routeNavigationTransport.ts`
- `architectureAllowlist.ts`

## 当前结论

- `student-analysis` 这条的路由职责主要分成两类：review workspace query 同步，以及班级学生页 / 题目详情 / 复盘归档 3 条薄导航。
- shared transports 已经足够承接 route `params / query` 读取与 `push / replace`；本轮不需要再造 student-analysis-specific route wrapper。
- 更适合新增的是本地 `studentAnalysisRoutes.ts`，统一表达三条薄导航目标。

## 设计边界

### student analysis page model 本轮负责

- 保留 review workspace query owner、student analysis 数据加载、breadcrumb、report export 和 submission review owner
- 保留 mounted / unmounted lifecycle 与 query watch
- 不再直接 import `vue-router`

### shared transports 本轮负责

- 提供 route `params / query` 读取
- 提供 `push()` / `replace()` transport
- 不承接 review workspace query schema、student analysis breadcrumb 或 role-aware page 逻辑

### student analysis routes 本轮负责

- 统一描述班级学生页、题目详情和复盘归档 route target
- 不承接 review query 写回、生命周期或数据加载

## 任务切片

- [ ] Slice 1：page model 改用 shared route transports
  - 目标：
    - `useStudentAnalysisPage.ts` 去掉 `vue-router`
    - params/query 读取和 review query replace 改为消费共享 transport
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`

- [ ] Slice 2：抽本地 route target helper
  - 目标：
    - 新增 `studentAnalysisRoutes.ts`
    - 班级学生页 / 题目详情 / 复盘归档改为 route target + navigation transport
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`

- [ ] Slice 3：allowlist / review / backlog 收尾
  - 目标：
    - 更新 allowlist、raw-source 护栏、todo、review
  - 验证：
    - `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/platform/__tests__/PlatformStudentAnalysis.test.ts src/__tests__/architectureBoundaries.test.ts`
    - `cd code/frontend && npm run typecheck`

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/platform/__tests__/PlatformStudentAnalysis.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/student-analysis-page-route-owner-cleanup.md docs/plan/impl-plan/2026-05-29-student-analysis-page-route-owner-cleanup-plan.md docs/reviews/frontend/2026-05-29-student-analysis-page-route-owner-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/student-analysis-workspace/model/studentAnalysisRoutes.ts code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisNavigation.ts code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts code/frontend/src/views/platform/__tests__/PlatformStudentAnalysis.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 本轮只收 transport owner，不重做 student analysis 的 review workspace state machine；如果后续 page model 继续增长，再另开切片拆更深层 workflow。
- 这轮 review 默认仍是同上下文 self-review；独立 reviewer gate 仍需单独说明。
