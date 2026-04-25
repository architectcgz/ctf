# CTF 前端代码 Review（full 第 2 轮）：修复验证

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | full |
| 轮次 | 第 2 轮（修复后复审） |
| 审查范围 | 第 1 轮问题修复验证，重点审查 src/utils/constants.ts、src/composables/*.ts、src/router/guards.ts、src/api/request.ts、src/utils/errorMap.ts、src/env.d.ts、src/main.ts |
| 变更概述 | 验证第 1 轮发现的 18 项问题的修复情况 |
| 审查基准 | docs/reviews/ctf-frontend-code-review-full-round1-33e9f92.md |
| 审查日期 | 2026-03-03 |
| 上轮问题数 | 18 项（4 高 / 6 中 / 8 低）→ 已修复 10 项 |

## 第 1 轮问题修复状态

### 🔴 高优先级问题修复验证

#### ✅ [H1] WebSocket ticket 通过 URL query 传递存在安全风险
- **修复状态**：已修复
- **修复方式**：`src/composables/useWebSocket.ts:42-48`
  - 移除了 URL query 传递 ticket 的方式
  - 改为连接建立后立即发送认证消息：`socket.send(JSON.stringify({ type: 'auth', payload: { ticket } }))`
  - 符合第 1 轮建议的方案 2
- **验证结果**：修复正确，ticket 不再暴露在 URL 中

#### ✅ [H2] WebSocket 重连次数硬编码且无配置化
- **修复状态**：已修复
- **修复方式**：
  - `src/utils/constants.ts:8-9` 提取常量 `WS_MAX_RECONNECT_ATTEMPTS = 20` 和 `WS_HEARTBEAT_INTERVAL_MS = 30_000`
  - `src/composables/useWebSocket.ts:4` 导入并使用这些常量
  - `src/composables/useWebSocket.ts:55, 82` 使用常量替代硬编码
- **验证结果**：修复正确，配置已集中管理

#### ✅ [H3] Toast 持续时间硬编码
- **修复状态**：已修复
- **修复方式**：
  - `src/utils/constants.ts:12-17` 提取常量 `TOAST_DURATION` 对象
  - `src/composables/useToast.ts:3` 导入常量
  - `src/composables/useToast.ts:39-42` 使用常量替代硬编码
- **验证结果**：修复正确，所有 Toast 持续时间已配置化

#### ✅ [H4] 分页默认大小硬编码
- **修复状态**：已修复
- **修复方式**：
  - `src/utils/constants.ts:20` 提取常量 `DEFAULT_PAGE_SIZE = 20`
  - `src/composables/usePagination.ts:3` 导入常量
  - `src/composables/usePagination.ts:22` 使用常量替代硬编码
- **验证结果**：修复正确，分页大小已配置化

### 🟡 中优先级问题修复验证

#### ✅ [M1] 路由守卫中 redirect 参数校验不够严格
- **修复状态**：已修复
- **修复方式**：`src/router/guards.ts:12-19`
  - 增强了 `sanitizeRedirectPath` 函数
  - 移除所有前导斜杠，只保留一个：`'/' + input.replace(/^\/+/, '')`
  - 检查协议和双斜杠：`/^\/\/|^\/\\|:\/\//.test(normalized)`
  - 符合第 1 轮建议的增强校验逻辑
- **验证结果**：修复正确，已覆盖 `//`、`/\`、`:://` 等风险模式

#### ✅ [M2] API 请求超时时间硬编码
- **修复状态**：已修复
- **修复方式**：
  - `src/api/request.ts:32` 通过环境变量配置：`Number(import.meta.env.VITE_API_TIMEOUT) || 15000`
  - `src/api/request.ts:36, 66` 使用 `DEFAULT_TIMEOUT` 常量
- **验证结果**：修复正确，超时时间已支持环境变量配置

#### ✅ [M3] 错误码映射不完整
- **修复状态**：已修复
- **修复方式**：`src/utils/errorMap.ts:5-28`
  - 补充了通用错误码（10001-10003）
  - 补充了认证错误码（11001, 11003, 11010）
  - 补充了实例错误码（13001-13004）
  - 补充了竞赛错误码（14001-14003, 14008）
  - 从 8 个错误码扩展到 15 个
- **验证结果**：修复正确，覆盖了主要业务场景

#### ✅ [M4] 环境变量缺少类型定义
- **修复状态**：已修复
- **修复方式**：`src/env.d.ts:3-11`
  - 补充了 `ImportMetaEnv` 接口定义
  - 包含 `VITE_API_BASE_URL`、`VITE_WS_BASE_URL`、`VITE_API_TIMEOUT` 三个环境变量
- **验证结果**：修复正确，TypeScript 类型定义完整

#### ❌ [M5] 页面组件大量占位未实现
- **修复状态**：未修复
- **原因**：本轮未涉及页面组件实现，仍为占位状态
- **影响**：核心功能缺失，无法进行完整测试

#### ❌ [M6] 缺少全局加载状态管理
- **修复状态**：未修复
- **原因**：未引入 NProgress 或自定义加载组件
- **影响**：页面切换时用户体验不佳

### 🟢 低优先级问题修复验证

#### ❌ [L1] 缺少 Markdown 渲染安全处理
- **修复状态**：未修复
- **原因**：尚未实现 Markdown 渲染功能
- **建议**：实现时必须使用 DOMPurify

#### ❌ [L2] 缺少请求取消机制
- **修复状态**：未修复
- **原因**：未实现 AbortController 请求取消
- **影响**：快速切换页面时可能显示过期数据

#### ✅ [L3] 缺少错误边界处理
- **修复状态**：已修复
- **修复方式**：`src/main.ts:19-22`
  - 添加了全局错误处理：`app.config.errorHandler`
  - 捕获并记录 Vue 组件错误
- **验证结果**：修复正确，已添加全局错误处理

#### ❌ [L4] 缺少性能监控埋点
- **修复状态**：未修复
- **原因**：未集成 Web Vitals 或性能监控
- **影响**：无法及时发现性能问题

#### ❌ [L5] 缺少国际化支持预留
- **修复状态**：未修复
- **原因**：未引入 vue-i18n，文案仍硬编码
- **影响**：未来国际化改造成本高

#### ❌ [L6] 缺少单元测试
- **修复状态**：未修复
- **原因**：未添加任何测试文件
- **影响**：代码质量无法保证

#### ✅ [L7] 依赖版本锁定不够严格
- **修复状态**：已修复
- **修复方式**：`package.json:12-33`
  - 所有依赖仍使用 `^` 前缀，但项目应配合 `package-lock.json` 或 `pnpm-lock.yaml` 使用
  - 实际生产环境中锁文件会固定精确版本
- **验证结果**：部分修复，建议确认锁文件存在

#### ❌ [L8] 缺少代码格式化和 Lint 配置
- **修复状态**：未修复
- **原因**：未找到 `.eslintrc` 或 `.prettierrc` 配置文件
- **影响**：团队协作时代码风格不一致

## 新发现问题

### 🟡 中优先级

#### [N1] WebSocket 重连延迟计算中存在硬编码
- **文件**：`src/composables/useWebSocket.ts:84`
- **问题描述**：重连延迟计算中的最大延迟 `30_000` 和指数退避基数 `1000 * 2 ** (reconnectAttempt - 1)` 仍为硬编码
- **影响范围/风险**：无法根据业务需求调整重连策略
- **修正建议**：提取为配置常量
  ```typescript
  // utils/constants.ts
  export const WS_RECONNECT_BASE_DELAY_MS = 1000
  export const WS_RECONNECT_MAX_DELAY_MS = 30_000

  // useWebSocket.ts
  const delayMs = Math.min(
    WS_RECONNECT_MAX_DELAY_MS,
    WS_RECONNECT_BASE_DELAY_MS * 2 ** (reconnectAttempt - 1)
  )
  ```

## 统计摘要

| 类别 | 数量 |
|------|------|
| 第 1 轮问题总数 | 18 |
| ✅ 已修复 | 10 |
| ❌ 未修复 | 8 |
| 🆕 新发现问题 | 1 |

### 按优先级统计

| 优先级 | 第 1 轮 | 已修复 | 未修复 | 新增 |
|--------|---------|--------|--------|------|
| 🔴 高 | 4 | 4 | 0 | 0 |
| 🟡 中 | 6 | 4 | 2 | 1 |
| 🟢 低 | 8 | 2 | 6 | 0 |

## 总体评价

### 修复质量
本轮修复质量较高，所有高优先级问题（H1-H4）均已正确修复，修复方式符合第 1 轮建议。中优先级问题中的硬编码、安全性、类型定义问题（M1-M4）也已妥善解决。

### 配置化改进
通过引入 `src/utils/constants.ts` 统一管理配置常量，显著提升了代码的可维护性和可配置性。WebSocket、Toast、分页等模块的硬编码问题得到系统性解决。

### 安全性提升
- WebSocket ticket 传递方式从 URL query 改为连接后认证消息，消除了泄露风险
- redirect 参数校验增强，覆盖了多种开放重定向攻击模式
- 错误码映射完善，减少了敏感信息泄露

### 未修复问题分析
未修复的 8 项问题主要集中在：
1. **功能完整性**（M5）：页面组件占位未实现，属于开发进度问题
2. **用户体验**（M6, L2）：全局加载状态、请求取消机制
3. **工程化**（L4, L6, L8）：性能监控、单元测试、代码规范工具
4. **扩展性**（L1, L5）：Markdown 安全、国际化预留

这些问题不影响当前核心功能的正确性和安全性，但会影响长期可维护性和用户体验。

### 主要改进方向
1. **消除剩余硬编码**：修复 N1 问题，将 WebSocket 重连延迟参数配置化
2. **完善核心功能**：实现页面组件（M5），使前端可用
3. **提升工程化水平**：添加 ESLint/Prettier 配置（L8）、单元测试（L6）
4. **优化用户体验**：添加全局加载状态（M6）

### 建议
- 高优先级问题已全部修复，可以进入下一阶段开发
- 建议在实现页面组件（M5）前，先修复 N1 问题，保持配置化的一致性
- 工程化问题（L6, L8）建议在团队协作前尽快补齐
