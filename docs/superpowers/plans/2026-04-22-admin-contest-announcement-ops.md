# Admin Contest Announcement Ops Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为管理员补齐竞赛公告的快速发布抽屉、单场公告管理页和对应入口，同时保持学生侧公告页只读。

**Architecture:** 在现有管理员竞赛目录和单场竞赛工作台上增加两个入口，公告业务收口到单独的 `useContestAnnouncementManagement` composable，并通过 admin 侧公告 API 统一处理列表、发布、删除和 ended 只读规则。竞赛列表抽屉和单场公告管理页复用同一套数据逻辑，避免双份状态机和未捕获的异步错误。

**Tech Stack:** Vue 3 Composition API、Vue Router、Vitest、Vue Test Utils、现有 `AdminSurfaceDrawer` / `CActionMenu` / `AppEmpty` / `AppLoading` 原语

---

### Task 1: 补齐管理员竞赛公告 API 契约

**Files:**
- Modify: `code/frontend/src/api/admin.ts`
- Modify: `code/frontend/src/api/contracts.ts`
- Test: `code/frontend/src/api/__tests__/admin.test.ts`

- [ ] **Step 1: 写管理员竞赛公告 API 的失败测试**

在 `code/frontend/src/api/__tests__/admin.test.ts` 中新增三个用例：
- `getAdminContestAnnouncements` 应调用 `GET /admin/contests/:id/announcements` 并归一化 `id`
- `createAdminContestAnnouncement` 应调用 `POST /admin/contests/:id/announcements`
- `deleteAdminContestAnnouncement` 应调用 `DELETE /admin/contests/:id/announcements/:aid`

- [ ] **Step 2: 运行测试确认失败**

Run: `npm run test:run -- src/api/__tests__/admin.test.ts`
Expected: FAIL，提示缺少新导出或行为不匹配。

- [ ] **Step 3: 写最小实现**

在 `code/frontend/src/api/contracts.ts` 复用或补齐管理员公告所需类型，在 `code/frontend/src/api/admin.ts` 增加三个 API 函数，保持和现有 admin API 的命名及归一化模式一致。

- [ ] **Step 4: 运行测试确认通过**

Run: `npm run test:run -- src/api/__tests__/admin.test.ts`
Expected: PASS


### Task 2: 收口公告管理业务状态到 composable

**Files:**
- Create: `code/frontend/src/composables/useContestAnnouncementManagement.ts`
- Create: `code/frontend/src/composables/__tests__/useContestAnnouncementManagement.test.ts`

- [ ] **Step 1: 写 composable 的失败测试**

覆盖最小行为：
- 加载公告列表成功后写入 `announcements`
- 发布成功后清空表单并刷新列表
- 删除成功后刷新列表
- ended 状态下 `canManageAnnouncements` 为 false
- 发布/删除失败时在本地消费异常，不向外抛出未处理 rejection

- [ ] **Step 2: 运行测试确认失败**

Run: `npm run test:run -- src/composables/__tests__/useContestAnnouncementManagement.test.ts`
Expected: FAIL

- [ ] **Step 3: 写最小实现**

在 composable 中收口：
- `loadAnnouncements`
- `publishAnnouncement`
- `deleteAnnouncement`
- `canManageAnnouncements`
- 局部错误和 loading 状态

要求：
- 以 `contest.status !== 'ended'` 作为前端只读判断
- 所有用户触发 async 都本地 catch

- [ ] **Step 4: 运行测试确认通过**

Run: `npm run test:run -- src/composables/__tests__/useContestAnnouncementManagement.test.ts`
Expected: PASS


### Task 3: 接入竞赛列表快速发布抽屉

**Files:**
- Create: `code/frontend/src/components/platform/contest/ContestAnnouncementManageDrawer.vue`
- Modify: `code/frontend/src/components/platform/contest/PlatformContestTable.vue`
- Modify: `code/frontend/src/components/platform/contest/ContestOrchestrationPage.vue`
- Modify: `code/frontend/src/views/platform/ContestManage.vue`
- Test: `code/frontend/src/components/platform/__tests__/PlatformContestTable.test.ts`
- Test: `code/frontend/src/views/platform/__tests__/ContestManage.test.ts`
- Optional reference: `code/frontend/src/components/notifications/AdminNotificationPublishDrawer.vue`

- [ ] **Step 1: 先给列表入口写失败测试**

在 `PlatformContestTable.test.ts` 增加：
- 非 ended 竞赛的更多菜单出现 `发布通知`
- ended 竞赛不出现
- 点击菜单项后触发新事件，如 `announce`

在 `ContestManage.test.ts` 增加：
- 接到 `announce` 后打开公告抽屉
- 抽屉拿到当前竞赛上下文

- [ ] **Step 2: 运行相关测试确认失败**

Run:
- `npm run test:run -- src/components/platform/__tests__/PlatformContestTable.test.ts`
- `npm run test:run -- src/views/platform/__tests__/ContestManage.test.ts`

Expected: FAIL

- [ ] **Step 3: 实现菜单入口和抽屉接线**

要求：
- 抽屉使用 `AdminSurfaceDrawer`
- 列表页只负责打开/关闭和传入 contest
- 抽屉内部复用 `useContestAnnouncementManagement`
- 抽屉中提供“进入完整管理页”按钮

- [ ] **Step 4: 运行相关测试确认通过**

Run:
- `npm run test:run -- src/components/platform/__tests__/PlatformContestTable.test.ts`
- `npm run test:run -- src/views/platform/__tests__/ContestManage.test.ts`

Expected: PASS


### Task 4: 新增单场竞赛公告管理页与路由

**Files:**
- Create: `code/frontend/src/views/platform/ContestAnnouncements.vue`
- Modify: `code/frontend/src/router/index.ts`
- Test: `code/frontend/src/views/platform/__tests__/ContestAnnouncements.test.ts`

- [ ] **Step 1: 写新页面路由与只读态失败测试**

覆盖最小行为：
- 路由 `/platform/contests/:id/announcements`
- 页面加载竞赛详情和公告列表
- ended 竞赛显示只读提示，不显示发布或删除操作

- [ ] **Step 2: 运行测试确认失败**

Run: `npm run test:run -- src/views/platform/__tests__/ContestAnnouncements.test.ts`
Expected: FAIL

- [ ] **Step 3: 写最小实现**

页面结构：
- 顶部竞赛标题、状态、返回按钮
- 发布区
- 公告列表

要求：
- 复用 `useContestAnnouncementManagement`
- 错误在页面内显示，不抛全局 500

- [ ] **Step 4: 运行测试确认通过**

Run: `npm run test:run -- src/views/platform/__tests__/ContestAnnouncements.test.ts`
Expected: PASS


### Task 5: 从单场工作台顶部增加“公告”入口

**Files:**
- Modify: `code/frontend/src/views/platform/ContestEdit.vue`
- Test: `code/frontend/src/views/platform/__tests__/ContestEdit.test.ts`

- [ ] **Step 1: 写顶部入口失败测试**

在 `ContestEdit.test.ts` 增加：
- 顶部显示“公告”入口按钮
- 点击后跳转到 `ContestAnnouncements` 路由

- [ ] **Step 2: 运行测试确认失败**

Run: `npm run test:run -- src/views/platform/__tests__/ContestEdit.test.ts`
Expected: FAIL

- [ ] **Step 3: 实现跳转按钮**

要求：
- 入口位于工作台顶部右侧工具区
- 不改变现有保存按钮和 AWD 操作入口的可见性逻辑

- [ ] **Step 4: 运行测试确认通过**

Run: `npm run test:run -- src/views/platform/__tests__/ContestEdit.test.ts`
Expected: PASS


### Task 6: 做最小回归验证

**Files:**
- Test only: `code/frontend/src/views/contests/__tests__/ContestDetail.test.ts`

- [ ] **Step 1: 跑学生侧公告页回归**

Run: `npm run test:run -- src/views/contests/__tests__/ContestDetail.test.ts`
Expected: 学生侧公告读取与 websocket 刷新相关用例继续通过；如果存在无关旧失败，单独记录。

- [ ] **Step 2: 跑本次新增/修改的最小测试集合**

Run:
- `npm run test:run -- src/api/__tests__/admin.test.ts`
- `npm run test:run -- src/composables/__tests__/useContestAnnouncementManagement.test.ts`
- `npm run test:run -- src/components/platform/__tests__/PlatformContestTable.test.ts`
- `npm run test:run -- src/views/platform/__tests__/ContestManage.test.ts`
- `npm run test:run -- src/views/platform/__tests__/ContestAnnouncements.test.ts`
- `npm run test:run -- src/views/platform/__tests__/ContestEdit.test.ts`

Expected: 本次改动相关用例通过

- [ ] **Step 3: 提交实现**

```bash
git add code/frontend/src/api/admin.ts \
  code/frontend/src/api/contracts.ts \
  code/frontend/src/composables/useContestAnnouncementManagement.ts \
  code/frontend/src/composables/__tests__/useContestAnnouncementManagement.test.ts \
  code/frontend/src/components/platform/contest/ContestAnnouncementManageDrawer.vue \
  code/frontend/src/components/platform/contest/PlatformContestTable.vue \
  code/frontend/src/components/platform/contest/ContestOrchestrationPage.vue \
  code/frontend/src/views/platform/ContestManage.vue \
  code/frontend/src/views/platform/ContestAnnouncements.vue \
  code/frontend/src/views/platform/ContestEdit.vue \
  code/frontend/src/router/index.ts \
  code/frontend/src/api/__tests__/admin.test.ts \
  code/frontend/src/components/platform/__tests__/PlatformContestTable.test.ts \
  code/frontend/src/views/platform/__tests__/ContestManage.test.ts \
  code/frontend/src/views/platform/__tests__/ContestAnnouncements.test.ts \
  code/frontend/src/views/platform/__tests__/ContestEdit.test.ts \
  docs/superpowers/plans/2026-04-22-admin-contest-announcement-ops.md
git commit -m "feat(前端): 补齐竞赛公告运营入口"
```
