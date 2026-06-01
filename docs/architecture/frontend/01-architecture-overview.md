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

- `code/frontend/src/pages/**`、`code/frontend/src/features/**/model`、`code/frontend/src/features/**/ui`、`code/frontend/src/widgets/**`、`code/frontend/src/shared/model/**`、`code/frontend/src/shared/lib/**`
  - 负责：路由入口统一落在 `pages/**`；页面级查询、导出、实时桥接和 query 同步下沉到 feature model / shared model / shared lib；`widgets/**` 负责跨 feature 页面区块组合，`features/**/ui` 负责单一能力 surface，`shared/model/common` 和 `shared/model/layout` 承接共享反馈、危险确认、复制反馈、工作区导航、面包屑细节与分页状态这类跨 feature 但不带业务 owner 的状态，`shared/model/theme` 承接全局主题与品牌 owner，`shared/model/reporting` 承接报告轮询这类共享 reporting workflow owner，`shared/model/navigation` 承接 route-aware transport 与 query/tab 同步 owner，`shared/model/realtime` 承接 WebSocket ticket、心跳、重连和 session 过期回退这类共享 realtime runtime owner，`shared/lib/*` 承接时间倒计时、请求取消、sanitize、键盘导航和 route target 契约等无业务语义的基础能力
  - 不负责：把 API 调用、路由状态和大段派生数据继续堆回单个 `.vue` 页面，或让 `features/*RoutePage.vue`、`widgets/*RoutePage.vue` 继续兼任页面层

- `code/frontend/src/stores/auth.ts`、`notification.ts`、`contest.ts`
  - 负责：登录快照、通知列表、竞赛共享状态这类跨页面共享数据
  - 不负责：单页筛选、局部表单和一次性临时状态

- `code/frontend/src/shared/ui/common/`、`code/frontend/src/shared/ui/common/modal-templates/`、`code/frontend/src/shared/ui/layout/`
  - 负责：共享 UI 原语、overlay 模板和应用总布局；Guardrail 见 `code/frontend/src/shared/ui/common/__tests__/ModalTemplates.test.ts`、`code/frontend/src/shared/ui/layout/__tests__/AppLayout.test.ts`
  - 不负责：承载具体业务 owner

## 1. 架构骨架

当前前端采用“薄 `pages/**` 路由入口 + feature model owner + 轻量 Pinia + 共享样式壳”的结构。

主链路：

1. `router` 决定路由命名空间、认证和标题
2. `pages/**` 作为路由页面入口
3. `features/**/model` 负责编排 API、query 同步、分页、导出和实时能力
4. `stores/**` 只承接跨页共享状态
5. `shared/ui/common/**` 和 `assets/styles/*.css` 负责统一交互骨架和视觉节奏

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
│   ├── pages/
│   │   └── **/
│   │       └── *RoutePage.vue
│   ├── stores/
│   │   ├── auth.ts
│   │   ├── notification.ts
│   │   └── contest.ts
│   ├── features/
│   │   └── */{model,ui}/
│   ├── shared/
│   │   ├── model/
│   │   │   ├── common/
│   │   │   ├── layout/
│   │   │   ├── navigation/
│   │   │   └── realtime/
│   │   └── ui/
│   │       ├── common/
│   │       └── layout/
│   ├── widgets/
│   │   └── */
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
- 旧 `/teacher/*` 不再作为前端运行时入口；登录 redirect 参数遇到旧教师端路径时，回退到角色默认首页

详情见：

- `02-routing.md`

### 3.2 状态 owner

- 全局共享状态只保留 `auth`、`notification`、`contest`
- 页面级状态默认进 `features/**/model`；跨 feature 的共享工作区状态进 `shared/model/**`，其中 route-aware transport 与 query/tab 同步进入 `shared/model/navigation`；无业务语义的基础能力进 `shared/lib/**`

详情见：

- `03-state-management.md`

### 3.3 请求与实时

- HTTP 统一走 `api/request.ts`，使用 session cookie、envelope 解包和 `ApiError`
- WebSocket 统一走 `shared/model/realtime/useWebSocket()`，ticket、心跳、重连和鉴权失败回退都在这一层

详情见：

- `04-api-layer.md`
- `05-websocket-composables.md`

### 3.4 共享原语与样式

- 共享组件集中在 `shared/ui/common/`
- overlay 模板集中在 `shared/ui/common/modal-templates/`
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

- 前端架构策略单点事实：`code/frontend/scripts/frontend-architecture-policy.json`
  - 分层、低层 forbidden imports、`pages/**` route entry 约束、增长守卫入口都以这份策略为准；以后如果要调整前端结构边界，先改这份策略，再同步对应测试或脚本。
- 前端分层：`code/frontend/src/__tests__/architectureBoundaries.test.ts`
- route page 边界：`code/frontend/src/__tests__/routePageArchitectureBoundary.test.ts`
- 后台导航命名空间：`code/frontend/src/config/__tests__/backofficeNavigation.test.ts`
- 共享弹窗模板：`code/frontend/src/shared/ui/common/__tests__/ModalTemplates.test.ts`
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
