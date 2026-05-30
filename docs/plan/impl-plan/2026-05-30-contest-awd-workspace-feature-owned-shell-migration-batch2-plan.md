> 状态：In Progress
> 事实源：`code/frontend/src/features/contest-awd-workspace/**`、`code/frontend/src/components/contests/awd/*`
> 替代：无

# Contest AWD Workspace Feature-Owned Shell 第二批迁移计划

## 目标

- 把 `components/contests/awd/*` 中仍只服务学生 AWD workspace 的历史壳，整体收口到 `features/contest-awd-workspace/ui/*`。

## 非目标

- 不调整 `contest-awd-workspace/model/*` 的状态 owner、轮询或动作流程。
- 不改 `ContestAWDWorkspacePanel.vue` 的功能语义。
- 不碰 `contest-awd-admin`、`contest-awd-config` 相关 feature。

## 输入依据

- `code/frontend/src/features/contest-awd-workspace/ui/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/features/contest-awd-workspace/ui/contestAwdWorkspaceUiStrategy.test.ts`
- `code/frontend/src/components/contests/awd/*.vue`
- `code/frontend/src/components/contests/awd/__tests__/AWDDefenseFileWorkbench.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `ContestAWDWorkspacePanel.vue` 是这批 `AWDAttack*`、`AWDWorkspace*` 壳的唯一运行时 consumer。
- `AWDDefenseFileWorkbench.vue` 当前虽然暂无运行时 consumer，但语义仍然是 AWD workspace feature 私有面板，不应继续留在全局历史组件目录。
- 最小正确切片是：补齐 feature UI owner、改 `ContestAWDWorkspacePanel.vue` 与 raw-source 测试、迁移邻近测试、删旧文件。

## 设计边界

### `features/contest-awd-workspace/ui` 本轮负责

- `AWDAttackResultFooter`
- `AWDAttackTargetGrid`
- `AWDAttackToolbar`
- `AWDAttackVectorPanel`
- `AWDDefenseFileWorkbench`
- `AWDWorkspaceHudStrip`
- `AWDWorkspaceIntelColumn`
- `ContestAWDWorkspacePanel` 对这批子壳的组合

### `features/contest-awd-workspace/model` 本轮继续负责

- attack / defense / summary / presentation workflow owner
- 远程调用、复制、打开、提交、刷新动作

## 任务切片

- [ ] Slice 1：迁移 AWD workspace 子壳到 feature UI
  - 目标：
    - 新建 `features/contest-awd-workspace/ui/*`
    - `ContestAWDWorkspacePanel.vue` 改为 feature 内部相对 import
    - 更新 `features/contest-awd-workspace/ui/index.ts`
  - 验证：
    - `rg -n "@/components/contests/awd" code/frontend/src/features/contest-awd-workspace`

- [ ] Slice 2：迁移邻近测试和类型声明
  - 目标：
    - `AWDDefenseFileWorkbench.test.ts` 迁到 feature UI 邻近测试
    - `contestAwdWorkspaceUiStrategy.test.ts` 与 `components.d.ts` 对齐新 owner
  - 验证：
    - `pnpm vitest run src/features/contest-awd-workspace/ui/contestAwdWorkspaceUiStrategy.test.ts src/features/contest-awd-workspace/ui/__tests__/AWDDefenseFileWorkbench.test.ts`

- [ ] Slice 3：删除旧 AWD 组件并更新 backlog
  - 目标：
    - 删除旧 `components/contests/awd/*`
    - 同步 backlog 当前事实
  - 验证：
    - `bash scripts/check-frontend-architecture.sh --quick`

## 验证计划

- `python3 harness/checks/check-reuse-decision.py`
- `bash scripts/check-task-intake.sh --reuse-decision contest-awd-workspace-feature-owned-shell-migration-batch2`
- `cd code/frontend && pnpm vitest run src/features/contest-awd-workspace/ui/contestAwdWorkspaceUiStrategy.test.ts src/features/contest-awd-workspace/ui/__tests__/AWDDefenseFileWorkbench.test.ts src/__tests__/architectureBoundaries.test.ts src/__tests__/routePageArchitectureBoundary.test.ts`
- `bash scripts/check-frontend-architecture.sh --quick`

## 残余风险

- 这轮之后，若 `components/contests` 仍有残余，就不应再是 AWD workspace 这一支。
- `AWDDefenseFileWorkbench` 当前主要靠邻近测试兜住 owner；后续若重新接回运行时 consumer，应继续保持在 `features/contest-awd-workspace/ui`。
