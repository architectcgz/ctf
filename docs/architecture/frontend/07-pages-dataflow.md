# 前端页面视图与数据流

> 对应：design-system/pages/*.md

---

## 1. 页面 → 组件 → API 映射

### 1.1 学员页面

| 页面 | 核心组件 | API 调用 | WebSocket |
|------|----------|----------|-----------|
| DashboardView | StatCard ×4, RadarChart, AppCard | `getProfile`, `getSkillProfile`, `getChallenges`, `getContests` | - |
| ChallengeList | AppInput(搜索), AppSelect(筛选) ×3, ChallengeCard(网格), AppPagination | `getChallenges(params)` | - |
| ChallengeDetail | MarkdownRenderer, FlagInput, InstanceControl, HintList | `getChallengeDetail`, `submitFlag`, `createInstance`, `destroyInstance`, `extendInstance`, `unlockHint` | - |
| ContestList | ContestCard(列表), StatusTabs | `getContests(params)` | - |
| ContestDetail | Tab(题目/公告/队伍/排行榜), ChallengeGrid, AnnouncementList, TeamPanel, ScoreboardTable, LineChart | `getContestDetail`, `getContestChallenges`, `getAnnouncements`, `getScoreboard` | `scoreboard/:id`, `contest/:id/announcements` |
| ScoreboardView | AppTable(highlightTop=3), LineChart, FrozenBanner | `getScoreboard` | `scoreboard/:id`(竞赛模式) |
| InstanceList | InstanceCard(列表), QueueStatus | `getMyInstances`, `destroyInstance`, `extendInstance` | - |
| SkillProfile | RadarChart, ProgressBar ×6, WeaknessAlert, RecommendCard ×3 | `getSkillProfile`, `getRecommendations` | - |
| UserProfile | AvatarDisplay, ProfileForm, PasswordForm | `getProfile`, `changePassword` | - |
| NotificationList | NotificationItem(列表), CategoryTabs | `getNotifications`, `markAsRead` | `notifications` |

### 1.2 教师页面

| 页面 | 核心组件 | API 调用 |
|------|----------|----------|
| TeacherDashboard | StatCard ×4, RadarChart(多组叠加), BarChart, AppCard | `getClasses`, `getClassStudents` |
| ClassManagement | AppSelect(班级), AppTable(学员), StudentDetailDrawer | `getClasses`, `getClassStudents`, `getStudentProgress` |
| ReportExport | ReportForm, HistoryList | `exportPersonalReport`, `exportClassReport` |

### 1.3 管理员页面

| 页面 | 核心组件 | API 调用 |
|------|----------|----------|
| AdminDashboard | StatCard ×4, GaugeChart ×2, LineChart, BarChart, LogList, AlertList | `getDashboard` |
| ChallengeManage | AppTable, FilterBar, ChallengeFormDrawer(分步表单) | `admin.getChallenges`, `createChallenge`, `updateChallenge`, `deleteChallenge` |
| ContestManage | AppTable, FilterBar, ContestFormDrawer | `admin.getContests`, `createContest`, `updateContest`, `deleteContest` |
| UserManage | AppTable, FilterBar, UserFormDrawer, ImportDialog | `admin.getUsers`, `createUser`, `updateUser`, `deleteUser`, `importUsers` |
| ImageManage | AppTable, UploadDialog | `admin.getImages`, `createImage`, `deleteImage` |
| CheatDetection | AlertCard(列表), StatusTabs, CheatDetailDrawer | 待定（后端 API 未在 04 中列出，需补充） |
| AuditLog | AppTable, FilterBar, DateRangePicker | `admin.getAuditLogs` |

---

## 2. 关键交互流程

### 2.1 Flag 提交流程

```
用户输入 Flag → 点击提交
  │
  ├─ 按钮 loading 态
  ├─ POST /challenges/:id/submissions
  │
  ├─ 成功 (correct: true)
  │   ├─ 输入框边框变绿 0.3s
  │   ├─ Toast success "恭喜！获得 {points} 分"
  │   ├─ 首血？ → Toast accent "首血！"
  │   └─ 刷新靶场卡片状态（已解出）
  │
  ├─ 失败 (correct: false)
  │   ├─ 输入框水平抖动 0.3s
  │   └─ Toast error "Flag 错误"
  │
  └─ 429 限流
      ├─ 按钮禁用 + 显示倒计时
      └─ Toast warning "提交过于频繁"
```

### 2.2 靶机实例启动流程

```
点击 "启动靶机"
  │
  ├─ 按钮 loading "正在部署..."
  ├─ POST /challenges/:id/instances
  │
  ├─ 201 成功
  │   ├─ 显示访问地址 + 复制按钮
  │   ├─ 启动倒计时 (useCountdown)
  │   └─ 显示延时/销毁按钮
  │
  ├─ 13002 实例超限
  │   └─ Toast warning + 引导跳转"我的实例"页
  │
  └─ 16002 资源不足（排队）
      ├─ 显示排队状态卡片
      ├─ 轮询 GET /instances 查询状态（5s 间隔）
      └─ 状态变为 running → 切换为正常显示
```

### 2.3 竞赛排行榜实时更新

```
进入竞赛排行榜 Tab
  │
  ├─ GET /contests/:id/scoreboard 初始加载
  ├─ useWebSocket('scoreboard/:id', handlers) 建立连接
  │
  ├─ 收到 scoreboard.update
  │   ├─ 更新 contestStore.scoreboard 对应行
  │   └─ 触发行位移动画 (300ms translateY)
  │
  ├─ 收到 scoreboard.frozen
  │   ├─ contestStore.isFrozen = true
  │   └─ 显示冻结横幅
  │
  └─ 离开 Tab → disconnect
```
