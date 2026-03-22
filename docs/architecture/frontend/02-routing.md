# 前端路由设计

> 对应：01-architecture-overview.md §2 目录结构

---

## 1. 路由表

### 1.1 公开路由（无需认证）

| 路径 | 名称 | 视图组件 | 说明 |
|------|------|----------|------|
| `/login` | Login | `auth/LoginView.vue` | 登录页 |
| `/register` | Register | `auth/RegisterView.vue` | 注册页 |

### 1.2 主布局路由（需认证，AppLayout 包裹）

| 路径 | 名称 | 视图组件 | 角色 | 说明 |
|------|------|----------|------|------|
| `/dashboard` | Dashboard | `dashboard/DashboardView.vue` | 全部 | 仪表盘（根据角色渲染不同内容） |
| `/challenges` | Challenges | `challenges/ChallengeList.vue` | student, teacher | 靶场列表 |
| `/challenges/:id` | ChallengeDetail | `challenges/ChallengeDetail.vue` | student, teacher | 靶场详情 + Flag 提交 |
| `/contests` | Contests | `contests/ContestList.vue` | 全部 | 竞赛列表 |
| `/contests/:id` | ContestDetail | `contests/ContestDetail.vue` | 全部 | 竞赛详情（Tab 子视图） |
| `/scoreboard` | Scoreboard | `scoreboard/ScoreboardView.vue` | 全部 | 全站排行榜 |
| `/instances` | Instances | `instances/InstanceList.vue` | student | 我的实例 |
| `/skill-profile` | SkillProfile | `profile/SkillProfile.vue` | student | 能力评估 |
| `/profile` | Profile | `profile/UserProfile.vue` | 全部 | 个人资料 + 安全设置 |
| `/notifications` | Notifications | `notifications/NotificationList.vue` | 全部 | 通知中心 |

### 1.3 教师路由（需 teacher/admin 角色）

| 路径 | 名称 | 视图组件 | 说明 |
|------|------|----------|------|
| `/teacher/dashboard` | TeacherDashboard | `teacher/TeacherDashboard.vue` | 教学概览 |
| `/teacher/classes` | ClassManagement | `teacher/ClassManagement.vue` | 班级管理 |
| `/teacher/reports` | ReportExport | `teacher/ReportExport.vue` | 报告导出 |

### 1.4 管理员路由（需 admin 角色）

| 路径 | 名称 | 视图组件 | 说明 |
|------|------|----------|------|
| `/admin/dashboard` | AdminDashboard | `admin/AdminDashboard.vue` | 系统概览 |
| `/admin/challenges` | ChallengeManage | `admin/ChallengeManage.vue` | 靶场管理 |
| `/admin/contests` | ContestManage | `admin/ContestManage.vue` | 竞赛管理 |
| `/admin/users` | UserManage | `admin/UserManage.vue` | 用户管理 |
| `/admin/images` | ImageManage | `admin/ImageManage.vue` | 镜像管理 |
| `/admin/cheat` | CheatDetection | `admin/CheatDetection.vue` | 作弊检测 |
| `/admin/audit` | AuditLog | `admin/AuditLog.vue` | 审计日志 |

### 1.5 系统路由（兜底）

| 路径 | 名称 | 视图组件 | 说明 |
|------|------|----------|------|
| `/403` | Forbidden | `errors/ForbiddenView.vue` | 无权限页面（可选；也可 Toast 后重定向仪表盘） |
| `/:pathMatch(.*)*` | NotFound | `errors/NotFoundView.vue` | 404 页面（SPA fallback） |

---

## 2. 路由守卫

### 2.1 全局前置守卫流程

```
beforeEach(to, from)
  │
  ├─ 白名单路由（/login, /register）？ → 放行
  │
  ├─ 无 Token？ → 重定向 /login?redirect=to.fullPath
  │
  ├─ Token 存在但用户信息未加载？ → 调用 GET /api/v1/auth/profile 加载
  │   └─ 失败（Token 过期）？ → 尝试 refresh → 仍失败 → 清除状态 → /login
  │
  ├─ 路由 meta.roles 存在？ → 校验当前用户角色
  │   └─ 角色不匹配？ → 重定向 /dashboard（或 /403）+ 无权限 Toast 提示
  │
  └─ 放行
```

### 2.2 路由 meta 定义

```ts
{
  path: '/admin/challenges',
  name: 'ChallengeManage',
  component: () => import('@/views/admin/ChallengeManage.vue'),
  meta: {
    roles: ['admin'],       // 允许的角色列表
    title: '靶场管理',       // 页面标题（document.title）
    icon: 'Settings'        // 侧边栏图标名
  }
}
```

### 2.3 竞赛详情 Tab 子路由

ContestDetail 内部使用组件级 Tab 切换（非路由级），通过 query 参数 `?tab=challenges|announcements|team|scoreboard` 控制当前 Tab，支持浏览器前进后退和直接链接分享。

---

## 3. 懒加载策略

- 所有视图组件使用 `() => import()` 动态导入
- 按路由分组自动 code splitting（Vite 默认行为）
- ECharts 按需引入（仅导入 radar/line/bar/gauge 组件）
- md-editor-v3 仅在管理端编辑页引入，学员端使用轻量 Markdown 渲染

---

## 4. 页面标题与菜单元信息（约定）

- `meta.title`：用于设置 `document.title`（建议统一前缀，例如 `CTF 靶场平台 - {title}`）。
- `meta.icon`：侧边栏图标名（Lucide icon key）。
- `meta.roles`：与后端一致的角色枚举：`student` / `teacher` / `admin`（文档中表格的 S/T/A 仅用于说明权限范围）。
