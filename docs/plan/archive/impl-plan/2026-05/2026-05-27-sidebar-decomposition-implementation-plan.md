> 状态：Current
> 事实源：`Sidebar.vue` 的 layout owner 与拆分边界
> 替代：无

# Sidebar Decomposition Implementation Plan

## 目标

- 把 `code/frontend/src/components/layout/Sidebar.vue` 从移动/桌面双模板的大文件拆成“layout owner + sidebar shells + nav tree”的结构。
- 保留 `Sidebar.vue` 对 route、workspace shell module、active 判定、展开态和导航跳转的 owner。

## 非目标

- 本轮不改 `useWorkspaceShellNavigation.ts` 的 route/module contract。
- 本轮不改 `AppLayout.vue` 与 `TopNav.vue` 的交互协议。
- 本轮不重做 sidebar 视觉风格，不新增权限能力或新的导航模块。

## 输入依据

- `docs/architecture/frontend/06-components.md`
- `docs/architecture/frontend/07-pages-dataflow.md`
- `code/frontend/src/components/layout/Sidebar.vue`
- `code/frontend/src/components/layout/AppLayout.vue`
- `code/frontend/src/components/layout/__tests__/Sidebar.test.ts`
- `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `code/frontend/src/composables/useWorkspaceShellNavigation.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `Sidebar.vue` 应继续留在 `components/layout/`，因为它是全局 layout 导航壳，不是单一 feature 的 UI。
- 当前最重的债是模板与样式的重复：移动端和桌面端分别内联 header、workspace label、nav item tree。
- route、module、expand/collapse、navigate owner 已经是集中且清楚的，不应该在这轮拆散。

## 设计边界

### `components/layout/Sidebar.vue` 继续负责

- `useWorkspaceShellNavigation()` 装配
- 当前用户角色、brand kicker、module items 派生
- active item / expanded menu 判定
- click 导航与移动端关闭逻辑

### `components/layout/sidebar/*` 本轮负责

- 移动端 sidebar shell
- 桌面端 sidebar shell
- 共享 nav tree / header / workspace label 渲染

### 本轮不负责

- 变更 backoffice module contract
- 变更 `AppLayout` 的 sidebar open/collapse owner
- 调整教师 / 管理员可见导航集合

## 任务切片

### Slice 1：抽出移动/桌面壳与 nav tree

- 目标：
  - 新增 `components/layout/sidebar/` 子组件
  - `Sidebar.vue` 只保留 owner、computed、行为函数与外层装配
- 验证：
  - `npm run test:run -- src/components/layout/__tests__/Sidebar.test.ts`
- Review focus：
  - route/expand/navigate owner 是否仍留在父组件
  - nav tree 子组件是否只做渲染，不额外重做判定逻辑

### Slice 2：同步 raw-source 护栏与 backlog

- 目标：
  - 更新 `Sidebar.test.ts` 与 `sharedThemeTokenAdoption.test.ts` 的聚合源码断言
  - 更新 backlog 中 `Sidebar.vue` 进展
- 验证：
  - `npm run test:run -- src/components/layout/__tests__/Sidebar.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- Review focus：
  - theme token、布局壳和 backoffice shell 护栏是否仍可覆盖子组件拆分后的源码

## 验证计划

- `cd code/frontend && npm run test:run -- src/components/layout/__tests__/Sidebar.test.ts src/components/layout/__tests__/TopNav.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`

## 残余风险

- `TopNav.vue` 仍是这组 `P2` 布局层大组件债里的后续重点。
- 如果 `Sidebar.vue` 拆分后还残留过多内联判定或 class helper，下一轮可能要进一步判断是否把展示判定再局部收口。
