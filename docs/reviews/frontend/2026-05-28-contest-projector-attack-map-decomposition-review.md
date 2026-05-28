# Contest Projector Attack Map Decomposition 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-28-contest-projector-attack-map-decomposition-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/contest-projector-attack-map-decomposition.md`
    - `docs/plan/impl-plan/2026-05-28-contest-projector-attack-map-decomposition-plan.md`
    - `docs/reviews/frontend/2026-05-28-contest-projector-attack-map-decomposition-review.md`
    - `code/frontend/src/features/contest-projector/model/projectorAttackMapSupport.ts`
    - `code/frontend/src/features/contest-projector/model/useProjectorAttackBoard.ts`
    - `code/frontend/src/features/contest-projector/model/index.ts`
    - `code/frontend/src/features/contest-projector/ui/ContestProjectorAttackMap.vue`
    - `code/frontend/src/features/contest-projector/ui/ContestProjectorAttackBoard.vue`
    - `code/frontend/src/features/contest-projector/ui/ContestProjectorAttackMapTeamSidebar.vue`
    - `code/frontend/src/features/contest-projector/ui/ContestProjectorAttackMapStatsSidebar.vue`
    - `code/frontend/src/features/contest-projector/ui/ContestProjectorAttackDetailOverlay.vue`
    - `code/frontend/src/features/contest-projector/ui/__tests__/ContestProjectorAttackMap.test.ts`
    - `code/frontend/src/features/contest-projector/ui/__tests__/contestProjectorAttackMapExtraction.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestProjectorPageStateExtraction.test.ts`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- Classification check：同意按 `contest-projector` feature 内部超大 attack map surface 收口处理，属于非 trivial frontend refactor。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `ContestProjectorAttackMap.vue` 现在只保留 map 展示派生、detail overlay 装配和 `Escape` 关闭 owner，不再继续混放左右侧栏、board shell 和 DOM mechanics。
- `ContestProjectorAttackBoard.vue` 与 `useProjectorAttackBoard.ts` 已明确承接 beam path、拖拽、`ResizeObserver`、team/service ref、localStorage 持久化这类 board 本地能力；这批逻辑不再散落在父组件模板旁边。
- `ContestProjectorAttackMapTeamSidebar.vue` 与 `ContestProjectorAttackMapStatsSidebar.vue` 已承接 legend / first blood / team drilldown、ranking / attack stats drilldown 这两块稳定展示区。父组件不再继续内联这些固定 panel。
- `projectorAttackMapSupport.ts` 把 `AttackMapDetailPanel`、service display/icon/key helper 收成单点，`ContestProjectorAttackDetailOverlay.vue` 与 board / sidebars 不再各写一份相似 helper。
- `ContestProjectorAttackMap.vue` 文件体量从原先约 `779` 行降到 `184` 行；这轮 touched surface 上的“map view-model + board mechanics + sidebars + board template 混写”债已经完成收口。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/features/contest-projector/ui/__tests__/ContestProjectorAttackMap.test.ts src/features/contest-projector/ui/__tests__/contestProjectorAttackMapExtraction.test.ts src/views/platform/__tests__/contestProjectorPageStateExtraction.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 只做了同上下文 self-review；按 `development-pipeline` 的独立 gate 要求，仍缺少单独 reviewer 上下文的复核证据。当前策略下未额外派生 subagent，因此这条缺口需要在交付说明里明确。
- `ContestProjectorAttackMap.css` 仍是 projector 线上的大体量样式文件，本轮没有继续把 CSS 再拆成 board / sidebar 粒度；如果后续 projector attack map 再增长，下一刀更适合从样式 owner 或 board 子 surface 着手，而不是回退到单文件模板堆叠。

## Touched known-debt status

- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md` 中这条 P2 对 `ContestProjectorAttackMap.vue` 的已知超大组件债，在 touched surface 内已经完成 owner 收口；当前该条 residual 重点已从 projector attack map 本体转移到其它未拆大组件，以及 projector attack map 后续若继续增长时的更深层 CSS / board 子 surface。
