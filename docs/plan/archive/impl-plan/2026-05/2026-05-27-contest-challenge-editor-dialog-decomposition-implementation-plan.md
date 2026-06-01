> 状态：Current
> 事实源：`ContestChallengeEditorDialog.vue` 当前 owner、`ContestChallengeOrchestrationPanel.vue` 调用面、现有 AWD 目录 section 拆分模式与对话框测试护栏
> 替代：无

# Contest Challenge Editor Dialog Decomposition Implementation Plan

## 目标

- 拆分 `code/frontend/src/components/platform/contest/ContestChallengeEditorDialog.vue`，把 AWD 题目目录选择区和题目设置区从单一超大对话框里拆成稳定 section。
- 保持 `ContestChallengeEditorDialog.vue` 继续拥有 `form`、selection state、校验、submit 和 `ContestChallengeOrchestrationPanel.vue` 的事件桥接。
- 同步更新该对话框相关的 raw-source 护栏和行为测试，让后续比赛编排需求不再继续堆回同一个 899 行组件。

## 非目标

- 本轮不改变 `ContestChallengeOrchestrationPanel.vue` 的数据加载、AWD 题目目录查询、分页和保存流程 owner。
- 本轮不改变赛事题目保存 payload、AWD 题目多选行为、普通赛事题目选择行为和用户可见文案。
- 本轮不引入新的 feature composable，也不把表单 owner 从对话框父组件迁移出去。

## 输入依据

- `code/frontend/src/components/platform/contest/ContestChallengeEditorDialog.vue`
- `code/frontend/src/components/platform/contest/ContestChallengeOrchestrationPanel.vue`
- `code/frontend/src/components/platform/awd-service/AwdChallengeLibrarySection.vue`
- `code/frontend/src/components/platform/awd-service/AWDChallengeEditorDialog.vue`
- `code/frontend/src/components/platform/__tests__/ContestChallengeEditorDialog.test.ts`
- `code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts`
- `code/frontend/src/views/__tests__/studentDirectoryTypographyBoundary.test.ts`
- `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts`
- `code/frontend/src/components/common/__tests__/BackofficeDialogAdoption.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `ContestChallengeEditorDialog.vue` 当前约 `899` 行，同时承担普通题目选择、AWD 题目目录筛选/分页/表格选择、分值顺序可见性设置、submit 校验和大段局部样式，已经是典型的超大对话框壳。
- `ContestChallengeOrchestrationPanel.vue` 已经拥有 AWD 目录数据、筛选值、分页值和保存动作的真实 owner，因此最小正确收口点不是改 orchestration feature，而是把对话框内部展示分区抽出来。
- 这个对话框既有行为测试，也有多条 raw-source 护栏，所以拆分时必须同步把断言改成“父对话框 + 子 section”组合视角。

## 设计边界

### 父对话框继续负责

- `form.challenge_id / awd_challenge_id / awd_challenge_ids / points / order / is_visible`
- `fieldErrors`、`clearErrors()`、`submit()` 和 `if (props.saving)` 的重复提交短路
- `showContestSelector / showAwdChallengeSelector / showContestSettings` 这类 mode / contestMode 驱动的 owner
- AWD 题目选择状态 `selectAwdChallenge()` / `isAwdChallengeSelected()`

### 子组件负责

- `ContestAwdChallengeSelectorSection.vue`
  - 负责 AWD 题目目录筛选栏、加载/错误/空态、表格、上一页/下一页按钮和选择按钮的展示
  - 不拥有实际筛选 state、分页 state 或保存逻辑
- `ContestChallengeSettingsSection.vue`
  - 负责普通题目选择只读/下拉展示、分值/顺序/可见性字段和字段错误展示
  - 不拥有 submit、校验或 props.draft 同步逻辑

### 本轮明确不负责

- 子组件不直接调用 `ContestChallengeOrchestrationPanel` 的方法，不持有 `reactive form` 或 `watch(props.draft)`。
- 本轮不改 `ContestChallengeOrchestrationPanel.vue` 的 API / feature owner。
- 本轮不触碰 `ContestAwdConfigWorkspaceShell.vue`。

## 任务切片

### Slice 1：提取 AWD 题目目录选择区

- 目标：
  - 从 `ContestChallengeEditorDialog.vue` 提取 AWD 目录筛选、表格、分页和错误/空态展示。
  - 保留 AWD 题目选择状态与筛选 emit owner 在父对话框。
- 预期改动：
  - `code/frontend/src/components/platform/contest/ContestChallengeEditorDialog.vue`
  - `code/frontend/src/components/platform/contest/ContestAwdChallengeSelectorSection.vue`
  - `code/frontend/src/components/platform/__tests__/ContestChallengeEditorDialog.test.ts`
  - `code/frontend/src/views/__tests__/studentDirectoryTypographyBoundary.test.ts`
- 验证：
  - `npm run test:run -- src/components/platform/__tests__/ContestChallengeEditorDialog.test.ts src/views/__tests__/studentDirectoryTypographyBoundary.test.ts`
- Review focus：
  - AWD 目录区是否没有接管 selection / query owner
  - `WorkspaceDataTable` 标题列纯净性护栏是否仍然成立

### Slice 2：提取题目设置区

- 目标：
  - 把普通题目选择、分值、顺序、可见性字段抽到独立 settings section。
  - 保留父对话框的 `form`、`fieldErrors`、`submit()` 与 `watch(props.draft)` owner。
- 预期改动：
  - `code/frontend/src/components/platform/contest/ContestChallengeEditorDialog.vue`
  - `code/frontend/src/components/platform/contest/ContestChallengeSettingsSection.vue`
  - `code/frontend/src/components/platform/__tests__/ContestChallengeEditorDialog.test.ts`
  - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts`
  - `code/frontend/src/components/common/__tests__/BackofficeDialogAdoption.test.ts`
- 验证：
  - `npm run test:run -- src/components/platform/__tests__/ContestChallengeEditorDialog.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts src/components/common/__tests__/BackofficeDialogAdoption.test.ts`
- Review focus：
  - settings section 是否只是字段展示 owner
  - 对话框仍然保留重复提交短路和 submit payload owner

### Slice 3：更新 guardrail 与收尾

- 目标：
  - 把 duplicate action 与 raw-source 护栏切到父对话框 + 子组件组合视角。
  - 确保 `ContestChallengeEditorDialog.vue` 不再保留整块被抽走的模板和样式。
- 预期改动：
  - `code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts`
  - 相关 raw-source 测试
  - `docs/reviews/frontend/*contest-challenge-editor-dialog*`
- 验证：
  - `npm run test:run -- src/views/__tests__/duplicateActionGuardAudit.test.ts src/components/platform/__tests__/ContestChallengeEditorDialog.test.ts`
- Review focus：
  - 护栏是否继续锁定真实 owner，而不是绑定旧文件布局

## 结构收口检查

- 父对话框保留真正的行为 owner，新增 section 不直接接触 feature / API / router。
- 新 section 不新增 `componentFeatureImportAllowlist` 例外。
- 当前 touched surface 上的 oversized dialog debt 要真实下降，不能只是把 899 行平移到另一个含糊的大组件里。

## 验证计划

- `npm run test:run -- src/components/platform/__tests__/ContestChallengeEditorDialog.test.ts src/views/__tests__/duplicateActionGuardAudit.test.ts src/views/__tests__/studentDirectoryTypographyBoundary.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts src/components/common/__tests__/BackofficeDialogAdoption.test.ts`
- `npm run typecheck`
- `bash scripts/check-workflow-complete.sh`
- 必要时：`git diff --check -- code/frontend/src/components/platform/contest code/frontend/src/components/platform/__tests__/ContestChallengeEditorDialog.test.ts code/frontend/src/views/__tests__/duplicateActionGuardAudit.test.ts code/frontend/src/views/__tests__/studentDirectoryTypographyBoundary.test.ts code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase3.test.ts code/frontend/src/components/common/__tests__/BackofficeDialogAdoption.test.ts`

## Review 关注点

- AWD 创建模式下的多选行为是否保持不变。
- 编辑模式下是否仍然只暴露分值 / 顺序 / 可见性，不回流 AWD 列表。
- `submit()` 的 in-flight guard、校验与 payload 组装 owner 是否仍留在父对话框。
- raw-source 护栏是否已经适应拆分后的结构。

## 回退 / 恢复说明

- 所有新 section 都应能按文件粒度回退，父对话框只需回退 imports、模板接线和局部 helper 迁移。
- 本轮不涉及 API、路由和持久化变更，回退主要是前端组件结构回退。

## 残余风险

- `ContestAwdConfigWorkspaceShell.vue` 仍是更大的下一块 debt，本轮不碰它。
- 如果 section props 切得不够干净，容易让 AWD 模式和普通模式的字段在子组件里互相泄漏；实现时要明确用 `show*` 约束可见面。
