# AWD Challenge Config Panel Decomposition 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-awd-challenge-config-panel-decomposition-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/awd-challenge-config-panel-decomposition.md`
    - `docs/plan/impl-plan/2026-05-29-awd-challenge-config-panel-decomposition-plan.md`
    - `docs/reviews/frontend/2026-05-29-awd-challenge-config-panel-decomposition-review.md`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
    - `code/frontend/src/features/platform-contests/ui/AWDChallengeConfigPanel.vue`
    - `code/frontend/src/features/platform-contests/ui/AWDChallengeConfigHeader.vue`
    - `code/frontend/src/features/platform-contests/ui/AWDChallengeConfigDirectorySection.vue`
    - `code/frontend/src/features/platform-contests/ui/AWDChallengeConfigDirectoryRow.vue`
    - `code/frontend/src/features/platform-contests/ui/awdChallengeConfigPanel.css`
    - `code/frontend/src/features/platform-contests/ui/awdChallengeConfigPanel.types.ts`
    - `code/frontend/src/components/platform/__tests__/AWDChallengeConfigPanel.test.ts`
    - `code/frontend/src/components/platform/__tests__/AWDChallengeConfigPanelExtraction.test.ts`
    - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase20.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase25.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts`
    - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- Classification check：同意按 `platform-contests` feature 内大 surface 收口处理，属于非 trivial frontend refactor。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `AWDChallengeConfigPanel.vue` 现在只保留 `challengeLinks` contract、排序、summary 派生、`useAwdCheckResultPresentation()` 接线，以及 directory row view model 组装，不再继续直接承接 header、table row 和整段 scoped CSS。
- `AWDChallengeConfigHeader.vue` 已明确承接 overline / title / description 与 summary strip，父层不再继续混放 metric cards。
- `AWDChallengeConfigDirectorySection.vue` 已明确承接 empty state、table shell 与 row 列表组合。
- `AWDChallengeConfigDirectoryRow.vue` 已明确承接 challenge identity、checker、score、rules summary、validation 与 row action primitive。
- `awdChallengeConfigPanel.css` 已承接 panel shell、table、row、validation 和 action 区样式，`AWDChallengeConfigPanel.vue` 不再保留 scoped style 尾巴。
- AWD config 相关 raw-source 护栏已统一切到 panel + 子组件 + CSS 的聚合源码视角，继续覆盖 primitive、theme token、workspace overline 与 dark surface 对齐。
- `AWDChallengeConfigPanel.vue` 文件体量从约 `562` 行降到 `213` 行；本轮 touched surface 上的“feature owner 正确但父层仍混放 header / row / CSS”债已经完成收口。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDChallengeConfigPanel.test.ts src/components/platform/__tests__/AWDChallengeConfigPanelExtraction.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase20.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase25.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；按 `development-pipeline` 的独立 gate 要求，依然缺少单独 reviewer 上下文的复核证据，这条缺口需要在交付说明里明确。
- `AWDChallengeConfigPanel.vue` 当前继续保留排序、summary 统计和 directory row view model 组装，规模可接受；如果后续再叠筛选或更多动作，更适合继续把 directory presentation model 拆成局部 composable，而不是把展示壳拉回父组件。
- 本轮不涉及 AWD 配置编辑对话框或 contest edit stage owner；如果上层再堆更多阶段逻辑，不应回灌到这块目录 panel。

## Touched known-debt status

- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md` 中这条 contest / AWD feature 内残余大 surface，已在 `AWDChallengeConfigPanel.vue` 这一块 touched surface 上完成一刀父壳收口；当前这条 P2 的剩余重点已进一步从 AWD config 目录收敛到 `AWDInstanceOrchestrationPanel.vue` 和其它仍混写 matrix / row / CSS 的 feature panel。
