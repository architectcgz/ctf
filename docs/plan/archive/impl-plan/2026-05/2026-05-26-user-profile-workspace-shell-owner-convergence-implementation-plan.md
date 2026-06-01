> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`UserProfile.vue` 当前剩余 workspace 壳、profile feature 已存在的数据与导出 owner
> 替代：无

# UserProfile Workspace Shell Owner Convergence Implementation Plan

## 目标

- 把 `UserProfile.vue` 里的页面壳、顶部资料区、摘要区、账号信息区、个人报告区和配套局部样式抽到独立 `UserProfileWorkspaceShell.vue`。
- 保持父页继续持有 `useUserProfilePage()` 的数据加载、导出、下载、错误和文案 owner。
- 让 `UserProfile.vue` 回到 route page 组合 owner，不再承接大块 workspace 模板。

## 非目标

- 本轮不改 `useUserProfilePage()` 的 API、导出轮询、下载流程或报告状态语义。
- 本轮不改个人资料业务文案、字段结构或报告格式选择规则。
- 本轮不改 `SkillProfile.vue`。

## 输入依据

- `code/frontend/src/views/profile/UserProfile.vue`
- `code/frontend/src/views/profile/__tests__/UserProfile.test.ts`
- `code/frontend/src/components/scoreboard/ScoreboardWorkspaceShell.vue`
- `.harness/reuse-decisions/user-profile-workspace-shell-owner-convergence.md`

## 当前结论

- `UserProfile.vue` 当前 719 行，超出 route view 阈值，但脚本层 owner 已经基本集中在 `useUserProfilePage()`。
- 当前剩余重量主要来自页面模板壳和局部样式，适合继续沿用“父页保留 owner，子组件承接稳定 workspace shell”的既有模式。

## 任务切片

### Slice 1：抽取 user profile workspace shell

- 目标：
  - 新增 `UserProfileWorkspaceShell.vue`，承接页面壳、顶部资料区、摘要区、账号信息区、个人报告区和对应局部样式。
  - `UserProfile.vue` 继续保留 profile 数据、报告导出、下载、错误态和文案 owner。
- 预期改动：
  - `code/frontend/src/views/profile/UserProfile.vue`
  - `code/frontend/src/components/profile/UserProfileWorkspaceShell.vue`
  - `code/frontend/src/__tests__/architectureAllowlist.ts`
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/profile/__tests__/UserProfile.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/surfaceBackground.test.ts src/views/__tests__/journalUserShellStyles.test.ts src/views/__tests__/profileJournalButtonStyles.test.ts src/views/__tests__/rootHeroLayout.test.ts src/views/__tests__/studentRootShellCleanup.test.ts src/views/__tests__/profileJournalUtilityStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts src/views/__tests__/profileJournalNoteStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/journalUserDirectoryStyles.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts`
  - `cd code/frontend && npm run typecheck`
- Review focus：
  - 父页是否仍然持有 profile/report owner
  - 新组件是否只承接稳定页面壳，而没有吸入 API 或导出流程逻辑
  - 路由页是否低于 view 阈值

### Slice 2：回写 TD-1 进展

- 目标：
  - 把 `UserProfile.vue` 从 oversized route view 推进到 workspace shell owner 收口的事实写回前端 review 索引。
- 预期改动：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- 验证：
  - `bash scripts/check-consistency.sh`
- Review focus：
  - 文档是否明确这是 route view 壳层收口，而不是 feature owner 迁移

## 风险

- 如果 props 设计过宽，可能把整份 profile 状态和导出流程整体透传到 shell，形成“减行数但没收 owner”的伪拆分。
- 多个源码断言测试会同时命中父页和新 shell，如果只改一边，容易把风格/结构断言改坏。

## 回退方式

- 如抽取后出现交互回归，可回退 `UserProfileWorkspaceShell.vue` 并把模板恢复到 `UserProfile.vue`。
- 本轮只影响前端视图层、测试护栏和 review 文档，不涉及后端或 API 契约。
