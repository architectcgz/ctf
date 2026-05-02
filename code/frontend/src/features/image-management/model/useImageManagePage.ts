import { computed, reactive, ref } from 'vue'
import { ArrowDownWideNarrow, Calendar, SortAsc } from 'lucide-vue-next'

import { getImages } from '@/api/admin/authoring'
import type { AdminImageListItem, ImageStatus } from '@/api/contracts'
import {
  createEmptyImageCreateForm,
  type ImageCreateForm,
} from '@/entities/image'
import type { WorkspaceDirectorySortOption } from '@/entities/workspace-directory'
import { usePagination } from '@/composables/usePagination'
import { useImageManageAutoRefresh } from './useImageManageAutoRefresh'
import { useImageManageMutations } from './useImageManageMutations'

type ImageSortKey = 'created_at' | 'name' | 'tag'
type ImageSortOption = WorkspaceDirectorySortOption & {
  key: ImageSortKey
  order: 'asc' | 'desc'
}
type ImageStatusSummaryItem = {
  key: string
  label: string
  value: number
  tone: 'success' | 'warning' | 'danger' | 'muted'
}

const imageStatusMeta: Record<
  ImageStatus,
  { label: string; color: string; backgroundColor: string }
> = {
  pending: {
    label: '等待中',
    color: 'color-mix(in srgb, var(--journal-muted) 84%, var(--journal-ink))',
    backgroundColor: 'color-mix(in srgb, var(--journal-muted) 14%, transparent)',
  },
  building: {
    label: '构建中',
    color: 'var(--color-warning)',
    backgroundColor: 'color-mix(in srgb, var(--color-warning) 14%, transparent)',
  },
  available: {
    label: '可用',
    color: 'var(--color-success)',
    backgroundColor: 'color-mix(in srgb, var(--color-success) 14%, transparent)',
  },
  failed: {
    label: '失败',
    color: 'var(--color-danger)',
    backgroundColor: 'color-mix(in srgb, var(--color-danger) 14%, transparent)',
  },
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
  const filteredRows = computed<AdminImageListItem[]>(() => {
    const normalizedKeyword = keyword.value.trim().toLowerCase()
    const nextRows = list.value.filter((row) => {
      const matchesKeyword =
        !normalizedKeyword ||
        row.name.toLowerCase().includes(normalizedKeyword) ||
        row.tag.toLowerCase().includes(normalizedKeyword) ||
        (row.description || '').toLowerCase().includes(normalizedKeyword)
      const matchesStatus = !statusFilter.value || row.status === statusFilter.value

      return matchesKeyword && matchesStatus
    })

    const sortedRows = [...nextRows]
    sortedRows.sort((left, right) => {
      switch (sortConfig.value.key) {
        case 'name': {
          const delta = left.name.localeCompare(right.name, 'zh-CN')
          return sortConfig.value.order === 'asc' ? delta : -delta
        }
        case 'tag': {
          const delta = left.tag.localeCompare(right.tag, 'zh-CN', { numeric: true })
          return sortConfig.value.order === 'asc' ? delta : -delta
        }
        case 'created_at':
        default: {
          const delta = new Date(left.created_at).getTime() - new Date(right.created_at).getTime()
          return sortConfig.value.order === 'asc' ? delta : -delta
        }
      }
    })

    return sortedRows
  })
  const filteredTotal = computed(() => filteredRows.value.length)
  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
  const { refreshHint } = useImageManageAutoRefresh({
    hasActiveImages,
    refresh,
  })
  const statusSummary = computed<ImageStatusSummaryItem[]>(() => {
    const counts = {
      available: 0,
      pending: 0,
      building: 0,
      failed: 0,
    }

    for (const row of list.value) {
      counts[row.status] += 1
    }

    const summary: ImageStatusSummaryItem[] = []

    if (counts.building > 0) {
      summary.push({ key: 'building', label: '构建中', value: counts.building, tone: 'warning' })
    }
    if (counts.pending > 0) {
      summary.push({ key: 'pending', label: '等待中', value: counts.pending, tone: 'muted' })
    }
    if (counts.failed > 0) {
      summary.push({ key: 'failed', label: '失败', value: counts.failed, tone: 'danger' })
    }

    return summary
  })
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

  function formatSize(bytes?: number): string {
    if (!bytes) return '未知大小'
    const units = ['B', 'KB', 'MB', 'GB', 'TB']
    let size = bytes
    let unitIndex = 0
    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024
      unitIndex++
    }
    return `${size.toFixed(2)} ${units[unitIndex]}`
  }

  function getStatusLabel(status: ImageStatus): string {
    return imageStatusMeta[status].label
  }

  function getStatusStyle(status: ImageStatus): Record<string, string> {
    const meta = imageStatusMeta[status]
    return {
      backgroundColor: meta.backgroundColor,
      color: meta.color,
    }
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

  function formatDateTime(value: string): string {
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return '--'
    return new Intl.DateTimeFormat('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    }).format(date)
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
    formatDateTime,
    formatSize,
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
