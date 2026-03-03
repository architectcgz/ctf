# CTF 前端架构 Review（frontend-arch 第 1 轮）：架构文档 vs 实际实现一致性审查

| 字段 | 说明 |
|------|------|
| 变更主题 | frontend-arch |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | 架构文档 `docs/architecture/frontend/01~08`（8 篇）vs `frontend/src/` 全部实现文件（4 .js + 20 .vue） |
| 变更概述 | 对比前端架构设计文档与当前代码实现的差距，识别偏离、缺失和风险 |
| 审查基准 | `docs/architecture/frontend/*.md` |
| 审查日期 | 2026-03-02 |

---

## 审查总览

架构文档共 8 篇，覆盖技术栈、路由、状态管理、API 层、WebSocket/Composables、组件体系、页面数据流、构建部署。当前代码处于早期脚手架阶段，已实现基础路由和布局，但与架构文档存在大量结构性差距。

| 级别 | 数量 |
|------|------|
| 高 | 5 |
| 中 | 6 |
| 低 | 3 |

---

## 高严重性问题

### H-1：TypeScript 完全缺失

- 文档要求：01-overview §5 明确要求 "TypeScript 基线"，所有文件使用 `.ts`/`.vue + lang="ts"`，08-build §3.2 要求 `<script setup lang="ts">`，§3.4 要求 CI 运行 `vue-tsc --noEmit`
- 实际情况：全部 4 个脚本文件均为 `.js`，所有 `.vue` 文件的 `<script setup>` 均无 `lang="ts"`，`package.json` 未安装 `typescript` 和 `vue-tsc`
- 影响：Props/Emits 无类型约束，API 响应无类型安全，重构风险高，06-components 中定义的所有 Props 类型表形同虚设
- 建议：尽早引入 TypeScript。当前文件少（24 个），迁移成本低；拖到 Phase 2 实现阶段再迁移代价会指数增长

### H-2：路由守卫与 meta 完全缺失

- 文档要求：02-routing 定义了完整的守卫流程（白名单 → token → profile → roles），每条路由需要 `meta: { requiresAuth, roles, title }` 字段
- 实际情况：`router/index.js` 无任何 `beforeEach` 守卫，无 `meta` 字段，无角色校验
- 影响：任何人可直接访问 `/admin/dashboard`，无认证保护；页面标题不会动态更新
- 建议：这是安全基线，必须在任何业务页面上线前完成

### H-3：API 层完全缺失

- 文档要求：04-api-layer 定义了 Axios 实例配置（baseURL、拦截器、token 注入、401 刷新队列）、8 个 API 模块（auth/challenge/contest/scoreboard/instance/user/admin/notification）、错误码映射
- 实际情况：`package.json` 未安装 `axios`，`src/` 下无 `api/` 目录，无任何 HTTP 请求封装
- 影响：所有页面视图目前只能使用 mock 数据，无法与后端联调；Token 刷新机制无载体
- 建议：API 层是所有业务页面的前置依赖，应作为 Phase 2 实现的第一个任务

### H-4：Auth Store 缺少关键功能

- 文档要求：03-state-management 定义 authStore 需包含 `refreshToken`、`updateTokens()`、`restore()`、`isStudent` computed
- 实际情况：`stores/auth.js` 仅有 `token`（单 token）、`setAuth()`、`logout()`，无 refresh token 存储、无 token 刷新方法、无会话恢复、无 `isStudent` 角色判断
- 影响：页面刷新后用户状态丢失（有 token 但无 user profile 恢复逻辑）；access token 过期后无法静默续期，直接登出
- 建议：配合 H-3 API 层一起补全，实现 `restore()` → 拉取 profile + 按需 refresh 的完整流程

### H-5：路由表严重不完整

- 文档要求：02-routing 定义了 4 组路由（公开、主区、教师、管理员），管理员下有 7 个子路由（dashboard/challenges/contests/users/images/cheat/audit），教师下有 3 个子路由，另有 404/403 错误页
- 实际情况：管理员仅 1 个路由（dashboard），教师路由完全缺失，无 404/403 页面，无错误路由兜底
- 影响：Phase 2 新增的管理员页面（拓扑编排、模板管理、标签管理、Writeup 管理、AWD 监控、作弊检测、数据导出）无路由挂载点
- 建议：补全路由表骨架（可先用空组件占位），确保路由结构与架构文档一致

---

## 中等严重性问题

### M-1：Element Plus 未安装

- 文档要求：01-overview 将 Element Plus 列为核心依赖，06-components 明确 "El* 组件优先直用"，08-build §1.2.1 规划了按需引入方案（unplugin-auto-import + unplugin-vue-components）
- 实际情况：`package.json` 中无 `element-plus`，无 `unplugin-auto-import`，无 `unplugin-vue-components`
- 影响：管理端大量表格/表单/弹窗场景缺少基础组件库支撑，自行实现成本高且不一致
- 建议：安装 Element Plus 并配置按需引入，同时在 `style.css` 中映射 `--el-color-primary` 等变量到现有 `@theme` token

### M-2：Vite 配置缺少 proxy 和分包策略

- 文档要求：08-build §1.2 定义了 `/api` 和 `/ws` 的 proxy 配置，§1.3 定义了 manualChunks（vendor + echarts 分包）
- 实际情况：`vite.config.js` 仅有 plugins 和 alias，无 server.proxy，无 build.rollupOptions
- 影响：开发环境无法代理后端 API（会遇到 CORS），生产构建无分包优化（echarts ~120KB gzip 会打入主包）
- 建议：补充 proxy 配置（开发联调前置条件）和 manualChunks

### M-3：Composables 完全缺失

- 文档要求：05-websocket-composables 定义了 5 个核心 composable（useWebSocket、useAuth、useToast、usePagination、useCountdown、useClipboard）
- 实际情况：`src/` 下无 `composables/` 目录
- 影响：各页面会重复实现分页、倒计时、剪贴板等逻辑；WebSocket 连接无统一管理
- 建议：至少先实现 useAuth（登录/登出/恢复）和 usePagination（列表页通用），其余按需补充

### M-4：Sidebar 硬编码且无角色区分

- 文档要求：02-routing 按角色分组路由（学员/教师/管理员），Sidebar 应根据用户角色动态渲染菜单项
- 实际情况：`Sidebar.vue` 硬编码所有导航项，Admin 区块始终可见（无 `v-if="isAdmin"` 判断），教师菜单完全缺失
- 影响：普通学员能看到管理入口（虽然路由守卫应拦截，但守卫也缺失——见 H-2）
- 建议：从 authStore 读取角色，条件渲染不同菜单组；配合 H-2 路由守卫形成完整的前端权限控制

### M-5：通知和竞赛 Store 缺失

- 文档要求：03-state-management 定义了 3 个全局 Store（auth、notification、contest），notification 需要 `unreadCount` + WebSocket 推送，contest 需要 `currentContest` + 排行榜数据
- 实际情况：仅有 `stores/auth.js`，无 notification store，无 contest store
- 影响：通知红点、竞赛实时数据无全局状态载体；Phase 2 AWD 竞赛页面强依赖 contest store
- 建议：创建 store 骨架文件，定义好接口类型，实现可在对接 API 时逐步填充

### M-6：组件体系与文档定义差距大

- 文档要求：06-components 定义了 10 个 App* 基础组件（AppButton、AppCard、AppInput、AppTable、AppPagination、AppTag、AppDialog、AppDrawer、AppToast、AppEmpty、AppSkeleton）
- 实际情况：`components/common/` 仅有 3 个组件（PageHeader、SectionCard、SkillRadar），均非文档定义的标准组件；缺少 `components/charts/` 目录
- 影响：各页面视图无法复用统一组件，样式一致性难以保证
- 建议：优先实现高频组件（AppEmpty、AppSkeleton、AppTag），其余在引入 Element Plus 后按需薄封装

---

## 低严重性问题

### L-1：文件命名不符合规范

- 文档要求：08-build §3.1 规定 Store 文件使用 `.ts` 后缀，Composable 使用 `use` 前缀 + `.ts`
- 实际情况：所有文件使用 `.js` 后缀（与 H-1 TypeScript 缺失关联）
- 建议：迁移 TypeScript 时统一重命名

### L-2：AppLayout 背景渐变与 MASTER.md 不一致

- 文档要求：MASTER.md 定义背景为纯色 `bg-base (#0f1117)`，layout.md 未提及径向渐变装饰
- 实际情况：`AppLayout.vue` 使用了 `radial-gradient` 叠加层（cyan + violet 渐变），属于额外视觉装饰
- 影响：轻微，视觉效果尚可，但未在设计文档中记录
- 建议：如保留此效果，在 layout.md 中补充说明；或移除以严格遵循 MASTER.md

### L-3：HelloWorld.vue 残留

- 实际情况：`components/HelloWorld.vue` 为 Vite 脚手架默认文件，未被任何组件引用
- 建议：删除，避免文件噪音

---

## 架构文档自身问题

在审查过程中也发现架构文档本身的几个问题：

### D-1：§3.4 编号跳跃（中）

- `08-build-deploy.md` 中 §3.2 之后直接跳到 §3.4（缺少 §3.3），实际 §3.3 Git 分支与提交出现在 §3.4 之后
- 建议：修正编号顺序

### D-2：Phase 2 页面未纳入 07-pages-dataflow（中）

- `07-pages-dataflow.md` 仅覆盖 Phase 1 页面，Phase 2 新增的 11 个页面（拓扑编排、AWD 竞赛、AWD 监控等）未列入页面→组件→API 映射表
- 建议：Phase 2 实现前补充对应映射，确保数据流设计先行

### D-3：05-websocket 缺少 AWD 实时端点（低）

- Phase 2 AWD 竞赛需要实时推送（轮次切换、服务状态变更、攻击事件），但 05-websocket §1.3 使用场景表仅列出 scoreboard/notifications/announcements 三个端点
- 建议：补充 `/ws/awd/:id` 端点定义及消息类型

---

## 实现优先级建议

基于依赖关系，建议按以下顺序补齐基础设施：

```
1. TypeScript 迁移（H-1）— 越早越好，当前文件少
   │
2. API 层 + Axios（H-3）— 所有业务页面的前置依赖
   │
3. Auth Store 补全 + 路由守卫（H-4 + H-2）— 安全基线
   │
4. Element Plus 安装 + 主题映射（M-1）— 管理端组件基础
   │
5. 路由表补全 + Sidebar 角色控制（H-5 + M-4）— Phase 2 页面挂载点
   │
6. 核心 Composables（M-3）— useAuth → usePagination → useCountdown
   │
7. Store 骨架 + 组件体系（M-5 + M-6）— 按页面开发节奏逐步补充
   │
8. Vite proxy + 分包（M-2）— 联调前完成
```

---

## 总结

当前前端代码处于 MVP 脚手架阶段，架构文档设计完整度较高，但实现落地率约 15~20%。5 个高严重性问题（TypeScript、路由守卫、API 层、Auth Store、路由表）构成 Phase 2 开发的阻塞项，需在业务页面实现前优先解决。架构文档本身质量良好，仅需小幅补充 Phase 2 相关内容。

