# Reuse Decision

## Change type

frontend / route-view / composable / shared panel

## Existing code searched

- `code/frontend/src/views/challenges/ChallengeDetail.vue`
- `code/frontend/src/features/challenge-detail/model/useChallengeDetailPage.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeDetailPresentation.ts`
- `code/frontend/src/views/dashboard/DashboardView.vue`
- `code/frontend/src/views/profile/SkillProfile.vue`
- `code/frontend/src/views/scoreboard/ScoreboardView.vue`
- `.harness/reuse-decisions/challenge-instance-card-copy-pruning.md`
- `.harness/reuse-decisions/writeup-submission-no-review-contract-frontend.md`
- `.harness/reuse-decisions/student-dashboard-panel-spacing-and-divider.md`
- `.harness/reuse-decisions/teacher-workspace-subpanel-token-fallbacks.md`

## Similar implementations found

- `challenge-instance-card-copy-pruning`
  - 已经把 challenge detail 视图继续收口在现有 route view 和 challenge-detail model owner 下，而不是拆出新的页面壳
- `writeup-submission-no-review-contract-frontend`
  - 已有 challenge detail 题解提交流程的前端 owner，可继续沿用按需加载和显式错误反馈的 composable 边界
- `student-dashboard-panel-spacing-and-divider`
  - 已经把 dashboard workspace 面板节奏收口到共享 token，而不是页面局部硬编码
- `teacher-workspace-subpanel-token-fallbacks`
  - 已有 workspace 子面板 overline / padding / token 对齐模式，可继续给 scoreboard、profile 等 tab 面板复用

## Decision

extend_existing

## Reason

这批前端改动不是新建页面体系，而是继续复用现有 owner：

- challenge detail 仍由 `useChallengeDetailPage` 统一拥有 route 级数据加载、tab 切换后的按需预取、错误态和重试
- `ChallengeSolutionsPanel` 与 `ChallengeSubmissionRecordsPanel` 只接收更明确的 loading 状态，不接管远程请求 owner
- dashboard / profile / scoreboard 继续沿用现有 workspace overline 和 panel spacing token，不新增并行样式语义

所以本次只扩展既有 composable、route view 和共享 panel token 模式，不新建新的 frontend slice 或页面壳。

## Files to modify

- `code/frontend/src/components/challenge/ChallengeSolutionsPanel.vue`
- `code/frontend/src/components/challenge/ChallengeSubmissionRecordsPanel.vue`
- `code/frontend/src/features/challenge-detail/model/useChallengeDetailDataLoader.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeDetailInteractions.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeDetailPage.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeWriteupSubmissionFlow.ts`
- `code/frontend/src/views/challenges/ChallengeDetail.vue`
- `code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`
- `code/frontend/src/views/dashboard/DashboardView.vue`
- `code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts`
- `code/frontend/src/views/profile/SkillProfile.vue`
- `code/frontend/src/views/profile/__tests__/SkillProfile.test.ts`
- `code/frontend/src/views/scoreboard/ScoreboardView.vue`
- `code/frontend/src/views/scoreboard/__tests__/ScoreboardView.test.ts`
- `code/frontend/src/assets/styles/journal-soft-surfaces.css`
- `code/frontend/vite.config.ts`
- `README.md`

## After implementation

- challenge detail 后续若继续扩 tab，优先走现有 `useChallengeDetailPage` 的按需加载 owner，不在子组件内各自发请求
- workspace tab 面板后续继续复用 `workspace-overline` 和共享 panel spacing token，不再回退到页面局部标题条
