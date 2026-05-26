# InstanceList Workspace Shell 收口复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - files reviewed：
    - `code/frontend/src/views/instances/InstanceList.vue`
    - `code/frontend/src/components/instance/InstanceListWorkspaceShell.vue`
    - `code/frontend/src/views/instances/__tests__/InstanceList.test.ts`
    - `code/frontend/src/views/__tests__/studentUserSurfaceAlignment.test.ts`
    - `code/frontend/src/views/__tests__/journalUserShellStyles.test.ts`
    - `code/frontend/src/views/__tests__/journalUserDirectoryButtonVariants.test.ts`
    - `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
    - `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
    - `code/frontend/src/views/__tests__/studentDirectoryTypographyBoundary.test.ts`
    - `code/frontend/src/views/__tests__/rootHeroLayout.test.ts`
    - `code/frontend/src/views/__tests__/studentRootShellCleanup.test.ts`
    - `code/frontend/src/views/__tests__/journalUserDirectoryStyles.test.ts`
    - `code/frontend/src/views/__tests__/journalEyebrowStyles.test.ts`
    - `.harness/reuse-decisions/instance-list-workspace-shell-owner-convergence.md`
    - `docs/plan/impl-plan/2026-05-26-instance-list-workspace-shell-owner-convergence-implementation-plan.md`
- Classification check：同意当前切片属于前端 `TD-1` 结构性收口，且本轮改动范围与计划一致。
- Gate verdict：Pass

## Findings

- 无 material finding。

## Material findings

- 无。

## Senior implementation assessment

- 当前实现把 `InstanceList.vue` 收回为 route view owner：实例列表加载、复制 / 打开 / 延时 / 销毁动作、即将过期提醒状态，以及 `useInstanceWarningFocus()` 的焦点管理仍由父页负责。
- 新增的 `InstanceListWorkspaceShell.vue` 只承接稳定页面壳、概况卡片、实例目录、过期提醒弹层和对应局部样式，没有重新吸入实例请求、`useToast` 或第二份 warning 状态 owner。
- 实例难度 pill 的样式计算已下沉到 shell，就地依赖 `entities/challenge` 的展示能力，父页不再保留空壳样式函数。
- `InstanceList.test.ts` 和相关共享样式断言已经改成区分“父页 owner 源码”和“父页 + shell 组合源码”，后续再细拆展示块时不会反向把模板和 scoped style 塞回 route view。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/instances/__tests__/InstanceList.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd code/frontend && npm run test:run -- src/views/__tests__/studentUserSurfaceAlignment.test.ts src/views/__tests__/journalUserShellStyles.test.ts src/views/__tests__/journalUserDirectoryButtonVariants.test.ts src/views/__tests__/rootHeroLayout.test.ts src/views/__tests__/studentRootShellCleanup.test.ts src/views/__tests__/journalUserDirectoryStyles.test.ts src/views/__tests__/journalEyebrowStyles.test.ts`
- `cd code/frontend && npm run test:run -- src/views/__tests__/studentDirectoryTypographyBoundary.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts -t "学生侧普通目录标题不应使用局部等宽字体变体|学生侧列表主标题列应保持纯净，不混入标签、序号或描述|首屏页面头部应使用共享 workspace-page-header 分隔线结构|不应在页面局部重复声明公共标题排版|不应在页面局部重复声明公共说明排版|典型工作区页头应优先复用共享 workspace-page-header 结构"`
- `git diff --check`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

## Residual risk

- `InstanceListWorkspaceShell.vue` 仍然是一个偏大的展示壳组件，但 route view owner 已收口；如果后续继续增加实例列表上的新交互，优先沿“概况卡片 / 实例目录 / 过期提醒弹层”继续拆展示区，不要把数据 owner 再抬回父页。
- `studentDirectoryTypographyBoundary.test.ts` 全量执行时仍有与本刀无关的既有失败，集中在 `ClassStudentsPage.vue` 的旧断言与当前源码不一致；本轮已改成定向执行实例页相关条目，不构成这次 `InstanceList` 抽壳的 blocker。

## Touched known-debt status

- 本轮 touched 的已知结构债是 `InstanceList.vue` route view 混合 owner / 模板 / 样式。
- 该债务在 touched surface 上已完成收口：父页本体降到 `69` 行，当前没有把实例请求、实例动作、warning 状态或提醒焦点再混回壳组件。
