# UserProfile Workspace Shell 收口复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - files reviewed：
    - `code/frontend/src/views/profile/UserProfile.vue`
    - `code/frontend/src/components/profile/UserProfileWorkspaceShell.vue`
    - `code/frontend/src/__tests__/architectureAllowlist.ts`
    - `code/frontend/src/views/profile/__tests__/UserProfile.test.ts`
    - `code/frontend/src/views/__tests__/journalUserShellStyles.test.ts`
    - `code/frontend/src/views/__tests__/journalUserDirectoryStyles.test.ts`
    - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- Classification check：同意当前切片属于前端 `TD-1` 结构性收口，且本轮改动范围与计划一致。
- Gate verdict：Pass

## Findings

- 无 material finding。

## Material findings

- 无。

## Senior implementation assessment

- 当前实现把个人资料页的加载、错误、导出、下载和文案 owner 继续保留在 `useUserProfilePage()` 与父页 `UserProfile.vue`。
- 新增的 `UserProfileWorkspaceShell.vue` 只承接页面模板、样式和局部事件转发，没有重新吸入 API、轮询或导出流程 owner，符合“先收页面壳，再看局部展示区是否继续细拆”的最小切片目标。
- 相关源码断言测试已经改成按父页 + shell 组合源码检查，避免后续继续抽壳时把页面模板回塞到 route view。

## Required re-validation

- `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- `npm run test:run -- src/views/profile/__tests__/UserProfile.test.ts src/views/__tests__/surfaceBackground.test.ts src/views/__tests__/rootHeroLayout.test.ts src/views/__tests__/studentRootShellCleanup.test.ts src/views/__tests__/profileJournalButtonStyles.test.ts src/views/__tests__/journalUserShellStyles.test.ts src/views/__tests__/journalUserDirectoryStyles.test.ts src/views/__tests__/journalEyebrowStyles.test.ts src/views/__tests__/profileJournalUtilityStyles.test.ts src/views/__tests__/profileJournalNoteStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/workspaceShellStyles.test.ts -t "首屏页面头部应使用共享 workspace-page-header 分隔线结构|应该在共享样式文件中声明 workspace shell 骨架样式|UserProfile|member pages should consume shared hero shell classes instead of duplicating formula|uses a section root that carries the hero background|student root shell cleanup|profile 页面不应继续在 scoped style 中重复声明 journal 按钮基础规则|profile 与 security 顶部概况应显式使用 metric-panel 类，旧共享 CSS 只保留变量桥接|profile 与 security 页顶部也应接入共享 topbar 与 summary 骨架|已切到 workspace overline 的页面不应继续携带旧 eyebrow 根节点修饰类|profile 页面不应继续在 scoped style 中重复声明 tech-font|profile 页面不应继续在局部样式里重写共享的基础 note 规则|个人资料与安全设置页不应回退到浅色状态块"`
- `npm run typecheck`
- `git diff --check`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

## Residual risk

- `UserProfileWorkspaceShell.vue` 仍然是一个偏大的展示壳组件，但 route view owner 已经收口；后续如果继续在这页追加展示逻辑，应优先继续沿账号区 / 报告区拆成更细的展示分区。
- `workspacePageHeaderStyles.test.ts` 和 `workspaceShellStyles.test.ts` 里仍有与本刀无关的既有失败，集中在 `StudentAnalysisPage` / `ChallengeTopologyStudioPage` 的旧断言，不构成这轮 `UserProfile` 抽壳的 blocker。

## Touched known-debt status

- 本轮 touched 的已知结构债是 `UserProfile.vue` oversized route view owner。
- 该债务在 touched surface 上已完成收口：父页已脱离 oversized allowlist，本体降到 `47` 行，当前没有把原本的数据、导出和下载 owner 再次混回去。
