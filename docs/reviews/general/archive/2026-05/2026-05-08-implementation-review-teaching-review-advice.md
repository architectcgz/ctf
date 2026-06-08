# 2026-05-08 教学复盘建议实现 Review

## Scope

- `code/backend/internal/module/teaching_readmodel/application/queries/service.go`
- `code/backend/internal/module/teaching_readmodel/application/queries/response_mapper.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- `code/backend/internal/module/assessment/infrastructure/report_repository.go`
- `code/frontend/src/api/assessment.ts`
- `code/frontend/src/api/teacher/students.ts`
- `code/frontend/src/utils/skillProfile.ts`
- `code/frontend/src/features/student-dashboard/model/useStudentDashboardData.ts`
- `code/frontend/src/features/skill-profile/model/useSkillProfilePage.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useTeacherStudentAnalysisPage.ts`
- `code/frontend/src/features/teacher-workspace/model/useTeacherWorkspace.ts`
- `code/frontend/src/views/profile/__tests__/SkillProfile.test.ts`

## Review Context

- 独立 reviewer: `Herschel`
- review type: implementation gate review
- review focus:
  - `weak_dimensions` owner 是否真正统一到 `internal/teaching/advice`
  - 教师端 recommendation 链路是否保留 `weak_dimensions`
  - review archive 是否按真实维度证据聚合，而不是把总提交数分摊到所有低分维度
  - 是否补上对应回归测试

## Verdict

- `Gate verdict: pass`

## Findings

无 findings。

## Checked Points

- 教师端 `/teacher/students/:id/recommendations` 已从仅返回 challenge 数组切为返回 `weak_dimensions + challenges`，并通过 `goverter` 生成的 mapper 装配到 `TeacherRecommendationResp`。
- 学生端和教师端前端页面都已统一消费 `RecommendationData.weak_dimensions`，不再根据 skill profile 分数阈值自行判定“薄弱项”。
- review archive 的维度事实已改为按 `evidence / writeups / manualReviews` 的真实 `Category` 聚合 `AttemptCount / SuccessCount / EvidenceCount`。
- 前端已补上“advice 没给出薄弱维度时，不因画像低分而自行判弱项”的回归测试。
- 后端已补上“只有具备真实 archive evidence 的低分维度才会被标成弱项”的回归测试。

## Validation Evidence

- `go test ./internal/teaching/advice/... ./internal/module/assessment/... ./internal/module/teaching_readmodel/... ./cmd/seed-teaching-review-data -count=1`
- `go run ./cmd/seed-teaching-review-data`
- `pnpm typecheck`
- `pnpm test:run src/utils/__tests__/skillProfile.test.ts src/views/profile/__tests__/SkillProfile.test.ts src/views/teacher/__tests__/TeacherDashboard.test.ts src/views/teacher/__tests__/TeacherClassStudents.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts`

## Residual Risks

- full-router 集成测试当前显式断言的是教师 recommendation 响应中的 `challenges` 非空，还没有单独断言 `weak_dimensions` 非空。
- 当前这不构成 blocker，因为 DTO、mapper、前端消费链路和回归测试已经闭合。
