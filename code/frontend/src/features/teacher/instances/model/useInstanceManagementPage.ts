import { onMounted } from 'vue'

import { confirmDestructiveAction } from '@/shared/model/common/useDestructiveConfirm'
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
    const confirmed = await confirmDestructiveAction({
      title: '确认销毁实例',
      message: '确定要销毁该实例吗？此操作不可恢复。',
      confirmButtonText: '确认销毁',
      cancelButtonText: '取消',
    })
    if (!confirmed) {
      return
    }
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
