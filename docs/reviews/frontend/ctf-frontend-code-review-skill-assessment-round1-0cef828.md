# CTF 前端代码 Review（skill-assessment 第 1 轮）：能力画像页面实现

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | skill-assessment |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | commit 0cef828，2 个文件，+244/-6 行 |
| 变更概述 | 实现能力画像页面，包含 ECharts 雷达图、推荐靶场、教师端功能 |
| 审查基准 | 项目 CLAUDE.md 前端规范 |
| 审查日期 | 2026-03-04 |
| 上轮问题数 | - |

## 问题清单

### 🔴 高优先级

#### [H1] ECharts 实例未在组件卸载时清理，存在内存泄漏风险
- **文件**：`code/frontend/src/views/profile/SkillProfile.vue:228-232`
- **问题描述**：`onMounted` 中创建了 ECharts 实例和 resize 监听器，但组件卸载时未清理，导致：
  1. ECharts 实例未 dispose，占用内存
  2. window resize 监听器未移除，持续触发回调
  3. 路由切换后可能导致内存泄漏
- **影响范围/风险**：长时间使用应用后内存占用持续增长，影响性能
- **修正建议**：添加 `onBeforeUnmount` 生命周期钩子清理资源

```typescript
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'

// 保存 resize 处理函数引用
const handleResize = () => chartInstance?.resize()

onMounted(() => {
  loadSkillProfile()
  loadRecommendations()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  chartInstance?.dispose()
  chartInstance = null
})
```

#### [H2] 教师端学员列表未加载，功能不可用
- **文件**：`code/frontend/src/views/profile/SkillProfile.vue:8-19`
- **问题描述**：模板中使用了 `students` 数组渲染下拉选项，但 `students` 初始化为空数组后从未赋值，导致：
  1. 教师端下拉框永远只有"我的能力画像"选项
  2. 无法选择学员查看其能力画像
  3. 教师端核心功能失效
- **影响范围/风险**：教师端功能完全不可用，需要补充学员列表加载逻辑
- **修正建议**：在 `onMounted` 中加载学员列表（需要先确认是否有对应 API）

```typescript
import { getClassStudents } from '@/api/teacher'

async function loadStudents() {
  if (!isTeacher.value) return
  try {
    // 假设教师有默认班级或需要先选择班级
    // 这里需要根据实际业务逻辑调整
    const className = authStore.user?.class_name
    if (className) {
      students.value = await getClassStudents(className)
    }
  } catch (error) {
    console.error('加载学员列表失败:', error)
  }
}

onMounted(() => {
  loadStudents()
  loadSkillProfile()
  loadRecommendations()
  window.addEventListener('resize', handleResize)
})
```

#### [H3] API 错误处理不完善，用户体验差
- **文件**：`code/frontend/src/views/profile/SkillProfile.vue:62-75, 78-91`
- **问题描述**：
  1. `loadSkillProfile` 和 `loadRecommendations` 中只有 `console.error`，用户看不到错误提示
  2. 网络错误时页面显示空状态，用户无法区分"无数据"和"加载失败"
  3. 缺少重试机制
- **影响范围/风险**：网络异常时用户体验差，无法感知错误原因
- **修正建议**：使用 Toast 或 Message 组件提示错误，区分空状态和错误状态

```typescript
const error = ref<string | null>(null)

async function loadSkillProfile() {
  loading.value = true
  error.value = null
  try {
    if (selectedStudentId.value) {
      skillProfile.value = await getStudentSkillProfile(selectedStudentId.value)
    } else {
      skillProfile.value = await getSkillProfile()
    }
    renderChart()
  } catch (err) {
    console.error('加载能力画像失败:', err)
    error.value = '加载能力画像失败，请稍后重试'
    skillProfile.value = null
    // 可选：使用 Toast 组件
    // ElMessage.error('加载能力画像失败')
  } finally {
    loading.value = false
  }
}

// 模板中添加错误状态显示
<div v-else-if="error" class="bg-error/10 border border-error/30 rounded-lg p-8 text-center">
  <p class="text-error mb-4">{{ error }}</p>
  <button @click="loadSkillProfile" class="px-4 py-2 bg-primary text-white rounded-lg">
    重试
  </button>
</div>
```

### 🟡 中优先级

#### [M1] 薄弱项阈值硬编码，缺乏灵活性
- **文件**：`code/frontend/src/views/profile/SkillProfile.vue:56-60`
- **问题描述**：薄弱项判断阈值 `60` 直接硬编码在计算属性中，无法根据业务需求调整
- **影响范围/风险**：后续需要调整阈值时需要修改代码重新部署
- **修正建议**：提取为常量或配置项

```typescript
// 在文件顶部定义常量
const WEAK_DIMENSION_THRESHOLD = 60

const weakDimensions = computed(() => {
  if (!skillProfile.value) return []
  return skillProfile.value.dimensions
    .filter(d => d.value < WEAK_DIMENSION_THRESHOLD)
    .map(d => d.name)
})
```

#### [M2] 雷达图配置硬编码颜色值，未使用主题系统
- **文件**：`code/frontend/src/views/profile/SkillProfile.vue:103-125`
- **问题描述**：
  1. 颜色值 `#666`、`#e5e7eb`、`#3b82f6` 等直接硬编码
  2. 未使用 Tailwind CSS 主题变量或 CSS 变量
  3. 主题切换时图表颜色不会跟随变化
- **影响范围/风险**：暗色模式下图表配色可能不协调
- **修正建议**：使用 CSS 变量或从计算样式中读取颜色

```typescript
function renderChart() {
  if (!chartRef.value || !skillProfile.value) return

  if (!chartInstance) {
    chartInstance = echarts.init(chartRef.value)
  }

  // 从 CSS 变量读取颜色
  const styles = getComputedStyle(document.documentElement)
  const primaryColor = styles.getPropertyValue('--color-primary').trim() || '#3b82f6'
  const borderColor = styles.getPropertyValue('--color-border').trim() || '#e5e7eb'
  const textSecondary = styles.getPropertyValue('--color-text-secondary').trim() || '#666'

  const dimensions = skillProfile.value.dimensions
  const option: EChartsOption = {
    radar: {
      indicator: dimensions.map(d => ({ name: d.name, max: 100 })),
      radius: '65%',
      splitNumber: 4,
      axisName: { color: textSecondary },
      splitLine: { lineStyle: { color: borderColor } },
      splitArea: { show: false },
      axisLine: { lineStyle: { color: borderColor } }
    },
    series: [{
      type: 'radar',
      data: [{
        value: dimensions.map(d => d.value),
        name: '能力值',
        areaStyle: { color: `${primaryColor}33` }, // 20% opacity
        lineStyle: { color: primaryColor, width: 2 },
        itemStyle: { color: primaryColor }
      }]
    }],
    tooltip: {
      trigger: 'item',
      formatter: (params: any) => {
        const data = params.data
        return dimensions.map((d, i) => `${d.name}: ${data.value[i]}`).join('<br/>')
      }
    }
  }

  chartInstance.setOption(option)
}
```

#### [M3] TypeScript 类型过于宽松，缺少类型安全
- **文件**：`code/frontend/src/views/profile/SkillProfile.vue:122`
- **问题描述**：tooltip formatter 参数使用 `any` 类型，失去类型检查
- **影响范围/风险**：运行时可能访问不存在的属性导致错误
- **修正建议**：定义明确的类型或使用 ECharts 提供的类型

```typescript
import type { EChartsOption, RadarSeriesOption } from 'echarts'

// 定义 tooltip 参数类型
interface TooltipParams {
  data: {
    value: number[]
    name: string
  }
}

const option: EChartsOption = {
  // ...
  tooltip: {
    trigger: 'item',
    formatter: (params: TooltipParams) => {
      const data = params.data
      return dimensions.map((d, i) => `${d.name}: ${data.value[i]}`).join('<br/>')
    }
  }
}
```

#### [M4] 缺少加载状态的骨架屏，用户体验可优化
- **文件**：`code/frontend/src/views/profile/SkillProfile.vue:18-20`
- **问题描述**：加载状态只有一个 spinner，首次加载时页面空白时间较长
- **影响范围/风险**：用户体验不够流畅
- **修正建议**：使用骨架屏占位

```vue
<div v-if="loading" class="space-y-6">
  <!-- 雷达图骨架屏 -->
  <div class="bg-surface rounded-lg p-6 border border-border">
    <div class="h-6 w-32 bg-background animate-pulse rounded mb-4"></div>
    <div class="w-full h-[400px] bg-background animate-pulse rounded"></div>
  </div>
  <!-- 推荐靶场骨架屏 -->
  <div class="bg-surface rounded-lg p-6 border border-border">
    <div class="h-6 w-24 bg-background animate-pulse rounded mb-4"></div>
    <div class="space-y-3">
      <div class="h-20 bg-background animate-pulse rounded"></div>
      <div class="h-20 bg-background animate-pulse rounded"></div>
    </div>
  </div>
</div>
```

### 🟢 低优先级

#### [L1] 难度映射函数重复定义，应提取为工具函数
- **文件**：`code/frontend/src/views/profile/SkillProfile.vue:128-145`
- **问题描述**：`difficultyClass` 和 `difficultyLabel` 函数在多个组件中重复定义（如靶场列表页）
- **影响范围/风险**：维护成本高，修改时需要同步多处
- **修正建议**：提取到 `@/utils/challenge.ts` 工具文件

```typescript
// utils/challenge.ts
export function getDifficultyClass(difficulty: ChallengeDifficulty): string {
  const map: Record<ChallengeDifficulty, string> = {
    beginner: 'bg-green-100 text-green-700',
    easy: 'bg-blue-100 text-blue-700',
    medium: 'bg-yellow-100 text-yellow-700',
    hard: 'bg-orange-100 text-orange-700',
    hell: 'bg-red-100 text-red-700'
  }
  return map[difficulty]
}

export function getDifficultyLabel(difficulty: ChallengeDifficulty): string {
  const map: Record<ChallengeDifficulty, string> = {
    beginner: '入门',
    easy: '简单',
    medium: '中等',
    hard: '困难',
    hell: '地狱'
  }
  return map[difficulty]
}

// 组件中使用
import { getDifficultyClass, getDifficultyLabel } from '@/utils/challenge'
```

#### [L2] 日期格式化函数可以提取为工具函数
- **文件**：`code/frontend/src/views/profile/SkillProfile.vue:149-151`
- **问题描述**：`formatDate` 函数功能单一，可能在其他组件中也需要使用
- **影响范围/风险**：代码复用性低
- **修正建议**：提取到 `@/utils/date.ts` 或使用第三方库（如 dayjs）

```typescript
// utils/date.ts
export function formatDateTime(isoString: string): string {
  return new Date(isoString).toLocaleString('zh-CN')
}

export function formatDate(isoString: string): string {
  return new Date(isoString).toLocaleDateString('zh-CN')
}

export function formatTime(isoString: string): string {
  return new Date(isoString).toLocaleTimeString('zh-CN')
}
```

#### [L3] 空状态提示文案可以更友好
- **文件**：`code/frontend/src/views/profile/SkillProfile.vue:23-26`
- **问题描述**：空状态提示较为简单，可以添加引导操作
- **影响范围/风险**：用户体验可优化
- **修正建议**：添加跳转到靶场列表的按钮

```vue
<div v-else-if="!skillProfile" class="bg-surface rounded-lg p-8 text-center border border-border">
  <svg class="w-16 h-16 mx-auto mb-4 text-text-tertiary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
  </svg>
  <p class="text-text-secondary mb-2">暂无能力画像数据</p>
  <p class="text-sm text-text-tertiary mb-4">完成更多靶场挑战后，系统将为你生成能力画像</p>
  <button
    @click="router.push('/challenges')"
    class="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 transition-colors"
  >
    前往靶场
  </button>
</div>
```

#### [L4] 推荐靶场卡片缺少 hover 反馈优化
- **文件**：`code/frontend/src/views/profile/SkillProfile.vue:52-68`
- **问题描述**：卡片有 `cursor-pointer` 和 `hover:border-primary`，但缺少其他视觉反馈
- **影响范围/风险**：交互体验可优化
- **修正建议**：添加 hover 时的背景色变化和过渡动画

```vue
<div
  v-for="item in recommendations"
  :key="item.challenge_id"
  class="flex items-start justify-between p-4 bg-background rounded-lg border border-border hover:border-primary hover:bg-surface transition-all cursor-pointer"
  @click="goToChallenge(item.challenge_id)"
>
  <!-- ... -->
</div>
```

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 3 |
| 🟡 中 | 4 |
| 🟢 低 | 4 |
| 合计 | 11 |

## 总体评价

代码整体结构清晰，功能实现完整，符合 Vue 3 Composition API 最佳实践。主要问题集中在：

1. **资源管理**：ECharts 实例和事件监听器未清理，存在内存泄漏风险（高优先级）
2. **功能完整性**：教师端学员列表未加载，核心功能不可用（高优先级）
3. **错误处理**：缺少用户友好的错误提示和重试机制（高优先级）
4. **代码质量**：存在硬编码、类型不严格、代码重复等问题（中/低优先级）

建议优先修复 3 个高优先级问题，确保功能可用性和稳定性，然后逐步优化中低优先级问题提升代码质量和用户体验。
