# Scoreboard Workspace Shell 收口复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - files reviewed：
    - `code/frontend/src/views/scoreboard/ScoreboardView.vue`
    - `code/frontend/src/components/scoreboard/ScoreboardWorkspaceShell.vue`
    - `code/frontend/src/views/scoreboard/__tests__/ScoreboardView.test.ts`
    - `code/frontend/src/views/__tests__/pageTabsStyles.test.ts`
    - `code/frontend/src/views/__tests__/journalUserDirectoryStyles.test.ts`
    - `code/frontend/src/__tests__/architectureAllowlist.ts`
- Classification check：同意当前切片属于前端 `TD-1` 结构性收口，且本轮改动范围与计划一致。
- Gate verdict：Pass

## Findings

- 无 material finding。

## Material findings

- 无。

## Senior implementation assessment

- 当前实现把 route tab、筛选、分页、刷新和数据装配继续保留在 `ScoreboardView.vue`，把顶部 tab rail、contest / points 两块 workspace 模板和局部样式下沉到 `ScoreboardWorkspaceShell.vue`，owner 边界比原先更清楚。
- 新 shell 只做模板装配、局部展示和事件转发，没有重新吸入 route/query 同步或请求 owner，也没有把 feature 依赖继续向下扩散到更深层的展示组件，符合这轮最小可审阅切片的目标。
- 相关测试已经改成按父页加 shell 的组合源码检查，避免后续再抽壳时把模板或样式 owner 回塞到父页。

## Required re-validation

- `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts src/views/scoreboard/__tests__/ScoreboardView.test.ts src/views/__tests__/pageTabsStyles.test.ts src/views/__tests__/journalUserDirectoryStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/journalEyebrowStyles.test.ts src/views/__tests__/journalUserShellStyles.test.ts`
- `npm run typecheck`
- `git diff --check -- code/frontend/src/components/scoreboard/ScoreboardWorkspaceShell.vue code/frontend/src/views/scoreboard/ScoreboardView.vue code/frontend/src/views/scoreboard/__tests__/ScoreboardView.test.ts code/frontend/src/views/__tests__/pageTabsStyles.test.ts code/frontend/src/views/__tests__/journalUserDirectoryStyles.test.ts code/frontend/src/__tests__/architectureAllowlist.ts .harness/reuse-decisions/scoreboard-workspace-shell-owner-convergence.md docs/plan/impl-plan/2026-05-26-scoreboard-workspace-shell-owner-convergence-implementation-plan.md docs/reviews/frontend/README.md docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

## Residual risk

- 本文档是当前实现上下文内的 review 归档，不是独立子上下文审查。
- `ScoreboardWorkspaceShell.vue` 仍然是一个偏大的壳组件，但这轮目标是先收口 route view owner；只要后续继续在这个 surface 上加逻辑，就需要继续按面板或目录区块拆更细的展示组件。

## Touched known-debt status

- 本轮 touched 的已知结构债是 `ScoreboardView.vue` oversized route view owner。
- 该债务在 touched surface 上已完成收口：父页已脱离 oversized allowlist，本体降到 `108` 行，当前没有把原本的页面 owner 再次混回去。
