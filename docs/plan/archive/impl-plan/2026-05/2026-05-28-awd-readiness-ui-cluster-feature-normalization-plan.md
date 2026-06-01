> 状态：Current
> 事实源：`AWDReadiness*` 当前 UI owner、`platform-contests` / `contest-awd-admin` readiness 消费边界
> 替代：无

# AWD Readiness UI Cluster Feature Normalization Plan

## 目标

- 新建 `features/awd-readiness`，承接 `AWDReadinessChecklist.vue`、`AWDReadinessDecisionHUD.vue`、`AWDReadinessSummary.vue`、`AWDReadinessOverrideDialog.vue`。
- 让 `ContestAwdPreflightPanel.vue`、`AWDOperationsPanel.vue`、`ContestManage.vue` 改为通过 `@/features/awd-readiness` public API 组合 readiness UI。
- 同步更新组件声明、raw-source 测试、架构事实文档和 backlog 记录。

## 非目标

- 本轮不迁 `AWDInstanceOrchestrationPanel.vue`。
- 本轮不合并 `useAwdStartOverrideFlow.ts` 与 `useAwdReadinessDecision.ts` 的 workflow / dialog state owner。
- 本轮不调整 AWD readiness 后端 contract、阻塞原因 owner 或前端交互文案。

## 输入依据

- `code/frontend/src/components/platform/contest/AWDReadinessChecklist.vue`
- `code/frontend/src/components/platform/contest/AWDReadinessDecisionHUD.vue`
- `code/frontend/src/components/platform/contest/AWDReadinessSummary.vue`
- `code/frontend/src/components/platform/contest/AWDReadinessOverrideDialog.vue`
- `code/frontend/src/features/platform-contests/ui/ContestAwdPreflightPanel.vue`
- `code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue`
- `code/frontend/src/views/platform/ContestManage.vue`
- `docs/architecture/features/AWD开赛就绪门禁设计.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- 这 4 个组件描述的是 AWD readiness capability 的 UI，不属于单一路由或单一 feature 私有面板。
- `ContestAwdPreflightPanel.vue` 只消费 checklist + decision HUD；`AWDOperationsPanel.vue` 与 `ContestManage.vue` 消费 summary / override dialog，复用已经跨 feature 形成 capability。
- 最小正确落点是新建 `features/awd-readiness/ui/*`，而不是继续把 owner 倾向到 `platform-contests` 或 `contest-awd-admin` 任一侧。

## 设计边界

### `features/awd-readiness/*` 本轮负责

- readiness checklist / decision HUD / summary / override dialog 的 UI owner
- readiness UI public API

### `features/platform-contests/*` 本轮继续负责

- contest edit / contest manage 的 route、stage、保存、导航与启动赛事 override workflow owner

### `features/contest-awd-admin/*` 本轮继续负责

- AWD 运维页的 readiness 查询、轮次 / 当前轮检查强制放行 workflow owner
- 实例编排、巡检结果、服务检查与攻击日志 runtime cluster

### `components/platform/contest/*` 本轮继续负责

- `AWDInstanceOrchestrationPanel.vue`
- `AWDServiceCheckDialog.vue`
- `AWDAttackLogDialog.vue`
- `AWDRoundCreateDialog.vue`

## 任务切片

### Slice 1：readiness UI cluster 迁位

- 目标：
  - 新增 `features/awd-readiness/index.ts`
  - 新增 `features/awd-readiness/ui/index.ts`
  - 把 4 个 readiness 组件整体迁入 `features/awd-readiness/ui`
- 验证：
  - `npm run test:run -- src/components/platform/__tests__/AWDReadinessSummary.test.ts`
- Review focus：
  - summary 对 checklist 的内部依赖是否切到 feature 内部相对 import
  - checklist / override dialog props、emits 和展示逻辑是否保持不变

### Slice 2：消费方与护栏同步

- 目标：
  - 更新 `ContestAwdPreflightPanel.vue`、`AWDOperationsPanel.vue`、`ContestManage.vue`
  - 更新 `components.d.ts`、raw-source / surface alignment / dialog adoption 测试引用路径
  - 更新 readiness 架构文档与 backlog 记录
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase2.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase19.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase25.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/components/common/__tests__/BackofficeDialogAdoption.test.ts src/views/platform/__tests__/ContestManage.test.ts src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - `platform-contests` / `contest-awd-admin` 是否都经由 public API 取 readiness UI
  - touched surface 是否不再留下旧 readiness 组件路径

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDReadinessSummary.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase2.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase19.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase25.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase26.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/components/common/__tests__/BackofficeDialogAdoption.test.ts src/views/platform/__tests__/ContestManage.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- readiness UI owner 收口后，`useAwdStartOverrideFlow.ts` 和 `useAwdReadinessDecision.ts` 仍各自维护 dialog state / 强制放行 workflow；后续如果这两条线的文案、state 结构或错误处理继续漂移，需要再单独评估 shared model owner。
- `src/features/__tests__/featureBoundaries.test.ts` 目前是仓库级红 baseline，原因是多个既有 feature 仍直接依赖 `@/components/*` 共享壳体；本轮把 readiness UI owner 收口到了独立 feature，但没有在同一刀内解决这条系统性层级债。
