import { onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useAuthStore } from '@/stores/auth'
import { resolveTeachingDashboardRouteName } from '@/utils/teachingWorkspaceRouting'

import { useInstances } from './useInstances'

export function useInstanceManagementPage() {
  const router = useRouter()
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

  function openDashboard(): void {
    router.push({ name: resolveTeachingDashboardRouteName(authStore.user?.role) })
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
    totalCount,
    runningCount,
    expiringSoonCount,
    instances,
    initialize,
    updateFilter,
    handlePageChange,
    handleDestroy,
    openDashboard,
  }
}
