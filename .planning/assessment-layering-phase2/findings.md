# Findings

- `assessment` 的 `ports` 已基本按用例划分完成，当前主要遗留在 composition：[`assessment_module.go`](/home/azhi/workspace/projects/ctf/code/backend/internal/app/composition/assessment_module.go) 仍直接 new concrete `assessmentinfra.Repository / ReportRepository` 并在单个函数内完成全部装配。
- `application` 调用面清晰：
  - `profile_service` 只依赖 `ProfileRepository`
  - `recommendation_service` 依赖 `RecommendationRepository + ChallengeRepository`
  - `report_service` 依赖 `ReportRepository + AssessmentProfileReader`
- 因此 `assessment` 更适合做“typed deps + builder 收口”的轻量 phase2，而不需要再拆主仓储接口。
