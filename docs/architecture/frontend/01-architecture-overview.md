# CTF 靶场平台 — 前端架构设计

> 版本：v1.0 | 日期：2026-03-02 | 状态：初稿
> 对应后端架构：../backend/01-system-architecture.md
> 对应 API 设计：../backend/04-api-design.md
> 对应设计系统：design-system/ctf-platform/MASTER.md
> 对应间距体系：./09-spacing-system.md

---

## 1. 技术选型

| 组件 | 选型 | 版本 | 选型理由 |
|------|------|------|----------|
| 框架 | Vue 3 | 3.5+ | Composition API，TypeScript 支持，校园项目学习成本低 |
| 构建 | Vite | 7.x | 开发热更新快，Tailwind CSS 官方插件支持 |
| 路由 | Vue Router 4 | 4.x | 路由守卫做权限控制，支持嵌套路由和懒加载 |
| 状态管理 | Pinia | 3.x | Vue 3 官方推荐，TypeScript 友好，DevTools 集成 |
| UI 组件库 | Element Plus | 2.x | 覆盖表格/表单/弹窗等高频组件，减少自研成本，提升一致性与交付速度 |
| 样式 | Tailwind CSS 4 | 4.x | 原子化 CSS，通过 `@theme` 定义设计令牌，并以 CSS 变量桥接 Element Plus 主题 |
| HTTP | Axios | 1.x | 拦截器统一处理 Token 刷新和错误码映射 |
| 图表 | ECharts + vue-echarts | 6.x + 8.x | 能力雷达图、排行榜趋势图、仪表盘统计 |
| 图标 | Lucide Vue Next | 0.x（锁版本） | 统一图标风格，Tree-shakable |
| Markdown | md-editor-v3 | 5.x（锁版本） | 题目描述编辑（管理端）与渲染（学员端） |
| 终端 | xterm.js | 5.x | WebSocket + xterm 实现浏览器内 SSH 连接靶机（Phase 2） |
| WebSocket | 原生 WebSocket | - | 封装 composable，支持心跳、重连、ticket 认证 |

### Element Plus 使用边界（必须）

- **优先直接使用 Element Plus 组件**（`ElButton/ElInput/ElTable/ElDialog/...`），减少“二次封装造成的 API 漂移”。
- **只在两类场景封装 `App*` 组件**：
  - 需要统一业务行为（如：权限控制按钮、提交 loading、防重复点击、统一错误展示）
  - 需要统一视觉 token 映射（如：Flag 输入、排行榜高亮行、冻结横幅）
- **主题统一**：以 Tailwind `@theme` 设计令牌为唯一来源，通过 CSS 变量桥接到 Element Plus（例如 `--el-color-primary: var(--color-primary)`），避免双套主题体系。

---

## 2. 项目目录结构

```
frontend/
├── public/
│   └── favicon.svg
├── src/
│   ├── api/                    # API 请求层
│   │   ├── request.ts          # Axios 实例 + 拦截器
│   │   ├── auth.ts             # 认证相关 API
│   │   ├── challenge.ts        # 靶场 API
│   │   ├── contest.ts          # 竞赛 API
│   │   ├── instance.ts         # 实例 API
│   │   ├── assessment.ts       # 评估 API
│   │   ├── notification.ts     # 通知 API
│   │   ├── admin.ts            # 管理后台 API
│   │   └── teacher.ts          # 教师 API
│   ├── assets/                 # 静态资源
│   │   └── images/
│   ├── components/             # 全局通用组件
│   │   ├── common/             # 基础 UI 组件
│   │   │   ├── AppButton.vue
│   │   │   ├── AppCard.vue
│   │   │   ├── AppInput.vue
│   │   │   ├── AppSelect.vue
│   │   │   ├── AppTable.vue
│   │   │   ├── AppPagination.vue
│   │   │   ├── AppTag.vue
│   │   │   ├── AppToast.vue
│   │   │   ├── AppDialog.vue
│   │   │   ├── AppDrawer.vue
│   │   │   ├── AppDropdown.vue
│   │   │   ├── AppSkeleton.vue
│   │   │   ├── AppEmpty.vue
│   │   │   └── AppLoading.vue
│   │   ├── layout/             # 布局组件
│   │   │   ├── AppLayout.vue   # 主布局壳（TopNav + Sidebar + Content）
│   │   │   ├── TopNav.vue
│   │   │   ├── Sidebar.vue
│   │   │   └── NotificationDropdown.vue
│   │   └── charts/             # 图表组件
│   │       ├── RadarChart.vue  # 能力雷达图
│   │       ├── LineChart.vue   # 趋势折线图
│   │       ├── BarChart.vue    # 柱状图
│   │       └── GaugeChart.vue  # 仪表盘环形图
│   ├── composables/            # 组合式函数
│   │   ├── useAuth.ts          # 认证状态 + Token 管理
│   │   ├── useWebSocket.ts     # WebSocket 连接管理
│   │   ├── useToast.ts         # Toast 通知
│   │   ├── usePagination.ts    # 分页逻辑
│   │   ├── useCountdown.ts     # 倒计时（实例/竞赛）
│   │   └── useClipboard.ts     # 复制到剪贴板
│   ├── router/
│   │   ├── index.ts            # 路由定义 + 守卫
│   │   └── guards.ts           # 路由守卫逻辑
│   ├── stores/                 # Pinia 状态管理
│   │   ├── auth.ts             # 用户认证状态
│   │   ├── notification.ts     # 通知状态
│   │   └── contest.ts          # 当前竞赛状态（WebSocket 驱动）
│   ├── utils/                  # 工具函数
│   │   ├── constants.ts        # 常量定义（角色、状态、分类色标等）
│   │   ├── format.ts           # 格式化（时间、数字、文件大小）
│   │   └── errorMap.ts         # 错误码 → 用户提示映射
│   ├── views/                  # 页面视图
│   │   ├── auth/
│   │   │   ├── LoginView.vue
│   │   │   └── RegisterView.vue
│   │   ├── dashboard/
│   │   │   └── DashboardView.vue
│   │   ├── challenges/
│   │   │   ├── ChallengeList.vue
│   │   │   └── ChallengeDetail.vue
│   │   ├── contests/
│   │   │   ├── ContestList.vue
│   │   │   └── ContestDetail.vue  # 内含 Tab: 题目/公告/队伍/排行榜
│   │   ├── scoreboard/
│   │   │   └── ScoreboardView.vue
│   │   ├── instances/
│   │   │   └── InstanceList.vue
│   │   ├── profile/
│   │   │   ├── SkillProfile.vue
│   │   │   └── UserProfile.vue
│   │   ├── notifications/
│   │   │   └── NotificationList.vue
│   │   ├── teacher/
│   │   │   ├── TeacherDashboard.vue
│   │   │   ├── ClassManagement.vue
│   │   │   └── ReportExport.vue
│   │   └── admin/
│   │       ├── AdminDashboard.vue
│   │       ├── ChallengeManage.vue
│   │       ├── ContestManage.vue
│   │       ├── UserManage.vue
│   │       ├── ImageManage.vue
│   │       ├── CheatDetection.vue
│   │       └── AuditLog.vue
│   ├── App.vue
│   ├── main.ts
│   └── style.css               # Tailwind 入口 + @theme 设计令牌
├── index.html
├── vite.config.ts
├── package.json
└── .env.development            # 开发环境变量
```

---

## 3. 版本锁定策略（必须）

- **禁止使用 `latest`**：所有依赖必须锁定到明确的 semver 区间（例如 `^7.3.1` / `~7.3.1`），避免“今天能跑、明天爆炸”。
- **框架/构建链升级优先级**：Vite/Vue/Pinia 属于底座依赖，统一在一个迭代周期内升级与回归验证；业务依赖（图表/编辑器）按需升级。
- **文档与代码一致性**：本文档以仓库当前前端基线（Vite 7 / Pinia 3 / ECharts 6）为准，后续升级需同步更新 `ctf/frontend/package.json` 与本目录文档。

---

## 4. 安全基线（必须）

- **富文本/Markdown 安全**：题目描述、公告、通知等内容渲染前必须做 sanitize（白名单策略），禁止任意 HTML 注入（否则可直接窃取 Token/接管账号）。
- **Refresh Token 存储**：必须由后端写入 HttpOnly Cookie（前端不落盘）；禁止前端以任何形式持久化 refresh token（例如 localStorage / IndexedDB）。
- **错误提示**：Toast/弹窗只展示用户可理解的文案；不直接透传后端 `message`（避免泄露内部信息），必要时附带 `request_id` 便于定位。
- **WebSocket ticket**：短期一次性 ticket 如通过 URL query 传递，网关必须避免记录 querystring（防止在日志/代理链路中泄露）。

---

## 5. TypeScript 基线（必须）

- **默认 TypeScript**：除非被第三方库限制，`src/` 下所有业务代码使用 `.ts`；Vue SFC 使用 `<script setup lang="ts">`。
- **类型边界**：API 返回值、Store 状态、路由 meta（如 `roles/title/icon`）必须有明确类型；禁止在核心链路长期使用 `any`。
