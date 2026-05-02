import { computed, onMounted, ref, type Component } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { FileChartColumnIncreasing, Rocket, ShieldAlert } from 'lucide-vue-next'

import { getMyProgress, getMyTimeline, getRecommendations, getSkillProfile } from '@/api/assessment'
import type {
  MyProgressData,
  RecommendationItem,
  SkillProfileData,
  TimelineEvent,
} from '@/api/contracts'
import StudentCategoryProgressPage from '@/components/dashboard/student/StudentCategoryProgressPage.vue'
import StudentDifficultyPage from '@/components/dashboard/student/StudentDifficultyPage.vue'
import StudentOverviewPage from '@/components/dashboard/student/StudentOverviewPage.vue'
import StudentRecommendationPage from '@/components/dashboard/student/StudentRecommendationPage.vue'
import StudentTimelinePage from '@/components/dashboard/student/StudentTimelinePage.vue'
import type { DashboardHighlightItem } from '@/components/dashboard/student/types'
import { useRouteQueryTabs } from '@/composables/useRouteQueryTabs'
import { useAuthStore } from '@/stores/auth'
import { getWeakDimensions } from '@/utils/skillProfile'

type DashboardPanelKey = 'overview' | 'category' | 'recommendation' | 'timeline' | 'difficulty'

export function useStudentDashboardPage() {
  const authStore = useAuthStore()
  const route = useRoute()
  const router = useRouter()

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
  const panelTabs: Array<{
    key: DashboardPanelKey
    label: string
    panelId: string
    tabId: string
  }> = [
    { key: 'overview', label: '训练总览', panelId: 'dashboard-panel-overview', tabId: 'dashboard-tab-overview' },
    {
      key: 'recommendation',
      label: '训练队列',
      panelId: 'dashboard-panel-recommendation',
      tabId: 'dashboard-tab-recommendation',
    },
    { key: 'category', label: '分类补强', panelId: 'dashboard-panel-category', tabId: 'dashboard-tab-category' },
    { key: 'timeline', label: '训练记录', panelId: 'dashboard-panel-timeline', tabId: 'dashboard-tab-timeline' },
    { key: 'difficulty', label: '强度推进', panelId: 'dashboard-panel-difficulty', tabId: 'dashboard-tab-difficulty' },
  ]

  const panelTabOrder = panelTabs.map((tab) => tab.key) as DashboardPanelKey[]
  const {
    activeTab: activePanel,
    setTabButtonRef,
    selectTab: switchPanel,
    handleTabKeydown,
  } = useRouteQueryTabs<DashboardPanelKey>({
    route,
    router,
    orderedTabs: panelTabOrder,
    defaultTab: 'overview',
  })

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

  function openChallenges(): void {
    router.push({ name: 'Challenges' })
  }

  function openCategoryChallenges(category: string): void {
    router.push({ name: 'Challenges', query: { category } })
  }

  function openDifficultyChallenges(difficulty: string): void {
    router.push({ name: 'Challenges', query: { difficulty } })
  }

  function openSkillProfile(): void {
    router.push({ name: 'SkillProfile' })
  }

  function openChallenge(challengeId: string): void {
    router.push(`/challenges/${challengeId}`)
  }

  function resolveDashboardPanelComponent(panelKey: DashboardPanelKey): Component {
    switch (panelKey) {
      case 'overview':
        return StudentOverviewPage
      case 'recommendation':
        return StudentRecommendationPage
      case 'category':
        return StudentCategoryProgressPage
      case 'timeline':
        return StudentTimelinePage
      case 'difficulty':
        return StudentDifficultyPage
    }
  }

  function resolveDashboardPanelBindings(panelKey: DashboardPanelKey): Record<string, unknown> {
    switch (panelKey) {
      case 'overview':
        return {
          embedded: true,
          displayName: displayName.value,
          className: authStore.user?.class_name,
          progress: progress.value,
          completionRate: completionRate.value,
          highlightItems: highlightItems.value,
          recommendations: recommendations.value,
          timeline: timeline.value,
          weakDimensions: weakDimensions.value,
          skillDimensions: skillProfile.value?.dimensions ?? [],
          onOpenChallenge: openChallenge,
          onOpenChallenges: openChallenges,
          onOpenSkillProfile: openSkillProfile,
        }
      case 'recommendation':
        return {
          embedded: true,
          weakDimensions: weakDimensions.value,
          recommendations: recommendations.value,
          onOpenChallenge: openChallenge,
          onOpenChallenges: openChallenges,
          onOpenSkillProfile: openSkillProfile,
        }
      case 'category':
        return {
          embedded: true,
          categoryStats: categoryStats.value,
          completionRate: completionRate.value,
          onOpenChallenges: openChallenges,
          onOpenCategoryChallenges: openCategoryChallenges,
          onOpenSkillProfile: openSkillProfile,
        }
      case 'timeline':
        return {
          embedded: true,
          timeline: timeline.value,
        }
      case 'difficulty':
        return {
          embedded: true,
          difficultyStats: difficultyStats.value,
          onOpenChallenges: openChallenges,
          onOpenDifficultyChallenges: openDifficultyChallenges,
        }
    }
  }

  onMounted(() => {
    void loadDashboard()
  })

  return {
    loading,
    error,
    progress,
    panelTabs,
    activePanel,
    setTabButtonRef,
    switchPanel,
    handleTabKeydown,
    loadDashboard,
    resolveDashboardPanelComponent,
    resolveDashboardPanelBindings,
  }
}
