import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useAuthStore } from '@/stores/auth'
import { DEFAULT_PAGE_SIZE } from '@/utils/constants'
import { resolveTeachingDashboardRouteName } from '@/utils/teachingWorkspaceRouting'

import { useTeacherInstances } from './useTeacherInstances'

export function useTeacherInstanceManagementPage() {
  const router = useRouter()
  const authStore = useAuthStore()
  const page = ref(1)
  const pageSize = ref(DEFAULT_PAGE_SIZE)

  const {
    classes,
    instances,
    filters,
    loadingClasses,
    loadingInstances,
    destroyingId,
    error,
    isAdmin,
    totalCount,
    runningCount,
    expiringSoonCount,
    initialize,
    updateFilter,
    removeInstance,
  } = useTeacherInstances()

  const totalPages = computed(() =>
    Math.max(1, Math.ceil(totalCount.value / Math.max(pageSize.value, 1)))
  )
  const paginatedInstances = computed(() => {
    const start = (page.value - 1) * pageSize.value
    return instances.value.slice(start, start + pageSize.value)
  })

  function handlePageChange(nextPage: number): void {
    const normalizedPage = Math.max(1, Math.floor(nextPage))
    if (normalizedPage === page.value || normalizedPage > totalPages.value) {
      return
    }
    page.value = normalizedPage
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

  watch(
    () => [filters.className, filters.keyword, filters.studentNo],
    () => {
      page.value = 1
    }
  )

  watch(totalCount, () => {
    if (page.value > totalPages.value) {
      page.value = totalPages.value
    }
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
    paginatedInstances,
    initialize,
    updateFilter,
    handlePageChange,
    handleDestroy,
    openDashboard,
  }
}
