# AWD Instance Orchestration Panel Decomposition 复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - plan：`docs/plan/impl-plan/2026-05-29-awd-instance-orchestration-panel-decomposition-plan.md`
  - files reviewed：
    - `.harness/reuse-decisions/awd-instance-orchestration-panel-decomposition.md`
    - `docs/plan/impl-plan/2026-05-29-awd-instance-orchestration-panel-decomposition-plan.md`
    - `docs/reviews/frontend/2026-05-29-awd-instance-orchestration-panel-decomposition-review.md`
    - `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
    - `code/frontend/src/features/contest-awd-admin/ui/AWDInstanceOrchestrationPanel.vue`
    - `code/frontend/src/features/contest-awd-admin/ui/AWDInstanceOrchestrationHeader.vue`
    - `code/frontend/src/features/contest-awd-admin/ui/AWDInstanceOrchestrationMatrix.vue`
    - `code/frontend/src/features/contest-awd-admin/ui/AWDInstanceOrchestrationRow.vue`
    - `code/frontend/src/features/contest-awd-admin/ui/awdInstanceOrchestrationPanel.css`
    - `code/frontend/src/features/contest-awd-admin/ui/awdInstanceOrchestration.types.ts`
    - `code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts`
    - `code/frontend/src/components/platform/__tests__/AWDInstanceOrchestrationPanelExtraction.test.ts`
    - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
- Classification check：同意按 `contest-awd-admin` feature 内大 surface 收口处理，属于非 trivial frontend refactor。
- Gate verdict：Pass with minor issues

## Findings

- 无新的 blocker finding。

## Material findings

- 无。

## Senior implementation assessment

- `AWDInstanceOrchestrationPanel.vue` 现在只保留 `orchestration/loading/startingKey` contract、实例 map、可见服务筛选、running summary、starting key 判定，以及 header / matrix / row view model 组装，不再继续直接内联 header、empty state、table row 和 scoped CSS。
- `AWDInstanceOrchestrationHeader.vue` 已明确承接标题、running summary、refresh 与 start-all action primitive。
- `AWDInstanceOrchestrationMatrix.vue` 已明确承接 empty state、table shell 和 row 列表组合。
- `AWDInstanceOrchestrationRow.vue` 已明确承接队伍行、服务单元格状态 / 访问链接 / 启动按钮，以及行级“启动本队”动作。
- `awdInstanceOrchestrationPanel.css` 已承接 panel shell、header、summary、matrix、row、status、action 与响应式样式，父面板不再保留 scoped style 尾巴。
- 实例编排相关 raw-source 护栏已统一切到 panel + 子组件 + CSS 的聚合源码视角，继续覆盖 extraction 边界与 ops action primitive。
- `AWDInstanceOrchestrationPanel.vue` 文件体量从约 `464` 行降到 `145` 行；本轮 touched surface 上“feature owner 正确但父层仍混放 header / matrix / row / CSS”的债已经完成收口。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDOperationsPanel.test.ts src/components/platform/__tests__/AWDInstanceOrchestrationPanelExtraction.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## Residual risk

- 这份 review 仍是同上下文 self-review；按 `development-pipeline` 的独立 gate 要求，依然缺少单独 reviewer 上下文的复核证据，这条缺口需要在交付说明里明确。
- `AWDInstanceOrchestrationPanel.vue` 当前继续保留实例 view model 和 starting key 判定，规模是合理的；如果后续再叠加筛选、批量状态或更多控制动作，更合适继续把 presentation model 拆成局部 composable，而不是把 table / row 壳重新拉回父组件。
- 本轮不改 `usePlatformContestAwd()` 的实例编排 workflow；如果上层再新增更多实例编排行为，不应把请求 / 异常 owner 下沉到 row 或 matrix 子组件。

## Touched known-debt status

- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md` 里这条 contest / AWD feature 内残余大 surface，已在 `AWDInstanceOrchestrationPanel.vue` 这一块 touched surface 上完成一刀父壳收口；当前 residual 重点已不再是实例编排 table 壳体本身，而转向其它仍在 feature 内混写 workflow owner / section / CSS 的 surface。
