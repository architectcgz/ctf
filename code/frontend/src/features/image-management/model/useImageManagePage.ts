import { computed, reactive, ref } from 'vue'
import { ArrowDownWideNarrow, Calendar, SortAsc } from 'lucide-vue-next'

import { getImages } from '@/api/admin/authoring'
import type { AdminImageListItem, ImageStatus } from '@/api/contracts'
import {
  createEmptyImageCreateForm,
  type ImageCreateForm,
} from '@/entities/image'
import type { WorkspaceDirectorySortOption } from '@/entities/workspace-directory'
import {
  buildImageStatusSummary,
  filterAndSortImages,
  formatImageDateTime,
  formatImageSize,
  getImageStatusLabel,
  getImageStatusStyle,
  type ImageSortConfig,
} from './imageManagePresentation'
import { usePagination } from '@/composables/usePagination'
import { useImageManageAutoRefresh } from './useImageManageAutoRefresh'
import { useImageManageMutations } from './useImageManageMutations'

type ImageSortOption = WorkspaceDirectorySortOption & {
  key: ImageSortConfig['key']
  order: 'asc' | 'desc'
}

const sortOptions: ImageSortOption[] = [
  { key: 'created_at', order: 'desc', label: '最近创建', icon: Calendar },
  { key: 'name', order: 'asc', label: '镜像名称 A-Z', icon: SortAsc },
  { key: 'tag', order: 'asc', label: '标签顺序', icon: ArrowDownWideNarrow },
]

export function useImageManagePage() {
  const dialogVisible = ref(false)
  const activeImage = ref<AdminImageListItem | null>(null)
  const keyword = ref('')
  const statusFilter = ref<ImageStatus | ''>('')
  const form = reactive<ImageCreateForm>(createEmptyImageCreateForm())

  const { list, total, page, pageSize, loading, changePage, refresh } = usePagination(getImages)
  const sortConfig = ref<ImageSortOption>(sortOptions[0]!)

  const hasActiveImages = computed(() =>
    list.value.some((row) => row.status === 'pending' || row.status === 'building')
  )
  const hasActiveFilters = computed(() => Boolean(keyword.value.trim() || statusFilter.value))
  const filteredRows = computed<AdminImageListItem[]>(() =>
    filterAndSortImages(list.value, keyword.value, statusFilter.value, sortConfig.value)
  )
  const filteredTotal = computed(() => filteredRows.value.length)
  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
  const { refreshHint } = useImageManageAutoRefresh({
    hasActiveImages,
    refresh,
  })
  const statusSummary = computed(() => buildImageStatusSummary(list.value))
  const { creating, handleCreate, handleDelete } = useImageManageMutations({
    form,
    dialogVisible,
    refresh,
  })

  function openDetail(row: AdminImageListItem): void {
    activeImage.value = row
  }

  function closeDetail(): void {
    activeImage.value = null
  }

  function getStatusLabel(status: ImageStatus): string {
    return getImageStatusLabel(status)
  }

  function getStatusStyle(status: ImageStatus): Record<string, string> {
    return getImageStatusStyle(status)
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

  function resetFilters(): void {
    keyword.value = ''
    statusFilter.value = ''
  }

  return {
    activeImage,
    changePage,
    closeDetail,
    creating,
    dialogVisible,
    filteredRows,
    filteredTotal,
    form,
    formatDateTime: formatImageDateTime,
    formatSize: formatImageSize,
    getStatusLabel,
    getStatusStyle,
    handleCreate,
    handleDelete,
    hasActiveFilters,
    keyword,
    list,
    loading,
    openDetail,
    page,
    refresh,
    refreshHint,
    resetFilters,
    selectedSortLabel: computed(() => sortConfig.value.label),
    setSort,
    sortOptions,
    statusFilter,
    statusSummary,
    total,
    totalPages,
  }
}
