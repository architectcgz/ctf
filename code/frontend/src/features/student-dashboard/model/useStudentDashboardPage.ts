import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useRouteQueryTabs } from '@/composables/useRouteQueryTabs'
import { useAuthStore } from '@/stores/auth'
import type { DashboardPanelKey, DashboardPanelTab } from './studentDashboardTypes'
import { useStudentDashboardData } from './useStudentDashboardData'
import { useStudentDashboardPanelBindings } from './useStudentDashboardPanelBindings'

export type { DashboardPanelKey }

export function useStudentDashboardPage() {
  const authStore = useAuthStore()
  const route = useRoute()
  const router = useRouter()
  const panelTabs: DashboardPanelTab[] = [
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
  const {
    loading,
    error,
    progress,
    timeline,
    recommendations,
    skillProfile,
    displayName,
    weakDimensions,
    categoryStats,
    difficultyStats,
    completionRate,
    highlightItems,
    loadDashboard,
  } = useStudentDashboardData({
    authStore,
    router,
  })

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
  const className = computed(() => authStore.user?.class_name)
  const { resolveDashboardPanelBindings } = useStudentDashboardPanelBindings({
    className,
    progress,
    timeline,
    recommendations,
    skillProfile,
    displayName,
    weakDimensions,
    categoryStats,
    difficultyStats,
    completionRate,
    highlightItems,
    openChallenge,
    openChallenges,
    openCategoryChallenges,
    openDifficultyChallenges,
    openSkillProfile,
  })

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
    resolveDashboardPanelBindings,
  }
}
