import { computed, ref, type Ref } from 'vue'
import { FileChartColumnIncreasing, Rocket, ShieldAlert } from 'lucide-vue-next'
import type { Router } from 'vue-router'

import { getMyProgress, getMyTimeline, getRecommendations, getSkillProfile } from '@/api/assessment'
import type {
  MyProgressData,
  RecommendationItem,
  SkillProfileData,
  TimelineEvent,
} from '@/api/contracts'
import type { useAuthStore } from '@/stores/auth'
import { getWeakDimensions } from '@/utils/skillProfile'
import type { DashboardHighlightItem } from './studentDashboardTypes'

interface UseStudentDashboardDataOptions {
  authStore: ReturnType<typeof useAuthStore>
  router: Router
}

export function useStudentDashboardData({
  authStore,
  router,
}: UseStudentDashboardDataOptions) {
  const loading = ref(false)
  const error = ref<string | null>(null)
  const progress = ref<MyProgressData | null>(null)
  const timeline = ref<TimelineEvent[]>([])
  const recommendations = ref<RecommendationItem[]>([])
  const skillProfile = ref<SkillProfileData | null>(null)

  const displayName = computed(() => authStore.user?.name || authStore.user?.username || '选手')
  const weakDimensions = computed(() => getWeakDimensions(skillProfile.value).slice(0, 3))
  const recommendationCount = computed(() => recommendations.value.length)
  const timelineCount = computed(() => timeline.value.length)
  const categoryStats = computed(() => progress.value?.category_stats ?? [])
  const difficultyStats = computed(() => progress.value?.difficulty_stats ?? [])
  const completionRate = computed(() => {
    const solved = progress.value?.total_solved ?? 0
    const total = progress.value?.category_stats?.reduce((sum, item) => sum + item.total, 0) ?? 0
    if (!total) return 0
    return Math.round((solved / total) * 100)
  })
  const highlightItems = computed<DashboardHighlightItem[]>(() => [
    {
      label: '训练完成率',
      value: `${completionRate.value}%`,
      description: '按当前分类题量计算的整体覆盖率',
      icon: Rocket,
    },
    {
      label: '推荐任务',
      value: `${recommendationCount.value} 项`,
      description:
        weakDimensions.value.length > 0
          ? `优先补强 ${weakDimensions.value.join(' / ')}`
          : '当前没有明显短板',
      icon: ShieldAlert,
    },
    {
      label: '近期动态',
      value: `${timelineCount.value} 条`,
      description: '最近实例和提交记录的浓缩视图',
      icon: FileChartColumnIncreasing,
    },
  ])

  async function loadDashboard(): Promise<void> {
    const role = authStore.user?.role
    if (role === 'teacher') {
      await router.replace({ name: 'TeacherDashboard' })
      return
    }
    if (role === 'admin') {
      await router.replace({ name: 'PlatformOverview' })
      return
    }

    loading.value = true
    error.value = null
    try {
      const [progressPayload, timelinePayload, recommendationPayload, profilePayload] =
        await Promise.all([getMyProgress(), getMyTimeline(), getRecommendations(), getSkillProfile()])

      progress.value = progressPayload
      timeline.value = timelinePayload.slice(0, 6)
      recommendations.value = recommendationPayload.slice(0, 4)
      skillProfile.value = profilePayload
    } catch (err) {
      console.error('加载学生仪表盘失败:', err)
      error.value = '加载仪表盘失败，请稍后重试'
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    error,
    progress,
    timeline,
    recommendations,
    skillProfile,
    displayName,
    weakDimensions,
    recommendationCount,
    timelineCount,
    categoryStats,
    difficultyStats,
    completionRate,
    highlightItems,
    loadDashboard,
  }
}
