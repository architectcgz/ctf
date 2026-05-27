> 状态：Current
> 事实源：contest announcements feature-owned UI normalization
> 替代：无

# Contest Announcements Feature UI Normalization Plan

## 目标

- 把 `ContestAnnouncementRealtimeBridge.vue` 与 `ContestAnnouncementManageDrawer.vue` 迁入 `features/contest-announcements/ui`。
- 让 `ContestDetail.vue` 和 `ContestManage.vue` 改为通过 `features/contest-announcements` public API 组合公告 UI。

## 非目标

- 本轮不改 `useContestAnnouncementRealtime.ts` 的 websocket owner。
- 本轮不改 `useContestAnnouncementManagement.ts` 的加载、发布、删除 owner。
- 本轮不重做公告管理 drawer 的视觉样式或交互流程。

## 输入依据

- `code/frontend/src/components/contests/ContestAnnouncementRealtimeBridge.vue`
- `code/frontend/src/components/platform/contest/ContestAnnouncementManageDrawer.vue`
- `code/frontend/src/features/contest-announcements/index.ts`
- `code/frontend/src/views/contests/ContestDetail.vue`
- `code/frontend/src/views/platform/ContestManage.vue`
- `code/frontend/src/__tests__/architectureAllowlist.ts`
- `docs/todos/2026-05-26-frontend-tech-debt-priority-backlog.md`

## 当前结论

- `contest-announcements` 已有明确的 feature model owner，但 `ui/` 目录缺位，导致 route/page 组合层仍要回头引用旧组件路径。
- realtime bridge 和管理 drawer 都是单一 feature UI，不是共享原语，也不该继续停在 `components/**`。
- 最小收口面是补 UI 落点、切换 route/page 组合层引用、同步 allowlist 和测试。

## 设计边界

### `features/contest-announcements/ui/*` 本轮负责

- 公告 websocket realtime bridge
- 公告管理 drawer

### `features/contest-announcements/model/*` 本轮继续负责

- websocket 连接与事件回调
- 公告列表加载、发布、删除、表单状态

### route / page 组合层本轮继续负责

- `ContestDetail.vue` 组合竞赛 workspace 和 realtime bridge
- `ContestManage.vue` 组合竞赛目录、公告 drawer、赛事表单和 AWD readiness dialog

## 任务切片

### Slice 1：feature UI 落位

- 目标：
  - 新增 `features/contest-announcements/ui/ContestAnnouncementRealtimeBridge.vue`
  - 新增 `features/contest-announcements/ui/ContestAnnouncementManageDrawer.vue`
  - `ContestDetail.vue`、`ContestManage.vue` 改从 feature public API 读取
- 验证：
  - `npm run test:run -- src/views/platform/__tests__/ContestManage.test.ts`
- Review focus：
  - realtime bridge 是否仍只负责启动订阅与转发更新事件
  - drawer 的发布 / 删除 / load / close owner 是否保持在 feature model

### Slice 2：allowlist 与类型声明同步

- 目标：
  - 收掉 `architectureAllowlist.ts` 里 `contest-announcements` 的 2 条 component->feature 例外
  - 更新 `components.d.ts`
- 验证：
  - `npm run test:run -- src/__tests__/architectureBoundaries.test.ts`
- Review focus：
  - route/page 组合层是否不再直连旧组件路径
  - feature public API 是否足够承接这组 UI

## 验证计划

- `cd code/frontend && npm run test:run -- src/views/platform/__tests__/ContestManage.test.ts src/__tests__/architectureBoundaries.test.ts`
- `cd code/frontend && npm run typecheck`
- `cd /home/azhi/workspace/projects/ctf && git diff --check`

## 残余风险

- `ContestAnnouncementsWorkspaceSection.vue`、`ContestAnnouncementsWorkspacePanel.vue` 等更大块公告 surface 仍会留在原目录，本轮只收口当前被 allowlist 冻结的两块单一 feature UI。
