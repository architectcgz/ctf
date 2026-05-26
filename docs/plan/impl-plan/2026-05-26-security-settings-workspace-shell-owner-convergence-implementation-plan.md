> 状态：Current
> 事实源：`TD-1` 超大组件专题拆分、`SecuritySettings.vue` 当前剩余 workspace 壳、profile feature 已存在的密码修改 owner
> 替代：无

# SecuritySettings Workspace Shell Owner Convergence Implementation Plan

## 目标

- 把 `SecuritySettings.vue` 里的页面壳、安全概况、密码修改区、安全提示区和配套局部样式抽到独立 `SecuritySettingsWorkspaceShell.vue`。
- 保持父页继续持有 `useSecuritySettingsPage()` 的密码表单、字段校验、提交流程和安全概况数据 owner。
- 让 `SecuritySettings.vue` 回到 route page 组合 owner，不再承接大块 workspace 模板和局部样式。

## 非目标

- 本轮不改 `useSecuritySettingsPage()` 的密码校验规则、提交 API、toast 文案或安全概况数据结构。
- 本轮不引入新的 profile 共享 shell 抽象，也不重排 profile feature 目录。
- 本轮不动 `InstanceList.vue`、`ContestDetail.vue` 或其他 `TD-1` 页面。

## 输入依据

- `code/frontend/src/views/profile/SecuritySettings.vue`
- `code/frontend/src/views/profile/__tests__/SecuritySettings.test.ts`
- `code/frontend/src/components/profile/UserProfileWorkspaceShell.vue`
- `code/frontend/src/components/profile/SkillProfileWorkspaceShell.vue`
- `.harness/reuse-decisions/security-settings-workspace-shell-owner-convergence.md`

## 当前结论

- `SecuritySettings.vue` 当前 491 行，虽未超过 500 行护栏，但已经明显承担 route view 之外的大块 workspace 模板与局部样式。
- `useSecuritySettingsPage()` 已经是明确的页面业务 owner，视图重量主要集中在稳定的展示壳，适合继续沿用“父页保留 owner，子组件承接 workspace shell”的既有模式。
- 现有源码断言测试大量直接读取 `SecuritySettings.vue?raw`，抽壳时需要同步改成父页 + shell 组合源码检查，否则后续继续细拆会反复把模板塞回 route view。

## 任务切片

### Slice 1：抽取 security settings workspace shell

- 目标：
  - 新增 `SecuritySettingsWorkspaceShell.vue`，承接页面壳、安全概况、密码修改区、安全提示区和对应局部样式。
  - `SecuritySettings.vue` 继续保留密码表单、字段错误、提交动作和安全概况数据 owner。
- 预期改动：
  - `code/frontend/src/views/profile/SecuritySettings.vue`
  - `code/frontend/src/components/profile/SecuritySettingsWorkspaceShell.vue`
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/profile/__tests__/SecuritySettings.test.ts`
  - `cd code/frontend && npm run typecheck`
- Review focus：
  - 父页是否仍然持有密码修改流程与字段校验 owner
  - 新组件是否只承接稳定模板和样式，没有吸入 `changePassword`、`useToast` 或表单校验逻辑

### Slice 2：同步源码断言测试

- 目标：
  - 把 profile 与共享样式类测试改成按父页 + shell 组合源码检查，避免因为抽壳导致旧断言误报。
- 预期改动：
  - `code/frontend/src/views/profile/__tests__/SecuritySettings.test.ts`
  - `code/frontend/src/views/__tests__/workspaceShellStyles.test.ts`
  - `code/frontend/src/views/__tests__/surfaceBackground.test.ts`
  - `code/frontend/src/views/__tests__/rootHeroLayout.test.ts`
  - `code/frontend/src/views/__tests__/studentRootShellCleanup.test.ts`
  - `code/frontend/src/views/__tests__/profileJournalButtonStyles.test.ts`
  - `code/frontend/src/views/__tests__/journalEyebrowStyles.test.ts`
  - `code/frontend/src/views/__tests__/profileJournalNoteStyles.test.ts`
  - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
  - `code/frontend/src/views/__tests__/journalUserShellStyles.test.ts`
  - `code/frontend/src/views/__tests__/profileJournalUtilityStyles.test.ts`
  - `code/frontend/src/views/__tests__/workspacePageHeaderStyles.test.ts`
- 验证：
  - `cd code/frontend && npm run test:run -- src/views/profile/__tests__/SecuritySettings.test.ts src/views/__tests__/surfaceBackground.test.ts src/views/__tests__/rootHeroLayout.test.ts src/views/__tests__/studentRootShellCleanup.test.ts src/views/__tests__/profileJournalButtonStyles.test.ts src/views/__tests__/journalEyebrowStyles.test.ts src/views/__tests__/profileJournalNoteStyles.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/journalUserShellStyles.test.ts src/views/__tests__/profileJournalUtilityStyles.test.ts src/views/__tests__/workspaceShellStyles.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts`
- Review focus：
  - 断言是否仍然锁定共享 shell / page header / note / button / token 收口事实
  - 是否避免把实现细节写死在 route view 单文件上

## 风险

- 如果 props 设计过宽，可能把整个密码修改流程直接透传进 shell，形成“减行数但没收 owner”的伪拆分。
- 由于现有多个共享样式测试直接读取 `SecuritySettings.vue` 源码，抽壳后如果漏改组合源码断言，容易出现大量与真实回归无关的假失败。

## 回退方式

- 如抽取后出现交互回归，可回退 `SecuritySettingsWorkspaceShell.vue` 并把模板恢复到 `SecuritySettings.vue`。
- 本轮只影响前端视图层和测试护栏，不涉及后端或 API 契约。
