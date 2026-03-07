<template>
  <div class="space-y-6">
    <!-- 教师端学员选择器 -->
    <div v-if="isTeacher" class="bg-surface rounded-lg p-4 border border-border">
      <label class="block text-sm font-medium mb-2">查看学员能力画像</label>
      <select
        v-model="selectedStudentId"
        class="w-full px-3 py-2 bg-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
      >
        <option value="">我的能力画像</option>
        <option v-for="student in students" :key="student.id" :value="student.id">
          {{ student.name || student.username }} ({{ student.username }})
        </option>
      </select>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="space-y-6">
      <div class="bg-surface rounded-lg p-6 border border-border">
        <div class="h-6 w-32 bg-background animate-pulse rounded mb-4"></div>
        <div class="w-full h-[400px] bg-background animate-pulse rounded"></div>
      </div>
      <div class="bg-surface rounded-lg p-6 border border-border">
        <div class="h-6 w-24 bg-background animate-pulse rounded mb-4"></div>
        <div class="space-y-3">
          <div class="h-20 bg-background animate-pulse rounded"></div>
          <div class="h-20 bg-background animate-pulse rounded"></div>
        </div>
      </div>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="bg-error/10 border border-error/30 rounded-lg p-8 text-center">
      <p class="text-error mb-4">{{ error }}</p>
      <button @click="loadSkillProfile" class="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary/90 transition-colors">
        重试
      </button>
    </div>

    <!-- 空状态 -->
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
        去做题
      </button>
    </div>

    <!-- 能力画像内容 -->
    <template v-else>
      <!-- 雷达图 -->
      <div class="bg-surface rounded-lg p-6 border border-border">
        <h2 class="text-lg font-semibold mb-4">能力维度分析</h2>
        <div ref="chartRef" class="w-full h-[400px]"></div>
        <p v-if="skillProfile.updated_at" class="text-xs text-text-tertiary mt-4 text-center">
          更新时间：{{ formatDate(skillProfile.updated_at) }}
        </p>
      </div>

      <!-- 薄弱项提示 -->
      <div v-if="weakDimensions.length > 0" class="bg-warning/10 border border-warning/30 rounded-lg p-4">
        <h3 class="text-sm font-medium text-warning mb-2">💡 薄弱项提示</h3>
        <p class="text-sm text-text-secondary">
          建议加强：<span class="font-medium">{{ weakDimensions.join('、') }}</span>
        </p>
      </div>

      <!-- 推荐靶场 -->
      <div class="bg-surface rounded-lg p-6 border border-border">
        <h2 class="text-lg font-semibold mb-4">推荐靶场</h2>
        <div v-if="loadingRecommendations" class="text-center py-8 text-text-secondary">加载中...</div>
        <div v-else-if="recommendations.length === 0" class="text-center py-8 text-text-secondary">
          暂无推荐靶场
        </div>
        <div v-else class="space-y-3">
          <div
            v-for="item in recommendations"
            :key="item.challenge_id"
            class="flex items-start justify-between p-4 bg-background rounded-lg border border-border hover:border-primary hover:bg-surface transition-all cursor-pointer"
            @click="goToChallenge(item.challenge_id)"
          >
            <div class="flex-1">
              <div class="flex items-center gap-2 mb-1">
                <h3 class="font-medium">{{ item.title }}</h3>
                <span
                  class="px-2 py-0.5 text-xs rounded"
                  :class="difficultyClass(item.difficulty)"
                >
                  {{ difficultyLabel(item.difficulty) }}
                </span>
              </div>
              <p class="text-sm text-text-secondary">{{ item.reason }}</p>
            </div>
            <svg class="w-5 h-5 text-text-tertiary flex-shrink-0 ml-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import * as echarts from 'echarts'
import type { EChartsOption } from 'echarts'

import { getRecommendations, getSkillProfile } from '@/api/assessment'
import { getClassStudents, getStudentRecommendations, getStudentSkillProfile } from '@/api/teacher'
import type { RecommendationItem, SkillProfileData, TeacherStudentItem } from '@/api/contracts'
import { useAuthStore } from '@/stores/auth'
import { difficultyClass, difficultyLabel } from '@/utils/challenge'
import { formatDate } from '@/utils/format'

const authStore = useAuthStore()
const router = useRouter()

const WEAK_DIMENSION_THRESHOLD = 60

const isTeacher = computed(() => authStore.isTeacher)
const selectedStudentId = ref('')
const students = ref<TeacherStudentItem[]>([])

const loading = ref(false)
const error = ref<string | null>(null)
const skillProfile = ref<SkillProfileData | null>(null)
const chartRef = ref<HTMLElement>()
let chartInstance: echarts.ECharts | null = null

const loadingRecommendations = ref(false)
const recommendations = ref<RecommendationItem[]>([])

const weakDimensions = computed(() => {
  if (!skillProfile.value) return []
  return skillProfile.value.dimensions
    .filter(d => d.value < WEAK_DIMENSION_THRESHOLD)
    .map(d => d.name)
})

// 加载学员列表
async function loadStudents() {
  if (!isTeacher.value) return
  try {
    const className = authStore.user?.class_name
    if (className) {
      students.value = await getClassStudents(className)
    }
  } catch (err) {
    console.error('加载学员列表失败:', err)
  }
}

// 加载能力画像
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
  } finally {
    loading.value = false
  }
}

// 加载推荐靶场
async function loadRecommendations() {
  loadingRecommendations.value = true
  try {
    if (selectedStudentId.value) {
      recommendations.value = await getStudentRecommendations(selectedStudentId.value)
    } else {
      recommendations.value = await getRecommendations()
    }
  } catch (err) {
    console.error('加载推荐靶场失败:', err)
    recommendations.value = []
  } finally {
    loadingRecommendations.value = false
  }
}

// 渲染雷达图
function renderChart() {
  if (!chartRef.value || !skillProfile.value) return

  if (!chartInstance) {
    chartInstance = echarts.init(chartRef.value)
  }

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
        areaStyle: { color: `${primaryColor}33` },
        lineStyle: { color: primaryColor, width: 2 },
        itemStyle: { color: primaryColor }
      }]
    }],
    tooltip: {
      trigger: 'item',
      formatter: (params: unknown) => {
        const data = (params as { data?: { value?: number[] } })?.data
        const values = Array.isArray(data?.value) ? data.value : []
        return dimensions.map((d, i) => `${d.name}: ${values[i] ?? 0}`).join('<br/>')
      }
    }
  }

  chartInstance.setOption(option)
}

// 跳转到靶场详情
function goToChallenge(id: string) {
  router.push(`/challenges/${id}`)
}

// 监听学员切换
watch(selectedStudentId, () => {
  loadSkillProfile()
  loadRecommendations()
})

const handleResize = () => chartInstance?.resize()

onMounted(() => {
  loadStudents()
  loadSkillProfile()
  loadRecommendations()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  chartInstance?.dispose()
  chartInstance = null
})
</script>
