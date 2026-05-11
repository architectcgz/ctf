# 前端构建与运行入口

> 状态：Current
> 事实源：`code/frontend/vite.config.ts`、`code/frontend/package.json`、`code/frontend/src/main.ts`
> 替代：无

## 定位

本文档只说明前端构建链、开发代理、运行入口和当前可执行的校验命令。

- 覆盖：Vite 配置、环境变量、插件、分包策略、启动脚本、全局样式入口。
- 不覆盖：反向代理和生产部署的具体运维脚本；那类内容应进入 `docs/operations/`。

## 当前设计

- `code/frontend/vite.config.ts`
  - 负责：Vite 插件装配、`@` alias、开发服务器、`/api` 与 `/ws` proxy、ECharts 与生态依赖的 manual chunk 规则
  - 不负责：运行时登录态恢复、页面错误处理或主题样式加载

- `code/frontend/package.json`
  - 负责：声明前端依赖版本和 `dev / build / preview / typecheck / lint / format / test / check:theme-tail` 脚本
  - 不负责：替代仓库级 workflow 检查

- `code/frontend/src/main.ts`
  - 负责：创建 Vue 应用、挂载 Pinia 和 Router、接入全局错误处理、提前恢复 session、导入全局样式
  - 不负责：每个页面自己的数据预取或业务重试策略

## 1. 构建入口

`vite.config.ts` 当前加载这些插件：

| 插件 | 当前作用 |
| --- | --- |
| `@vitejs/plugin-vue` | Vue SFC 编译 |
| `@tailwindcss/vite` | Tailwind 4 集成 |
| `unplugin-auto-import` | 自动导入 `vue`、`vue-router`、`pinia`，生成 `src/auto-imports.d.ts` |
| `unplugin-vue-components` | 本地组件类型声明，生成 `src/components.d.ts` |

固定配置：

- alias：`@ -> code/frontend/src`
- dev server：`0.0.0.0:5173`

## 2. 环境变量与开发代理

当前配置里真正生效的环境变量：

| 变量 | 当前用途 | 默认值 |
| --- | --- | --- |
| `VITE_DEV_PROXY_TARGET` | 开发时 `/api` 和 `/ws` 的代理目标 | `http://127.0.0.1:8080` |
| `VITE_API_BASE_URL` | Axios base URL | `/api/v1` |
| `VITE_API_TIMEOUT` | Axios 请求超时 | `15000` |
| `VITE_WS_BASE_URL` | WebSocket base path | `/ws` |

dev proxy 规则：

- `/api` -> `proxyTarget`
- `/ws` -> `wsProxyTarget`
- `wsProxyTarget` 由 `proxyTarget.replace(/^http/i, 'ws')` 推导

这意味着当前本地开发默认是：

- 浏览器访问 Vite
- Vite 代转 API 和 WebSocket
- 前端仍按同源路径 `/api/v1`、`/ws` 发请求

## 3. 当前分包策略

`vite.config.ts` 当前没有用静态 chunk 表，而是按依赖路径做 `manualChunks(id)`：

| Chunk | 当前内容 |
| --- | --- |
| `vue-core` | Vue 本体 |
| `vue-ecosystem` | `vue-router`、`pinia`、`@vueuse/core` |
| `network` | `axios` |
| `vue-echarts` | `vue-echarts` |
| `echarts-charts` | `echarts/charts` |
| `echarts-components` | `echarts/components`、`echarts/features` |
| `echarts-renderers` | `echarts/renderers` |
| `echarts-runtime` | `echarts` 核心运行时 |
| `zrender` | `zrender` |

当前目标不是“每个页面一个手写 chunk”，而是先把重依赖拆清，再让路由懒加载处理页面级切分。

## 4. 运行时启动链

`src/main.ts` 当前启动顺序：

1. `createApp(App)`
2. `createPinia()`
3. `app.use(pinia)`
4. `app.use(router)`
5. 注册 `app.config.errorHandler`
6. 提前执行 `useAuthStore(pinia).restore()`
7. `app.mount('#app')`

全局错误处理规则：

- `ApiError` 直接放过，由请求层和页面自己处理
- 非 `ApiError` 的 Vue 运行时异常跳 `/500`

说明：

- 登录态恢复不会阻塞应用挂载，但 Router guard 会等待同一个 `restorePromise`
- 页面级数据预取仍放在各自 feature model，不挪到 `main.ts`

## 5. 全局样式入口

`main.ts` 当前按顺序加载：

1. `style.css`
2. `assets/styles/theme.css`
3. `assets/styles/surface-shell-background.css`
4. `assets/styles/teacher-surface.css`
5. `assets/styles/page-tabs.css`
6. `assets/styles/workspace-shell.css`
7. `assets/styles/journal-eyebrows.css`
8. `assets/styles/journal-notes.css`
9. `assets/styles/journal-soft-surfaces.css`
10. `assets/styles/journal-admin-shell.css`
11. `assets/styles/journal-user-shell.css`
12. `assets/styles/journal-user-directory.css`

这说明当前前端运行时依赖的是“样式层组合”，而不是运行时 UI 组件库主题注入。

## 6. 当前脚本与最小校验

`package.json` 当前脚本：

| 命令 | 当前用途 |
| --- | --- |
| `npm run dev` | 本地开发 |
| `npm run build` | 生产构建 |
| `npm run preview` | 预览构建产物 |
| `npm run typecheck` | `vue-tsc --noEmit` |
| `npm run check:theme-tail` | 检查硬编码主题尾部 token |
| `npm run lint` | ESLint |
| `npm run format` | Prettier |
| `npm run test` / `npm run test:run` | Vitest |

当前文档任务关联度最高的运行校验：

- `npm run check:theme-tail`
- `npm run typecheck`
- `npm run test:run`

## 7. 部署边界

当前前端代码只假设：

- 静态资源由前端服务器或反向代理提供
- `/api` 指向后端 HTTP
- `/ws` 指向后端 WebSocket

具体 Nginx、容器或 systemd 配置不在当前事实源里，避免把未维护的运维样例继续写成活动文档。
