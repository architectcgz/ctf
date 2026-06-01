> 状态：Current
> 事实源：`ContestChallengeOrchestrationPanel.vue` 当前 orchestration owner、`useContestChallengeOrchestration.ts`、现有 extraction / primitive 护栏
> 替代：无

# Contest Challenge Orchestration Panel Decomposition Plan

## 目标

- 把 `ContestChallengeOrchestrationPanel.vue` 从“model owner + header + directory table + 大段样式”收口成明确的 orchestration owner
- 在 `features/contest-workbench/ui` 内补齐 orchestration header / directory section cluster
- 保持对外 props、事件行为和既有测试语义不变，让 `ContestEditWorkspacePanel.vue` 与现有挂载测试继续按原 contract 工作

## 非目标

- 本轮不改 `useContestChallengeOrchestration.ts` 的数据加载、保存、删除和对话框状态 owner
- 本轮不改变 `ContestChallengeEditorDialog.vue` 的 form / selection / save contract
- 本轮不顺手处理 `ContestProjectorAttackMap.vue`、`AWDOperationsPanel.vue` 或 `AWDChallengeConfigPanel.vue`

## 输入依据

- `code/frontend/src/features/contest-workbench/ui/ContestChallengeOrchestrationPanel.vue`
- `code/frontend/src/features/contest-workbench/model/useContestChallengeOrchestration.ts`
- `code/frontend/src/features/contest-workbench/ui/ContestChallengeEditorDialog.vue`
- `code/frontend/src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts`
- `code/frontend/src/components/platform/__tests__/contestChallengeOrchestrationExtraction.test.ts`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase2.test.ts`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase24.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `ContestChallengeOrchestrationPanel.vue` 的真实行为 owner 在 `useContestChallengeOrchestration.ts`，父组件只是桥接这个 model 和 dialog
- 当前 header、题目目录表、空态 / 加载态与行操作 DOM 结构已经稳定，适合继续拆成 feature 内局部子组件
- 既有测试对用户可见契约已有较好覆盖，本轮应优先复用这些护栏，并把 raw-source 检查改成聚合源码

## 设计边界

### `ContestChallengeOrchestrationPanel.vue` 本轮继续负责

- `useContestChallengeOrchestration()` model wiring
- `ContestChallengeSummaryStrip` / `ContestChallengeFilterStrip` 的组合顺序
- `ContestChallengeEditorDialog` props / emits 桥接
- `openActionMenuId`、`editingChallenge`、`dialogOpen` 等 model state 与子区块的桥接

### `ContestChallengeOrchestrationHeader.vue` 本轮负责

- 标题、副标题和顶部主操作按钮
- refresh / create 动作的纯展示层发射

### `ContestChallengeDirectorySection.vue` 本轮负责

- 加载态、空态、目录表与行操作渲染
- 题目标题 route link、可见性 / 分值 / 顺序展示
- edit / remove / update:openActionMenuId 这些纯展示层事件桥接

### 本轮不动

- orchestration model 的 API 调用与异常处理
- dialog 内部题目表单与 AWD 题库选择逻辑
- `ContestEditWorkspacePanel.vue` 的 route shell owner

## 任务切片

### Slice 1：提取 orchestration header 与 directory section

- 目标：
  - 新增 `ContestChallengeOrchestrationHeader.vue`
  - 新增 `ContestChallengeDirectorySection.vue`
  - 父面板改为只组合 header / summary / filter / directory / dialog
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts`
- Review focus：
  - 父组件是否仍然是唯一 model wiring / dialog owner
  - 子组件是否只消费 props 和回发事件，没有偷偷接入 API / composable

### Slice 2：更新 raw-source 护栏到聚合源码

- 目标：
  - 调整 extraction / primitive adoption 测试，适配 header / directory section 下沉后的聚合源码检查
  - 保持 shared UI primitive 和既有 summary strip 护栏不回退
- 验证：
  - `cd code/frontend && npm run test:run -- src/components/platform/__tests__/contestChallengeOrchestrationExtraction.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase2.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase24.test.ts`
- Review focus：
  - 新测试是否仍能真正约束 parent owner，而不是被拆分后失效
  - 没有把 UI primitive / row action / action menu 的约束遗漏到子组件之外

### Slice 3：backlog 与 review 收口

- 目标：
  - 更新 backlog 里 `ContestChallengeOrchestrationPanel.vue` 的状态
  - 补 frontend review 记录
- 验证：
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- Review focus：
  - touched surface 上的超大 orchestration panel 债是否真的收口
  - 没有把 owner 从父面板转移成新的 feature 内大组件

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts src/components/platform/__tests__/contestChallengeOrchestrationExtraction.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase2.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase24.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 如果目录 section 拆完后 `ContestChallengeEditorDialog.vue` 又继续膨胀，本轮不会顺手处理；它需要单独按 dialog 内部 owner 再开一刀。
- 若后续 AWD challenge config 的展示再次并回到这条面板，需要继续坚持“父层保留 orchestration model，具体展示分区下沉”的边界，不回退到单文件堆模板。
