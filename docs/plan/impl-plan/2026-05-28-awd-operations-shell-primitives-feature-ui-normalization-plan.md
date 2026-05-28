> 状态：Current
> 事实源：`AWDContestSelectorField.vue`、`AWDRuntimePendingState.vue` 当前 owner、`contest-awd-admin` 既有 feature-owned UI 收口模式
> 替代：无

# AWD Operations Shell Primitives Feature UI Normalization Plan

## 目标

- 把 `AWDContestSelectorField.vue`、`AWDRuntimePendingState.vue` 迁入 `features/contest-awd-admin/ui`。
- 让 `AWDOperationsPanel.vue` 改为 feature 内部相对 import。
- 同步更新组件声明、相关 raw-source / theme token 测试和 backlog 记录。

## 非目标

- 本轮不调整 `AWDOperationsPanel.vue` 的 tab state、运行态 workflow 或 page owner。
- 本轮不把 selector / pending state 抽成 shared 层通用组件。
- 本轮不继续处理 `contest-awd-admin` 以外的 AWD 路由或 capability。

## 输入依据

- `code/frontend/src/components/platform/contest/AWDContestSelectorField.vue`
- `code/frontend/src/components/platform/contest/AWDRuntimePendingState.vue`
- `code/frontend/src/features/contest-awd-admin/ui/AWDOperationsPanel.vue`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `code/frontend/src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- 这两个子件都只服务 `AWDOperationsPanel.vue`，属于 `contest-awd-admin` 单一 feature 的 operations shell primitives。
- 它们没有像 readiness UI 那样的跨 feature 复用面，不需要新建独立 capability feature。
- 最小正确落点是 `features/contest-awd-admin/ui/*`。

## 设计边界

### `features/contest-awd-admin/ui/*` 本轮负责

- contest selector field
- runtime pending state
- `AWDOperationsPanel.vue` 对这两个子件的 feature 内部组合

### `features/contest-awd-admin/model/*` 本轮继续负责

- contest 选择
- runtime readiness / pending stage 判定
- tab 切换与后续运维动作 owner

### 本轮不动

- `features/awd-inspector/*`
- `features/awd-readiness/*`
- `AWDOperationsPanel.vue` 内部其它已迁位子件

## 任务切片

### Slice 1：operations shell primitives 迁位

- 目标：
  - 把 `AWDContestSelectorField.vue`、`AWDRuntimePendingState.vue` 迁入 `features/contest-awd-admin/ui`
  - `AWDOperationsPanel.vue` 改为 feature 内部相对 import
- 验证：
  - `npm run test:run -- src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - props / emits contract 是否保持不变
  - runtime pending state 是否仍只负责展示，不吞掉 workflow owner

### Slice 2：护栏与引用同步

- 目标：
  - 更新 `components.d.ts`
  - 更新 `contestUiPrimitiveAdoptionPhase4.test.ts` 与 `sharedThemeTokenAdoption.test.ts` 的 raw-source 引用路径
  - backlog 记录 `contest-awd-admin` operations shell 子件进展
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts`
- Review focus：
  - touched surface 是否不再留下旧 selector / pending state 路径
  - theme token 护栏是否继续生效

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/components/platform/__tests__/awdOperationsPanelTabsExtraction.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮只收 operations shell primitives owner，不继续处理 `contest-awd-admin` 外围更高层的跨 feature 结构耦合；如果后续继续下钻，需要回到更高层 backlog 单独分刀。
