> 状态：Current
> 事实源：`AWDInstanceOrchestrationPanel.vue` 当前 owner、`contest-awd-admin` 既有 feature-owned UI 收口模式
> 替代：无

# AWD Instance Orchestration Panel Feature UI Normalization Plan

## 目标

- 把 `AWDInstanceOrchestrationPanel.vue` 迁入 `features/contest-awd-admin/ui`。
- 让 `AWDOperationsPanel.vue` 改为 feature 内部相对 import。
- 同步更新组件声明、相关测试和 backlog 记录。

## 非目标

- 本轮不迁 `AWDServiceCheckDialog.vue`、`AWDAttackLogDialog.vue`、`AWDRoundCreateDialog.vue`。
- 本轮不调整 `usePlatformContestAwd()` 的实例编排 workflow、刷新策略或按钮交互逻辑。
- 本轮不继续处理 `contest-awd-admin` 里的其他 runtime 子件 cluster。

## 输入依据

- `code/frontend/src/components/platform/contest/AWDInstanceOrchestrationPanel.vue`
- `code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue`
- `code/frontend/src/components/platform/__tests__/AWDOperationsPanel.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `AWDInstanceOrchestrationPanel.vue` 只服务 `AWDOperationsPanel.vue`，是 `contest-awd-admin` 单一 feature 的运行态子 panel。
- 它不具备像 `AWDReadiness*` 那样的跨 feature 共享面，不需要新建独立 capability feature。
- 最小正确落点是 `features/contest-awd-admin/ui/AWDInstanceOrchestrationPanel.vue`。

## 设计边界

### `features/contest-awd-admin/ui/*` 本轮负责

- instance orchestration panel owner
- `AWDOperationsPanel.vue` 对该 panel 的 feature 内部组合

### `components/platform/contest/*` 本轮继续负责

- `AWDServiceCheckDialog.vue`
- `AWDAttackLogDialog.vue`
- `AWDRoundCreateDialog.vue`
- 其它未切分的 AWD runtime 子件

### `features/contest-awd-admin/model/*` 本轮继续负责

- 实例编排数据加载、启动动作、刷新动作和 pending key owner

## 任务切片

### Slice 1：instance orchestration panel 迁位

- 目标：
  - 新增 `features/contest-awd-admin/ui/AWDInstanceOrchestrationPanel.vue`
  - `AWDOperationsPanel.vue` 改为 feature 内部相对 import
- 验证：
  - `npm run test:run -- src/components/platform/__tests__/AWDOperationsPanel.test.ts`
- Review focus：
  - props / emits contract 是否保持不变
  - runtime panel owner 是否仍留在 `AWDOperationsPanel` / `usePlatformContestAwd()`

### Slice 2：护栏与引用同步

- 目标：
  - 更新 `components.d.ts`
  - backlog 记录 `contest-awd-admin` 线上剩余 runtime 子件进展
- 验证：
  - `npm run test:run -- src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - touched surface 是否不再留下旧 `AWDInstanceOrchestrationPanel.vue` 路径

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDOperationsPanel.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `contest-awd-admin` 运行态子件里仍有 `AWDServiceCheckDialog.vue`、`AWDAttackLogDialog.vue`、`AWDRoundCreateDialog.vue` 等 legacy component 路径；如果后续继续清这条线，应按 runtime / dialog cluster 再单独分刀。
