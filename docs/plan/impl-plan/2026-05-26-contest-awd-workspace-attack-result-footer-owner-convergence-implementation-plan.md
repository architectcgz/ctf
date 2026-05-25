> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ContestAWDWorkspacePanel.vue` 当前结构、AWD 学员战场既有 `awd/*` 子组件分层模式
> 替代：无

# Contest AWD Workspace Attack Result Footer Owner Convergence Implementation Plan

## 目标

- 让 `ContestAWDWorkspacePanel.vue` 把中区攻击结果 footer 抽成独立子组件。
- 保持父组件继续拥有 `submitResult`、`formatAttackResultToast`、`getSubmitResultMessage` 和提交流程 owner。
- 让新子组件只承接成功 / 失败结果提示模板与局部样式。

## 非目标

- 本轮不改 `submitAttack`、`handleSubmit`、`flagInputs`、目标列表或左侧服务编排。
- 本轮不改 `ContestDetail.vue`、路由、feature model 或请求层。
- 本轮不顺手调整提示文案。

## 输入依据

- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `.harness/reuse-decisions/contest-awd-workspace-attack-result-footer-owner-convergence.md`

## 当前结论

- `ContestAWDWorkspacePanel.vue` 中区还保留一个纯展示 footer，依赖面只有 `submitResult` 和格式化后的提示消息。
- 这是进入左侧服务编排壳之前最后一个低风险切片。
- 当前最小安全方式是让 footer 子组件只接收 `isSuccess` 和 message，不直接感知攻击提交 workflow。

## 任务切片

### Slice 1：抽出 attack result footer 子组件

- 目标：
  - 新建 `AWDAttackResultFooter.vue`，承接攻击结果提示模板与样式。
  - `ContestAWDWorkspacePanel.vue` 保留 `submitResult` owner 和消息格式化。
- 预期改动：
  - `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
  - `code/frontend/src/components/contests/awd/AWDAttackResultFooter.vue`
- 组件边界：
  - 父组件继续拥有 `submitResult` 和消息构造
  - 子组件只接收 `success` 和 `message`
- 验证：
  - `git diff --check -- code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue code/frontend/src/components/contests/awd/AWDAttackResultFooter.vue`
  - `cd code/frontend && npm run test:run -- src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts src/views/contests/__tests__/ContestDetail.test.ts -t AWD`
- Review focus：
  - 父组件是否继续持有攻击结果 owner，没有把文案格式化下沉进子组件
  - 新子组件是否保持纯展示，不引入本地逻辑

### Slice 2：回写 TD-1 进展

- 目标：
  - 把攻击结果 footer 切片写回前端主索引。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
  - `rg -n "ContestAWDWorkspacePanel|result footer|攻击结果|提交" docs/reviews/frontend/README.md docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- Review focus：
  - 文档是否明确中区展示块已经基本收口，剩余重心在左侧服务编排壳

## 风险

- raw source 护栏需要同步更新到新 footer 子组件。
- 如果把 `getSubmitResultMessage` 一起搬到子组件，会让消息格式化 owner 漂移。
- 这轮若顺手去碰左侧服务编排壳，会放大切片边界。

## 回退方式

- 如 `AWDAttackResultFooter.vue` 抽层引入回归，可回退新组件并恢复父组件的 footer 模板。
- 本轮仍只影响前端组件层、测试护栏和文档，不涉及 API、route 或攻击动作链。
