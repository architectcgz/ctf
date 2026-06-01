> 状态：Current
> 事实源：`api/contracts.ts`、writeups API client、teacher / platform writeup workflow、fronted backlog
> 替代：无

# Writeup Submission Contract Naming Neutralization Implementation Plan

## 目标

- 把已经跨 teacher / platform 共用的 `TeacherSubmissionWriteupItemData` 收口成中性命名 `WriteupSubmissionItemData`。
- 保持 teacher 与 platform 现有题解查看 / 管理行为不变，只收 contract 命名语义。
- 用一个单 DTO 切片开始消化更深层 `Teacher*` contract 命名债。

## 非目标

- 本轮不改 `TeacherSubmissionWriteupDetailData`。
- 本轮不改 `TeacherManualReviewSubmissionItemData` / `TeacherManualReviewSubmissionDetailData`。
- 本轮不改 `TeacherClassItem`、`TeacherAWDReviewContestItemData`、`TeacherAttackSessionQuery` 等其他共享 contract。
- 本轮不调整 HTTP path、response schema 或 teacher / platform 题解功能边界。

## 输入依据

- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/teacher/writeups.ts`
- `code/frontend/src/api/teaching/writeups.ts`
- `code/frontend/src/features/challenge-writeup-editor/model/useChallengeWriteupManagement.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useSubmissionReviewFlows.ts`
- `code/frontend/src/components/teacher/StudentInsightPanel.vue`
- `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightWriteupsSection.vue`
- `code/frontend/src/components/teacher/student-insight/studentInsightShared.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `TeacherSubmissionWriteupItemData` 现在同时服务于 platform 题解管理和 teacher 学生分析，已经不是教师专属 DTO。
- 继续只靠 API owner wrapper 兜底，无法解决读代码时“平台功能看起来还在借 teacher 模型”的认知噪音。
- 最小安全切片是只改这一组 item DTO 名称，并同步所有直接引用；manual review 与 detail DTO 留到后续独立切片。

## 任务切片

### Slice 1：收口共享 writeup submission item contract 名称

- 目标：
  - 在 `api/contracts.ts` 提供中性 `WriteupSubmissionItemData`，移除直接消费面上的 `TeacherSubmissionWriteupItemData`。
- 预期改动：
  - `code/frontend/src/api/contracts.ts`
  - `code/frontend/src/api/teacher/writeups.ts`
  - `code/frontend/src/api/teaching/writeups.ts`
- review focus：
  - API client 返回类型和 normalize 逻辑不变
  - 不误伤 manual review / writeup detail 相关类型

### Slice 2：同步 teacher / platform writeup workflow 与组件引用

- 目标：
  - 让 platform 题解管理与 teacher 学生分析的 writeup submission item 引用都切到中性 DTO。
- 预期改动：
  - `code/frontend/src/features/challenge-writeup-editor/model/useChallengeWriteupManagement.ts`
  - `code/frontend/src/features/teacher-student-analysis/model/useSubmissionReviewFlows.ts`
  - `code/frontend/src/components/teacher/StudentInsightPanel.vue`
  - `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
  - `code/frontend/src/components/teacher/student-insight/StudentInsightWriteupsSection.vue`
  - `code/frontend/src/components/teacher/student-insight/studentInsightShared.ts`
- review focus：
  - teacher / platform 题解列表、分页、状态标签和人工评阅关联行为不回归

### Slice 3：同步事实文档与 review 证据

- 目标：
  - 在 contract / backlog / review 里记录这轮“更深层 Teacher* contract 命名”收口进展。
- 预期改动：
  - `docs/contracts/api-contract-v1.md`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `docs/reviews/frontend/2026-05-27-writeup-submission-contract-naming-neutralization-review.md`

## 验证

- `npm run test:run -- src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts src/views/platform/__tests__/PlatformStudentAnalysis.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `npm run typecheck`
- `bash scripts/check-workflow-complete.sh`

## 回退 / 恢复说明

- 本轮回退按 DTO 命名切：整体撤销 `WriteupSubmissionItemData` 重命名和对应引用更新即可，不涉及运行时行为或接口路径回退。

## 残余风险

- `TeacherSubmissionWriteupDetailData` 和 manual review 两个 DTO 仍保留 teacher 前缀，本轮不覆盖。
- `TeacherClassItem`、`TeacherAWDReviewContestItemData`、`TeacherAttackSessionQuery` 等其他共享 contract 还需要后续分 slice 处理。
