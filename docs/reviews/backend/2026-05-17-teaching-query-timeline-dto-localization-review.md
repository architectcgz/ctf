# Teaching Query Timeline DTO Localization Review

- Review target: `ctf` repo，本地 `main` 工作区；review 范围为 teaching query teacher timeline DTO 模块内化相关 diff，重点覆盖 `teaching_query/application/queries`、`internal/dto/timeline.go`、app 集成测试与对应文档
- Files reviewed:
  - `code/backend/internal/module/teaching_query/application/queries/contracts.go`
  - `code/backend/internal/module/teaching_query/application/queries/response_mapper.go`
  - `code/backend/internal/module/teaching_query/application/queries/response_mapper_gen.go`
  - `code/backend/internal/module/teaching_query/application/queries/student_review_service.go`
  - `code/backend/internal/module/teaching_query/application/queries/student_review_service_test.go`
  - `code/backend/internal/module/teaching_query/application/queries/timeline_types.go`
  - `code/backend/internal/app/full_router_state_matrix_integration_test.go`
  - `code/backend/internal/dto/timeline.go`
  - `docs/architecture/backend/04-api-design.md`
  - `docs/plan/impl-plan/2026-05-17-teaching-query-timeline-dto-localization-implementation-plan.md`
  - `.harness/reuse-decisions/teaching-query-timeline-dto-localization.md`
- Classification check: agree with pipeline，属于 non-trivial backend refactor + review gate
- Initial gate verdict: pass with minor issues

## Current Status

- 2026-05-17 补充修复状态：已完成。独立 review 没有发现 blocker；补充项主要是验证深度不足，当前已新增 `TestStudentReviewQueryServiceGetStudentTimelineMapsFields`、`TestStudentReviewQueryServiceGetStudentTimelineNormalizesNilEventsToEmptySlice`，并加强 `full_router_state_matrix_integration_test.go` 对 `challenge_id` 与 `timestamp` 的断言。

## Findings

- no findings

## Validation Evidence

- `cd code/backend && go test ./internal/module/teaching_query/... -count=1`
- `cd code/backend && go test ./internal/app -run 'TestFullRouter_TeacherAccessAndRecommendationStateMatrix|TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge' -count=1`

## Final Review Verdict

- Gate verdict after follow-up tests: pass
- 结论：`TimelineResp` / `TimelineEvent` owner 已收口到 `teaching_query/application/queries`，旧的全局 `internal/dto/timeline.go` 已退出；`GET /api/v1/teacher/students/:id/timeline` 的路径、查询参数、JSON 字段和时间输出口径保持不变。
