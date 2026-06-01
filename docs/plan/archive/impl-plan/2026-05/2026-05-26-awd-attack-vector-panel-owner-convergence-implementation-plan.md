> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ContestAWDWorkspacePanel.vue` 当前结构、AWD 工作台既有子组件分层模式
> 替代：无

# AWD Attack Vector Panel Owner Convergence Implementation Plan

## 目标

- 让 `ContestAWDWorkspacePanel.vue` 把中区 `攻击向量` 装配壳抽成独立子组件。
- 保持父组件继续拥有 challenge 筛选、目标筛选、Flag 输入、打开目标和提交攻击的 owner。
- 让新子组件只承接攻击列标题、空状态、筛选条、目标网格和结果 footer 的装配。

## 非目标

- 本轮不改防守列与情报列。
- 本轮不调整 `useContestAWDWorkspace` 的数据 owner。
- 本轮不引入新的攻击筛选能力，也不处理 `showOnlyReachableTargets` 的遗留行为。

## 输入依据

- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDAttackToolbar.vue`
- `code/frontend/src/components/contests/awd/AWDAttackTargetGrid.vue`
- `code/frontend/src/components/contests/awd/AWDAttackResultFooter.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `.harness/reuse-decisions/awd-attack-vector-panel-owner-convergence.md`

## 当前结论

- `ContestAWDWorkspacePanel.vue` 现阶段剩余最显眼的大模板块是中区攻击列。
- 这块只是把 `AWDAttackToolbar`、`AWDAttackTargetGrid`、`AWDAttackResultFooter` 以及空状态拼在一起，适合继续下沉。
- 最小安全切片是：父页继续持有筛选状态、Flag 输入和提交动作；子组件只消费现成 props 并向上透传事件。

## 任务切片

### Slice 1：抽出攻击列装配壳

- 目标：
  - 新建 `AWDAttackVectorPanel.vue`，承接 `攻击向量` 中区壳层与状态分支。
  - `ContestAWDWorkspacePanel.vue` 只保留战场级布局壳与 workflow owner。
- 预期改动：
  - `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
  - `code/frontend/src/components/contests/awd/AWDAttackVectorPanel.vue`
- 组件边界：
  - 父组件继续拥有 `activeChallengeKey / targetKeyword / flagInputs`
  - 父组件继续拥有 `openTarget / handleSubmit / submitResult`
  - 子组件只接收现成值并透传事件
- 验证：
  - `git diff --check -- code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue code/frontend/src/components/contests/awd/AWDAttackVectorPanel.vue`
  - `cd code/frontend && npm run test:run -- src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- Review focus：
  - 是否只是抽装配壳，没有把 workflow owner 一起下沉
  - 中区空状态和 footer 是否仍保持原有文案与结构

### Slice 2：回写 TD-1 进展

- 目标：
  - 把 AWD 工作台这次中区攻击列切片写回前端 review 索引。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
  - `rg -n "AWDAttackVectorPanel|攻击向量|ContestAWDWorkspacePanel" docs/reviews/frontend/README.md docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- Review focus：
  - 文档是否明确这轮只收口中区装配壳，剩余仍是 script / workflow owner 密度

## 风险

- source 护栏要把中区 `攻击向量` 标题和 `AWDAttackToolbar` / `AWDAttackTargetGrid` / `AWDAttackResultFooter` 断言迁到新组件。
- 如果顺手把 `handleSubmit`、`flagInputs` 或 `openTarget` 下沉到子组件，会模糊 parent owner 边界。

## 回退方式

- 如 `AWDAttackVectorPanel.vue` 抽层引入回归，可回退新组件并恢复父页中区模板。
- 本轮只影响前端组件层、测试护栏和 review 文档，不涉及 API、route 或服务端逻辑。
