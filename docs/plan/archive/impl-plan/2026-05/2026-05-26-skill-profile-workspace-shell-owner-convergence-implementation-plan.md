> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`SkillProfile.vue` 当前剩余 workspace 壳、skill-profile feature 已存在的数据与 tab owner
> 替代：无

# SkillProfile Workspace Shell Owner Convergence Implementation Plan

## 目标

- 把 `SkillProfile.vue` 里的页面壳、顶部 tab rail、教师选择区、分析/薄弱项/推荐三块内容面板和配套局部样式抽到独立 `SkillProfileWorkspaceShell.vue`。
- 保持父页继续持有 `useSkillProfilePage()` 的加载、错误、学员切换、推荐跳转 owner，以及 `useUrlSyncedTabs<SkillProfileTabKey>()` 的 route tab owner。
- 让 `SkillProfile.vue` 回到 route page 组合 owner，不再承接大块 workspace 模板和局部样式。

## 非目标

- 本轮不改 `useSkillProfilePage()` 的 API、加载策略、推荐逻辑或跳转契约。
- 本轮不改 tab key、query 同步协议、键盘导航逻辑或教师/学生文案。
- 本轮不动 `ContestAwdConfig.vue`。

## 输入依据

- `code/frontend/src/views/profile/SkillProfile.vue`
- `code/frontend/src/views/profile/__tests__/SkillProfile.test.ts`
- `code/frontend/src/views/profile/__tests__/skillProfileTabsAdoption.test.ts`
- `code/frontend/src/components/profile/UserProfileWorkspaceShell.vue`
- `.harness/reuse-decisions/skill-profile-workspace-shell-owner-convergence.md`

## 当前结论

- `SkillProfile.vue` 当前 741 行，超过 route view 阈值，但 page owner 已经主要集中在 `useSkillProfilePage()` 与 `useUrlSyncedTabs<SkillProfileTabKey>()`。
- 当前剩余重量主要来自稳定的页面模板壳和局部样式，适合继续沿用“父页保留 owner，子组件承接 workspace shell”的既有模式。
- `skillProfileTabsAdoption.test.ts` 已明确要求 tab sync owner 继续留在父页，这一层不能在抽壳时下沉。

## 任务切片

### Slice 1：抽取 skill profile workspace shell

- 目标：
  - 新增 `SkillProfileWorkspaceShell.vue`，承接页面壳、tab rail、教师选择区、三块内容面板和对应局部样式。
  - `SkillProfile.vue` 继续保留能力画像数据、推荐数据、教师学员切换、tab owner 和跳转动作 owner。
- 预期改动：
  - `code/frontend/src/views/profile/SkillProfile.vue`
  - `code/frontend/src/components/profile/SkillProfileWorkspaceShell.vue`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
- 验证：
  - `cd code/frontend && npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts src/views/profile/__tests__/SkillProfile.test.ts src/views/profile/__tests__/skillProfileTabsAdoption.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/rootHeroLayout.test.ts src/views/__tests__/studentRootShellCleanup.test.ts src/views/__tests__/profileJournalButtonStyles.test.ts src/views/__tests__/profileJournalUtilityStyles.test.ts src/views/__tests__/profileJournalNoteStyles.test.ts src/views/__tests__/journalUserShellStyles.test.ts src/views/__tests__/journalEyebrowStyles.test.ts src/views/__tests__/pageTabsStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
  - `cd code/frontend && npm run typecheck`
- Review focus：
  - 父页是否仍然持有 route tab、学员切换、远端数据和跳转 owner
  - 新组件是否只承接稳定页面壳，而没有重新吸入 `useUrlSyncedTabs` 或 API 逻辑
  - 路由页是否脱离 oversized allowlist

### Slice 2：回写 TD-1 进展

- 目标：
  - 把 `SkillProfile.vue` 从 oversized route view 推进到 workspace shell owner 收口的事实写回前端 review 索引。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
- Review focus：
  - 文档是否明确这是 route view 壳层收口，而不是数据 owner 迁移

## 风险

- 如果 props 设计过宽，可能把整份画像状态和 tab 逻辑整体透传到 shell，形成“减行数但没收 owner”的伪拆分。
- 多个源码断言测试会同时命中父页和新 shell；若不改成组合源码检查，后续继续拆壳会再次卡住。

## 回退方式

- 如抽取后出现交互回归，可回退 `SkillProfileWorkspaceShell.vue` 并把模板恢复到 `SkillProfile.vue`。
- 本轮只影响前端视图层、测试护栏和 review 文档，不涉及后端或 API 契约。
