> 状态：Current
> 事实源：challenge list / detail feature UI owner、前端架构 backlog、raw-source 护栏
> 替代：无

# Challenge Feature Boundary Cleanup Plan

## 目标

- 把 challenge list / challenge detail 已明确 owner 的 UI 从历史 `components/challenge` 目录回收到对应 `features/*/ui`

## 非目标

- 不重做 `challenge-list` 与 `challenge-detail` 的交互、异步流程、样式结构和 page model。
- 不调整 `ChallengeSolutionsPanel.vue`、`ChallengeSubmissionRecordsPanel.vue`、`ChallengeMetaStrip.vue` 等已经落在合理 owner 的文件。
- 不顺手处理 `components/challenge` 目录里其他 consumer 仍不明确的文件。

## 输入依据

- `ChallengeListRoutePage.vue`
- `features/challenge-list/*`
- `features/challenge-detail/ui/ChallengeWorkspaceShell.vue`
- `ChallengeList.test.ts`
- `ChallengeDetail.test.ts`
- `challengePageUiStrategy.test.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`
- `docs/reviews/frontend/2026-05-28-challenge-detail-feature-ui-batch-review.md`

## 当前结论

- `ChallengeDirectoryPanel.vue` 只由 student challenge list route page 使用，owner 已经是 `challenge-list`。
- `ChallengeActionAside.vue`、`ChallengeQuestionPanel.vue`、`ChallengeWriteupPanel.vue`、`ChallengeInstanceCard.vue` 只服务 `ChallengeWorkspaceShell.vue` 这条 feature 内部装配链，owner 已经是 `challenge-detail`。
- 当前残留问题主要不是行为错误，而是文件落点、public API、raw-source 护栏和 backlog 事实还停在旧历史目录。

## 设计边界

### `features/challenge-list` 本轮负责

- 持有 `ChallengeDirectoryPanel.vue` 的 UI owner 与对外导出
- 不承接 page model 以外的新路由 / 状态 owner

### `features/challenge-detail` 本轮负责

- 持有 workspace shell 内部 question / writeup / action / instance 四个子块
- 继续通过 `ChallengeWorkspaceShell.vue` 组合这些私有 UI
- 不新增跨 feature 共享出口

### route page / tests / docs 本轮负责

- 只同步 import 路径、raw-source 断言、组件声明和 backlog / review 事实
- 不改功能行为断言口径

## 任务切片

- [x] Slice 1：challenge list UI owner 迁位
  - 目标：
    - `ChallengeDirectoryPanel.vue` 迁入 `features/challenge-list/ui`
    - `ChallengeListRoutePage.vue` 与相关 raw-source 测试改为消费新路径 / public API
  - 验证：
    - `cd code/frontend && npm run test:run -- src/pages/challenges/__tests__/ChallengeList.test.ts src/pages/__tests__/sharedPaginationControls.test.ts src/pages/__tests__/journalUserDirectoryStyles.test.ts`

- [x] Slice 2：challenge detail 内部子块迁位
  - 目标：
    - `ChallengeActionAside.vue`、`ChallengeQuestionPanel.vue`、`ChallengeWriteupPanel.vue`、`ChallengeInstanceCard.vue` 迁入 `features/challenge-detail/ui`
    - `ChallengeWorkspaceShell.vue` 与相关 raw-source 测试改为消费新路径 / public API
  - 验证：
    - `cd code/frontend && npm run test:run -- src/pages/challenges/__tests__/ChallengeDetail.test.ts src/pages/challenges/__tests__/challengePageUiStrategy.test.ts src/pages/__tests__/workspacePageHeaderStyles.test.ts`

- [x] Slice 3：文档 / 声明 / review 收尾
  - 目标：
    - 更新 `components.d.ts`、backlog 与 review 归档
    - 删除旧 `components/challenge/*` 路径残片
  - 验证：
    - `cd code/frontend && npm run test:run -- src/pages/challenges/__tests__/ChallengeList.test.ts src/pages/challenges/__tests__/ChallengeDetail.test.ts src/pages/challenges/__tests__/challengePageUiStrategy.test.ts src/pages/__tests__/sharedPaginationControls.test.ts src/pages/__tests__/journalUserDirectoryStyles.test.ts src/pages/__tests__/workspacePageHeaderStyles.test.ts src/features/__tests__/featureBoundaries.test.ts`
    - `cd code/frontend && npm run typecheck`

## 验证计划

- `cd code/frontend && npm run test:run -- src/pages/challenges/__tests__/ChallengeList.test.ts src/pages/challenges/__tests__/ChallengeDetail.test.ts src/pages/challenges/__tests__/challengePageUiStrategy.test.ts src/pages/__tests__/sharedPaginationControls.test.ts src/pages/__tests__/journalUserDirectoryStyles.test.ts src/pages/__tests__/workspacePageHeaderStyles.test.ts src/features/__tests__/featureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/.worktrees/ctf-challenge-feature-boundary-cleanup && git diff --check -- .harness/reuse-decisions/challenge-feature-boundary-cleanup.md docs/plan/impl-plan/2026-05-30-challenge-feature-boundary-cleanup-plan.md docs/reviews/frontend/2026-05-30-challenge-feature-boundary-cleanup-review.md docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md code/frontend/src/components.d.ts code/frontend/src/pages/challenges/ChallengeListRoutePage.vue code/frontend/src/pages/challenges/__tests__/ChallengeList.test.ts code/frontend/src/pages/challenges/__tests__/ChallengeDetail.test.ts code/frontend/src/pages/challenges/__tests__/challengePageUiStrategy.test.ts code/frontend/src/pages/__tests__/sharedPaginationControls.test.ts code/frontend/src/pages/__tests__/journalUserDirectoryStyles.test.ts code/frontend/src/pages/__tests__/workspacePageHeaderStyles.test.ts code/frontend/src/features/challenge-list/index.ts code/frontend/src/features/challenge-list/ui/index.ts code/frontend/src/features/challenge-list/ui/ChallengeDirectoryPanel.vue code/frontend/src/features/challenge-detail/ui/index.ts code/frontend/src/features/challenge-detail/ui/ChallengeWorkspaceShell.vue code/frontend/src/features/challenge-detail/ui/ChallengeActionAside.vue code/frontend/src/features/challenge-detail/ui/ChallengeQuestionPanel.vue code/frontend/src/features/challenge-detail/ui/ChallengeWriteupPanel.vue code/frontend/src/features/challenge-detail/ui/ChallengeInstanceCard.vue`
- `cd /home/azhi/workspace/projects/.worktrees/ctf-challenge-feature-boundary-cleanup && bash scripts/check-task-intake.sh --reuse-decision challenge-feature-boundary-cleanup`

## 残余风险

- `components.d.ts` 是生成文件，当前仓库通常直接提交更新后的声明；如果本地生成策略与手工编辑不一致，需要在 typecheck 后再核对是否被工具覆盖。
- 旧 review / audit 文档中的历史路径属于历史事实，本轮只补新增 review 与 backlog 现状，不回写所有旧文档。
