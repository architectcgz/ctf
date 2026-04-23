<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { ArrowDownWideNarrow, Calendar, SortAsc } from 'lucide-vue-next'

import { createImage, deleteImage, getImages, type AdminImagePayload } from '@/api/admin'
import type { AdminImageListItem, ImageStatus } from '@/api/contracts'
import ImageCreateModal from '@/components/platform/images/ImageCreateModal.vue'
import ImageDetailModal from '@/components/platform/images/ImageDetailModal.vue'
import ImageDirectoryPanel from '@/components/platform/images/ImageDirectoryPanel.vue'
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

const statusSummary = computed(() => {
  const counts = {
    available: 0,
    pending: 0,
    building: 0,
    failed: 0,
  }

  for (const row of list.value) {
    counts[row.status] += 1
  }

  const summary = []

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
      <header class="image-header">
        <div class="image-header__intro">
          <div class="workspace-overline">
            Image Registry
          </div>
          <h1 class="image-title">
            镜像管理
          </h1>
          <p class="image-copy">
            集中查看镜像构建状态、描述与创建时间。
          </p>
        </div>

        <div class="image-header__side">
          <div
            class="image-header__actions"
            role="group"
            aria-label="镜像列表操作"
          >
            <button
              :disabled="loading"
              class="ui-btn ui-btn--ghost"
              data-testid="image-refresh-button"
              @click="handleManualRefresh"
            >
              立即刷新
            </button>
            <button
              class="ui-btn ui-btn--primary"
              @click="dialogVisible = true"
            >
              创建镜像
            </button>
          </div>
          <div
            class="image-status-strip"
            aria-label="镜像状态摘要"
          >
            <div
              v-if="statusSummary.length > 0"
              class="image-status-strip__row"
            >
              <div
                v-for="item in statusSummary"
                :key="item.key"
                :class="['image-status-pill', `image-status-pill--${item.tone}`]"
                data-testid="image-status-pill"
              >
                <span>{{ item.label }}</span>
                <strong>{{ item.value }}</strong>
              </div>
            </div>
            <div class="image-status-strip__note">
              {{ refreshHint }}
            </div>
          </div>
        </div>
      </header>

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

.image-header__actions > .ui-btn {
  --ui-btn-height: 2.45rem;
  --ui-btn-radius: 0.75rem;
  --ui-btn-padding: var(--space-2) var(--space-4);
  --ui-btn-font-size: var(--font-size-0-875);
  --ui-btn-font-weight: 600;
  --ui-btn-primary-background: var(--journal-accent);
  --ui-btn-primary-hover-background: color-mix(in srgb, var(--journal-accent) 88%, black);
  --ui-btn-primary-hover-shadow: 0 10px 24px color-mix(in srgb, var(--journal-accent) 18%, transparent);
  --ui-btn-ghost-color: var(--journal-ink);
  --ui-btn-ghost-hover-color: var(--journal-accent);
  --ui-btn-ghost-hover-background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
}

.image-header {
  display: grid;
  gap: var(--space-6);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.image-copy {
  max-width: 48rem;
}

.image-header__side {
  display: grid;
  gap: var(--space-3);
  justify-items: start;
}

.image-header__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
}

.image-status-strip {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3) var(--space-4);
}

.image-status-strip__row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2-5);
}

.image-status-strip__note {
  font-size: var(--font-size-0-82);
  line-height: 1.6;
  color: var(--journal-muted);
}

.image-status-pill {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  min-height: 2.25rem;
  padding: 0 var(--space-3);
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-radius: 999px;
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
  color: var(--journal-muted);
  font-size: var(--font-size-0-82);
  line-height: 1;
}

.image-status-pill strong {
  font-size: var(--font-size-0-9);
  font-weight: 700;
  color: var(--journal-ink);
}

.image-status-pill--success {
  border-color: color-mix(in srgb, var(--color-success) 22%, transparent);
  background: color-mix(in srgb, var(--color-success) 12%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-success) 82%, var(--journal-ink));
}

.image-status-pill--warning {
  border-color: color-mix(in srgb, var(--color-warning) 24%, transparent);
  background: color-mix(in srgb, var(--color-warning) 12%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-warning) 84%, var(--journal-ink));
}

.image-status-pill--danger {
  border-color: color-mix(in srgb, var(--color-danger) 24%, transparent);
  background: color-mix(in srgb, var(--color-danger) 10%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-danger) 84%, var(--journal-ink));
}

.image-status-pill--muted {
  border-color: color-mix(in srgb, var(--journal-muted) 18%, transparent);
  background: color-mix(in srgb, var(--journal-muted) 10%, var(--journal-surface));
  color: var(--journal-muted);
}

@media (max-width: 720px) {
  .image-status-strip {
    align-items: flex-start;
  }

  .image-status-strip__note {
    width: 100%;
  }
}
</style>
