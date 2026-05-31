import { computed } from 'vue'

import { useRouteQueryTransport } from '@/shared/model/navigation/useRouteQueryTransport'
import { useAuthStore } from '@/stores/auth'
import {
  buildTeacherDashboardPanelQuery,
  resolveTeacherDashboardPanel,
  type TeacherDashboardPanelKey,
} from './teacherDashboardPanelRoute'
import { teacherClassManagementRoute } from './teacherDashboardRoutes'
import { useTeacherOverviewData } from './useTeacherOverviewData'

export function useDashboardPage() {
  const { query, replaceQuery } = useRouteQueryTransport()
  const authStore = useAuthStore()
  const overviewData = useTeacherOverviewData()
  const classManagementRoute = computed(() => teacherClassManagementRoute(authStore.user?.role))
  const activePanel = computed<TeacherDashboardPanelKey>(() =>
    resolveTeacherDashboardPanel(query.value.panel)
  )

  async function switchPanel(panel: TeacherDashboardPanelKey): Promise<void> {
    if (activePanel.value === panel) {
      return
    }
    await replaceQuery(buildTeacherDashboardPanelQuery(query.value, panel))
  }

  return {
    overview: overviewData.overview,
    error: overviewData.error,
    classManagementRoute,
    activePanel,
    switchPanel,
    initialize: overviewData.initialize,
  }
}
