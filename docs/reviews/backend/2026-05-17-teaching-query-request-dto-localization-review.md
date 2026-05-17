# Teaching Query Request DTO Localization Review

- Review target: `ctf` repo，本地 `main` 工作区；review 范围为 teaching query request DTO 模块内化相关 diff，重点覆盖 `teaching_query/api/http`、`teaching_query/application/queries`、`internal/dto/teacher.go` 与对应文档
- Files reviewed:
  - `code/backend/internal/module/teaching_query/api/http/handler.go`
  - `code/backend/internal/module/teaching_query/api/http/handler_test.go`
  - `code/backend/internal/module/teaching_query/api/http/request_mapper.go`
  - `code/backend/internal/module/teaching_query/api/http/request_types.go`
  - `code/backend/internal/module/teaching_query/application/queries/contracts.go`
  - `code/backend/internal/module/teaching_query/application/queries/input_types.go`
  - `code/backend/internal/module/teaching_query/application/queries/service.go`
  - `code/backend/internal/module/teaching_query/application/queries/class_insight_service.go`
  - `code/backend/internal/module/teaching_query/application/queries/student_review_service.go`
  - `code/backend/internal/module/teaching_query/application/queries/class_insight_service_test.go`
  - `code/backend/internal/module/teaching_query/application/queries/student_review_service_test.go`
  - `code/backend/internal/dto/teacher.go`
  - `docs/architecture/backend/04-api-design.md`
  - `docs/plan/impl-plan/2026-05-17-teaching-query-request-dto-localization-implementation-plan.md`
  - `.harness/reuse-decisions/teaching-query-request-dto-localization.md`
- Classification check: agree with pipeline，属于 non-trivial backend refactor + review gate
- Initial gate verdict: pass with minor issues

## Current Status

- 2026-05-17 补充修复状态：已完成。独立 review 没有发现 blocker；补充项主要是验证口径和 trust-boundary 断言偏弱，当前已补齐 `TestGetClassReviewBindsWindowIntoTeachingInput`，并把 implementation plan 的 app 验证清单扩到 `TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge`。

## Findings

- no findings

## Validation Evidence

- `cd code/backend && go test ./internal/module/teaching_query/... -count=1`
- `cd code/backend && go test ./internal/app -run 'TestFullRouter_TeacherAccessAndRecommendationStateMatrix|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge|TestFullRouter_InstanceHintAndProxyStateMatrix' -count=1`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

## Final Review Verdict

- Gate verdict after follow-up tests: pass
- 结论：教师侧 class / student / evidence / attack session request DTO 已不再依赖全局 `internal/dto/teacher.go`；`teaching_query/api/http` 成为 HTTP request owner，`teaching_query/application/queries` 只消费模块内纯 input，现有 query 参数、默认值、权限口径和 response shape 保持不变。
