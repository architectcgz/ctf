> 状态：Current
> 事实源：student analysis page owner、review query helper、前端架构 allowlist、TeacherStudentAnalysis 测试
> 替代：无

# Student Analysis Router Helper Cleanup Plan

## 目标

- 把 `useStudentAnalysisNavigation.ts` 和 `useStudentAnalysisReviewQuerySync.ts` 从 route-aware helper 收口回纯 page 下游 helper。
- 让 `useStudentAnalysisPage.ts` 保留唯一 route/query owner。
- 删除对应两条 `featureRouterImportAllowlist`。

## 非目标

- 不重构 `useStudentAnalysisPage.ts` 的整体加载流程。
- 不改 student analysis 的 review workspace、writeup review 或 report export 业务语义。
- 不处理 `featureRouterImportAllowlist` 其它 feature 条目。

## 输入依据

- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisNavigation.ts`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisReviewQuerySync.ts`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts`
- `code/frontend/src/features/student-analysis-workspace/model/__tests__/useStudentAnalysisReviewQuerySync.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`

## 当前结论

- `useStudentAnalysisNavigation.ts` 只是把若干页面动作组装成导航 helper，不应该继续直接依赖 `Router` 类型。
- `useStudentAnalysisReviewQuerySync.ts` 只是 route query 与 review workspace state 的同步 helper，应该消费本地 `route-like` / `replaceQuery` contract，而不是 `vue-router` 类型。
- `useStudentAnalysisPage.ts` 已经持有 `useRoute()` 与 `useRouter()`，继续作为唯一 route-aware owner 更合理。

## 设计边界

### `useStudentAnalysisPage.ts` 本轮负责

- `useRoute()` / `useRouter()` 获取
- `openClassStudents` / `openChallenge` / `openReviewArchivePage` 的具体导航动作
- `reviewMode` / `reviewResult` / `reviewChallengeId` 的 query 写回

### `useStudentAnalysisNavigation.ts` 本轮负责

- 导航 helper 组合
- 调用外部注入的 `openClassStudentsRoute` / `openChallengeRoute` / `openReviewArchiveRoute`

### `useStudentAnalysisReviewQuerySync.ts` 本轮负责

- 从 `route.query` 解析 review workspace query
- 对 session query 做 state compare / sync
- 调用外部注入的 `replaceReviewWorkspaceQuery`

## 任务切片

### Slice 1：navigation helper 去掉 router 类型依赖

- 目标：
  - 从 `useStudentAnalysisNavigation.ts` 移除 `vue-router` 类型 import
  - 改为 callback 注入导航动作
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- Review focus：
  - helper 是否已经不再 import `vue-router`
  - 班级页、题目详情、复盘归档导航行为是否保持不变

### Slice 2：review query helper 去掉 router 类型依赖

- 目标：
  - 从 `useStudentAnalysisReviewQuerySync.ts` 移除 `vue-router` 类型 import
  - 改为本地 route-like / replaceQuery contract
- 验证：
  - `cd code/frontend && npm run test:run -- src/features/student-analysis-workspace/model/__tests__/useStudentAnalysisReviewQuerySync.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- Review focus：
  - helper 是否已经不再 import `vue-router`
  - query 同步和 challenge filter reload 语义是否保持不变

### Slice 3：allowlist / 护栏 / backlog 收尾

- 目标：
  - 删除两条 allowlist
  - 更新 raw-source 护栏与 backlog 记录
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/features/student-analysis-workspace/model/__tests__/useStudentAnalysisReviewQuerySync.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - route/query owner 是否只回到 page，而没有漂到别的 helper
  - allowlist 是否真实下降

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/features/student-analysis-workspace/model/__tests__/useStudentAnalysisReviewQuerySync.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check -- .harness/reuse-decisions/student-analysis-router-helper-cleanup.md docs/plan/impl-plan/2026-05-29-student-analysis-router-helper-cleanup-plan.md docs/reviews/frontend/2026-05-29-student-analysis-router-helper-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/__tests__/architectureAllowlist.ts code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisNavigation.ts code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisReviewQuerySync.ts code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisPage.ts code/frontend/src/features/student-analysis-workspace/model/__tests__/useStudentAnalysisReviewQuerySync.test.ts code/frontend/src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮只收口两条 `featureRouterImportAllowlist`，不代表剩余条目都不合理；仍需逐条判定。
- 这轮 review 默认仍是同上下文 self-review；独立 reviewer gate 仍需单独说明。
