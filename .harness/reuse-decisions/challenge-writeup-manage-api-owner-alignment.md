# Reuse Decision

## Change type
+api / feature / component / test / docs

## Existing code searched
- `code/frontend/src/components/platform/writeup/ChallengeWriteupManagePanel.vue`
- `code/frontend/src/features/challenge-writeup-editor/model/useChallengeWriteupManagement.ts`
- `code/frontend/src/api/admin/authoring.ts`
- `code/frontend/src/api/teacher/writeups.ts`
- `code/frontend/src/api/teaching/writeups.ts`
- `code/frontend/src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/contracts/api-contract-v1.md`
- `docs/reviews/frontend/2026-05-27-admin-awd-review-api-owner-alignment-review.md`

## Similar implementations found
- `useAwdReviewIndex` / `useAwdReviewDetailPage` / `useAwdReviewExportFlow` 刚刚已经通过 `api/admin/*` 薄 wrapper 收口过共享 workflow 的 admin / teacher API owner。
- `api/admin/authoring.ts` 已经是 challenge admin surface 的既有 owner，`getChallengeWriteup`、`deleteChallengeWriteup` 都放在这里，说明学员题解投稿目录的 admin 语义 wrapper 继续落在同一文件最符合当前结构。
- `TeacherStudentAnalysis` 与 `useSubmissionReviewFlows` 仍然承接教师侧题解查看 / 评阅链路，说明这轮不能把“teacher 能查看题解”错误收窄成 admin-only 能力，而应只收口 platform surface 的 API owner。

## Decision
extend_existing

## Reason
- `ChallengeWriteupManagePanel` 本身已经是 platform 目录下的管理面板，但 `useChallengeWriteupManagement` 仍直接 import `@/api/teaching` 的 `getTeacherWriteupSubmissions`，属于 backlog 里典型的 platform surface 继续依赖 teacher 语义 owner。
- 最小正确方案不是重写面板，也不是拆 `challenge-writeup-editor` feature，而是在 `api/admin/authoring.ts` 补一个 platform 语义的薄 wrapper，然后让 `useChallengeWriteupManagement` 切到这个 owner。
- 这样只收口 platform 面板的 API owner，不改 HTTP contract、不复制 feature，也不触碰教师侧已有的题解查看 / 评阅流程。

## Files to modify
- `.harness/reuse-decisions/challenge-writeup-manage-api-owner-alignment.md`
- `docs/plan/impl-plan/2026-05-27-challenge-writeup-manage-api-owner-alignment-implementation-plan.md`
- `code/frontend/src/api/admin/authoring.ts`
- `code/frontend/src/features/challenge-writeup-editor/model/useChallengeWriteupManagement.ts`
- `code/frontend/src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts`
- `docs/contracts/api-contract-v1.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/frontend/2026-05-27-challenge-writeup-manage-api-owner-alignment-review.md`

## After implementation
- 如果这层收口稳定，后续 platform challenge 相关共享 feature 也应优先通过 `api/admin/authoring.ts` 等 platform/admin owner 暴露薄 wrapper，而不是直接从 `@/api/teaching` 消费 teacher 命名函数；teacher 侧查看 / 评阅链路则继续保留自己的 owner。
