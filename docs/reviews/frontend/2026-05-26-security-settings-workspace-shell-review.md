# SecuritySettings Workspace Shell 收口复核

- Review target：
  - repository：`ctf`
  - branch：`main`
  - diff source：working tree against `HEAD`
  - files reviewed：
    - `code/frontend/src/views/profile/SecuritySettings.vue`
    - `code/frontend/src/components/profile/SecuritySettingsWorkspaceShell.vue`
    - `code/frontend/src/views/profile/__tests__/SecuritySettings.test.ts`
    - `.harness/reuse-decisions/security-settings-workspace-shell-owner-convergence.md`
    - `docs/plan/impl-plan/2026-05-26-security-settings-workspace-shell-owner-convergence-implementation-plan.md`
- Classification check：同意当前切片属于前端 `TD-1` 结构性收口，且本轮改动范围与计划一致。
- Gate verdict：Pass

## Findings

- 无 material finding。

## Material findings

- 无。

## Senior implementation assessment

- 当前实现继续把 `SecuritySettings.vue` 保持为 route view owner：密码表单、字段校验、提交流程和安全概况数据仍由 `useSecuritySettingsPage()` 负责。
- 新增的 `SecuritySettingsWorkspaceShell.vue` 只承接稳定页面壳、安全概况、密码修改区、安全提示区和对应局部样式，没有重新吸入 `changePassword`、`useToast` 或第二份表单校验 owner。
- `SecuritySettings.test.ts` 与相关共享样式断言已经改成区分“父页 owner 源码”和“父页 + shell 组合源码”，能防止后续继续细拆时把整页模板和样式回塞到 route view。

## Required re-validation

- `cd code/frontend && npm run test:run -- src/views/profile/__tests__/SecuritySettings.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd code/frontend && npm run test:run -- src/views/profile/__tests__/SecuritySettings.test.ts src/views/__tests__/surfaceBackground.test.ts src/views/__tests__/rootHeroLayout.test.ts src/views/__tests__/studentRootShellCleanup.test.ts src/views/__tests__/profileJournalButtonStyles.test.ts src/views/__tests__/journalEyebrowStyles.test.ts src/views/__tests__/profileJournalNoteStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/journalUserShellStyles.test.ts src/views/__tests__/profileJournalUtilityStyles.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/journalUserDirectoryStyles.test.ts -t "首屏页面头部应使用共享 workspace-page-header 分隔线结构|典型工作区页头应优先复用共享 workspace-page-header 结构|profile 与 security 页顶部也应接入共享 topbar 与 summary 骨架|profile 与 security 顶部概况应显式使用 metric-panel 类|应该在共享样式文件中声明 workspace shell 骨架样式|member pages should consume shared hero shell classes instead of duplicating formula|uses a section root that carries the hero background|student root shell cleanup|profile 页面不应继续在 scoped style 中重复声明 journal 按钮基础规则|已切到 workspace overline 的页面不应继续携带旧 eyebrow 根节点修饰类|profile 页面不应继续在局部样式里重写共享的基础 note 规则|SecuritySettings|profile 页面不应继续在 scoped style 中重复声明 tech-font|SecuritySettings 不应包含"`
- `git diff --check`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

## Residual risk

- `SecuritySettingsWorkspaceShell.vue` 仍然是一个偏大的展示壳组件，但 route view owner 已经收口；如果后续继续追加新的安全分区，优先沿“安全概况 / 密码修改 / 安全提示”继续拆成更细的展示区，而不是把表单 owner 再抬回父页。
- `workspaceShellStyles.test.ts` 与 `workspacePageHeaderStyles.test.ts` 全量执行时仍有与本刀无关的既有失败，集中在 `StudentAnalysisPage.vue` 与 `ChallengeTopologyStudioPage.vue` 的旧断言，不构成这轮 `SecuritySettings` 抽壳的 blocker。

## Touched known-debt status

- 本轮 touched 的已知结构债是 `SecuritySettings.vue` route view 混合 owner / 模板 / 样式。
- 该债务在 touched surface 上已完成收口：父页本体降到 `26` 行，当前没有把密码表单、字段校验、提交流程或安全概况 owner 再次混回壳组件。
