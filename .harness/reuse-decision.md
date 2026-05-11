# Reuse Decision

This file is only for the current task and may be overwritten.
Durable reuse knowledge belongs in `harness/reuse/index.yaml`; append-only summaries belong in `harness/reuse/history.md`.

## Change type
- page
- component
- styling-token

## Existing code searched
- code/frontend/src/assets/styles/theme.css
- code/frontend/src/entities/challenge/model/presentation.ts
- code/frontend/src/entities/challenge/ui/ChallengeDirectoryRow.vue
- code/frontend/src/entities/challenge/ui/ChallengeMetaStrip.vue
- code/frontend/src/components/dashboard/student/StudentRecommendationPage.vue
- code/frontend/src/views/challenges/ChallengeList.vue
- code/frontend/src/views/challenges/ChallengeDetail.vue
- code/frontend/src/views/__tests__/studentUserSurfaceAlignment.test.ts
- code/frontend/src/views/challenges/__tests__/ChallengeList.test.ts
- code/frontend/src/views/dashboard/__tests__/DashboardView.test.ts

## Similar implementations found
- code/frontend/src/entities/challenge/model/presentation.ts
- code/frontend/src/entities/challenge/ui/ChallengeDirectoryRow.vue
- code/frontend/src/entities/challenge/ui/ChallengeMetaStrip.vue
- code/frontend/src/views/challenges/ChallengeList.vue

## Decision
- refactor_existing

## Reason
- 本次不是新增题目分类展示组件，而是复用 challenge entity 的分类展示入口 `getChallengeCategoryColor()`，把旧的 `--challenge-tone-*` 收口为语义更明确的 `--challenge-category-pill-*`。
- `StudentRecommendationPage.vue` 原本本地写了灰色 `journal-category-chip`，现在改为通过 challenge entity 取色，和题目列表、题目详情的分类胶囊共享同一组题目分类胶囊变量。
- `ChallengeList.vue` 和 `ChallengeDetail.vue` 只移除页面局部 `--challenge-tone-*` 定义，避免同一分类胶囊色在不同页面分散维护。
- 回归测试沿用学生仪表盘与题目列表现有测试文件扩充，不新增独立测试 harness。

## Files to modify
- code/frontend/src/assets/styles/theme.css
- code/frontend/src/components/dashboard/student/StudentRecommendationPage.vue
- code/frontend/src/entities/challenge/model/presentation.ts
- code/frontend/src/entities/challenge/ui/ChallengeDirectoryRow.vue
- code/frontend/src/entities/challenge/ui/ChallengeMetaStrip.vue
- code/frontend/src/views/__tests__/studentUserSurfaceAlignment.test.ts
- code/frontend/src/views/challenges/ChallengeDetail.vue
- code/frontend/src/views/challenges/ChallengeList.vue

## After implementation
- No new durable reuse entry was added. The reusable rule was recorded in the active CTF frontend theme skill; project code now uses the challenge entity presentation helper as the shared owner for challenge category pill colors.
