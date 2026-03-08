# CTF 前端 UI 明日待办

## 目的

- 记录当前未提交的 UI 改造状态
- 明确明天继续开发的优先顺序
- 避免重新进入代码时先花时间回忆上下文

## 当前状态

- 学员端仪表盘已拆成独立页面：
  - `总览`
  - `训练建议`
  - `分类进度`
  - `近期动态`
  - `难度分布`
- 教师端首页已改成 `教学介入台`
- 管理员首页已改成 `系统值守台`
- 通知下拉层级、透明背景、裁剪问题已处理
- 当前仓库仍有一批前端 UI 改造处于未提交状态，明天应继续在这批改动上收口

## 明日优先级

### P0 先做收口验证

1. 手工检查学生端 5 个仪表盘子页
   - 路由：
     - `/dashboard`
     - `/dashboard?panel=recommendation`
     - `/dashboard?panel=category`
     - `/dashboard?panel=timeline`
     - `/dashboard?panel=difficulty`
   - 检查点：
     - 左侧导航选中态是否正确
     - 顶栏标题是否和子页一致
     - 首屏是否仍残留主页内容
     - 移动端是否出现横向溢出

2. 手工检查教师端与管理员首页
   - 页面：
     - `TeacherDashboard`
     - `AdminDashboard`
   - 检查点：
     - 独立页面语言是否成立，不像“同一模板换文案”
     - 卡片层级、按钮位置、空态和 loading 态是否一致
     - 深色背景下文字对比度是否足够

3. 回归通知链路
   - 检查顶部通知下拉在以下页面是否仍正常：
     - 学员仪表盘
     - 教师首页
     - 管理员首页
   - 重点确认：
     - 不被内容区遮挡
     - 不透明背景是否稳定
     - 小屏下定位是否正确

### P1 继续做页面专属化

1. 改造 `ClassManagement`
   - 目标：从“普通管理页”继续往“教师工作台子页”收
   - 方向：
     - 班级列表区更像教学分组导航
     - 学员列表和班级详情分层更明确
     - 减少通用后台卡片感

2. 改造 `UserManage`
   - 目标：从标准 CRUD 页改成更像“用户治理台”
   - 方向：
     - 用户状态、角色、批量导入做更清晰分区
     - 筛选和动作区不要继续沿用通用表格页心智

3. 改造 `ContestManage`
   - 目标：让竞赛管理页更像“赛事编排页”
   - 方向：
     - 列表、状态、时间、操作做更明显的赛事编排层次
     - 明确保留“无删除能力”的边界提示

### P2 视觉与工程收尾

1. 清测试 warning
   - 当前已知：
     - `ElButton` 在测试里仍有组件解析 warning
     - 管理员测试里仍有 router injection warning
   - 目标：让 UI 相关测试输出更干净

2. 统一标题与页面命名
   - 检查 `routeTitle`、`TopNav`、页面标题、副标题是否完全一致
   - 避免“导航叫 A、页头叫 B、浏览器标题叫 C”

3. 评估构建包体 warning
   - 当前 `vite build` 仍提示大 chunk
   - 这不是明天必须做，但可以评估：
     - 是否对重型页面做动态分包
     - 是否把大图表/重依赖延迟加载

## 建议执行顺序

1. 先做 P0 手工验收
2. 再做 `ClassManagement`
3. 然后做 `UserManage`
4. 再做 `ContestManage`
5. 最后处理 warning、标题统一和构建包体评估

## 明天开始前先跑

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
npm run dev
```

如需先确认当前代码仍可编译：

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
npm run typecheck
npx vitest run src/views/dashboard/__tests__/DashboardView.test.ts src/views/teacher/__tests__/TeacherDashboard.test.ts src/views/admin/__tests__/AdminDashboard.test.ts
```

## 相关文件入口

- 任务总文档：
  - [frontend-task-breakdown.md](/home/azhi/workspace/projects/ctf/code/docs/tasks/frontend-task-breakdown.md)
- 学员仪表盘：
  - [DashboardView.vue](/home/azhi/workspace/projects/ctf/code/frontend/src/views/dashboard/DashboardView.vue)
  - [dashboard/](/home/azhi/workspace/projects/ctf/code/frontend/src/components/dashboard/)
- 教师首页：
  - [TeacherDashboard.vue](/home/azhi/workspace/projects/ctf/code/frontend/src/views/teacher/TeacherDashboard.vue)
  - [TeacherDashboardPage.vue](/home/azhi/workspace/projects/ctf/code/frontend/src/components/teacher/dashboard/TeacherDashboardPage.vue)
- 管理员首页：
  - [AdminDashboard.vue](/home/azhi/workspace/projects/ctf/code/frontend/src/views/admin/AdminDashboard.vue)
  - [AdminDashboardPage.vue](/home/azhi/workspace/projects/ctf/code/frontend/src/components/admin/dashboard/AdminDashboardPage.vue)

## 完成标准

- 三类首页和学生端子页都通过手工视觉验收
- `ClassManagement`、`UserManage`、`ContestManage` 至少完成其中 1 到 2 个页面的专属化改造
- `typecheck` 和核心 UI 测试通过
- 若准备提交，再补 review 和任务文档同步
