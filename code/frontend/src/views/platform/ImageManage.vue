<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { ArrowDownWideNarrow, Calendar, SortAsc } from 'lucide-vue-next'

import { createImage, deleteImage, getImages, type AdminImagePayload } from '@/api/admin'
import type { AdminImageListItem, ImageStatus } from '@/api/contracts'
import ImageCreateModal from '@/components/platform/images/ImageCreateModal.vue'
import ImageDetailModal from '@/components/platform/images/ImageDetailModal.vue'
import ImageDirectoryPanel from '@/components/platform/images/ImageDirectoryPanel.vue'
import ImageManageHeroPanel from '@/components/platform/images/ImageManageHeroPanel.vue'
import {
  type WorkspaceDirectorySortOption,
} from '@/components/common/WorkspaceDirectoryToolbar.vue'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'

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

const toast = useToast()
const dialogVisible = ref(false)
const activeImage = ref<AdminImageListItem | null>(null)
const creating = ref(false)
const keyword = ref('')
const statusFilter = ref<ImageStatus | ''>('')
const form = reactive<AdminImagePayload>({
  name: '',
  tag: '',
  description: '',
})

const { list, total, page, pageSize, loading, changePage, refresh } = usePagination(getImages)

const sortOptions: ImageSortOption[] = [
  { key: 'created_at', order: 'desc', label: '最近创建', icon: Calendar },
  { key: 'name', order: 'asc', label: '镜像名称 A-Z', icon: SortAsc },
  { key: 'tag', order: 'asc', label: '标签顺序', icon: ArrowDownWideNarrow },
]
const sortConfig = ref<ImageSortOption>(sortOptions[0]!)

let pollTimer: number | null = null

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

const refreshHint = computed(() =>
  hasActiveImages.value ? '构建中镜像会每 10 秒自动刷新' : '当前无进行中镜像，可手动刷新'
)

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

function stopPolling() {
  if (pollTimer !== null) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

function startPolling() {
  if (pollTimer !== null) return
  pollTimer = window.setInterval(() => {
    void refresh()
  }, 10000)
}

async function handleCreate() {
  if (creating.value) {
    return
  }

  if (!form.name || !form.tag) {
    toast.error('请填写完整信息')
    return
  }

  creating.value = true
  try {
    await createImage(form)
    toast.success('镜像创建成功')
    dialogVisible.value = false
    Object.assign(form, { name: '', tag: '', description: '' })
    await refresh()
  } catch {
    toast.error('创建失败')
  } finally {
    creating.value = false
  }
}

async function handleDelete(id: string) {
  const confirmed = await confirmDestructiveAction({
    message: '确定要删除此镜像吗？',
  })
  if (!confirmed) {
    return
  }

  try {
    await deleteImage(id)
    toast.success('删除成功')
    await refresh()
  } catch (error) {
    const message = error instanceof Error && error.message.trim() ? error.message : '删除失败'
    toast.error(message)
  }
}

async function handleManualRefresh() {
  await refresh()
}

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

watch(
  hasActiveImages,
  (active) => {
    if (active) {
      startPolling()
      return
    }
    stopPolling()
  },
  { immediate: true }
)

onMounted(() => {
  void refresh()
})

onUnmounted(() => {
  stopPolling()
})
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-admin journal-notes-rail journal-hero flex min-h-full flex-1 flex-col"
  >
    <main class="content-pane">
      <ImageManageHeroPanel
        :loading="loading"
        :refresh-hint="refreshHint"
        :status-summary="statusSummary"
        @refresh="void handleManualRefresh()"
        @create="dialogVisible = true"
      />

      <!--
        class="image-board workspace-directory-section"
        class="image-list workspace-directory-list"
        class="admin-pagination workspace-directory-pagination"
        PlatformPaginationControls
      -->
      <ImageDirectoryPanel
        :list="list"
        :rows="filteredRows"
        :total="total"
        :filtered-total="filteredTotal"
        :page="page"
        :total-pages="totalPages"
        :loading="loading"
        :keyword="keyword"
        :status-filter="statusFilter"
        :has-active-filters="hasActiveFilters"
        :sort-options="sortOptions"
        :selected-sort-label="sortConfig.label"
        :get-status-label="getStatusLabel"
        :get-status-style="getStatusStyle"
        :format-date-time="formatDateTime"
        @update:keyword="keyword = $event"
        @update:status-filter="statusFilter = $event"
        @select-sort="setSort"
        @reset-filters="resetFilters"
        @open-detail="openDetail"
        @delete-image="handleDelete"
        @change-page="void changePage($event)"
      />
    </main>

    <ImageDetailModal
      :open="!!activeImage"
      :image="activeImage"
      :format-size="formatSize"
      :format-date-time="formatDateTime"
      :get-status-label="getStatusLabel"
      :get-status-style="getStatusStyle"
      @close="closeDetail"
      @update:open="!$event && closeDetail()"
    />

    <ImageCreateModal
      :open="dialogVisible"
      :creating="creating"
      :form="form"
      @close="dialogVisible = false"
      @update:open="dialogVisible = $event"
      @update:name="form.name = $event"
      @update:tag="form.tag = $event"
      @update:description="form.description = $event"
      @submit="handleCreate"
    />
  </section>
</template>

<style scoped>
.journal-shell {
  --admin-summary-grid-columns: repeat(2, minmax(0, 1fr));
  --admin-control-border: color-mix(in srgb, var(--journal-border) 78%, transparent);
  --journal-divider-border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  --journal-shell-hero-radial-strength: 7%;
  --journal-shell-hero-radial-size: 22rem;
  --journal-shell-hero-end: var(--journal-surface);
  --journal-shell-hero-shadow: 0 22px 50px var(--color-shadow-soft);
  --journal-shell-dark-ink: var(--color-text-primary);
  --journal-shell-dark-accent: var(--color-primary-hover);
  --journal-shell-dark-surface: color-mix(
    in srgb,
    var(--color-bg-surface) 92%,
    var(--color-bg-base)
  );
  --journal-shell-dark-surface-subtle: color-mix(
    in srgb,
    var(--color-bg-surface) 78%,
    var(--color-bg-base)
  );
  --journal-shell-dark-hero-radial-strength: 10%;
  --journal-shell-dark-hero-top: color-mix(
    in srgb,
    var(--journal-surface) 97%,
    var(--color-bg-base)
  );
  --journal-shell-dark-hero-end: color-mix(
    in srgb,
    var(--journal-surface-subtle) 95%,
    var(--color-bg-base)
  );
}
</style>
