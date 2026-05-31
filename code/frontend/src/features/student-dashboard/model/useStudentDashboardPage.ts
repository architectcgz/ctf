import { computed, onMounted } from 'vue'
import { useRouteNavigationTransport } from '@/shared/model/navigation/useRouteNavigationTransport'
import { useRouteQueryTabs } from '@/shared/model/navigation/useRouteQueryTabs'
import { useAuthStore } from '@/stores/auth'
import type { DashboardPanelKey, DashboardPanelTab } from './studentDashboardTypes'
import {
  studentDashboardCategoryChallengesRoute,
  studentDashboardChallengeDetailRoute,
  studentDashboardChallengesRoute,
  studentDashboardDifficultyChallengesRoute,
  studentDashboardRoleRedirectRoute,
  studentDashboardSkillProfileRoute,
} from './studentDashboardRoutes'
import { useStudentDashboardData } from './useStudentDashboardData'
import { useStudentDashboardPanelBindings } from './useStudentDashboardPanelBindings'

export type { DashboardPanelKey }

export function useStudentDashboardPage() {
  const authStore = useAuthStore()
  const { push, replace } = useRouteNavigationTransport()
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
    roleRedirectTarget,
    displayName,
    weakDimensions,
    categoryStats,
    difficultyStats,
    completionRate,
    highlightItems,
    loadDashboard,
  } = useStudentDashboardData({
    authStore,
  })

  const panelTabOrder = panelTabs.map((tab) => tab.key) as DashboardPanelKey[]
  const {
    activeTab: activePanel,
    setTabButtonRef,
    selectTab: switchPanel,
    handleTabKeydown,
  } = useRouteQueryTabs<DashboardPanelKey>({
    orderedTabs: panelTabOrder,
    defaultTab: 'overview',
  })

  function openChallenges(): void {
    void push(studentDashboardChallengesRoute)
  }

  function openCategoryChallenges(category: string): void {
    void push(studentDashboardCategoryChallengesRoute(category))
  }

  function openDifficultyChallenges(difficulty: string): void {
    void push(studentDashboardDifficultyChallengesRoute(difficulty))
  }

  function openSkillProfile(): void {
    void push(studentDashboardSkillProfileRoute)
  }

  function openChallenge(challengeId: string): void {
    void push(studentDashboardChallengeDetailRoute(challengeId))
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
    if (roleRedirectTarget.value) {
      void replace(studentDashboardRoleRedirectRoute(roleRedirectTarget.value))
      return
    }
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
