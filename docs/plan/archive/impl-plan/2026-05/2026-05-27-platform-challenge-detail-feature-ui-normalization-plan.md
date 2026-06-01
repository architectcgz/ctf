> 状态：Current
> 事实源：platform challenge detail feature-owned UI normalization
> 替代：无

# Platform Challenge Detail Feature UI Normalization Plan

## 目标

- 把 `AdminChallengeTopbarPanel.vue`、`AdminChallengeWorkspaceTabs.vue`、`AdminChallengeProfilePanel.vue` 从 `components/platform/challenge` 迁入 `features/platform-challenge-detail/ui`。
- 让 `PlatformChallengeDetailWorkspace.vue` 改为通过 `features/platform-challenge-detail` public API 组合题目详情 UI。

## 非目标

- 本轮不改 `usePlatformChallengeDetailPage.ts`、`usePlatformChallengeDetailRoutePage.ts` 的加载、保存、路由同步 owner。
- 本轮不改 `ChallengeWriteupManagePanel` 的 slot 合成方式。
- 本轮不重做题目详情视觉结构、tab 行为或 flag 配置交互。

## 输入依据

- `code/frontend/src/components/platform/challenge/AdminChallengeTopbarPanel.vue`
- `code/frontend/src/components/platform/challenge/AdminChallengeWorkspaceTabs.vue`
- `code/frontend/src/components/platform/challenge/AdminChallengeProfilePanel.vue`
- `code/frontend/src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.vue`
- `code/frontend/src/features/platform-challenge-detail/ui/index.ts`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- 这组三件套只服务 platform challenge detail workspace，本质上已经是单一 feature 的 topbar / tabs / profile surface。
- 目前之所以还停在 `components/platform/challenge`，主要是历史拆分先把 route view 减重，但没有继续把 feature-owned UI 落位到 `features/*/ui`。
- 最小收口面是迁 UI owner 和引用路径，不改 widget / route / model 的职责。

## 设计边界

### `features/platform-challenge-detail/ui/*` 本轮负责

- 题目详情 topbar
- 题目详情 tabs rail / panel shell
- 题目详情 profile surface
- 已存在的 flag config 子 UI

### widget 本轮继续负责

- 把 topbar、tabs、writeup panel 组合成 workspace
- 转发顶层交互事件

### feature model / route view 本轮继续负责

- challenge 加载、flag 保存、query tab 同步
- 打开拓扑、返回题库、打开题解页等导航 owner

## 任务切片

### Slice 1：feature UI 落位

- 目标：
  - 新增 `features/platform-challenge-detail/ui/AdminChallengeTopbarPanel.vue`
  - 新增 `features/platform-challenge-detail/ui/AdminChallengeWorkspaceTabs.vue`
  - 新增 `features/platform-challenge-detail/ui/AdminChallengeProfilePanel.vue`
  - `PlatformChallengeDetailWorkspace.vue` 改从 feature public API 读取
- 验证：
  - `npm run test:run -- src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.test.ts src/views/platform/__tests__/ChallengeDetail.test.ts`
- Review focus：
  - widget 是否仍然只是组合 / 转发层
  - tabs / profile 的 props、slot、emit 契约是否保持稳定

### Slice 2：allowlist 与 raw-source 护栏同步

- 目标：
  - 收掉 `architectureAllowlist.ts` 里 `platform-challenge-detail` 的 2 条 component->feature 例外和 2 条 widget->legacy component 例外
  - 切换 challenge detail 相关 raw-source 测试与 `components.d.ts` 到新路径
- 验证：
  - `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeQueryTabsAdoption.test.ts src/views/platform/__tests__/challengeDetailWorkspaceExtraction.test.ts src/views/platform/__tests__/challengeDetailPanelExtraction.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts`
- Review focus：
  - route page / widget 是否继续只组合 feature public API
  - raw-source 护栏是否仍覆盖 topbar / tabs / profile / flag panel 这条 surface

## 验证计划

- `cd code/frontend && npm run test:run -- src/widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.test.ts src/views/platform/__tests__/ChallengeDetail.test.ts src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeQueryTabsAdoption.test.ts src/views/platform/__tests__/challengeDetailWorkspaceExtraction.test.ts src/views/platform/__tests__/challengeDetailPanelExtraction.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/views/__tests__/workspacePageHeaderStyles.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`

## 残余风险

- `ChallengeDescriptionPanel.vue`、`ChallengeManageDirectoryPanel.vue` 等其他 challenge 业务组件仍会留在原目录，本轮不扩大到 challenge 全族迁移。
