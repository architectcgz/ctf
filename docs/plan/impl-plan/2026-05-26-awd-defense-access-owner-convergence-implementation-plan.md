> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`ContestAWDWorkspacePanel.vue` 当前脚本结构、AWD 防守连接现有交互
> 替代：无

# AWD Defense Access Owner Convergence Implementation Plan

## 目标

- 把 `ContestAWDWorkspacePanel.vue` 里的防守 SSH / 复制链路局部 owner 收口到独立 composable。
- 保持 `useContestAWDWorkspace` 继续拥有远端 access 请求、SSH 票据生成和打开目标入口。
- 让父页只保留服务选择、工作区布局和防守列装配。

## 非目标

- 本轮不改 `AWDDefenseOperationsPanel.vue` 和 `AWDDefenseConnectionPanel.vue` 的模板结构。
- 本轮不改情报列事件标题映射和攻击结果 toast 格式化。
- 本轮不改 `useAwdWorkspaceAccessActions.ts` 的远端请求 owner。

## 输入依据

- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseColumn.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseOperationsPanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseConnectionPanel.vue`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAccessActions.ts`
- `code/frontend/src/features/contest-awd-workspace/model/sshAccessPresentation.ts`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `.harness/reuse-decisions/awd-defense-access-owner-convergence.md`

## 当前结论

- 攻击向量 state 已经收口后，父页里剩余最成组的逻辑是防守 access 解析与复制。
- 这块只依赖 `selectedServiceId`、`servicesByServiceId`、`sshAccessByServiceId` 和 `openService`，适合独立成 composable。
- 最小安全切片是：composable 持有复制态、选中 access 派生和复制动作；父页继续透传 `openDefenseSSH`、`restartService` 和防守列 props。

## 任务切片

### Slice 1：抽出 defense access composable

- 目标：
  - 新建 `useAwdDefenseAccessPanel.ts`，收口 access 解析、打开服务和复制动作。
  - 父页消费 composable 返回值，不再手写 `copyTextToClipboard` 等逻辑。
- 预期改动：
  - `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
  - `code/frontend/src/features/contest-awd-workspace/model/useAwdDefenseAccessPanel.ts`
  - `code/frontend/src/features/contest-awd-workspace/model/index.ts`
  - `code/frontend/src/features/contest-awd-workspace/index.ts`
- 边界：
  - composable 只处理防守 access 局部 owner
  - `openDefenseSSH` 仍由 `useContestAWDWorkspace` 提供
  - 父页仍负责把结果装配给 `AWDDefenseColumn.vue`
- 验证：
  - `git diff --check -- code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue code/frontend/src/features/contest-awd-workspace/model/useAwdDefenseAccessPanel.ts code/frontend/src/features/contest-awd-workspace/model/index.ts code/frontend/src/features/contest-awd-workspace/index.ts`
  - `cd code/frontend && npm run test:run -- src/features/contest-awd-workspace/model/useAwdDefenseAccessPanel.test.ts`
- Review focus：
  - composable 是否只收口 access UI owner，没有把远端 SSH 请求挪走
  - 复制成功 / 失败提示与 copied 态是否保持一致

### Slice 2：补护栏与 review 索引

- 目标：
  - 给新 composable 增加复制与打开服务测试。
  - 更新源码护栏和 `TD-1` review 索引。
- 预期改动：
  - `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `cd code/frontend && npm run test:run -- src/features/contest-awd-workspace/model/useAwdDefenseAccessPanel.test.ts src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
  - `bash scripts/check-consistency.sh`
- Review focus：
  - 文档是否明确这轮收口的是防守 access 脚本 owner

## 风险

- `copiedCommand` / `copiedPassword` 必须仍按当前选中服务驱动。
- 剪贴板失败提示不能被吞掉，也不能误改成全局异常。
- `openDefenseService` 只能桥接到实例 access，不应顺手改 `openDefenseSSH`。

## 回退方式

- 如 composable 抽层引入回归，可回退 `useAwdDefenseAccessPanel.ts` 并恢复父页原有 access 逻辑。
- 本轮只影响前端 feature/composable、父页装配和 review 文档，不涉及 API、route 或服务端逻辑。
