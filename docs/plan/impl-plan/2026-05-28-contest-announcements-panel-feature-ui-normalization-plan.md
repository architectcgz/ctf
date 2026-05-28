# Contest Announcements Panel Feature UI Normalization Plan

## 目标

- 把 `ContestAnnouncementsTopbarPanel.vue` 与 `ContestAnnouncementsWorkspacePanel.vue` 迁入 `features/platform-contests/ui`。
- 让 `ContestAnnouncements.vue` 改为通过 `@/features/platform-contests` public API 组合 panel。
- 同步组件声明、raw-source 护栏测试和 backlog 记录。

## 非目标

- 本轮不改 `useContestAnnouncementsPage()` 的加载、返回工作台、提交或删除逻辑。
- 本轮不改 `useContestAnnouncementManagement()` 的权限、校验、toast 或 API owner。
- 本轮不重排 `ContestAnnouncements.vue` 的 loading / error / empty 结构。

## 输入依据

- `code/frontend/src/views/platform/ContestAnnouncements.vue`
- `code/frontend/src/features/platform-contests/model/useContestAnnouncementsPage.ts`
- `code/frontend/src/features/platform-contests/ui/index.ts`
- `code/frontend/src/components/platform/contest/ContestAnnouncementsTopbarPanel.vue`
- `code/frontend/src/components/platform/contest/ContestAnnouncementsWorkspacePanel.vue`
- `code/frontend/src/views/platform/__tests__/ContestAnnouncements.test.ts`
- `code/frontend/src/views/platform/__tests__/contestAnnouncementsPanelExtraction.test.ts`
- `code/frontend/src/views/platform/__tests__/contestAnnouncementsWorkspaceExtraction.test.ts`
- `code/frontend/src/components.d.ts`

## 当前结论

- `ContestAnnouncements.vue` 是 route shell，owner 已在 `useContestAnnouncementsPage()`。
- 顶栏与工作区 panel 只服务这一路由，属于 `platform-contests` 的 feature-owned UI。
- 最小正确落点是 `features/platform-contests/ui/*`，并通过 `features/platform-contests` public API 暴露。

## 设计边界

### `features/platform-contests/ui/*` 本轮负责

- `ContestAnnouncementsTopbarPanel.vue`
- `ContestAnnouncementsWorkspacePanel.vue`
- `ContestAnnouncements.vue` 对这两个 panel 的 public API 组合

### `features/platform-contests/model/*` 本轮继续负责

- `useContestAnnouncementsPage()` 的 route / page workflow

### `features/contest-announcements/*` 本轮继续负责

- 公告列表加载、发布、删除、权限判定、表单错误与 toast owner

### `components/platform/contest/*` 本轮不再负责

- `ContestAnnouncementsTopbarPanel.vue`
- `ContestAnnouncementsWorkspacePanel.vue`

## 任务切片

### Slice 1：feature UI owner 迁位

- 目标：
  - 新增 `features/platform-contests/ui/ContestAnnouncementsTopbarPanel.vue`
  - 新增 `features/platform-contests/ui/ContestAnnouncementsWorkspacePanel.vue`
  - `ContestAnnouncements.vue` 改为从 `@/features/platform-contests` public API 引用 panel
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/ContestAnnouncements.test.ts`
- Review focus：
  - route shell 是否继续只保留 page owner 与 panel 组合
  - panel 行为与 props / emits contract 是否保持不变

### Slice 2：护栏同步

- 目标：
  - 更新 `features/platform-contests/ui/index.ts`
  - 更新 `components.d.ts`
  - 更新 panel extraction / workspace extraction 测试
  - backlog 记录 ContestAnnouncements feature UI 收口进展
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/contestAnnouncementsPanelExtraction.test.ts src/views/platform/__tests__/contestAnnouncementsWorkspaceExtraction.test.ts`
- Review focus：
  - touched surface 是否已不再依赖旧 `components/platform/contest/ContestAnnouncements*.vue` 路径
  - `features/platform-contests` public API 是否成为唯一组合入口

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestAnnouncements.test.ts src/views/platform/__tests__/contestAnnouncementsPanelExtraction.test.ts src/views/platform/__tests__/contestAnnouncementsWorkspaceExtraction.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh`
- `cd /home/azhi/workspace/projects/ctf && bash scripts/check-workflow-complete.sh`

## 残余风险

- 这轮只处理 contest announcements 的 page-sized panel owner，不继续拆公告 feature 内部更细的表单或历史列表 primitive。
