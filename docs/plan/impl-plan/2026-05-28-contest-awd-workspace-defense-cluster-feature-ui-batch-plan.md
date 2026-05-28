> 状态：Current
> 事实源：`ContestAWDWorkspacePanel.vue` 当前 owner、`contest-awd-workspace` 既有 feature model / UI 收口模式
> 替代：无

# Contest AWD Workspace Defense Cluster Feature UI Batch Plan

## 目标

- 把 `AWDDefenseColumn.vue`、`AWDDefenseAlertsPanel.vue`、`AWDDefenseOperationsPanel.vue`、`AWDDefenseConnectionPanel.vue`、`AWDDefenseServiceList.vue` 迁入 `features/contest-awd-workspace/ui`。
- 让 `ContestAWDWorkspacePanel.vue` 改为 feature 内部相对 import。
- 同步更新 feature UI 导出、allowlist、raw-source 护栏、`components.d.ts` 与 backlog / review 事实源。

## 非目标

- 本轮不调整 `ContestAWDWorkspacePanel.vue` 的数据加载、轮询、服务动作、SSH access 或攻击向量 owner。
- 本轮不继续迁 `AWDWorkspaceHudStrip.vue`、`AWDWorkspaceIntelColumn.vue`、`AWDAttackVectorPanel.vue`。
- 本轮不重构 `AWDDefenseOperationsPanel.vue`、`AWDDefenseServiceList.vue` 内部样式或交互，只做 owner 迁位。

## 输入依据

- `code/frontend/src/components/contests/awd/AWDDefenseColumn.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseAlertsPanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseOperationsPanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseConnectionPanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseServiceList.vue`
- `code/frontend/src/features/contest-awd-workspace/ui/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/features/contest-awd-workspace/index.ts`
- `code/frontend/src/features/contest-awd-workspace/ui/index.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/components.d.ts`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `code/frontend/src/views/contests/__tests__/contestStudentActionPrimitives.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `ContestAWDWorkspacePanel.vue` 已经持有防守区的 page / workflow owner，左列三件套只是稳定 UI cluster。
- 这组三件套只被 `ContestAWDWorkspacePanel.vue` 组合消费，不是跨 feature 共享 contest primitive。
- 最小正确改动是整体迁位到 `features/contest-awd-workspace/ui`，并通过 feature 内部相对 import / public API 收口。

## 设计边界

### `features/contest-awd-workspace/ui/*` 本轮负责

- defense column
- defense alerts panel
- defense operations panel
- defense connection panel
- defense service list
- `ContestAWDWorkspacePanel.vue` 对 defense cluster 的 feature 内部组合

### `features/contest-awd-workspace/model/*` 本轮继续负责

- AWD workspace 数据加载、刷新与动作 owner
- defense service selection / access / presentation / summary
- attack vector 与 event presentation owner

### `views/contests/ContestDetail.vue` 本轮不负责

- 本轮不触达 route shell；AWD workspace 仍由 `ContestDetail.vue` 通过既有 feature public API 组合

## 任务切片

### Slice 1：defense cluster 迁位

- 目标：
  - 新增 `features/contest-awd-workspace/ui/AWDDefense*.vue`
  - `ContestAWDWorkspacePanel.vue` 改为 feature 内部相对 import
  - `features/contest-awd-workspace/ui/index.ts` 补 defense cluster export
  - `AWDDefenseConnectionPanel.test.ts` 跟随组件 owner 迁到 feature UI tests
- 验证：
  - `npm run test:run -- src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts src/features/contest-awd-workspace/ui/__tests__/AWDDefenseConnectionPanel.test.ts`
- Review focus：
  - `ContestAWDWorkspacePanel.vue` 是否仍只保留 workspace 组合 owner
  - 迁位后 props / emits contract 是否保持不变

### Slice 2：护栏与事实源同步

- 目标：
  - 更新 `architectureAllowlist.ts`、`components.d.ts`
  - 更新 raw-source 测试与 backlog / review 文档中的当前路径事实
- 验证：
  - `npm run test:run -- src/views/contests/__tests__/contestStudentActionPrimitives.test.ts src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - 旧 `components/contests/awd/AWDDefense*.vue` 路径是否已从 touched surface 消失
  - `componentFeatureImportAllowlist` 是否只剩 layout / topology model consumer 残留

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts src/views/contests/__tests__/contestStudentActionPrimitives.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `contest-awd-workspace` 在这轮之后仍会保留 `AWDAttackVectorPanel.vue`、`AWDWorkspaceHudStrip.vue`、`AWDWorkspaceIntelColumn.vue` 等 UI 在旧目录；如果后续确认它们也只服务单一 feature，应继续按独立切片判断 owner。
