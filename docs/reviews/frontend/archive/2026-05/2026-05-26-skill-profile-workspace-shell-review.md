# SkillProfile Workspace Shell 收口复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - files reviewed：
    - `code/frontend/src/views/profile/SkillProfile.vue`
    - `code/frontend/src/components/profile/SkillProfileWorkspaceShell.vue`
    - `code/frontend/src/__tests__/architectureAllowlist.ts`
    - `code/frontend/src/views/profile/__tests__/SkillProfile.test.ts`
    - `code/frontend/src/views/profile/__tests__/skillProfileTabsAdoption.test.ts`
    - `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
    - `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
    - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- Classification check：同意当前切片属于前端 `TD-1` 结构性收口，且本轮改动范围与计划一致。
- Gate verdict：Pass

## Findings

- 无 material finding。

## Material findings

- 无。

## Senior implementation assessment

- 当前实现把能力画像页的加载、推荐跳转、学员切换和 route tab owner 继续保留在 `useSkillProfilePage()`、`useUrlSyncedTabs<SkillProfileTabKey>()` 与父页 `SkillProfile.vue`。
- 新增的 `SkillProfileWorkspaceShell.vue` 只承接页面模板、样式和局部事件转发，没有重新吸入 API、跳转或 tab 同步逻辑，符合“先收页面壳，再看面板展示区是否继续细拆”的最小切片目标。
- 相关源码断言测试已经改成按父页 + shell 组合源码检查，避免后续继续抽壳时把页面模板回塞到 route view。

## Required re-validation

- `npm run test:run -- src/views/profile/__tests__/SkillProfile.test.ts src/views/profile/__tests__/skillProfileTabsAdoption.test.ts`
- `npm run typecheck`
- `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts src/views/profile/__tests__/SkillProfile.test.ts src/views/profile/__tests__/skillProfileTabsAdoption.test.ts src/views/__tests__/rootHeroLayout.test.ts src/views/__tests__/studentRootShellCleanup.test.ts src/views/__tests__/profileJournalButtonStyles.test.ts src/views/__tests__/profileJournalUtilityStyles.test.ts src/views/__tests__/profileJournalNoteStyles.test.ts src/views/__tests__/journalUserShellStyles.test.ts src/views/__tests__/journalEyebrowStyles.test.ts src/views/__tests__/pageTabsStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
- `npm run test:run -- src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts -t "首屏页面头部应使用共享 workspace-page-header 分隔线结构|带顶部 tab 的页面不应继续在 tab 面板内重复渲染 eyebrow|不应在页面局部重复声明公共标题排版"`
- `git diff --check`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

## Residual risk

- `SkillProfileWorkspaceShell.vue` 仍然是一个偏大的展示壳组件，但 route view owner 已经收口；后续如果继续在这页追加展示逻辑，应优先继续沿分析区 / 薄弱项区 / 推荐区拆成更细的展示分区。
- `workspaceShellStyles.test.ts` 与 `workspacePageHeaderStyles.test.ts` 全量执行时仍有与本刀无关的既有失败，集中在 `StudentAnalysisPage.vue` 的旧断言，不构成这轮 `SkillProfile` 抽壳的 blocker。

## Touched known-debt status

- 本轮 touched 的已知结构债是 `SkillProfile.vue` oversized route view owner。
- 该债务在 touched surface 上已完成收口：父页已脱离 oversized allowlist，本体降到 `82` 行，当前没有把原本的数据、推荐跳转、学员切换或 route tab owner 再次混回去。
