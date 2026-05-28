> 状态：Current
> 事实源：`AWDRoundCreateDialog.vue`、`AWDServiceCheckDialog.vue`、`AWDAttackLogDialog.vue` 当前 owner、`contest-awd-admin` 既有 feature-owned UI 收口模式
> 替代：无

# AWD Runtime Dialog Cluster Feature UI Normalization Plan

## 目标

- 把 `AWDRoundCreateDialog.vue`、`AWDServiceCheckDialog.vue`、`AWDAttackLogDialog.vue` 迁入 `features/contest-awd-admin/ui`。
- 让 `AWDOperationsPanel.vue` 改为 feature 内部相对 import。
- 同步更新组件声明、相关 raw-source / duplicate-action / dialog adoption 测试和 backlog 记录。

## 非目标

- 本轮不调整 `usePlatformContestAwd()` 的创建轮次、录入服务检查、补录攻击日志 workflow owner。
- 本轮不继续拆 dialog 内部表单逻辑或抽 shared form composable。
- 本轮不处理 `contest-awd-admin` 之外的 AWD route / capability 边界。

## 输入依据

- `code/frontend/src/components/platform/contest/AWDRoundCreateDialog.vue`
- `code/frontend/src/components/platform/contest/AWDServiceCheckDialog.vue`
- `code/frontend/src/components/platform/contest/AWDAttackLogDialog.vue`
- `code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts`
- `code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts`
- `code/frontend/src/components/common/__tests__/BackofficeDialogAdoption.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- 这 3 个 dialog 都只服务 `AWDOperationsPanel.vue`，属于 `contest-awd-admin` 单一 feature 的 runtime/dialog cluster。
- 它们不像 `AWDReadiness*` 那样有跨 feature 复用面，不需要新建独立 capability feature。
- 最小正确落点是 `features/contest-awd-admin/ui/*`。

## 设计边界

### `features/contest-awd-admin/ui/*` 本轮负责

- round create dialog
- service check dialog
- attack log dialog
- `AWDOperationsPanel.vue` 对这 3 个 dialog 的 feature 内部组合

### `features/contest-awd-admin/model/*` 本轮继续负责

- 创建轮次
- 录入服务检查
- 补录攻击日志
- dialog open state / saving state / submit handlers owner

### 其它模块本轮不动

- `features/awd-readiness/*`
- `features/awd-inspector/*`
- `components/platform/contest/AWDContestSelectorField.vue`
- `components/platform/contest/AWDRuntimePendingState.vue`

## 任务切片

### Slice 1：runtime dialogs 迁位

- 目标：
  - 把 3 个 dialog 迁入 `features/contest-awd-admin/ui`
  - `AWDOperationsPanel.vue` 改为 feature 内部相对 import
- 验证：
  - `npm run test:run -- src/components/platform/__tests__/AWDOperationsPanel.test.ts`
- Review focus：
  - dialog props / emits contract 是否保持不变
  - workflow owner 是否仍留在 `AWDOperationsPanel.vue` / `usePlatformContestAwd()`

### Slice 2：护栏与引用同步

- 目标：
  - 更新 `components.d.ts`
  - 更新 raw-source / duplicate-action / dialog adoption 测试引用路径
  - backlog 记录 `contest-awd-admin` runtime dialog cluster 进展
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts src/views/__tests__/duplicateActionGuardAudit.test.ts src/components/common/__tests__/BackofficeDialogAdoption.test.ts src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - touched surface 是否不再留下旧 runtime dialog 路径
  - duplicate-action / dialog adoption guardrail 是否继续生效

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDOperationsPanel.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts src/views/__tests__/duplicateActionGuardAudit.test.ts src/components/common/__tests__/BackofficeDialogAdoption.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮只收 runtime dialog owner，不继续清 `AWDContestSelectorField.vue`、`AWDRuntimePendingState.vue` 等仍在旧目录的 AWD operations 子件；如果后续继续下钻，应再按 `contest-awd-admin` owner 分刀。
