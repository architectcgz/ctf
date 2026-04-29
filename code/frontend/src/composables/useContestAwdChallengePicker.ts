import { computed, reactive, watch, type Ref } from 'vue'
import { useDebounceFn } from '@vueuse/core'

import { listAdminAwdChallenges } from '@/api/admin'
import { ApiError } from '@/api/request'
import { usePagination } from '@/composables/usePagination'
import type {
  AdminAwdChallengeData,
  AWDDeploymentMode,
  AWDReadinessStatus,
  AWDServiceType,
} from '@/api/contracts'

export interface ContestAwdChallengePickerFilters {
  keyword: string
  serviceType: AWDServiceType | ''
  deploymentMode: AWDDeploymentMode | ''
  readinessStatus: AWDReadinessStatus | ''
}

interface UseContestAwdChallengePickerOptions {
  existingChallengeIds: Readonly<Ref<string[]>>
  pageSize?: number
}

function humanizeRequestError(error: unknown, fallback: string): string {
  if (error instanceof ApiError && error.message.trim()) {
    return error.message
  }
  if (error instanceof Error && error.message.trim()) {
    return error.message
  }
  return fallback
}

export function useContestAwdChallengePicker(options: UseContestAwdChallengePickerOptions) {
  const filters = reactive<ContestAwdChallengePickerFilters>({
    keyword: '',
    serviceType: '',
    deploymentMode: '',
    readinessStatus: '',
  })

  const pagination = usePagination<AdminAwdChallengeData>(({ page, page_size }) =>
    listAdminAwdChallenges({
      page,
      page_size,
      keyword: filters.keyword.trim() || undefined,
      service_type: filters.serviceType || undefined,
      deployment_mode: filters.deploymentMode || undefined,
      readiness_status: filters.readinessStatus || undefined,
      status: 'published',
    })
  )

  pagination.pageSize.value = Math.max(1, Math.floor(options.pageSize ?? pagination.pageSize.value))

  const selectableList = computed(() => {
    const existingIds = new Set(options.existingChallengeIds.value.map(String))
    return pagination.list.value.filter((item) => !existingIds.has(String(item.id)))
  })

  const loadError = computed(() =>
    pagination.error.value ? humanizeRequestError(pagination.error.value, 'AWD 题目加载失败') : ''
  )

  type DebouncedRefresh = ReturnType<typeof useDebounceFn> & {
    cancel?: () => void
  }
  const scheduleKeywordRefresh = useDebounceFn(() => {
    void pagination.changePage(1)
  }, 250) as DebouncedRefresh

  watch(
    () => filters.keyword,
    () => {
      scheduleKeywordRefresh()
    }
  )

  watch(
    [() => filters.serviceType, () => filters.deploymentMode, () => filters.readinessStatus],
    () => {
      void resetImmediateFilters()
    }
  )

  function setKeyword(value: string) {
    filters.keyword = value
  }

  async function resetImmediateFilters() {
    scheduleKeywordRefresh.cancel?.()
    await pagination.changePage(1)
  }

  function setServiceType(value: AWDServiceType | '') {
    filters.serviceType = value
  }

  function setDeploymentMode(value: AWDDeploymentMode | '') {
    filters.deploymentMode = value
  }

  function setReadinessStatus(value: AWDReadinessStatus | '') {
    filters.readinessStatus = value
  }

  async function reset() {
    filters.keyword = ''
    filters.serviceType = ''
    filters.deploymentMode = ''
    filters.readinessStatus = ''
    scheduleKeywordRefresh.cancel?.()
    await pagination.changePage(1)
  }

  return {
    filters,
    list: pagination.list,
    selectableList,
    total: pagination.total,
    page: pagination.page,
    pageSize: pagination.pageSize,
    loading: pagination.loading,
    loadError,
    refresh: pagination.refresh,
    changePage: pagination.changePage,
    setKeyword,
    setServiceType,
    setDeploymentMode,
    setReadinessStatus,
    reset,
  }
}
