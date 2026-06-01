# Challenge Detail Workspace Tab Owner Convergence Review

## Review target

- Repository: `/home/azhi/workspace/projects/ctf`
- Branch: `main`
- Diff source: working tree uncommitted change set `challenge-detail workspace tab owner convergence`
- Files reviewed:
  - `.harness/reuse-decisions/challenge-detail-workspace-tab-owner-convergence.md`
  - `docs/plan/impl-plan/2026-05-31-challenge-detail-workspace-tab-owner-convergence-plan.md`
  - `code/frontend/src/features/challenge-detail/model/useChallengeDetailPage.ts`
  - `code/frontend/src/pages/challenges/ChallengeDetailRoutePage.vue`
  - `code/frontend/src/widgets/challenge-detail-workspace/ChallengeDetailWorkspace.vue`
  - `code/frontend/src/pages/challenges/__tests__/ChallengeDetail.test.ts`
  - `code/frontend/src/pages/challenges/__tests__/challengePageUiStrategy.test.ts`
  - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Validation actually run:
  - `git diff --check -- .harness/reuse-decisions/challenge-detail-workspace-tab-owner-convergence.md docs/plan/impl-plan/2026-05-31-challenge-detail-workspace-tab-owner-convergence-plan.md code/frontend/src/features/challenge-detail/model/useChallengeDetailPage.ts code/frontend/src/pages/challenges/ChallengeDetailRoutePage.vue code/frontend/src/widgets/challenge-detail-workspace/ChallengeDetailWorkspace.vue code/frontend/src/pages/challenges/__tests__/ChallengeDetail.test.ts code/frontend/src/pages/challenges/__tests__/challengePageUiStrategy.test.ts docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
  - `cd code/frontend && npm run test:run -- src/pages/challenges/__tests__/ChallengeDetail.test.ts src/pages/challenges/__tests__/challengePageUiStrategy.test.ts`
  - `bash scripts/check-consistency.sh`
  - `bash scripts/check-workflow-complete.sh`

## Classification check

- 结论：同意按 `HARNESS / frontend non-trivial refactor review` 审查。
- 依据：改动集中在 challenge detail workspace 的 panel query owner 收口、widget 键盘导航 owner 下沉和对应护栏测试，没有扩散到 challenge 数据加载、实例 workflow 或 solution 子 tab 本地 owner。

## Gate verdict

- `pass`
- blocker：无

## Findings

- no findings

## Material findings

- 无 material findings。

## Senior implementation assessment

- `useChallengeDetailPage.ts` 现在通过 `useRouteQueryTabs()` 单点持有 `question/solution/records/writeup` 的 `panel` query owner，和最近几轮 student-analysis / class-students / skill-profile 的模式一致。
- `ChallengeDetailWorkspace.vue` 只保留 workspace tab 的键盘焦点导航与展示桥接，没有重新接管 route-aware query 状态。
- `ChallengeDetailRoutePage.vue` 去掉了纯转手的 tab wrapper，route page 继续保持薄组合层。

## Required re-validation

- 当前无需额外修复后复验。

## Residual risk

- 当前没有发现 correctness、回归或 owner 边界问题。
- 轻微测试空白：workspace tab 键盘导航 owner 已移动到 widget 内部，但当前运行态护栏主要覆盖 query hydrate / click 回写 / 默认 tab 清理 query，还没有直接覆盖 workspace tab 的 `Arrow/Home/End` 路径。

## Touched known-debt status

- 已触达并实质收口 `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md` 中 challenge detail workspace panel owner tightening 这条 debt。
- 当前 touched surface 上，`panel` query owner 已从旧 `useUrlSyncedTabs()` 模式收回 page model，widget 与 route page 都不再额外改写 query 状态。
