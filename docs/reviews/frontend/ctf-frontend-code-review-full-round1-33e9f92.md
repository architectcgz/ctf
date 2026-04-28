# CTF 前端代码 Review（full 第 1 轮）：全量前端代码审查

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | full |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | 前端全部源码（src/ 目录下所有 .vue/.ts/.js 文件） |
| 变更概述 | 对 CTF 平台前端代码进行全面架构一致性与代码质量审查 |
| 审查基准 | docs/architecture/frontend/01-architecture-overview.md<br>docs/architecture/frontend/02-routing.md<br>docs/architecture/frontend/03-state-management.md<br>docs/architecture/frontend/04-api-layer.md |
| 审查日期 | 2026-03-03 |
| 上轮问题数 | - |

## 问题清单

### 🔴 高优先级

#### [H1] WebSocket ticket 通过 URL query 传递存在安全风险
- **文件**：`src/composables/useWebSocket.ts:40`
- **问题描述**：WebSocket 连接时将 ticket 作为 query 参数拼接到 URL 中（`?ticket=${encodeURIComponent(ticket)}`），这会导致 ticket 暴露在浏览器历史记录、代理日志、网关日志中
- **影响范围/风险**：ticket 泄露可能导致未授权的 WebSocket 连接，攻击者可以劫持实时通信通道
- **修正建议**：
  - 方案 1（推荐）：WebSocket 握手时通过 `Sec-WebSocket-Protocol` 子协议头传递 ticket
  - 方案 2：连接建立后立即发送认证消息，而非通过 URL 传递
  ```typescript
  // 方案 1 示例
  socket = new WebSocket(resolveWsUrl(path), [`ticket.${ticket}`])

  // 方案 2 示例
  socket = new WebSocket(resolveWsUrl(path))
  socket.addEventListener('open', () => {
    socket.send(JSON.stringify({ type: 'auth', payload: { ticket } }))
  })
  ```

#### [H2] WebSocket 重连次数硬编码且无配置化
- **文件**：`src/composables/useWebSocket.ts:79`
- **问题描述**：最大重连次数 `20` 和心跳间隔 `30_000` 直接硬编码在代码中，无法根据环境或业务需求调整
- **影响范围/风险**：生产环境可能需要不同的重连策略，硬编码导致无法灵活配置
- **修正建议**：提取为配置常量或通过 composable 参数传入
  ```typescript
  // utils/constants.ts
  export const WS_MAX_RECONNECT_ATTEMPTS = 20
  export const WS_HEARTBEAT_INTERVAL_MS = 30_000

  // 或通过参数传入
  export function useWebSocket(
    endpoint: string,
    handlers: WebSocketHandlers,
    options?: { maxReconnectAttempts?: number; heartbeatIntervalMs?: number }
  )
  ```

#### [H3] Toast 持续时间硬编码
- **文件**：`src/composables/useToast.ts:37-40`
- **问题描述**：不同类型 Toast 的持续时间（3000/4000/5000ms）直接硬编码，无法统一调整或配置
- **影响范围/风险**：用户体验调优时需要修改代码，且分散在多处
- **修正建议**：提取为常量配置
  ```typescript
  // utils/constants.ts
  export const TOAST_DURATION = {
    SUCCESS: 3000,
    INFO: 3000,
    WARNING: 4000,
    ERROR: 5000,
  } as const

  // useToast.ts
  const fallbackToast: ToastApi = {
    success: (message) => add('success', message, TOAST_DURATION.SUCCESS),
    // ...
  }
  ```

#### [H4] 分页默认大小硬编码
- **文件**：`src/composables/usePagination.ts:20`
- **问题描述**：默认分页大小 `20` 硬编码在 composable 中，无法全局统一配置
- **影响范围/风险**：不同页面可能需要不同的默认分页大小，且无法统一调整
- **修正建议**：提取为常量并支持参数覆盖
  ```typescript
  // utils/constants.ts
  export const DEFAULT_PAGE_SIZE = 20

  // usePagination.ts
  export function usePagination<T>(
    fetchFn: ...,
    options?: { initialPageSize?: number }
  ): PaginationState<T> {
    const pageSize = ref(options?.initialPageSize ?? DEFAULT_PAGE_SIZE)
    // ...
  }
  ```

### 🟡 中优先级

#### [M1] 路由守卫中 redirect 参数校验不够严格
- **文件**：`src/router/guards.ts:12-16` 和 `src/views/auth/LoginView.vue:39-44`
- **问题描述**：`sanitizeRedirectPath` 函数只检查了 `//` 开头的情况，但未检查其他潜在的开放重定向风险（如 `///`、`/\`、`/http://`）
- **影响范围/风险**：可能存在开放重定向漏洞，攻击者可以构造恶意 URL 进行钓鱼攻击
- **修正建议**：增强校验逻辑，使用白名单或更严格的正则
  ```typescript
  function sanitizeRedirectPath(input: unknown): string {
    if (typeof input !== 'string') return '/dashboard'
    // 移除所有前导斜杠，只保留一个
    const normalized = '/' + input.replace(/^\/+/, '')
    // 检查是否包含协议或双斜杠
    if (/^\/\/|^\/\\|:\/\//.test(normalized)) return '/dashboard'
    return normalized
  }
  ```

#### [M2] API 请求超时时间硬编码
- **文件**：`src/api/request.ts:35` 和 `src/api/request.ts:65`
- **问题描述**：Axios 实例的 timeout 设置为 `15000` 硬编码，无法根据不同环境或接口类型调整
- **影响范围/风险**：某些接口（如报告导出、文件上传）可能需要更长的超时时间
- **修正建议**：通过环境变量配置，并支持单个请求覆盖
  ```typescript
  // .env
  VITE_API_TIMEOUT=15000

  // request.ts
  const DEFAULT_TIMEOUT = Number(import.meta.env.VITE_API_TIMEOUT) || 15000

  const instance = axios.create({
    baseURL,
    timeout: DEFAULT_TIMEOUT,
    // ...
  })

  // 特殊接口可覆盖
  export async function exportReport(data: unknown): Promise<ReportExportData> {
    return request<ReportExportData>({
      method: 'POST',
      url: '/reports/personal',
      data,
      timeout: 60000 // 长时间操作
    })
  }
  ```

#### [M3] 错误码映射不完整
- **文件**：`src/utils/errorMap.ts:5-14`
- **问题描述**：ERROR_MAP 只覆盖了 8 个错误码，但架构文档中提到需要覆盖更多业务错误码，且缺少通用错误码（如 10001 参数校验失败、10002 资源不存在等）
- **影响范围/风险**：未映射的错误码会显示通用失败文案，用户体验不佳
- **修正建议**：补充常见错误码映射，并在架构文档中明确错误码规范
  ```typescript
  const ERROR_MAP: Record<number, string> = {
    // 通用错误码 (10xxx)
    10001: '参数校验失败，请检查输入',
    10002: '资源不存在',
    10003: '操作被拒绝',

    // 认证错误码 (11xxx)
    11001: '用户名或密码错误',
    11002: '登录状态已过期，请重新登录',
    11003: '账号已被锁定，请联系管理员',
    11010: '登录失败次数过多，请稍后再试',

    // 实例错误码 (13xxx)
    13001: '靶场不存在或已下线',
    13002: '实例数量已达上限，请先销毁已有实例',
    13003: 'Flag 错误，请检查后重试',
    13004: '提交过于频繁，请稍后再试',

    // 竞赛错误码 (14xxx)
    14001: '竞赛未开始或已结束',
    14002: '未报名该竞赛',
    14003: '队伍人数已满',
    14008: '邀请码无效',
  }
  ```

#### [M4] 环境变量缺少类型定义
- **文件**：`src/api/request.ts:31`、`src/composables/useWebSocket.ts:19`
- **问题描述**：`import.meta.env.VITE_API_BASE_URL` 和 `VITE_WS_BASE_URL` 没有 TypeScript 类型定义，可能导致类型错误
- **影响范围/风险**：开发时缺少类型提示，容易出错
- **修正建议**：在 `env.d.ts` 中补充类型定义
  ```typescript
  // src/env.d.ts
  /// <reference types="vite/client" />

  interface ImportMetaEnv {
    readonly VITE_API_BASE_URL?: string
    readonly VITE_WS_BASE_URL?: string
    readonly VITE_API_TIMEOUT?: string
  }

  interface ImportMeta {
    readonly env: ImportMetaEnv
  }
  ```

#### [M5] 页面组件大量占位未实现
- **文件**：
  - `src/views/challenges/ChallengeDetail.vue`
  - `src/views/challenges/ChallengeList.vue`
  - `src/views/contests/ContestDetail.vue`
  - `src/views/dashboard/DashboardView.vue`
  - `src/views/admin/ChallengeManage.vue`
  - `src/views/admin/UserManage.vue`
  - 等多个页面组件
- **问题描述**：大量页面组件只有占位文本，未实现实际功能，导致前端无法正常使用
- **影响范围/风险**：核心功能缺失，无法进行完整的前后端联调和测试
- **修正建议**：按优先级实现核心页面：
  1. 高优先级：ChallengeList、ChallengeDetail（Flag 提交）、InstanceList
  2. 中优先级：ContestList、ContestDetail、Scoreboard
  3. 低优先级：管理后台页面

#### [M6] 缺少全局加载状态管理
- **文件**：架构层面问题
- **问题描述**：当前没有全局加载状态（如页面切换时的顶部进度条），用户体验不佳
- **影响范围/风险**：页面切换时用户无法感知加载状态，可能误以为系统卡死
- **修正建议**：使用 NProgress 或自定义全局加载组件
  ```typescript
  // router/index.ts
  import NProgress from 'nprogress'
  import 'nprogress/nprogress.css'

  router.beforeEach(() => {
    NProgress.start()
  })

  router.afterEach(() => {
    NProgress.done()
  })
  ```

### 🟢 低优先级

#### [L1] 缺少 Markdown 渲染安全处理
- **文件**：架构层面问题（未找到 Markdown 渲染实际使用）
- **问题描述**：架构文档中提到使用 md-editor-v3 渲染题目描述，但未在代码中找到实际使用，且未见 DOMPurify 等 sanitize 库
- **影响范围/风险**：如果后续实现 Markdown 渲染时未做 sanitize，可能导致 XSS 攻击
- **修正建议**：在实现 Markdown 渲染时，必须使用 DOMPurify 进行内容清洗
  ```typescript
  import DOMPurify from 'dompurify'

  // 渲染前清洗
  const cleanHtml = DOMPurify.sanitize(markdownHtml, {
    ALLOWED_TAGS: ['p', 'br', 'strong', 'em', 'code', 'pre', 'ul', 'ol', 'li', 'a', 'img'],
    ALLOWED_ATTR: ['href', 'src', 'alt', 'title', 'class'],
  })
  ```

#### [L2] 缺少请求取消机制
- **文件**：`src/api/request.ts`
- **问题描述**：当用户快速切换页面或重复点击时，之前的请求没有被取消，可能导致资源浪费和状态混乱
- **影响范围/风险**：用户体验问题，可能导致页面显示过期数据
- **修正建议**：使用 Axios 的 AbortController 实现请求取消
  ```typescript
  // composables/useRequest.ts
  export function useRequest<T>(requestFn: () => Promise<T>) {
    const loading = ref(false)
    const data = ref<T>()
    const error = ref<Error>()
    let controller: AbortController | null = null

    async function execute() {
      controller?.abort()
      controller = new AbortController()
      loading.value = true
      try {
        data.value = await requestFn()
      } catch (e) {
        if (e.name !== 'AbortError') error.value = e
      } finally {
        loading.value = false
      }
    }

    onUnmounted(() => controller?.abort())
    return { loading, data, error, execute }
  }
  ```

#### [L3] 缺少错误边界处理
- **文件**：`src/App.vue`
- **问题描述**：Vue 应用缺少全局错误边界，组件渲染错误会导致整个应用崩溃
- **影响范围/风险**：局部组件错误可能导致整个页面白屏
- **修正建议**：添加全局错误处理
  ```typescript
  // main.ts
  app.config.errorHandler = (err, instance, info) => {
    console.error('Vue error:', err, info)
    // 可选：上报到监控系统
  }

  // 或使用 ErrorBoundary 组件
  ```

#### [L4] 缺少性能监控埋点
- **文件**：架构层面问题
- **问题描述**：缺少页面加载时间、API 请求耗时等性能监控埋点
- **影响范围/风险**：无法及时发现性能问题
- **修正建议**：集成性能监控（如 Web Vitals）
  ```typescript
  // utils/performance.ts
  import { onCLS, onFID, onLCP } from 'web-vitals'

  export function initPerformanceMonitoring() {
    onCLS(console.log)
    onFID(console.log)
    onLCP(console.log)
  }
  ```

#### [L5] 缺少国际化支持预留
- **文件**：架构层面问题
- **问题描述**：当前所有文案硬编码为中文，未预留国际化扩展能力
- **影响范围/风险**：如果未来需要支持多语言，改造成本高
- **修正建议**：如有国际化需求，建议使用 vue-i18n 并提取文案到语言包

#### [L6] 缺少单元测试
- **文件**：整体项目
- **问题描述**：项目中未找到任何单元测试文件，关键工具函数和 composables 缺少测试覆盖
- **影响范围/风险**：代码重构时容易引入 bug，且难以保证质量
- **修正建议**：至少为以下模块添加单元测试：
  - `utils/errorMap.ts`
  - `composables/usePagination.ts`
  - `router/guards.ts` 中的 `sanitizeRedirectPath`
  - `stores/auth.ts`

#### [L7] 依赖版本锁定不够严格
- **文件**：`package.json`
- **问题描述**：所有依赖使用 `^` 前缀，允许自动升级次版本，可能导致"今天能跑、明天爆炸"
- **影响范围/风险**：CI/CD 环境中可能因依赖升级导致构建失败
- **修正建议**：
  - 方案 1：使用 `~` 前缀只允许补丁版本升级
  - 方案 2：使用 `package-lock.json` 或 `pnpm-lock.yaml` 锁定精确版本
  - 方案 3：定期手动升级依赖并回归测试

#### [L8] 缺少代码格式化和 Lint 配置
- **文件**：项目根目录
- **问题描述**：未找到 `.eslintrc`、`.prettierrc` 等代码规范配置文件
- **影响范围/风险**：团队协作时代码风格不一致，容易产生无意义的 diff
- **修正建议**：添加 ESLint + Prettier 配置
  ```json
  // .eslintrc.json
  {
    "extends": [
      "plugin:vue/vue3-recommended",
      "@vue/typescript/recommended",
      "prettier"
    ]
  }

  // .prettierrc.json
  {
    "semi": false,
    "singleQuote": true,
    "printWidth": 120
  }
  ```

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 4 |
| 🟡 中 | 6 |
| 🟢 低 | 8 |
| 合计 | 18 |

## 总体评价

### 架构一致性
前端代码整体遵循了架构文档的设计规范，目录结构清晰，分层合理。API 层、状态管理、路由守卫的实现与架构文档基本一致。

### 代码质量
核心基础设施（Axios 拦截器、Token 刷新、路由守卫、WebSocket 封装）实现质量较高，逻辑清晰，TypeScript 类型定义完整。但存在大量硬编码问题，缺少配置化和可维护性考虑。

### 安全性
Token 存储策略正确（Access Token 存 localStorage，Refresh Token 由后端 HttpOnly Cookie 管理），未发现 XSS 风险（未使用 v-html/innerHTML）。但 WebSocket ticket 传递方式存在安全隐患，redirect 参数校验不够严格。

### 完成度
项目处于早期阶段，大量页面组件只有占位未实现，核心业务功能（靶场列表、Flag 提交、竞赛详情等）缺失，无法进行完整的功能测试。

### 主要改进方向
1. **消除硬编码**：将所有魔法数字、超时时间、重连次数等提取为配置常量
2. **完善核心功能**：优先实现靶场和竞赛相关的核心页面
3. **增强安全性**：修复 WebSocket ticket 传递方式，加强 redirect 参数校验
4. **提升可维护性**：添加单元测试、代码规范工具、错误边界处理
