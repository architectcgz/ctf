import { ArrowDownWideNarrow, Calendar, UserRound } from 'lucide-vue-next'
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getAuditLogs } from '@/api/admin'
import type { AuditLogItem } from '@/api/contracts'
import type { WorkspaceDirectorySortOption } from '@/components/common/WorkspaceDirectoryToolbar.vue'
import { useAbortController } from '@/composables/useAbortController'

type AuditSortKey = 'created_at' | 'action' | 'actor'
type AuditSortOption = WorkspaceDirectorySortOption & {
  key: AuditSortKey
  order: 'asc' | 'desc'
}

const sortOptions: AuditSortOption[] = [
  { key: 'created_at', order: 'desc', label: '最近操作', icon: Calendar },
  { key: 'action', order: 'asc', label: '动作顺序', icon: ArrowDownWideNarrow },
  { key: 'actor', order: 'asc', label: '执行人顺序', icon: UserRound },
]

export function useAuditLogPage() {
  const route = useRoute()
  const router = useRouter()

  const filters = reactive({
    action: '',
    resource_type: '',
    actor_user_id: '',
  })

  const list = ref<AuditLogItem[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(20)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const keyword = ref('')
  const activeActorLog = ref<AuditLogItem | null>(null)
  const autoApplyReady = ref(false)
  const sortConfig = ref<AuditSortOption>(sortOptions[0]!)

  const { createController, abort } = useAbortController()
  let textFilterTimer: ReturnType<typeof setTimeout> | null = null
  let suppressAutoApply = false
  let latestLogsRequestId = 0

  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
  const hasActiveFilters = computed(() =>
    Boolean(keyword.value.trim() || filters.action || filters.resource_type || filters.actor_user_id)
  )
  const filteredRows = computed<AuditLogItem[]>(() => {
    const nextRows = [...list.value]

    nextRows.sort((left, right) => {
      switch (sortConfig.value.key) {
        case 'action': {
          const delta = left.action.localeCompare(right.action, 'zh-CN')
          return sortConfig.value.order === 'asc' ? delta : -delta
        }
        case 'actor': {
          const delta = (left.actor_username || '').localeCompare(right.actor_username || '', 'zh-CN')
          return sortConfig.value.order === 'asc' ? delta : -delta
        }
        case 'created_at':
        default: {
          const delta = new Date(left.created_at).getTime() - new Date(right.created_at).getTime()
          return sortConfig.value.order === 'asc' ? delta : -delta
        }
      }
    })

    return nextRows
  })

  function formatDate(value: string): string {
    return new Date(value).toLocaleString('zh-CN')
  }

  function normalizeQueryValue(value: unknown): string {
    if (Array.isArray(value)) return typeof value[0] === 'string' ? value[0] : ''
    return typeof value === 'string' ? value : ''
  }

  function hydrateFromRoute(): void {
    filters.action = normalizeQueryValue(route.query.action)
    filters.resource_type = normalizeQueryValue(route.query.resource_type)
    filters.actor_user_id = normalizeQueryValue(route.query.actor_user_id)

    const nextPage = Number.parseInt(normalizeQueryValue(route.query.page), 10)
    page.value = Number.isFinite(nextPage) && nextPage > 0 ? nextPage : 1
  }

  async function syncRouteQuery(): Promise<void> {
    const query: Record<string, string> = {}

    if (filters.action) query.action = filters.action
    if (filters.resource_type) query.resource_type = filters.resource_type
    if (filters.actor_user_id) query.actor_user_id = filters.actor_user_id
    if (page.value > 1) query.page = String(page.value)

    await router.replace({ name: 'AuditLog', query })
  }

  function detailPreview(detail: Record<string, unknown> | undefined): string {
    if (!detail) return '-'
    return Object.entries(detail)
      .slice(0, 3)
      .map(([key, value]) => `${key}: ${String(value)}`)
      .join(' / ')
  }

  function actorDisplayName(item: AuditLogItem): string {
    return item.actor_username || '未知执行人'
  }

  function resourceDisplayName(item: AuditLogItem): string {
    return item.resource_id ? `${item.resource_type} #${item.resource_id}` : item.resource_type
  }

  function openActorDetail(item: AuditLogItem): void {
    activeActorLog.value = item
  }

  function closeActorDetail(): void {
    activeActorLog.value = null
  }

  async function loadLogs(): Promise<void> {
    const requestId = ++latestLogsRequestId
    const controller = createController()
    loading.value = true
    error.value = null
    try {
      const payload = await getAuditLogs(
        {
          page: page.value,
          page_size: pageSize.value,
          action: filters.action || undefined,
          resource_type: filters.resource_type || undefined,
          actor_user_id: filters.actor_user_id ? Number(filters.actor_user_id) : undefined,
        },
        {
          signal: controller.signal,
        }
      )
      if (requestId !== latestLogsRequestId) {
        return
      }
      list.value = payload.list
      total.value = payload.total
      page.value = payload.page
      pageSize.value = payload.page_size
    } catch (err) {
      if (requestId !== latestLogsRequestId) {
        return
      }
      if (
        err &&
        typeof err === 'object' &&
        'code' in err &&
        (err as { code?: unknown }).code === 'ERR_CANCELED'
      ) {
        return
      }
      console.error('加载审计日志失败:', err)
      error.value = '加载审计日志失败，请稍后重试'
    } finally {
      if (requestId !== latestLogsRequestId) {
        return
      }
      loading.value = false
    }
  }

  async function applyFilters(): Promise<void> {
    page.value = 1
    await syncRouteQuery()
    await loadLogs()
  }

  function clearTextFilterTimer(): void {
    if (textFilterTimer !== null) {
      clearTimeout(textFilterTimer)
      textFilterTimer = null
    }
  }

  function scheduleTextFilterApply(): void {
    if (!autoApplyReady.value || suppressAutoApply) {
      return
    }
    clearTextFilterTimer()
    textFilterTimer = setTimeout(() => {
      void applyFilters()
    }, 250)
  }

  async function changePage(next: number): Promise<void> {
    page.value = Math.max(1, Math.floor(next))
    await syncRouteQuery()
    await loadLogs()
  }

  function setSort(option: WorkspaceDirectorySortOption): void {
    const matchedOption =
      sortOptions.find((item) => item.key === option.key && item.label === option.label) ??
      sortOptions[0]

    if (!matchedOption) {
      return
    }

    sortConfig.value = matchedOption
  }

  async function resetFilters(): Promise<void> {
    clearTextFilterTimer()
    suppressAutoApply = true
    keyword.value = ''
    filters.action = ''
    filters.resource_type = ''
    filters.actor_user_id = ''
    suppressAutoApply = false
    page.value = 1
    await syncRouteQuery()
    await loadLogs()
  }

  onMounted(() => {
    hydrateFromRoute()
    autoApplyReady.value = true
    void loadLogs()
  })

  onBeforeUnmount(() => {
    clearTextFilterTimer()
    abort()
  })

  watch(
    () => filters.action,
    () => {
      if (!autoApplyReady.value || suppressAutoApply) {
        return
      }
      clearTextFilterTimer()
      void applyFilters()
    }
  )

  watch(
    () => [filters.resource_type, filters.actor_user_id] as const,
    () => {
      scheduleTextFilterApply()
    }
  )

  return {
    activeActorLog,
    actorDisplayName,
    changePage,
    closeActorDetail,
    detailPreview,
    error,
    filteredRows,
    filters,
    formatDate,
    hasActiveFilters,
    keyword,
    list,
    loadLogs,
    loading,
    openActorDetail,
    page,
    resetFilters,
    resourceDisplayName,
    selectedSortLabel: computed(() => sortConfig.value.label),
    setSort,
    sortOptions,
    total,
    totalPages,
  }
}
