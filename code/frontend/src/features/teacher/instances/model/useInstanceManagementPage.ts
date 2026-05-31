import { onMounted } from 'vue'

import { useAuthStore } from '@/stores/auth'
import { teacherInstanceDashboardRoute } from './teacherInstanceManagementRoutes'

import { useInstances } from './useInstances'

export function useInstanceManagementPage() {
  const authStore = useAuthStore()

  const {
    classes,
    instances,
    page,
    filters,
    loadingClasses,
    loadingInstances,
    destroyingId,
    error,
    isAdmin,
    totalCount,
    runningCount,
    expiringSoonCount,
    totalPages,
    initialize,
    updateFilter,
    loadInstances,
    removeInstance,
  } = useInstances()
  const dashboardRoute = teacherInstanceDashboardRoute(authStore.user?.role)

  function handlePageChange(nextPage: number): void {
    const normalizedPage = Math.max(1, Math.floor(nextPage))
    if (normalizedPage === page.value || normalizedPage > totalPages.value) {
      return
    }
    page.value = normalizedPage
    void loadInstances()
  }

  async function handleDestroy(id: string): Promise<void> {
    await removeInstance(id)
  }

  onMounted(() => {
    void initialize()
  })

  return {
    classes,
    filters,
    page,
    totalPages,
    loadingClasses,
    loadingInstances,
    destroyingId,
    error,
    isAdmin,
    dashboardRoute,
    totalCount,
    runningCount,
    expiringSoonCount,
    instances,
    initialize,
    updateFilter,
    handlePageChange,
    handleDestroy,
  }
}
