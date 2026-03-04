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
          {{ student.name }} ({{ student.username }})
        </option>
      </select>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
    </div>

    <!-- 空状态 -->
    <div v-else-if="!skillProfile" class="bg-surface rounded-lg p-8 text-center border border-border">
      <p class="text-text-secondary mb-4">暂无能力画像数据</p>
      <p class="text-sm text-text-tertiary">完成更多靶场挑战后，系统将为你生成能力画像</p>
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
            class="flex items-start justify-between p-4 bg-background rounded-lg border border-border hover:border-primary transition-colors cursor-pointer"
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
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import * as echarts from 'echarts'
import type { EChartsOption } from 'echarts'

import { getRecommendations, getSkillProfile } from '@/api/assessment'
import { getStudentRecommendations, getStudentSkillProfile } from '@/api/teacher'
import type { RecommendationItem, SkillProfileData, TeacherStudentItem } from '@/api/contracts'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const router = useRouter()

const isTeacher = computed(() => authStore.isTeacher)
const selectedStudentId = ref('')
const students = ref<TeacherStudentItem[]>([])

const loading = ref(false)
const skillProfile = ref<SkillProfileData | null>(null)
const chartRef = ref<HTMLElement>()
let chartInstance: echarts.ECharts | null = null

const loadingRecommendations = ref(false)
const recommendations = ref<RecommendationItem[]>([])

// 薄弱项（分数低于 60）
const weakDimensions = computed(() => {
  if (!skillProfile.value) return []
  return skillProfile.value.dimensions
    .filter(d => d.value < 60)
    .map(d => d.name)
})

// 加载能力画像
async function loadSkillProfile() {
  loading.value = true
  try {
    if (selectedStudentId.value) {
      skillProfile.value = await getStudentSkillProfile(selectedStudentId.value)
    } else {
      skillProfile.value = await getSkillProfile()
    }
    renderChart()
  } catch (error) {
    console.error('加载能力画像失败:', error)
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
  } catch (error) {
    console.error('加载推荐靶场失败:', error)
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

  const dimensions = skillProfile.value.dimensions
  const option: EChartsOption = {
    radar: {
      indicator: dimensions.map(d => ({ name: d.name, max: 100 })),
      radius: '65%',
      splitNumber: 4,
      axisName: { color: '#666' },
      splitLine: { lineStyle: { color: '#e5e7eb' } },
      splitArea: { show: false },
      axisLine: { lineStyle: { color: '#e5e7eb' } }
    },
    series: [{
      type: 'radar',
      data: [{
        value: dimensions.map(d => d.value),
        name: '能力值',
        areaStyle: { color: 'rgba(59, 130, 246, 0.2)' },
        lineStyle: { color: '#3b82f6', width: 2 },
        itemStyle: { color: '#3b82f6' }
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

// 难度样式
function difficultyClass(difficulty: string) {
  const map: Record<string, string> = {
    beginner: 'bg-green-100 text-green-700',
    easy: 'bg-blue-100 text-blue-700',
    medium: 'bg-yellow-100 text-yellow-700',
    hard: 'bg-orange-100 text-orange-700',
    hell: 'bg-red-100 text-red-700'
  }
  return map[difficulty] || 'bg-gray-100 text-gray-700'
}

function difficultyLabel(difficulty: string) {
  const map: Record<string, string> = {
    beginner: '入门',
    easy: '简单',
    medium: '中等',
    hard: '困难',
    hell: '地狱'
  }
  return map[difficulty] || difficulty
}

// 跳转到靶场详情
function goToChallenge(id: string) {
  router.push(`/challenges/${id}`)
}

// 格式化日期
function formatDate(isoString: string) {
  return new Date(isoString).toLocaleString('zh-CN')
}

// 监听学员切换
watch(selectedStudentId, () => {
  loadSkillProfile()
  loadRecommendations()
})

onMounted(() => {
  loadSkillProfile()
  loadRecommendations()

  // 响应式调整图表大小
  window.addEventListener('resize', () => chartInstance?.resize())
})
</script>
