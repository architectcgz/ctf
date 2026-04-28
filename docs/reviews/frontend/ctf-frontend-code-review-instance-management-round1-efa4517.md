# CTF 前端代码 Review（instance-management 第 1 轮）：实例管理页面功能实现

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | instance-management |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | commit efa4517，1 个文件，+192/-5 行 |
| 变更概述 | 实现实例管理页面的倒计时、复制地址、延时、销毁、超时提醒功能 |
| 审查基准 | API contracts (`code/frontend/src/api/contracts.ts`) |
| 审查日期 | 2026-03-04 |
| 上轮问题数 | - |

## 问题清单

### 🔴 高优先级

#### [H1] 未调用后端 API，使用硬编码 Mock 数据
- **文件**：`code/frontend/src/views/instances/InstanceList.vue:115-118`
- **问题描述**：
  - 第 115-118 行直接硬编码了两条 Mock 实例数据
  - 未调用 `api/instance.ts` 中已定义的 `getMyInstances()` API
  - `loading` 状态始终为 `false`，未实际发起网络请求
- **影响范围/风险**：
  - 页面无法显示真实数据
  - 用户操作（延时、销毁）不会同步到后端
  - 功能完全不可用
- **修正建议**：
```typescript
// 在 onMounted 中调用 API
import { getMyInstances, destroyInstance as apiDestroyInstance, extendInstance } from '@/api/instance'

onMounted(async () => {
  loading.value = true
  try {
    const data = await getMyInstances()
    instances.value = data.map(item => ({
      ...item,
      remaining: calculateRemaining(item.expires_at)
    }))
  } catch (error) {
    console.error('加载实例失败:', error)
  } finally {
    loading.value = false
  }
  timer = window.setInterval(updateCountdown, 1000)
})

// 计算剩余秒数
function calculateRemaining(expiresAt: string): number {
  return Math.max(0, Math.floor((new Date(expiresAt).getTime() - Date.now()) / 1000))
}
```

#### [H2] 延时和销毁操作未调用后端 API
- **文件**：`code/frontend/src/views/instances/InstanceList.vue:154-165`
- **问题描述**：
  - `extendTime()` 只修改本地状态，未调用 `extendInstance()` API
  - `destroyInstance()` 只过滤本地数组，未调用 `destroyInstance()` API
  - 操作不会持久化到后端
- **影响范围/风险**：
  - 刷新页面后操作丢失
  - 后端实例继续运行，造成资源浪费
  - 用户误以为操作成功
- **修正建议**：
```typescript
async function extendTime(id: string) {
  try {
    const result = await extendInstance(id)
    const instance = instances.value.find(i => i.id === id)
    if (instance) {
      instance.remaining = calculateRemaining(result.expires_at)
      instance.remaining_extends = result.remaining_extends
      warnedInstances.delete(id)
    }
  } catch (error) {
    console.error('延时失败:', error)
    // 显示错误提示
  }
}

async function destroyInstance(id: string) {
  try {
    await apiDestroyInstance(id)
    instances.value = instances.value.filter(i => i.id !== id)
    warnedInstances.delete(id)
  } catch (error) {
    console.error('销毁失败:', error)
    // 显示错误提示
  }
}
```

#### [H3] 类型定义不匹配 API contracts
- **文件**：`code/frontend/src/views/instances/InstanceList.vue:115-118`
- **问题描述**：
  - Mock 数据使用的字段名与 `InstanceListItem` 类型不一致
  - Mock: `title`, `category`, `difficulty`, `address`, `remaining`
  - API: `challenge_title`, `category`, `difficulty`, `access_url`, `expires_at`
  - `status` 值 `'running'` 不在 `InstanceStatus` 枚举中（应为 `'running'` 但类型定义中实际存在）
- **影响范围/风险**：
  - 接入真实 API 后字段映射错误
  - TypeScript 类型检查失效
  - 运行时可能出现 undefined 错误
- **修正建议**：
```typescript
// 定义本地使用的 ViewModel 类型
interface InstanceViewModel extends InstanceListItem {
  remaining: number // 计算得出的剩余秒数
}

const instances = ref<InstanceViewModel[]>([])

// 在 API 调用后转换
instances.value = data.map(item => ({
  ...item,
  remaining: calculateRemaining(item.expires_at)
}))
```

#### [H4] 缺少错误处理和用户反馈
- **文件**：`code/frontend/src/views/instances/InstanceList.vue:150-165`
- **问题描述**：
  - `copyAddress()` 未处理 clipboard API 失败情况
  - `extendTime()` 和 `destroyInstance()` 无成功/失败提示
  - 无网络错误、权限错误的用户反馈
- **影响范围/风险**：
  - 用户不知道操作是否成功
  - 复制失败时无提示（某些浏览器需要 HTTPS）
  - 体验差
- **修正建议**：
```typescript
async function copyAddress(address: string) {
  try {
    await navigator.clipboard.writeText(address)
    // 显示成功提示（使用 Toast 组件）
  } catch (error) {
    console.error('复制失败:', error)
    // 降级方案：使用 document.execCommand('copy')
    // 或显示错误提示
  }
}
```

### 🟡 中优先级

#### [M1] 硬编码魔法数字未提取为常量
- **文件**：`code/frontend/src/views/instances/InstanceList.vue:51, 114, 157, 179`
- **问题描述**：
  - `300`（5 分钟阈值）出现在第 51、179 行
  - `3`（最大实例数）硬编码在第 114 行
  - `1800`（延时 30 分钟）硬编码在第 157 行
  - 违反全局 CLAUDE.md 中的"禁止硬编码"规范
- **影响范围/风险**：
  - 修改阈值需要多处改动
  - 可维护性差
  - 与后端配置不一致时难以排查
- **修正建议**：
```typescript
// 提取为常量
const MAX_INSTANCES = 3
const WARNING_THRESHOLD_SECONDS = 300 // 5 分钟
const EXTEND_DURATION_SECONDS = 1800 // 30 分钟

// 或从配置/环境变量读取
const config = {
  maxInstances: import.meta.env.VITE_MAX_INSTANCES || 3,
  warningThreshold: 300,
  extendDuration: 1800
}
```

#### [M2] 倒计时逻辑存在时间漂移风险
- **文件**：`code/frontend/src/views/instances/InstanceList.vue:174-186`
- **问题描述**：
  - 使用 `setInterval` 每秒递减 `remaining`
  - 未考虑 JavaScript 定时器不精确的问题
  - 长时间运行后可能累积误差（页面挂起、CPU 节流）
- **影响范围/风险**：
  - 显示的剩余时间与实际不符
  - 可能提前或延后触发超时提醒
- **修正建议**：
```typescript
// 基于服务器时间计算，而非递减
function updateCountdown() {
  const now = Date.now()
  instances.value.forEach(instance => {
    if (instance.status === 'running') {
      const expiresAt = new Date(instance.expires_at).getTime()
      instance.remaining = Math.max(0, Math.floor((expiresAt - now) / 1000))

      if (instance.remaining < WARNING_THRESHOLD_SECONDS && !warnedInstances.has(instance.id)) {
        warnedInstances.add(instance.id)
        warningInstance.value = instance
        showWarning.value = true
      }
    }
  })
}
```

#### [M3] 缺少 TypeScript 类型注解
- **文件**：`code/frontend/src/views/instances/InstanceList.vue:121`
- **问题描述**：
  - `warningInstance` 使用 `ref<any>(null)`，丢失类型安全
  - 应使用明确的类型定义
- **影响范围/风险**：
  - 失去 TypeScript 类型检查保护
  - IDE 无法提供准确的代码补全
- **修正建议**：
```typescript
const warningInstance = ref<InstanceViewModel | null>(null)
```

#### [M4] 字段映射不一致
- **文件**：`code/frontend/src/views/instances/InstanceList.vue:20, 40`
- **问题描述**：
  - 模板中使用 `instance.title`，但 API 返回的是 `challenge_title`
  - 模板中使用 `instance.address`，但 API 返回的是 `access_url`
  - 需要在数据加载时做字段映射
- **影响范围/风险**：
  - 接入真实 API 后显示为空
  - 运行时错误
- **修正建议**：
```typescript
// 在 API 调用后映射字段
instances.value = data.map(item => ({
  ...item,
  title: item.challenge_title, // 映射字段
  address: item.access_url || item.ssh_info ? `${item.ssh_info?.host}:${item.ssh_info?.port}` : '',
  remaining: calculateRemaining(item.expires_at)
}))
```

#### [M5] 状态标签映射不完整
- **文件**：`code/frontend/src/views/instances/InstanceList.vue:128-141`
- **问题描述**：
  - `getStatusLabel()` 只处理了 `running`, `starting`, `stopping` 三种状态
  - API 定义的 `InstanceStatus` 包含 8 种状态：`pending`, `creating`, `running`, `expired`, `destroying`, `destroyed`, `failed`, `crashed`
  - 未处理的状态会直接显示英文原值
- **影响范围/风险**：
  - 用户看到未翻译的状态值
  - 体验不一致
- **修正建议**：
```typescript
function getStatusLabel(status: InstanceStatus): string {
  const labels: Record<InstanceStatus, string> = {
    pending: '等待中',
    creating: '创建中',
    running: '运行中',
    expired: '已过期',
    destroying: '销毁中',
    destroyed: '已销毁',
    failed: '失败',
    crashed: '崩溃'
  }
  return labels[status] || status
}

function getStatusClass(status: InstanceStatus): string {
  const classes: Record<InstanceStatus, string> = {
    pending: 'text-[#f59e0b]',
    creating: 'text-[#f59e0b]',
    running: 'text-[#22c55e]',
    expired: 'text-[var(--color-text-muted)]',
    destroying: 'text-[#f59e0b]',
    destroyed: 'text-[var(--color-text-muted)]',
    failed: 'text-[#ef4444]',
    crashed: 'text-[#ef4444]'
  }
  return classes[status] || 'text-[var(--color-text-muted)]'
}
```

#### [M6] 销毁操作缺少二次确认
- **文件**：`code/frontend/src/views/instances/InstanceList.vue:66-70`
- **问题描述**：
  - 点击"销毁"按钮直接执行，无确认弹窗
  - 误操作风险高（销毁是不可逆操作）
- **影响范围/风险**：
  - 用户误点击导致实例被销毁
  - 丢失工作进度
- **修正建议**：
```typescript
// 添加确认弹窗
async function confirmDestroy(id: string) {
  if (confirm('确定要销毁该实例吗？此操作不可恢复。')) {
    await destroyInstance(id)
  }
}

// 或使用自定义确认对话框组件
```

### 🟢 低优先级

#### [L1] 空状态提示不够友好
- **文件**：`code/frontend/src/views/instances/InstanceList.vue:75-77`
- **问题描述**：
  - 空状态只显示"暂无运行中的实例"
  - 缺少引导用户创建实例的入口
- **影响范围/风险**：
  - 新用户不知道如何创建实例
  - 体验不够友好
- **修正建议**：
```vue
<div v-if="instances.length === 0" class="flex flex-col items-center justify-center py-12 text-center">
  <div class="text-[var(--color-text-muted)] mb-4">暂无运行中的实例</div>
  <router-link to="/challenges" class="text-[var(--color-primary)] hover:underline">
    前往靶场列表创建实例
  </router-link>
</div>
```

#### [L2] 弹窗关闭逻辑不完整
- **文件**：`code/frontend/src/views/instances/InstanceList.vue:81-106`
- **问题描述**：
  - 弹窗只能通过点击背景或按钮关闭
  - 缺少 ESC 键关闭、右上角关闭按钮
- **影响范围/风险**：
  - 用户体验略差
  - 不符合常见弹窗交互习惯
- **修正建议**：
```typescript
// 添加 ESC 键监听
function handleEscape(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    showWarning.value = false
  }
}

watch(showWarning, (show) => {
  if (show) {
    document.addEventListener('keydown', handleEscape)
  } else {
    document.removeEventListener('keydown', handleEscape)
  }
})
```

#### [L3] 延时按钮未显示剩余延时次数
- **文件**：`code/frontend/src/views/instances/InstanceList.vue:58-64`
- **问题描述**：
  - API 返回的 `remaining_extends` 字段未使用
  - 用户不知道还能延时几次
  - 延时次数用尽后按钮仍可点击
- **影响范围/风险**：
  - 用户体验不佳
  - 可能触发后端错误
- **修正建议**：
```vue
<button
  v-if="instance.status === 'running'"
  @click="extendTime(instance.id)"
  :disabled="instance.remaining_extends <= 0"
  class="rounded-lg border border-[var(--color-border-default)] bg-[#21262d] px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] transition-colors duration-150 hover:bg-[#30363d] disabled:opacity-50 disabled:cursor-not-allowed"
>
  延时 +30min ({{ instance.remaining_extends }})
</button>
```

#### [L4] 定时器清理时机不够严谨
- **文件**：`code/frontend/src/views/instances/InstanceList.vue:192-194`
- **问题描述**：
  - 只在 `onUnmounted` 时清理定时器
  - 如果组件在 `onMounted` 执行前就被销毁，`timer` 为 `null`，清理逻辑正常
  - 但代码风格上可以更严谨
- **影响范围/风险**：
  - 极端情况下可能有内存泄漏（概率极低）
- **修正建议**：
```typescript
onUnmounted(() => {
  if (timer !== null) {
    clearInterval(timer)
    timer = null
  }
})
```

#### [L5] 颜色值硬编码，未使用 CSS 变量
- **文件**：`code/frontend/src/views/instances/InstanceList.vue:22-27, 51, 138-140`
- **问题描述**：
  - 部分颜色使用硬编码的 Hex 值（如 `#06b6d4`, `#34d399`, `#f59e0b`, `#22c55e`）
  - 未统一使用 CSS 变量，不利于主题切换
- **影响范围/风险**：
  - 主题切换时颜色不一致
  - 可维护性差
- **修正建议**：
```typescript
// 定义语义化的 CSS 变量或使用 Tailwind 的主题色
// 例如：text-cyan-500, text-emerald-400 等
// 或在 CSS 中定义：
// --color-category-web: #06b6d4;
// --color-difficulty-easy: #34d399;
```

#### [L6] 缺少加载骨架屏
- **文件**：`code/frontend/src/views/instances/InstanceList.vue:8-10`
- **问题描述**：
  - 加载状态只显示一个 spinner
  - 可以使用骨架屏提升体验
- **影响范围/风险**：
  - 体验略差（非关键问题）
- **修正建议**：
```vue
<!-- 使用骨架屏替代 spinner -->
<div v-if="loading" class="space-y-4">
  <div v-for="i in 2" :key="i" class="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-5 animate-pulse">
    <div class="h-6 bg-[var(--color-bg-hover)] rounded w-1/3 mb-3"></div>
    <div class="h-4 bg-[var(--color-bg-hover)] rounded w-1/2"></div>
  </div>
</div>
```

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 4 |
| 🟡 中 | 6 |
| 🟢 低 | 6 |
| 合计 | 16 |

## 总体评价

本次变更实现了实例管理页面的核心 UI 和交互逻辑，包括倒计时、复制地址、延时、销毁、超时提醒等功能。代码结构清晰，Vue Composition API 使用规范，定时器清理正确。

**主要问题**：

1. **功能完整性严重不足**：未调用任何后端 API，所有操作都是本地 Mock，功能完全不可用（4 个高优先级问题）
2. **类型安全缺失**：字段映射与 API contracts 不一致，存在运行时错误风险
3. **硬编码问题**：魔法数字未提取为常量，违反项目规范
4. **用户体验待完善**：缺少错误提示、二次确认、空状态引导等

**必须修复的问题**：

- **[H1] ~ [H4]**：所有高优先级问题必须在下一轮修复，否则功能无法使用
- **[M1] ~ [M6]**：所有中优先级问题必须修复，涉及类型安全、数据一致性、用户体验
- **[L1] ~ [L6]**：所有低优先级问题也需要修复，提升代码质量和用户体验

**下一步建议**：

1. 优先接入后端 API（[H1], [H2]）
2. 修复类型定义和字段映射（[H3], [M4]）
3. 添加错误处理和用户反馈（[H4]）
4. 提取硬编码常量（[M1]）
5. 完善状态处理和交互细节（[M5], [M6], [L1] ~ [L6]）
