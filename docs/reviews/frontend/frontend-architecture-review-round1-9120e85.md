# CTF 前端架构 Review（第 1 轮）：架构文档与代码实现一致性审查

| 字段 | 内容 |
|------|------|
| 变更主题 | frontend-architecture |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | 前端架构文档 + 代码实现全量审查 |
| 变更概述 | 审查前端架构设计文档与当前代码实现的一致性，识别偏离、遗漏和潜在风险 |
| 审查基准 | `docs/architecture/frontend/*.md` (8 个文档) |
| 审查日期 | 2026-03-05 |
| Commit Hash | 9120e85 |

---

## 审查总结

**整体评价**：前端架构设计文档完整且规范，代码实现基本遵循架构设计，但存在以下关键问题需要修复：

- **高优先级问题**：3 项（安全风险、架构偏离）
- **中优先级问题**：5 项（功能缺失、不一致）
- **低优先级问题**：4 项（优化建议）

---

## 高优先级问题（必须修复）

### 🔴 H1. WebSocket 认证方式与架构文档不一致

**问题描述**：
- **架构文档要求**（`05-websocket-composables.md`）：ticket 应通过 URL query 参数传递
  ```
  ws://host/ws/{endpoint}?ticket={ticket}
  ```
- **实际实现**（`useWebSocket.ts:53`）：ticket 通过 WebSocket 消息体发送
  ```typescript
  socket?.send(JSON.stringify({ type: 'auth', payload: { ticket }, timestamp: new Date().toISOString() }))
  ```

**影响**：
- 架构文档与代码实现不一致，导致后端开发者按文档实现时无法与前端对接
- 如果后端按文档实现了 query 参数认证，当前前端代码将无法通过认证

**修复建议**：
1. **推荐方案**：修改代码实现，改为通过 URL query 传递 ticket（与文档一致）
   ```typescript
   const path = `${wsBase.replace(/\/$/, '')}/${endpoint.replace(/^\//, '')}?ticket=${encodeURIComponent(ticket)}`
   socket = new WebSocket(resolveWsUrl(path))
   ```
2. **备选方案**：如果消息体认证更安全，则更新架构文档以匹配实现

**优先级**：高 - 影响前后端联调

---

### 🔴 H2. 路由守卫临时禁用了认证和权限检查

**问题描述**：
`router/guards.ts` 中存在大量注释掉的认证和权限检查代码：

```typescript
// 临时禁用登录检查，用于查看页面效果
// if (to.meta?.requiresAuth && !authStore.isLoggedIn) {
//   next({ path: '/login', query: { redirect: to.fullPath } })
//   return
// }

// 临时禁用角色权限检查
// const userRole = authStore.user?.role
// const requiredRoles = to.meta?.roles
// if (!hasRole(requiredRoles, userRole)) {
//   toast.warning('无权限访问该页面')
//   next('/dashboard')
//   return
// }
```

**影响**：
- **严重安全风险**：任何人都可以访问所有页面，包括管理员页面
- 违反架构文档 `02-routing.md` 中的路由守卫设计
- 如果这段代码被部署到生产环境，将导致严重的权限绕过漏洞

**修复建议**：
立即恢复认证和权限检查逻辑，移除临时注释

**优先级**：高 - 严重安全风险

---

### 🔴 H3. 缺少 Markdown/富文本内容安全过滤

**问题描述**：
- **架构文档要求**（`01-architecture-overview.md §4`）：题目描述、公告、通知等内容渲染前必须做 sanitize（白名单策略）
- **实际情况**：
  - `package.json` 中已安装 `dompurify: ^3.3.1`
  - 但在代码中未找到 `useSanitize` composable 的实际使用
  - 未在视图组件中看到对用户生成内容的 sanitize 处理

**影响**：
- **XSS 攻击风险**：恶意用户可以在题目描述、公告中注入 JavaScript 代码
- 可能导致 Token 窃取、账号接管等严重安全问题

**修复建议**：
1. 在所有渲染用户生成内容的地方使用 `useSanitize`
2. 重点检查：
   - 靶场详情页的题目描述渲染
   - 竞赛公告渲染
   - 通知内容渲染
3. 示例：
   ```typescript
   import { useSanitize } from '@/composables/useSanitize'
   const { sanitize } = useSanitize()
   const safeHtml = sanitize(challenge.description)
   ```

**优先级**：高 - 严重安全风险

---

## 中优先级问题（建议修复）

### 🟡 M1. 缺少架构文档中定义的部分 API 模块函数

**问题描述**：
对比 `04-api-layer.md` 与实际 `src/api/` 实现，发现以下缺失：

| 模块 | 缺失函数 | 文档位置 |
|------|----------|----------|
| `notification.ts` | `markAllRead()` | §2.6 |
| `assessment.ts` | `getMyTimeline()` | §2.5 |
| `teacher.ts` | 部分函数未实现 | §2.7 |

**影响**：
- 功能不完整，部分页面可能无法正常工作
- 前后端接口对接时可能出现遗漏

**修复建议**：
补充缺失的 API 函数，确保与架构文档一致

**优先级**：中 - 影响功能完整性

---

### 🟡 M2. Element Plus 未按架构文档建议配置按需引入

**问题描述**：
- **架构文档建议**（`08-build-deploy.md §1.2.1`）：使用 `unplugin-auto-import` + `unplugin-vue-components` 实现按需引入
- **实际实现**（`main.ts:19`）：使用全量引入
  ```typescript
  app.use(ElementPlus)
  ```

**影响**：
- 打包体积增大约 30-50KB（gzip 后）
- 首屏加载时间增加

**修复建议**：
1. 移除 `main.ts` 中的 `app.use(ElementPlus)`
2. 依赖 `vite.config.ts` 中已配置的 `ElementPlusResolver` 自动按需引入
3. 验证所有 Element Plus 组件是否正常工作

**优先级**：中 - 影响性能

---

### 🟡 M3. 缺少架构文档中定义的部分基础组件

**问题描述**：
对比 `06-components.md` 与实际 `src/components/common/` 实现，发现以下缺失：

| 组件 | 状态 | 说明 |
|------|------|------|
| `AppButton.vue` | ❌ 缺失 | 文档建议封装，但未实现 |
| `AppCard.vue` | ❌ 缺失 | 文档定义，但未实现 |
| `AppInput.vue` | ❌ 缺失 | 文档定义，但未实现 |
| `AppTable.vue` | ❌ 缺失 | 文档定义，但未实现 |
| `AppPagination.vue` | ❌ 缺失 | 文档定义，但未实现 |
| `AppTag.vue` | ❌ 缺失 | 文档定义，但未实现 |
| `AppDialog.vue` | ❌ 缺失 | 文档定义，但未实现 |
| `AppDrawer.vue` | ❌ 缺失 | 文档定义，但未实现 |

**实际情况**：
- 存在 `SectionCard.vue`（未在文档中定义）
- 存在 `PageHeader.vue`（未在文档中定义）

**影响**：
- 架构文档与实际实现不一致
- 可能导致后续开发者困惑

**修复建议**：
选择以下方案之一：
1. **推荐**：更新架构文档，说明优先直接使用 Element Plus 组件，只在必要时封装 `App*` 组件
2. 补充实现文档中定义的组件
3. 移除文档中不必要的组件定义

**优先级**：中 - 影响架构一致性

---

### 🟡 M4. 缺少 WebSocket 超时检测机制

**问题描述**：
- **架构文档要求**（`05-websocket-composables.md §1.2`）：超时 60s 未收到 pong → 视为断开
- **实际实现**（`useWebSocket.ts`）：只发送 ping，但未实现超时检测逻辑

**影响**：
- 当服务端无响应时，客户端无法及时检测到连接异常
- 可能导致用户长时间停留在"已连接"状态，但实际无法接收消息

**修复建议**：
添加超时检测逻辑：
```typescript
let lastPongTime = Date.now()

socket.addEventListener('message', (evt) => {
  const msg = JSON.parse(String(evt.data))
  if (msg.type === 'pong') {
    lastPongTime = Date.now()
  }
  // ... 其他处理
})

// 在心跳定时器中检查超时
heartbeatTimer = window.setInterval(() => {
  if (Date.now() - lastPongTime > 60000) {
    socket?.close()
    return
  }
  socket?.send(JSON.stringify({ type: 'ping', ... }))
}, WS_HEARTBEAT_INTERVAL_MS)
```

**优先级**：中 - 影响可靠性

---

### 🟡 M5. 错误码映射不完整

**问题描述**：
- `utils/errorMap.ts` 中只定义了 14 个错误码
- 架构文档 `04-api-layer.md §3` 中提到的部分错误码未映射
- 缺少对未映射错误码的通用处理说明

**影响**：
- 部分错误无法给用户友好的提示
- 可能直接透传后端 message（违反安全基线）

**修复建议**：
1. 补充常见错误码映射
2. 在 `request.ts` 响应拦截器中，对未映射的错误码使用通用提示 + request_id

**优先级**：中 - 影响用户体验

---

## 低优先级问题（优化建议）

### 🟢 L1. 缺少架构文档中建议的性能优化

**问题描述**：
架构文档 `08-build-deploy.md §4` 中建议的部分优化未实现：
- 请求去重（Axios 拦截器对相同 GET 请求去重）
- 图片懒加载（`loading="lazy"` 属性）
- 虚拟滚动（审计日志等大列表场景）

**修复建议**：
按需实现，优先级：请求去重 > 图片懒加载 > 虚拟滚动

**优先级**：低 - 性能优化

---

### 🟢 L2. TypeScript 类型覆盖不完整

**问题描述**：
- 部分 API 返回值类型使用 `unknown` 或 `any`
- 例如：`contest.ts` 中的 `myTeam: ref<unknown>(null)`

**修复建议**：
补充明确的类型定义，提升类型安全

**优先级**：低 - 代码质量

---

### 🟢 L3. 缺少架构文档中建议的 .env 文件

**问题描述**：
架构文档 `08-build-deploy.md §1.1` 中定义了 `.env.development` 和 `.env.production`，但项目中未找到这些文件

**修复建议**：
添加环境变量配置文件，避免硬编码

**优先级**：低 - 配置管理

---

### 🟢 L4. 部分视图组件未实现

**问题描述**：
路由表中定义了 25 个视图组件，但部分可能尚未完整实现（需逐个检查）

**修复建议**：
按优先级实现缺失的页面

**优先级**：低 - 功能完整性

---

## 架构亮点

以下方面实现良好，值得保持：

✅ **Token 刷新机制**：实现了无感刷新，避免并发请求重复刷新
✅ **路由懒加载**：所有视图组件使用动态导入
✅ **分包策略**：合理拆分 vendor 和 echarts chunk
✅ **Tailwind CSS 集成**：使用 CSS 变量桥接 Element Plus 主题
✅ **Composables 设计**：职责清晰，复用性好
✅ **Store 划分**：遵循最小化原则，只存储全局共享状态
✅ **错误处理**：统一的错误拦截和 Toast 提示

---

## 修复优先级建议

1. **立即修复**（本周内）：
   - H2: 恢复路由守卫的认证和权限检查
   - H3: 实现 Markdown/富文本内容安全过滤

2. **短期修复**（2 周内）：
   - H1: 统一 WebSocket 认证方式
   - M1: 补充缺失的 API 函数
   - M4: 实现 WebSocket 超时检测

3. **中期优化**（1 个月内）：
   - M2: 配置 Element Plus 按需引入
   - M3: 统一架构文档与代码实现
   - M5: 完善错误码映射

4. **长期优化**（按需）：
   - L1-L4: 性能优化和代码质量提升

---

## 总结

CTF 前端架构设计文档质量高，代码实现基本遵循架构设计。主要问题集中在：

1. **安全性**：路由守卫被临时禁用、缺少 XSS 防护
2. **一致性**：WebSocket 认证方式、组件定义与文档不一致
3. **完整性**：部分 API 函数、组件未实现

建议优先修复高优先级安全问题，然后逐步完善功能完整性和架构一致性。
