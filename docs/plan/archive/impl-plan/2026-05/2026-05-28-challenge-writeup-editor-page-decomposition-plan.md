> 状态：Current
> 事实源：`ChallengeWriteupEditorPage.vue` 当前 owner、`useChallengeWriteupEditorPage()` workflow owner、题解编辑页护栏测试
> 替代：无

# Challenge Writeup Editor Page Decomposition Plan

## 目标

- 把 `ChallengeWriteupEditorPage.vue` 从“page shell + 多 section 模板 + 大段样式”收口成明确的 feature page shell owner
- 为题解编辑页补齐 editor form section、snapshot section、challenge rail 和独立 CSS 文件
- 保持平台题解管理入口与嵌入态编辑页的 public API、保存/删除/推荐行为和页面文案不变

## 非目标

- 本轮不调整 `useChallengeWriteupEditorPage()` 的 workflow owner
- 本轮不改 `ChallengeWriteupManagePanel.vue`、`ChallengeWriteupViewPage.vue` 或路由壳组合方式
- 本轮不改变题解保存、删除、推荐、恢复已保存版本的数据源和提示策略

## 输入依据

- `code/frontend/src/features/challenge-writeup-editor/ui/ChallengeWriteupEditorPage.vue`
- `code/frontend/src/features/challenge-writeup-editor/model/useChallengeWriteupEditorPage.ts`
- `code/frontend/src/views/platform/__tests__/ChallengeWriteup.test.ts`
- `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `useChallengeWriteupEditorPage()` 已经是题解编辑 workflow 的明确 owner，当前主问题不在 workflow，而在父 SFC 继续承载 editor form、snapshot、challenge rail 和大段 page 样式。
- `ChallengeWriteupEditorPage.vue` 的脚本 owner 比较清楚：`embedded` / `back` contract、shell 组合与 composable 装配。继续把稳定展示区块下沉，不会破坏 workflow 边界。
- 测试目前直接读取 `ChallengeWriteupEditorPage.vue?raw`；如果把 section 和样式拆走，必须同步改成聚合源码视角，否则 raw-source 护栏会误报。

## 设计边界

### `ChallengeWriteupEditorPage.vue` 本轮继续负责

- `challengeId` / `embedded` props contract
- `back` emits contract
- topbar、`PageHeader` 与 embedded heading 组合
- `useChallengeWriteupEditorPage(props.challengeId)` workflow wiring
- page shell class 与 `main/aside` 布局组合

### `ChallengeWriteupEditorFormSection.vue` 本轮负责

- editor header、badge
- title / visibility / content 表单壳
- visibility note
- save / toggle recommendation / restore / delete 动作入口

### `ChallengeWriteupSnapshotSection.vue` 本轮负责

- 已保存版本 snapshot grid
- 无管理员题解时的 empty state

### `ChallengeWriteupChallengeRail.vue` 本轮负责

- challenge meta rail
- challenge loading fallback 文案

### `challengeWriteupEditorPage.css` 本轮负责

- page shell 变量
- topbar / embedded heading / workspace layout
- editor form / snapshot / rail 样式
- 响应式样式

## 任务切片

### Slice 1：抽出三个稳定 section

- 目标：
  - 新增 `ChallengeWriteupEditorFormSection.vue`
  - 新增 `ChallengeWriteupSnapshotSection.vue`
  - 新增 `ChallengeWriteupChallengeRail.vue`
  - 父页改为只组合 section
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeWriteup.test.ts`
- Review focus：
  - 父组件是否仍是唯一 shell / workflow wiring owner
  - 子组件是否只消费 props / action handler，不重新接管 workflow

### Slice 2：抽出 page CSS 与 raw-source 护栏

- 目标：
  - 新增 `challengeWriteupEditorPage.css`
  - 更新题解编辑页相关 raw-source 护栏为聚合源码视角
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeWriteup.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts`
- Review focus：
  - 样式 owner 是否从父 SFC 收口
  - 护栏是否仍覆盖 `PageHeader`、`ui-btn` 原语和 workspace title/copy 结构

### Slice 3：同步 backlog、review 与终态验证

- 目标：
  - 更新 backlog 当前进展
  - 补 frontend review
  - 完成类型检查与 harness gate
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeWriteup.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/__tests__/architectureBoundaries.test.ts`
  - `cd code/frontend && npm run typecheck`
  - `cd /home/azhi/workspace/projects/ctf && git diff --check`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
  - `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`
- Review focus：
  - touched surface 上的 editor/snapshot/rail/style 混写是否真的收口
  - 当前 feature 内是否留下新的大子组件替代旧大父组件

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengeWriteup.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `ChallengeWriteupEditorFormSection.vue` 仍可能偏大，因为它同时承接表单区和 editor actions；如果题解编辑页继续增长，下一刀更适合在 feature 内按“form shell / action strip”继续分，而不是回退到父页再混写。
- 当前工作树里已经有一批未提交的 layout P2 和 class report export 收口改动；本轮实现时需要继续按路径保持可拆分提交，不把无关残留混进后续 feature 提交。
