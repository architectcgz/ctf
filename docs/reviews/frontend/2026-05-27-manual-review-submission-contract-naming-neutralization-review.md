# Manual Review Submission Contract Naming Neutralization 收口复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-27-manual-review-submission-contract-naming-neutralization-implementation-plan.md`
  - files reviewed：
    - `code/frontend/src/api/contracts.ts`
    - `code/frontend/src/api/teacher/writeups.ts`
    - `code/frontend/src/api/teaching/writeups.ts`
    - `code/frontend/src/features/teacher-student-analysis/model/useSubmissionReviewFlows.ts`
    - `code/frontend/src/components/teacher/StudentInsightPanel.vue`
    - `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue`
    - `code/frontend/src/components/teacher/student-insight/StudentInsightWriteupsSection.vue`
    - `code/frontend/src/components/teacher/student-insight/StudentInsightManualReviewSection.vue`
    - `code/frontend/src/components/teacher/student-insight/studentInsightShared.ts`
    - `docs/contracts/api-contract-v1.md`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：同意本轮属于前端 P1 级结构收口，范围限定在 manual review 这组已被 teacher / platform 共用的 DTO 中性化命名。
- Gate verdict：Pass

## Findings

- 无 material finding。

## Material findings

- 无。

## Senior implementation assessment

- 本轮没有把 `TeacherSubmissionWriteupDetailData`、`TeacherClassItem`、`TeacherAttackSessionQuery` 等其他残余命名一起打包，而是只收 manual review item/detail 这组已经明确进入共享 student-analysis workflow 的 DTO，切片边界合理。
- `api/contracts.ts`、teacher / teaching writeups client、submission review flow 以及 student insight 组件消费面都同步切到了中性命名，避免出现 contract 已收口但共享组件还继续引用旧 teacher 前缀的半收口状态。
- teacher API function 命名与 HTTP path 保持原样，只收 contract 语义，不会误导这轮变成 owner 或权限边界调整。

## Required re-validation

- `npm run test:run -- src/views/platform/__tests__/PlatformStudentAnalysis.test.ts src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `npm run typecheck`
- `bash scripts/check-workflow-complete.sh`
- `git diff --check -- code/frontend/src/api/contracts.ts code/frontend/src/api/teacher/writeups.ts code/frontend/src/api/teaching/writeups.ts code/frontend/src/features/teacher-student-analysis/model/useSubmissionReviewFlows.ts code/frontend/src/components/teacher/StudentInsightPanel.vue code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue code/frontend/src/components/teacher/student-insight/StudentInsightWriteupsSection.vue code/frontend/src/components/teacher/student-insight/StudentInsightManualReviewSection.vue code/frontend/src/components/teacher/student-insight/studentInsightShared.ts docs/contracts/api-contract-v1.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md .harness/reuse-decisions/manual-review-submission-contract-naming-neutralization.md docs/plan/impl-plan/2026-05-27-manual-review-submission-contract-naming-neutralization-implementation-plan.md docs/reviews/frontend/2026-05-27-manual-review-submission-contract-naming-neutralization-review.md`

## Residual risk

- `TeacherSubmissionWriteupDetailData` 仍保留 teacher 前缀，本轮没有覆盖。
- `TeacherClassItem`、`TeacherAWDReviewContestItemData`、`TeacherAttackSessionQuery` 等其他共享 contract 命名残余还在，需要后续独立切片。
- 本轮 review 在同一主会话里完成，没有独立子 agent 复核；因此这里只作为同上下文 gate review 证据，不表述为独立审查。

## Touched known-debt status

- 本轮 touched 的已知结构债是共享 student-analysis workflow 里的 manual review DTO 仍保留 teacher 前缀。
- 在本轮 touched surface 上，这条债务已经完成当前阶段收口：manual review item/detail contract 已切到 `ManualReviewSubmissionItemData` / `ManualReviewSubmissionDetailData`；剩余更深层 teacher 命名已明确转移到 writeup detail / class / AWD review / attack session query 等其他 contract。
