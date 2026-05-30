# Reuse Decision

## Change type
+component / widget / docs / test

## Existing code searched
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewWorkspace.vue`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewWorkspace.test.ts`
- `code/frontend/src/widgets/awd-review-workspace/index.ts`
- `code/frontend/src/components/teacher/awd-review/AwdReviewRoundSelector.vue`
- `code/frontend/src/components/teacher/awd-review/AwdReviewAnalysisSection.vue`
- `code/frontend/src/components/teacher/awd-review/AwdReviewEvidenceGrid.vue`
- `code/frontend/src/components/teacher/awd-review/AwdReviewTeamDrawer.vue`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/teacher/__tests__/teacherAwdReviewWorkspaceExtraction.test.ts`
- `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
- `code/frontend/src/views/platform/__tests__/PlatformAwdReviewDetail.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## Similar implementations found
- `ReviewArchiveWorkspace`、`AwdReviewWorkspace` 这类 route-level workspace 已经稳定作为 shared widget owner；它们内部的 detail 子区块更适合直接并入 widget 本身，而不是再挂一层 legacy `components/*` 中转。
- `AwdReviewWorkspace` 当前被 teacher / platform 两边复用，teacher 组件入口只是历史目录归属残留。

## Decision
refactor_existing

## Reason
- `AwdReviewWorkspace` 仍通过 `components/awd-review` -> `components/teacher/awd-review/*` 这条过渡桥读取四个 detail 子组件，会继续保留 shared widget 对 legacy 组件目录的结构依赖。
- 这次直接把 round selector / analysis / evidence / team drawer 并入 `widgets/awd-review-workspace`，同步更新测试与类型索引，并删除 `components/awd-review` 与 `components/teacher/awd-review` 的残余入口，不保留中间桥接层。

## Files to modify
- `.harness/reuse-decisions/awd-review-component-public-owner-neutralization.md`
- `docs/plan/impl-plan/2026-05-27-awd-review-component-public-owner-neutralization-implementation-plan.md`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewWorkspace.vue`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewAnalysisSection.vue`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewEvidenceGrid.vue`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewRoundSelector.vue`
- `code/frontend/src/widgets/awd-review-workspace/AwdReviewTeamDrawer.vue`
- `code/frontend/src/widgets/awd-review-workspace/index.ts`
- `code/frontend/src/widgets/awd-review-workspace/awdReviewWorkspaceUiStrategy.test.ts`
- `code/frontend/src/components/awd-review/index.ts`
- `code/frontend/src/components/teacher/awd-review/*`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/teacher/__tests__/teacherAwdReviewWorkspaceExtraction.test.ts`
- `code/frontend/src/components.d.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/frontend/2026-05-27-awd-review-component-public-owner-neutralization-review.md`

## After implementation
- `AwdReviewWorkspace` 将直接从 `widgets/awd-review-workspace` 内部读取 round selector / analysis / evidence / team drawer，相关 legacy component import 依赖会从这条 shared workspace 上彻底清掉。
