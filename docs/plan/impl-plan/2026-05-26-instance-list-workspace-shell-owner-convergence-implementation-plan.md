> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`InstanceList.vue` 当前剩余 workspace 壳、instance-list feature 已存在的页面 owner
> 替代：无

# InstanceList Workspace Shell Owner Convergence Implementation Plan

## 目标

- 把 `InstanceList.vue` 里的页面壳、概况卡片、实例目录和过期提醒弹层以及对应局部样式抽到独立 `InstanceListWorkspaceShell.vue`。
- 保持父页继续持有 `useInstanceListPage()` 的实例数据加载、定时刷新、延时/销毁/打开/复制动作、过期提醒状态，以及 `useInstanceWarningFocus()` 的按钮聚焦 owner。
- 让 `InstanceList.vue` 回到 route page 组合 owner，不再承接大块 workspace 模板和 scoped style。

## 非目标

- 本轮不改 `useInstanceListPage()` 的轮询策略、实例操作契约或 toast 行为。
- 本轮不改实例列表的数据来源，不处理 architecture review 中更大的 server-side query owner 收口。
- 本轮不动 `ContestDetail.vue`、`ChallengeTopologyStudioPage.vue` 或 teacher 侧页面。

## 输入依据

- `code/frontend/src/views/instances/InstanceList.vue`
- `code/frontend/src/views/instances/__tests__/InstanceList.test.ts`
- `code/frontend/src/features/instance-list/model/useInstanceListPage.ts`
- `code/frontend/src/components/profile/SecuritySettingsWorkspaceShell.vue`
- `.harness/reuse-decisions/instance-list-workspace-shell-owner-convergence.md`

## 当前结论

- `InstanceList.vue` 当前 493 行，虽然未超 500 行护栏，但 route view 已经同时承担业务 owner、整页模板和局部样式。
- 业务 owner 已主要收口在 `useInstanceListPage()`；route view 里额外残留的本地责任是过期提醒关闭按钮聚焦和难度标签样式辅助函数。
- 最小安全切片是延续 shell 方案：父页继续保留 feature owner 和本地按钮 ref，shell 只承接稳定展示层，并通过 prop 接收按钮 ref setter 和业务动作。

## 任务切片

### Slice 1：抽取 instance list workspace shell

- 目标：
  - 新增 `InstanceListWorkspaceShell.vue`，承接页面壳、概况卡片、目录列表、过期提醒弹层和对应局部样式。
  - `InstanceList.vue` 继续保留实例数据、轮询、操作方法、warning 状态和 `warningCloseButton` ref owner。
- 预期改动：
  - `code/frontend/src/views/instances/InstanceList.vue`
  - `code/frontend/src/components/instance/InstanceListWorkspaceShell.vue`
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/instances/__tests__/InstanceList.test.ts`
  - `cd code/frontend && npm run typecheck`
- Review focus：
  - 父页是否仍然持有数据加载、定时刷新、动作调用和 warning owner
  - 新 shell 是否只承接展示层，没有吸入 `useInstanceListPage()` 或 toast/API 调用

### Slice 2：同步源码断言测试

- 目标：
  - 把直接读取 `InstanceList.vue?raw` 的源码断言改成区分“父页 owner 源码”和“父页 + shell 组合源码”。
- 预期改动：
  - `code/frontend/src/views/instances/__tests__/InstanceList.test.ts`
  - `code/frontend/src/views/__tests__/studentUserSurfaceAlignment.test.ts`
  - `code/frontend/src/views/__tests__/journalUserShellStyles.test.ts`
  - `code/frontend/src/views/__tests__/journalEyebrowStyles.test.ts`
  - `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
  - `code/frontend/src/views/__tests__/journalUserDirectoryButtonVariants.test.ts`
  - `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
  - `code/frontend/src/views/__tests__/studentDirectoryTypographyBoundary.test.ts`
  - `code/frontend/src/views/__tests__/rootHeroLayout.test.ts`
  - `code/frontend/src/views/__tests__/studentRootShellCleanup.test.ts`
  - `code/frontend/src/views/__tests__/journalUserDirectoryStyles.test.ts`
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/instances/__tests__/InstanceList.test.ts src/views/__tests__/studentUserSurfaceAlignment.test.ts src/views/__tests__/journalUserShellStyles.test.ts src/views/__tests__/journalEyebrowStyles.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/journalUserDirectoryButtonVariants.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/studentDirectoryTypographyBoundary.test.ts src/views/__tests__/rootHeroLayout.test.ts src/views/__tests__/studentRootShellCleanup.test.ts src/views/__tests__/journalUserDirectoryStyles.test.ts`
- Review focus：
  - 断言是否仍然锁定实例页的 shared shell / page header / directory / warning dialog 事实
  - 是否避免继续把整页模板和样式绑定死在 route view 单文件

## 风险

- 如果 ref 透传设计不清晰，过期提醒关闭按钮的自动聚焦可能在抽壳后失效。
- `InstanceList.vue` 的源码断言分散在多个共享样式测试里，漏改任意一处都可能造成与实际回归无关的假失败。

## 回退方式

- 如抽取后出现交互回归，可回退 `InstanceListWorkspaceShell.vue` 并把模板恢复到 `InstanceList.vue`。
- 本轮只影响前端视图层和测试护栏，不涉及 API 或后端契约。
