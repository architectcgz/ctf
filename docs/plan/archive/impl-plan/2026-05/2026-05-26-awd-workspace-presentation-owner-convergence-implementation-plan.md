> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ContestAWDWorkspacePanel.vue` 当前脚本结构、AWD 工作台现有 feature/composable 分层
> 替代：无

# AWD Workspace Presentation Owner Convergence Implementation Plan

## 目标

- 把 `ContestAWDWorkspacePanel.vue` 里的 challenge 标题映射、事件结果标签、服务引用文案和攻击结果 toast 文案收口到独立 composable。
- 把 AWD 工作台 touched surface 的 runtime challenge 身份统一收紧到 `awd_service_id + awd_challenge_id`，不再回退到历史 `challenge_id`。
- 保持 `useContestAWDWorkspace` 继续拥有攻击提交副作用，只消费 presentation 回调。
- 让父页继续保留工作区主数据装配、布局和防守 / 攻击 owner。

## 非目标

- 本轮不改 `AWDWorkspaceIntelColumn.vue` 和 `AWDAttackVectorPanel.vue` 模板结构。
- 本轮不改 `useContestAWDWorkspace.ts` 的请求、刷新和 toast owner。
- 本轮不重排 `defenseAlerts` 的展示结构，也不改 `formatRoundStatusLabel`。

## 输入依据

- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDWorkspaceIntelColumn.vue`
- `code/frontend/src/features/contest-awd-workspace/model/awdDefensePresentation.ts`
- `code/frontend/src/features/contest-awd-workspace/model/awdChallengeIdentity.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAttackVector.ts`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `.harness/reuse-decisions/awd-workspace-presentation-owner-convergence.md`

## 当前结论

- 经过前几刀后，父页中仍然成组存在的脚本主要是情报与结果文案 presentation。
- 这组逻辑只依赖挑战列表和攻击结果，不适合继续留在 route-level panel。
- 用户已确认 AWD 工作台里的 challenge 身份不该再回退到 `challenge_id`；只要进入 runtime challenge 语义，就应只认 `awd_service_id` 和 `awd_challenge_id`。
- 最小安全切片是：抽 `useAwdWorkspacePresentation.ts`，并补 `isAwdRuntimeChallenge` 守卫，让父页、攻击向量与防守 presentation 一起只消费 AWD 专用 challenge 身份。

## 任务切片

### Slice 1：抽出 workspace presentation composable

- 目标：
  - 新建 `useAwdWorkspacePresentation.ts`，收口 challenge 映射、事件标签和攻击结果文案格式化。
  - 父页消费 composable 返回值，不再手写这组 helper。
  - 新建 `awdChallengeIdentity.ts`，让 AWD runtime challenge 身份从父页、攻击向量到防守卡片保持同一守卫。
- 预期改动：
  - `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
  - `code/frontend/src/features/contest-awd-workspace/model/awdChallengeIdentity.ts`
  - `code/frontend/src/features/contest-awd-workspace/model/awdDefensePresentation.ts`
  - `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAttackVector.ts`
  - `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspacePresentation.ts`
  - `code/frontend/src/features/contest-awd-workspace/model/index.ts`
  - `code/frontend/src/features/contest-awd-workspace/index.ts`
- 边界：
  - composable 只处理文案和展示映射
  - AWD runtime challenge 只在 `awd_service_id` 与 `awd_challenge_id` 都存在时成立
  - `submitResult` 和 `submitAttack` 仍由 `useContestAWDWorkspace` 持有
  - 父页仍负责把函数透传给 `AWDWorkspaceIntelColumn.vue`
- 验证：
  - `git diff --check -- code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue code/frontend/src/features/contest-awd-workspace/model/awdChallengeIdentity.ts code/frontend/src/features/contest-awd-workspace/model/awdDefensePresentation.ts code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAttackVector.ts code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspacePresentation.ts code/frontend/src/features/contest-awd-workspace/model/index.ts code/frontend/src/features/contest-awd-workspace/index.ts`
  - `cd code/frontend && npm run test:run -- src/features/contest-awd-workspace/model/useAwdWorkspacePresentation.test.ts src/features/contest-awd-workspace/model/awdDefensePresentation.test.ts`
- Review focus：
  - composable 是否只收口 presentation owner，没有把副作用迁进去
  - AWD runtime challenge 身份是否已经统一改成专用字段，不再混用 `challenge_id`
  - 情报列和攻击结果文案是否保持一致

### Slice 2：补源码护栏与 review 索引

- 目标：
  - 给新 composable 增加 challenge 映射和结果文案测试。
  - 给 AWD 专用 challenge 身份补源码护栏。
  - 更新源码护栏与 `TD-1` review 索引。
- 预期改动：
  - `code/frontend/src/features/contest-awd-workspace/model/awdDefensePresentation.test.ts`
  - `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `cd code/frontend && npm run test:run -- src/features/contest-awd-workspace/model/useAwdWorkspacePresentation.test.ts src/features/contest-awd-workspace/model/awdDefensePresentation.test.ts src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
  - `cd code/frontend && npm run typecheck`
  - `bash scripts/check-consistency.sh`
- Review focus：
  - 文档是否明确这轮收口的是 presentation owner 与 AWD challenge 身份 owner
  - 源码护栏是否能阻止 `awd_challenge_id || challenge_id` 一类回退再次回流

## 风险

- `formatAttackResultToast` 仍要兼容 `useContestAWDWorkspace` 里的格式化回调签名。
- `eventDirectionLabel`、`eventResultLabel` 和 `formatServiceRef` 不能改变现有文案。
- challenge 标题优先级仍应保持 `service_id -> awd_challenge_id`。
- AWD runtime challenge touched surface 不能再退回 `challenge_id` 兜底，否则会重新混淆历史 challenge 标识与运行态服务标识。

## 回退方式

- 如 composable 抽层引入回归，可回退 `useAwdWorkspacePresentation.ts` 并恢复父页 helper。
- 本轮只影响前端 feature/composable、父页装配和 review 文档，不涉及 API、route 或服务端逻辑。
