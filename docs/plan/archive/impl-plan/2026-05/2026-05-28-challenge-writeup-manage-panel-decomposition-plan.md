> 状态：Current
> 事实源：`ChallengeWriteupManagePanel.vue` 当前 owner、`useChallengeWriteupManagement()` workflow owner、题解目录页护栏测试
> 替代：无

# Challenge Writeup Manage Panel Decomposition Plan

## 目标

- 把 `ChallengeWriteupManagePanel.vue` 从“workflow wiring + 多 section 模板 + row action menu + 大段样式”收口成明确的 feature panel owner
- 为题解目录页补齐 header、summary strip、directory section、directory row 和独立 CSS 文件
- 保持平台题目详情里题解目录的 public API、删除行为、分页行为和按钮文案不变

## 非目标

- 本轮不调整 `useChallengeWriteupManagement()` 的 workflow owner
- 本轮不改 `ChallengeWriteupEditorPage.vue`、`ChallengeWriteupViewPage.vue` 或题目详情 route workspace
- 本轮不改变官方题解删除、学员题解分页、open writeup 事件和 action menu 行为

## 输入依据

- `code/frontend/src/features/challenge-writeup-editor/ui/ChallengeWriteupManagePanel.vue`
- `code/frontend/src/features/challenge-writeup-editor/model/useChallengeWriteupManagement.ts`
- `code/frontend/src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts`
- `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `useChallengeWriteupManagement()` 已经是题解目录 workflow 的明确 owner，当前主问题不在 workflow，而在父 SFC 继续承载 header、summary strip、directory row、pagination 和大段 panel 样式。
- `ChallengeWriteupManagePanel.vue` 的脚本 owner 比较清楚：`openWriteup()`、`handleDelete()`、`actionMenuOpen`。继续把稳定展示区块下沉，不会破坏 workflow 边界。
- 测试目前直接读取 `ChallengeWriteupManagePanel.vue?raw`；如果把 section 和样式拆走，必须同步改成聚合源码视角，否则 raw-source、theme-token 和 surface-alignment 护栏会误报。

## 设计边界

### `ChallengeWriteupManagePanel.vue` 本轮继续负责

- `challengeId` / `challengeTitle` props contract
- `openWriteup` emits contract
- `useChallengeWriteupManagement({ challengeId })` workflow wiring
- `actionMenuOpen`
- `openWriteup()` / `closeActionMenu()` / `setActionMenuOpen()` / `handleDelete()`

### `ChallengeWriteupManageHeader.vue` 本轮负责

- 目录页头
- “编写题解”主动作

### `ChallengeWriteupSummaryStrip.vue` 本轮负责

- 官方题解 / 学员题解 summary strip

### `ChallengeWriteupDirectorySection.vue` 本轮负责

- directory heading
- submission loading
- empty state
- directory row 列表组合
- pagination

### `ChallengeWriteupDirectoryRow.vue` 本轮负责

- 单条题解 row 展示
- 官方题解查看按钮
- 官方题解 action menu
- 学员题解 placeholder

### `challengeWriteupManagePanel.css` 本轮负责

- panel shell
- header / summary strip / directory / row / action 样式
- 响应式样式

## 任务切片

### Slice 1：抽出 header / summary / directory row 组合

- 目标：
  - 新增 `ChallengeWriteupManageHeader.vue`
  - 新增 `ChallengeWriteupSummaryStrip.vue`
  - 新增 `ChallengeWriteupDirectorySection.vue`
  - 新增 `ChallengeWriteupDirectoryRow.vue`
  - 父 panel 改为只组合 section
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts`
- Review focus：
  - 父组件是否仍是唯一 workflow / menu / open-delete owner
  - 子组件是否只消费 props / callbacks，不重新接管 delete 或 pagination workflow

### Slice 2：抽出 panel CSS 与 raw-source 护栏

- 目标：
  - 新增 `challengeWriteupManagePanel.css`
  - 更新题解目录页相关 raw-source 护栏为聚合源码视角
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
- Review focus：
  - 样式 owner 是否从父 SFC 收口
  - 护栏是否仍覆盖 workspace title、theme token、surface divider 和 action menu primitive

### Slice 3：同步 backlog、review 与终态验证

- 目标：
  - 更新 backlog 当前进展
  - 补 frontend review
  - 完成类型检查与 harness gate
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/__tests__/architectureBoundaries.test.ts`
  - `cd code/frontend && npm run typecheck`
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - touched surface 上的 manage header / summary / directory / style 混写是否真的收口
  - 当前 feature 内是否留下新的大子组件替代旧大父组件

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeWriteupManagePanel.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `ChallengeWriteupDirectorySection.vue` 仍可能偏大，因为它继续同时承接 loading/empty/list/pagination；如果目录页后续继续增长，下一刀更适合在 feature 内继续把 directory shell 与 row group 再分，而不是回退到父 panel 再混写。
- 当前工作树里已经有一批未提交的 layout P2、class report export 和 writeup editor 收口改动；本轮实现时需要继续按路径保持可拆分提交，不把无关残留混进后续 feature 提交。
