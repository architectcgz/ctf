# 页面设计：竞赛公告 (Contest Announcements)

> 继承：../design-system/MASTER.md | 角色：学生 / 管理员
> 当前覆盖：
> - 学生页 `/contests/:id?panel=announcements`
> - 管理员抽屉 `ContestAnnouncementManageDrawer`
> - 管理员页 `/platform/contests/:id/announcements`

---

## 页面定位

- 学生侧只读消费竞赛公告，和题目、队伍并列存在于同一竞赛工作区。
- 管理员侧提供两层运营入口：
  - 竞赛目录内的快速发布抽屉
  - 单场竞赛的完整公告管理页
- 教师端当前没有发布、删除竞赛公告的入口。

---

## 1. 学生侧公告面板

入口位于 `ContestDetail.vue` 的 `announcements` panel，保持和其他竞赛子页一致的 `top-tabs + workspace panel` 结构。

页面内容分为两层：

- 面板头部显示 `Announcements` overline、标题 `公告` 和当前公告数量。
- 正文复用 `ContestAnnouncementsPanel.vue`，按时间倒序展示公告列表。

列表项只展示当前已经稳定存在的三项信息：

- 标题
- 发布时间
- 正文内容（为空时不额外补占位块）

状态表现：

- 请求失败时，使用页内 warning alert，不跳离竞赛页。
- 没有公告时，展示 `AppEmpty` 或内联空态。
- 公告到达后，通过 `ContestAnnouncementRealtimeBridge` 触发刷新，学生侧不出现任何管理动作。

---

## 2. 管理员侧快速发布抽屉

入口位于 `ContestManage.vue` 的竞赛目录表格：

- 行内主按钮仍保留 `编辑` / `运维`
- `更多` 菜单增加 `发布通知`
- 点击后打开 `ContestAnnouncementManageDrawer`

抽屉结构：

1. 头部
   - 标题使用 `{竞赛标题} · 公告`
   - 显示当前赛事状态
   - 提供 `进入完整管理页` 按钮

2. 发布区
   - 字段只有 `标题` 和 `内容`
   - 两项都为必填
   - 提交按钮文案为 `发布公告`

3. 历史公告区
   - 显示当前公告条数
   - 列表项展示标题、发布时间、正文
   - 非结束赛事允许删除

状态规则：

- `contest.status !== 'ended'` 时允许发布和删除。
- 已结束赛事只保留历史查看，不再渲染发布表单和删除动作。
- 加载、错误、空状态都在抽屉内局部消化，不跳出竞赛目录。

---

## 3. 管理员完整管理页

完整页路由为 `/platform/contests/:id/announcements`，角色限制为 `admin`。

页面骨架：

- 顶部使用 `ContestAnnouncementsTopbarPanel`
  - 返回入口回到 `ContestEdit`
  - 显示 `Contest Announcements` overline
  - 标题使用竞赛名称
  - 右侧显示当前赛事状态

- 正文使用 `ContestAnnouncementsWorkspacePanel`
  - 上半区为发布表单或只读提示
  - 下半区为历史公告列表

完整页和抽屉共用同一套公告管理状态：

- 公告读取：`getAdminContestAnnouncements`
- 发布：`createAdminContestAnnouncement`
- 删除：`deleteAdminContestAnnouncement`
- 表单校验、loading、toast 和只读判断保持一致

---

## 4. 共享交互规则

发布流程：

- 提交前校验标题和内容非空
- 成功后清空表单、刷新列表，并提示 `公告已发布`
- 失败后保留当前输入，在当前上下文内提示错误

删除流程：

- 仅管理员且赛事未结束时可执行
- 成功后刷新列表，并提示 `公告已删除`
- 失败时保留当前列表，不跳页

权限与读取路径：

- 学生侧读取 `/api/v1/contests/:id/announcements`
- 管理员侧读取和写入 `/api/v1/admin/contests/:id/announcements`
- 当前页面设计不承接全站通知，也不扩展到教师运营页

---

## 代码落点

- 学生侧：
  - `code/frontend/src/views/contests/ContestDetail.vue`
  - `code/frontend/src/components/contests/ContestAnnouncementsPanel.vue`
  - `code/frontend/src/components/contests/ContestAnnouncementRealtimeBridge.vue`
- 管理端：
  - `code/frontend/src/views/platform/ContestManage.vue`
  - `code/frontend/src/components/platform/contest/ContestAnnouncementManageDrawer.vue`
  - `code/frontend/src/views/platform/ContestAnnouncements.vue`
  - `code/frontend/src/components/platform/contest/ContestAnnouncementsTopbarPanel.vue`
  - `code/frontend/src/components/platform/contest/ContestAnnouncementsWorkspacePanel.vue`
  - `code/frontend/src/features/platform-contests/model/useContestAnnouncementsPage.ts`
  - `code/frontend/src/features/contest-announcements/model/useContestAnnouncementManagement.ts`
