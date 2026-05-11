# CTF 靶场平台 — 前端架构总览

> 状态：Current
> 事实源：`code/frontend/src/`、`code/frontend/vite.config.ts`、`code/frontend/package.json`
> 替代：无
> 对应后端架构：`docs/architecture/backend/01-system-architecture.md`
> 对应 API 设计：`docs/architecture/backend/04-api-design.md`

## 当前设计

- `code/frontend/src/router/`、`code/frontend/src/config/backofficeNavigation.ts`
  - 负责：注册学生端、`/academy/*`、`/platform/*` 路由，维护登录态守卫、默认首页映射和后台导航归属
  - 不负责：页面内部数据流和业务状态机

- `code/frontend/src/views/**`、`code/frontend/src/features/**/model`、`code/frontend/src/composables/use*.ts`
  - 负责：route view 只保留页面壳，页面级查询、导出、实时桥接和 query 同步下沉到 feature model / composable
  - 不负责：把 API 调用、路由状态和大段派生数据继续堆回单个 `.vue` 页面

- `code/frontend/src/stores/auth.ts`、`notification.ts`、`contest.ts`
  - 负责：登录快照、通知列表、竞赛共享状态这类跨页面共享数据
  - 不负责：单页筛选、局部表单和一次性临时状态

- `code/frontend/src/components/common/`、`code/frontend/src/components/common/modal-templates/`、`code/frontend/src/components/layout/`
  - 负责：共享 UI 原语、overlay 模板和应用总布局；Guardrail 见 `code/frontend/src/components/common/__tests__/ModalTemplates.test.ts`、`code/frontend/src/components/layout/__tests__/AppLayout.test.ts`
  - 不负责：承载具体业务 owner

## 1. 架构骨架

当前前端采用“薄 route view + feature model owner + 轻量 Pinia + 共享样式壳”的结构。

主链路：

1. `router` 决定路由命名空间、认证和标题
2. `views/**` 作为路由页面入口
3. `features/**/model` 负责编排 API、query 同步、分页、导出和实时能力
4. `stores/**` 只承接跨页共享状态
5. `components/common/**` 和 `assets/styles/*.css` 负责统一交互骨架和视觉节奏

## 2. 当前目录骨架

```text
code/frontend/
├── src/
│   ├── api/
│   │   ├── request.ts
│   │   ├── auth.ts
│   │   ├── challenge.ts
│   │   ├── contest.ts
│   │   ├── instance.ts
│   │   ├── notification.ts
│   │   ├── scoreboard.ts
│   │   ├── teacher/
│   │   └── admin/
│   ├── router/
│   │   ├── index.ts
│   │   ├── guards.ts
│   │   └── routes/
│   ├── stores/
│   │   ├── auth.ts
│   │   ├── notification.ts
│   │   └── contest.ts
│   ├── features/
│   │   └── */model/
│   ├── composables/
│   │   └── use*.ts
│   ├── components/
│   │   ├── common/
│   │   ├── common/modal-templates/
│   │   ├── layout/
│   │   ├── teacher/
│   │   ├── platform/
│   │   └── contests/ / scoreboard/ ...
│   ├── views/
│   │   ├── auth/
│   │   ├── dashboard/
│   │   ├── challenges/
│   │   ├── contests/
│   │   ├── notifications/
│   │   ├── profile/
│   │   ├── scoreboard/
│   │   ├── teacher/
│   │   └── platform/
│   ├── assets/styles/
│   ├── __tests__/
│   ├── main.ts
│   └── style.css
├── vite.config.ts
└── package.json
```

## 3. 关键边界

### 3.1 路由与命名空间

- 教师工作区正式 URL：`/academy/*`
- 平台工作区正式 URL：`/platform/*`
- 学生端当前是混合命名，不是统一 `/student/*` 前缀
- 旧 `/teacher/*` 当前只保留 redirect 兼容，不再作为活动入口

详情见：

- `02-routing.md`

### 3.2 状态 owner

- 全局共享状态只保留 `auth`、`notification`、`contest`
- 页面级状态默认进 `features/**/model` 或 `composables/use*.ts`

详情见：

- `03-state-management.md`

### 3.3 请求与实时

- HTTP 统一走 `api/request.ts`，使用 session cookie、envelope 解包和 `ApiError`
- WebSocket 统一走 `useWebSocket()`，ticket、心跳、重连和鉴权失败回退都在这一层

详情见：

- `04-api-layer.md`
- `05-websocket-composables.md`

### 3.4 共享原语与样式

- 共享组件集中在 `components/common/`
- overlay 模板集中在 `components/common/modal-templates/`
- 全局节奏和主题由 `theme.css`、`style.css`、`workspace-shell.css`、`page-tabs.css`、`teacher-surface.css` 等样式文件共同维护

详情见：

- `06-components.md`
- `09-spacing-system.md`

## 4. 当前技术基线

| 类别 | 当前基线 |
| --- | --- |
| 框架 | Vue 3 |
| 构建 | Vite 7 |
| 路由 | Vue Router 4 |
| 状态管理 | Pinia 3 |
| 样式 | Tailwind CSS 4 + CSS 变量 |
| HTTP | Axios |
| 图表 | ECharts + vue-echarts |
| 图标 | lucide-vue-next |
| Markdown / sanitize | marked + DOMPurify |
| 测试 | Vitest + Vue Test Utils |

说明：

- 当前前端没有活动的外部 UI 组件库事实源；共享原语和样式壳都在仓库内维护。
- 运行入口、代理和分包策略见 `08-build-deploy.md`。

## 5. Guardrail

- 前端分层：`code/frontend/src/__tests__/architectureBoundaries.test.ts`
- route view 边界：`code/frontend/src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- 后台导航命名空间：`code/frontend/src/config/__tests__/backofficeNavigation.test.ts`
- 共享弹窗模板：`code/frontend/src/components/common/__tests__/ModalTemplates.test.ts`
- 主题尾部硬编码检查：`cd code/frontend && npm run check:theme-tail`

## 6. 读取顺序

1. `02-routing.md`
2. `03-state-management.md`
3. `04-api-layer.md`
4. `05-websocket-composables.md`
5. `06-components.md`
6. `07-pages-dataflow.md`
7. `08-build-deploy.md`
8. `09-spacing-system.md`
