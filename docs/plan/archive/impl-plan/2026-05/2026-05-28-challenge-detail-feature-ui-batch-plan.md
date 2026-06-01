> 状态：Current
> 事实源：`challenge-detail` 当前 UI owner、feature-owned UI 迁移既有模式
> 替代：无

# Challenge Detail Feature UI Batch Plan

## 目标

- 把 `ChallengeWorkspaceShell.vue`、`ChallengeSolutionsPanel.vue`、`ChallengeSubmissionRecordsPanel.vue` 迁入 `features/challenge-detail/ui`。
- 让 `ChallengeDetail.vue` 改为从 `@/features/challenge-detail` public API 引入 `ChallengeWorkspaceShell`。
- 同步更新 feature UI 导出、raw-source 护栏、`components.d.ts`、allowlist 和 backlog 记录。

## 非目标

- 本轮不拆 `ChallengeWorkspaceShell.vue` 内部 question / writeup / action aside 子件。
- 本轮不改 `useChallengeDetailPage()` 的 route/query、数据加载、题解提交流程和实例操作 owner。
- 本轮不处理 `ChallengeQuestionPanel.vue`、`ChallengeWriteupPanel.vue`、`ChallengeActionAside.vue` 是否还应进一步 feature 化。

## 输入依据

- `code/frontend/src/components/challenge/ChallengeWorkspaceShell.vue`
- `code/frontend/src/components/challenge/ChallengeSolutionsPanel.vue`
- `code/frontend/src/components/challenge/ChallengeSubmissionRecordsPanel.vue`
- `code/frontend/src/views/challenges/ChallengeDetail.vue`
- `code/frontend/src/features/challenge-detail/index.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`
- `code/frontend/src/views/challenges/__tests__/challengeDetailPanelExtraction.test.ts`
- `code/frontend/src/views/challenges/__tests__/challengeDetailSharedShell.test.ts`
- `code/frontend/src/views/__tests__/pageTabsStyles.test.ts`
- `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `challenge-detail` 当前 route page owner 已经收回 `useChallengeDetailPage()`，剩下的 legacy 路径问题主要是 feature 私有 UI 仍留在 `components/challenge/*`。
- 这组三件套只被 `ChallengeDetail.vue` 和彼此组合消费，不属于共享 challenge primitive。
- 最小正确改动是整体迁位到 `features/challenge-detail/ui`，并通过 feature public API 暴露给 route view。

## 设计边界

### `features/challenge-detail/ui/*` 本轮负责

- challenge detail workspace shell
- solutions panel
- submission records panel

### `features/challenge-detail/model/*` 本轮继续负责

- route/query 与 tab owner
- challenge / solution / submission / writeup / instance 数据与动作 owner
- 展示用格式化与状态映射

### `views/challenges/ChallengeDetail.vue` 本轮继续负责

- 组合 `useChallengeDetailPage()`
- 加载态 / 错误态 / 空态切换
- feature page shell 的顶层装配

## 任务切片

### Slice 1：feature UI 迁位

- 目标：
  - 新增 `features/challenge-detail/ui/index.ts`
  - 把 3 个 challenge detail 私有 UI 文件迁入 feature
  - `ChallengeDetail.vue` 与 `ChallengeWorkspaceShell.vue` 切到 feature public API / feature 内部相对 import
- 验证：
  - `npm run test:run -- src/views/challenges/__tests__/ChallengeDetail.test.ts src/views/challenges/__tests__/challengeDetailPanelExtraction.test.ts src/views/challenges/__tests__/challengeDetailSharedShell.test.ts`
- Review focus：
  - route page 是否仍只保留 page owner，不把 UI 迁位反向拉回 view 层
  - 迁位后 props / emits contract 是否保持不变

### Slice 2：护栏与 backlog 同步

- 目标：
  - 更新 `architectureAllowlist.ts`、`components.d.ts` 和 raw-source 测试引用
  - 在前端 backlog 中记录 `challenge-detail` 这组 allowlist 的收口进展
- 验证：
  - `npm run test:run -- src/views/__tests__/pageTabsStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - 旧 `components/challenge/*` 路径是否已从 touched surface 消失
  - `componentFeatureImportAllowlist` 是否只剩非本轮范围的残留

## 验证计划

- `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/challenges/__tests__/ChallengeDetail.test.ts src/views/challenges/__tests__/challengeDetailPanelExtraction.test.ts src/views/challenges/__tests__/challengeDetailSharedShell.test.ts src/views/__tests__/pageTabsStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- `challenge-detail` 迁位后仍有 `ChallengeQuestionPanel.vue`、`ChallengeWriteupPanel.vue`、`ChallengeActionAside.vue` 留在 `components/challenge/*`；如果后续它们也只继续服务单一 feature，应按下一刀继续判断 owner，而不是在这轮顺手混进来。
