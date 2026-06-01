> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ChallengeDetail.vue` 当前剩余 workspace 壳、challenge 已抽出的 question / solutions / records / writeup / action aside 组件
> 替代：无

# Challenge Detail Workspace Shell Owner Convergence Implementation Plan

## 目标

- 把 `ChallengeDetail.vue` 里剩余的 tab workspace 装配壳抽到独立 `ChallengeWorkspaceShell.vue`。
- 保持父页继续持有 `useChallengeDetailPage` 的 route/query 同步、远端数据、实例动作、Flag 提交、题解编辑和异常处理 owner。
- 让 `ChallengeDetail.vue` 从“仍然承接大块 workspace 模板和布局样式”进一步收敛到 route page 组合 owner。

## 非目标

- 本轮不改 `useChallengeDetailPage`、`useChallengeDetailDataLoader`、`useChallengeDetailInteractions`、`useChallengeDetailPresentation` 的数据流或 API 协议。
- 本轮不改 `ChallengeQuestionPanel.vue`、`ChallengeSolutionsPanel.vue`、`ChallengeSubmissionRecordsPanel.vue`、`ChallengeWriteupPanel.vue`、`ChallengeActionAside.vue` 的业务语义。
- 本轮不改题目页视觉方向、主题变量语义或交互规则。

## 输入依据

- `code/frontend/src/views/challenges/ChallengeDetail.vue`
- `code/frontend/src/components/challenge/ChallengeQuestionPanel.vue`
- `code/frontend/src/components/challenge/ChallengeSolutionsPanel.vue`
- `code/frontend/src/components/challenge/ChallengeSubmissionRecordsPanel.vue`
- `code/frontend/src/components/challenge/ChallengeWriteupPanel.vue`
- `code/frontend/src/components/challenge/ChallengeActionAside.vue`
- `code/frontend/src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.vue`
- `code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`
- `code/frontend/src/views/challenges/__tests__/challengeDetailPanelExtraction.test.ts`
- `.harness/reuse-decisions/challenge-detail-workspace-shell-owner-convergence.md`

## 当前结论

- `ChallengeDetail.vue` 当前 574 行，超出 route view 阈值，但业务内容已经大部分下沉到独立 panel / aside。
- 剩余重量主要来自题目页 tabbar、主区分支切换、右侧实例工具栏装配，以及配套的布局样式。
- 最小安全切片是把 workspace shell 抽走，父页继续只做数据 / 动作 owner 和状态分支，不碰 feature model。

## 任务切片

### Slice 1：抽取 challenge workspace shell

- 目标：
  - 新增 `ChallengeWorkspaceShell.vue`，承接 tabbar、panel 切换、右侧 `ChallengeActionAside` 装配和对应布局样式。
  - `ChallengeDetail.vue` 继续保留 loading / error / empty / challenge state 分支，以及 `useChallengeDetailPage` 返回的 owner。
- 预期改动：
  - `code/frontend/src/views/challenges/ChallengeDetail.vue`
  - `code/frontend/src/components/challenge/ChallengeWorkspaceShell.vue`
- 验证：
  - `git diff --check -- code/frontend/src/views/challenges/ChallengeDetail.vue code/frontend/src/components/challenge/ChallengeWorkspaceShell.vue`
  - `cd code/frontend && npm run test:run -- src/views/challenges/__tests__/ChallengeDetail.test.ts src/views/challenges/__tests__/challengeDetailPanelExtraction.test.ts`
  - `cd code/frontend && npm run typecheck`
- Review focus：
  - 父页是否继续持有 route/data/action owner
  - 新子组件是否只承接 workspace shell，而没有吸入 API 或路由逻辑
  - 路由页是否回到 page-size threshold 内

### Slice 2：回写 TD-1 进展

- 目标：
  - 把 `ChallengeDetail.vue` 这一刀从“已抽 panel”推进到“workspace shell owner 收口”的事实写回前端 review 索引。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
  - `rg -n "ChallengeDetail.vue|workspace shell|ChallengeWorkspaceShell" docs/reviews/frontend/README.md docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- Review focus：
  - 文档是否明确这是 route page shell 收口，而不是数据 owner 迁移

## 风险

- workspace shell 抽走后，tab ref / keyboard owner 可能被误迁到子组件，导致 route-query tab 行为边界变模糊。
- 如果 props / emits 设计不清楚，`ChallengeWorkspaceShell.vue` 可能变成“把整个 page state 再包装一层”的伪抽取。
- panel extraction 测试和架构阈值测试需要同步更新，否则会把合理收口误判成回归。

## 回退方式

- 如抽取后出现交互回归，可回退 `ChallengeWorkspaceShell.vue`，把模板恢复到 `ChallengeDetail.vue`。
- 本轮只影响题目详情前端视图层、测试护栏和 review 文档，不涉及后端或 API 契约。
