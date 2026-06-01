> 状态：Current
> 事实源：`api/contracts.ts`、writeups API client、student analysis workflow、fronted backlog
> 替代：无

# Manual Review Submission Contract Naming Neutralization Implementation Plan

## 目标

- 把已经跨 teacher / platform 共用的 `TeacherManualReviewSubmissionItemData`、`TeacherManualReviewSubmissionDetailData` 收口成中性命名。
- 保持 teacher 与 platform 学员分析里的人工评阅列表、详情与提交评阅行为不变，只收 contract 命名语义。
- 继续按最小可审阅切片消化更深层 `Teacher*` contract 命名债。

## 非目标

- 本轮不改 `TeacherSubmissionWriteupDetailData`。
- 本轮不改 `TeacherClassItem`、`TeacherAWDReviewContestItemData`、`TeacherAttackSessionQuery`。
- 本轮不改 teacher / platform 学员分析页面的 route owner、API path 或权限边界。
- 本轮不调整人工评阅接口的 request / response schema。

## 输入依据

- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/teacher/writeups.ts`
- `code/frontend/src/api/teaching/writeups.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useSubmissionReviewFlows.ts`
- `code/frontend/src/components/teacher/StudentInsightPanel.vue`
- `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightWriteupsSection.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightManualReviewSection.vue`
- `code/frontend/src/components/teacher/student-insight/studentInsightShared.ts`
- `code/frontend/src/views/platform/PlatformStudentAnalysis.vue`
- `code/frontend/src/views/teacher/TeacherStudentAnalysis.vue`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- manual review DTO 现在通过共享 `student-analysis-workspace` 同时服务 teacher 与 platform 两边的学员分析页，已经不是教师专属数据模型。
- 继续只从 API owner 层兜底，无法消除消费面代码里“platform 页面仍在拿 teacher DTO”这一层语义噪音。
- 最小安全切片是只改 manual review item/detail 这组 DTO 名称，并同步 API client、shared workflow、teacher 组件与事实文档。

## 任务切片

### Slice 1：收口共享 manual review contract 名称

- 目标：
  - 在 `api/contracts.ts` 提供中性 `ManualReviewSubmissionItemData`、`ManualReviewSubmissionDetailData`。
  - 同步 `api/teacher/writeups.ts` 与 `api/teaching/writeups.ts` 返回类型和 raw normalize 类型名。
- 预期改动：
  - `code/frontend/src/api/contracts.ts`
  - `code/frontend/src/api/teacher/writeups.ts`
  - `code/frontend/src/api/teaching/writeups.ts`
- review focus：
  - API client normalize 逻辑不变
  - 不误伤 writeup submission item / detail DTO

### Slice 2：同步 student analysis workflow 与组件引用

- 目标：
  - 让共享 `student-analysis-workspace` 以及 teacher 组件消费面切到中性 manual review DTO。
- 预期改动：
  - `code/frontend/src/features/teacher-student-analysis/model/useSubmissionReviewFlows.ts`
  - `code/frontend/src/components/teacher/StudentInsightPanel.vue`
  - `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
  - `code/frontend/src/components/teacher/student-insight/StudentInsightWriteupsSection.vue`
  - `code/frontend/src/components/teacher/student-insight/StudentInsightManualReviewSection.vue`
  - `code/frontend/src/components/teacher/student-insight/studentInsightShared.ts`
- review focus：
  - teacher / platform 学员分析里的人工评阅列表、详情打开与评阅提交行为不回归

### Slice 3：同步事实文档与 review 证据

- 目标：
  - 在 contract / backlog / review 里记录这轮 manual review contract 命名收口进展。
- 预期改动：
  - `docs/contracts/api-contract-v1.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-27-manual-review-submission-contract-naming-neutralization-review.md`

## 验证

- `npm run test:run -- src/views/platform/__tests__/PlatformStudentAnalysis.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `npm run typecheck`
- `bash scripts/check-workflow-complete.sh`

## 回退 / 恢复说明

- 本轮回退按 manual review DTO 命名切：整体撤销 `ManualReviewSubmissionItemData`、`ManualReviewSubmissionDetailData` 及其引用更新即可，不涉及运行时接口路径或权限回退。

## 残余风险

- `TeacherSubmissionWriteupDetailData` 仍保留 teacher 前缀，后续仍需独立切片。
- `TeacherClassItem`、`TeacherAWDReviewContestItemData`、`TeacherAttackSessionQuery` 等其他共享 contract 命名残余还在。
