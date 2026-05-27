# Reuse Decision

## Change type
+api / feature / component / widget / docs / test

## Existing code searched
- `code/frontend/src/api/teacher/students.ts`
- `code/frontend/src/api/teaching/students.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useReviewWorkspace.ts`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisReviewQuerySync.ts`
- `code/frontend/src/components/teacher/StudentInsightPanel.vue`
- `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightAttackSessionsSection.vue`
- `code/frontend/src/widgets/teacher-student-review-workspace/StudentReviewWorkspace.vue`
- `docs/contracts/api-contract-v1.md`
- `docs/architecture/features/攻击会话读模型与复盘工作台架构.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- 题解域和班级目录的共享 contract 已按最小切片完成中性化命名，说明当前可以继续沿“共享 query / DTO 一组一刀”的方式推进。
- `TeacherAttackSessionQuery` 当前通过 shared student-analysis / review-workspace 同时服务 teacher 与 platform 两边的复盘筛选流，已经不是教师专属 query contract。

## Decision
refactor_existing

## Reason
- `TeacherAttackSessionQuery` 只是攻击会话查询参数，不承载教师专属行为；保留 teacher 前缀会继续误导 platform 学员分析和复盘工作台仍在借 teacher 语义 owner。
- 这次只改 query contract 名称和相关消费面，不顺手动 response DTO、session data 或 AWD review contest item，能控制 blast radius。

## Files to modify
- `.harness/reuse-decisions/attack-session-query-contract-naming-neutralization.md`
- `docs/plan/impl-plan/2026-05-27-attack-session-query-contract-naming-neutralization-implementation-plan.md`
- `code/frontend/src/api/teacher/students.ts`
- `code/frontend/src/api/teaching/students.ts`
- `code/frontend/src/features/teacher-student-analysis/model/useReviewWorkspace.ts`
- `code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisReviewQuerySync.ts`
- `code/frontend/src/components/teacher/StudentInsightPanel.vue`
- `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
- `code/frontend/src/components/teacher/student-insight/StudentInsightAttackSessionsSection.vue`
- `code/frontend/src/widgets/teacher-student-review-workspace/StudentReviewWorkspace.vue`
- `docs/contracts/api-contract-v1.md`
- `docs/architecture/features/攻击会话读模型与复盘工作台架构.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/frontend/2026-05-27-attack-session-query-contract-naming-neutralization-review.md`

## After implementation
- 这组 query 收口后，当前更深层 teacher 前缀共享 contract 将主要剩下 `TeacherAWDReviewContestItemData`。
