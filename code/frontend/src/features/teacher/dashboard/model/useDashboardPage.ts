import { computed } from 'vue'

import { useAuthStore } from '@/stores/auth'
import { teacherClassManagementRoute } from './teacherDashboardRoutes'
import { useTeacherOverviewData } from './useTeacherOverviewData'

export function useDashboardPage() {
  const authStore = useAuthStore()
  const overviewData = useTeacherOverviewData()
  const classManagementRoute = computed(() => teacherClassManagementRoute(authStore.user?.role))

  return {
    overview: overviewData.overview,
    error: overviewData.error,
    classManagementRoute,
    initialize: overviewData.initialize,
  }
}
