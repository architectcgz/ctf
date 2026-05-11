# Reuse Decision

This file is only for the current task and may be overwritten.
Durable reuse knowledge belongs in `harness/reuse/index.yaml`; append-only summaries belong in `harness/reuse/history.md`.

## Change type
- page
- styling-token

## Existing code searched
- code/frontend/src/assets/styles/theme.css
- code/frontend/src/assets/styles/workspace-shell.css
- code/frontend/src/entities/challenge/model/presentation.ts
- code/frontend/src/entities/challenge/ui/ChallengeDirectoryRow.vue
- code/frontend/src/entities/challenge/ui/ChallengeMetaStrip.vue
- code/frontend/src/components/dashboard/student/StudentRecommendationPage.vue
- code/frontend/src/views/challenges/ChallengeList.vue
- code/frontend/src/views/__tests__/studentUserSurfaceAlignment.test.ts

## Similar implementations found
- code/frontend/src/assets/styles/workspace-shell.css
- code/frontend/src/entities/challenge/model/presentation.ts
- code/frontend/src/views/challenges/ChallengeList.vue

## Decision
- refactor_existing

## Reason
- 本次不是新增难度展示组件，而是在已有 `difficulty-chip` 和 challenge entity 的基础上，把题目难度胶囊色收口为 `--challenge-difficulty-pill-*`。
- `--color-diff-*` 继续作为基础色保留；面向题目难度胶囊的导出变量使用更明确的 `challenge-difficulty-pill` 语义，避免继续暴露 `--challenge-diff-*` 这种缩写命名。
- 推荐页难度胶囊继续复用现有 `difficultyClass()` 和共享 `.difficulty-chip` 样式，不新增本地页面样式。
- 回归测试沿用学生仪表盘和题目列表现有测试文件扩充。

## Files to modify
- code/frontend/src/assets/styles/theme.css
- code/frontend/src/assets/styles/workspace-shell.css
- code/frontend/src/entities/challenge/model/presentation.ts
- code/frontend/src/views/__tests__/studentUserSurfaceAlignment.test.ts
- code/frontend/src/views/challenges/ChallengeList.vue

## After implementation
- No new durable reuse entry was added. The rule has been recorded in the active CTF frontend theme skill; project code now uses challenge entity presentation helpers and shared difficulty chip styles as the owner for challenge difficulty pill colors.
