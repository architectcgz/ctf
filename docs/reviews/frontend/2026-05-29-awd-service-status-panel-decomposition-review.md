# AWD Service Status Panel Decomposition 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-awd-service-status-panel-decomposition-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/awd-service-status-panel-decomposition.md`
    - `docs/plan/impl-plan/2026-05-29-awd-service-status-panel-decomposition-plan.md`
    - `docs/reviews/frontend/2026-05-29-awd-service-status-panel-decomposition-review.md`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
    - `code/frontend/src/features/awd-inspector/ui/AWDServiceStatusPanel.vue`
    - `code/frontend/src/features/awd-inspector/ui/AWDServiceStatusToolbar.vue`
    - `code/frontend/src/features/awd-inspector/ui/AWDServiceStatusMatrix.vue`
    - `code/frontend/src/features/awd-inspector/ui/AWDServiceRoundPerformanceTable.vue`
    - `code/frontend/src/features/awd-inspector/ui/awdServiceStatusPanel.css`
    - `code/frontend/src/features/awd-inspector/ui/awdServiceStatusPanel.types.ts`
    - `code/frontend/src/components/platform/__tests__/AWDServiceStatusPanel.test.ts`
    - `code/frontend/src/components/platform/__tests__/AWDServiceStatusPanelExtraction.test.ts`
    - `code/frontend/src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts`
    - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- Classification check：同意按 `awd-inspector` feature 内大 surface 收口处理，属于非 trivial frontend refactor。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `AWDServiceStatusPanel.vue` 现在只保留 `AWDServiceStatusPanelProps/Emits` contract、挑战列/队伍行/状态展示 presentation model、filter forwarding 与空态文案推导，不再继续直接内联 toolbar、matrix table、round performance summary 和 scoped CSS。
- `AWDServiceStatusToolbar.vue` 已明确承接标题、队伍统计、team/status/source/alert filters 与 export action。
- `AWDServiceStatusMatrix.vue` 已明确承接服务矩阵 table shell、状态单元格和空态。
- `AWDServiceRoundPerformanceTable.vue` 已明确承接本轮得分与健康表现 summary table。
- `awdServiceStatusPanel.css` 已承接 panel shell、toolbar、filters、matrix、status card、summary table 与响应式样式，父面板不再保留 scoped style 尾巴。
- 服务状态相关 raw-source 护栏已统一切到 panel + 子组件 + CSS 的聚合源码视角，继续覆盖 extraction 边界与 theme token 约束。
- `AWDServiceStatusPanel.vue` 文件体量从约 `509` 行降到 `193` 行；本轮 touched surface 上“feature owner 正确但父层仍混放 toolbar / matrix / summary / CSS”的债已经完成收口。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDServiceStatusPanel.test.ts src/components/platform/__tests__/AWDServiceStatusPanelExtraction.test.ts src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；按 `development-pipeline` 的独立 gate 要求，依然缺少单独 reviewer 上下文的复核证据，这条缺口需要在交付说明里明确。
- `AWDServiceStatusPanel.vue` 当前继续保留服务矩阵和 round performance 的 presentation model，规模是合理的；如果后续再叠排序、drilldown 或更复杂交互，更合适继续把 presentation model 切成局部 composable，而不是把 matrix / summary 壳重新拉回父组件。
- 本轮不改 `AWDRoundInspector.vue` 或 `awdInspector.types.ts` 的外部 contract；如果上层以后要重塑服务状态输入结构，不应把这类 contract 变更和本轮展示壳拆分混在一起。

## Touched known-debt status

- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md` 里这条 contest / AWD feature 内残余大 surface，已在 `AWDServiceStatusPanel.vue` 这一块 touched surface 上完成一刀父壳收口；当前 residual 重点已不再是服务状态 toolbar / matrix / summary 壳体本身，而转向 `AWDTrafficPanel.vue`、`AWDAttackLogDialog.vue` 等仍在 feature 内混写 section / CSS 的 surface。
