# Attack Session Query Contract Naming Neutralization 收口复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-27-attack-session-query-contract-naming-neutralization-implementation-plan.md`
  - files reviewed：
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
- Classification check：同意本轮属于前端 P1 级结构收口，范围限定在共享 attack session query contract 的中性化命名。
- Gate verdict：Pass

## Findings

- 无 material finding。

## Material findings

- 无。

## Senior implementation assessment

- `TeacherAttackSessionQuery` 已经通过共享 review workspace 同时服务 teacher / platform 两边的复盘筛选流，本轮改成 `AttackSessionQuery` 能直接消除角色语义噪音，而且不改行为面。
- 这刀只收 query contract，没有把 `TeacherAttackSessionResponseData` 或 `TeacherAWDReviewContestItemData` 混进来，边界合理。
- route query sync、筛选 UI 和 `getStudentAttackSessions()` 的请求参数结构都保持原样，风险主要停留在类型引用层。

## Required re-validation

- `npm run test:run -- src/views/platform/__tests__/PlatformStudentAnalysis.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `npm run typecheck`
- `bash scripts/check-workflow-complete.sh`
- `git diff --check -- code/frontend/src/api/teacher/students.ts code/frontend/src/api/teaching/students.ts code/frontend/src/features/teacher-student-analysis/model/useReviewWorkspace.ts code/frontend/src/features/student-analysis-workspace/model/useStudentAnalysisReviewQuerySync.ts code/frontend/src/components/teacher/StudentInsightPanel.vue code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue code/frontend/src/components/teacher/student-insight/StudentInsightAttackSessionsSection.vue code/frontend/src/widgets/teacher-student-review-workspace/StudentReviewWorkspace.vue docs/contracts/api-contract-v1.md docs/architecture/features/攻击会话读模型与复盘工作台架构.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md .harness/reuse-decisions/attack-session-query-contract-naming-neutralization.md docs/plan/impl-plan/2026-05-27-attack-session-query-contract-naming-neutralization-implementation-plan.md docs/reviews/frontend/2026-05-27-attack-session-query-contract-naming-neutralization-review.md`

## Residual risk

- `TeacherAWDReviewContestItemData` 仍保留 teacher 前缀，后续需要独立切片。
- 本轮 review 在同一主会话里完成，没有独立子 agent 复核；因此这里只作为同上下文 gate review 证据，不表述为独立审查。

## Touched known-debt status

- 本轮 touched 的已知结构债是共享复盘筛选 contract 仍保留 `TeacherAttackSessionQuery` 前缀。
- 在本轮 touched surface 上，这条债务已经完成当前阶段收口；当前更深层 teacher 前缀共享 contract 主要剩下 AWD review contest item 这一组。
