> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ContestAWDWorkspacePanel.vue` 当前脚本结构、`contest-awd-workspace` feature 现有 composable 分层
> 替代：无

# AWD Attack Vector State Owner Convergence Implementation Plan

## 目标

- 把 `ContestAWDWorkspacePanel.vue` 里的攻击向量局部状态和派生逻辑收口到独立 composable。
- 保持 `useContestAWDWorkspace` 继续拥有远端工作区数据、攻击提交副作用和刷新链路。
- 让父页只保留战场级布局装配、情报列映射和防守侧脚本。

## 非目标

- 本轮不改 `AWDAttackVectorPanel.vue` 的模板结构。
- 本轮不改防守列、情报列和 SSH 复制链路。
- 本轮不改 `useAwdWorkspaceAttackSubmission.ts` 的请求与 toast owner。

## 输入依据

- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/features/contest-awd-workspace/model/useContestAWDWorkspace.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAttackSubmission.ts`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdDefenseServiceSelection.ts`
- `.harness/reuse-decisions/awd-attack-vector-state-owner-convergence.md`

## 当前结论

- 模板装配壳已经抽到 `AWDAttackVectorPanel.vue`，剩余脚本里最完整的局部域就是攻击向量状态。
- 这组逻辑只依赖挑战列表、工作区 targets 和 `submitAttack`，适合做成 feature composable。
- 最小安全切片是：composable 持有 `activeChallengeKey / targetKeyword / flagInputs` 与相关派生；父页继续消费结果，不改远端副作用 owner。

## 任务切片

### Slice 1：抽出攻击向量状态 composable

- 目标：
  - 新建 `useAwdWorkspaceAttackVector.ts`，收口 challenge 选择、目标筛选、Flag 输入和提交后清空逻辑。
  - 父页改为消费 composable 返回值，不再手写这组局部逻辑。
- 预期改动：
  - `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
  - `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAttackVector.ts`
  - `code/frontend/src/features/contest-awd-workspace/model/index.ts`
  - `code/frontend/src/features/contest-awd-workspace/index.ts`
- 边界：
  - composable 只处理攻击向量局部 UI state
  - `submitAttack` 仍由 `useContestAWDWorkspace` 提供
  - 父页仍负责把结果装配给 `AWDAttackVectorPanel.vue`
- 验证：
  - `git diff --check -- code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAttackVector.ts code/frontend/src/features/contest-awd-workspace/model/index.ts code/frontend/src/features/contest-awd-workspace/index.ts`
  - `cd code/frontend && npm run test:run -- src/features/contest-awd-workspace/model/useAwdWorkspaceAttackVector.test.ts`
- Review focus：
  - composable 是否只收口攻击向量 state，没有把远端请求 owner 误挪走
  - 提交成功后 Flag 输入清空行为是否保持不变

### Slice 2：补 feature 测试与 review 索引

- 目标：
  - 为新 composable 增加本地状态与提交流程测试。
  - 把 TD-1 进展写回 review 索引。
- 预期改动：
  - `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAttackVector.test.ts`
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `cd code/frontend && npm run test:run -- src/features/contest-awd-workspace/model/useAwdWorkspaceAttackVector.test.ts src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
  - `bash scripts/check-consistency.sh`
- Review focus：
  - 文档是否明确这是 script owner 收口，而不是新增行为

## 风险

- `activeChallengeKey` 默认选中和 challenge 列表刷新回退规则不能变。
- 目标筛选只能按现有队伍名称关键字匹配，不能顺手改行为。
- 如果 composable 顺手接管 `submitResult` 或 toast，会超出本轮边界。

## 回退方式

- 如 composable 抽层引入回归，可回退 `useAwdWorkspaceAttackVector.ts` 并恢复父页原有攻击向量 state。
- 本轮只影响前端 feature/composable、父页装配和 review 文档，不涉及 API、route 或服务端逻辑。
