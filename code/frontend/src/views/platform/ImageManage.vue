<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import {
  ArrowDownWideNarrow,
  Calendar,
  Clock,
  Database,
  FileText,
  Fingerprint,
  Info,
  Maximize2,
  RefreshCw,
  SortAsc,
  Tag,
} from 'lucide-vue-next'

import { createImage, deleteImage, getImages, type AdminImagePayload } from '@/api/admin'
import type { AdminImageListItem, ImageStatus } from '@/api/contracts'
import PlatformPaginationControls from '@/components/platform/PlatformPaginationControls.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryToolbar, {
  type WorkspaceDirectorySortOption,
} from '@/components/common/WorkspaceDirectoryToolbar.vue'
import AdminSurfaceModal from '@/components/common/modal-templates/AdminSurfaceModal.vue'
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
const imageTableColumns = [
  {
    key: 'name',
    label: '镜像名称',
    widthClass: 'w-[18%] min-w-[10rem]',
    cellClass: 'image-table__name-cell',
  },
  {
    key: 'tag',
    label: '标签',
    widthClass: 'w-[14%] min-w-[7rem]',
    cellClass: 'image-table__tag-cell',
  },
  {
    key: 'description',
    label: '描述',
    widthClass: 'w-[28%] min-w-[14rem]',
    cellClass: 'image-table__description-cell',
  },
  {
    key: 'status',
    label: '状态',
    align: 'center' as const,
    widthClass: 'w-[12%] min-w-[7rem]',
    cellClass: 'image-table__status-cell',
  },
  {
    key: 'created_at',
    label: '创建时间',
    widthClass: 'w-[18%] min-w-[10rem]',
    cellClass: 'image-table__time-cell',
  },
  {
    key: 'actions',
    label: '操作',
    align: 'right' as const,
    widthClass: 'w-[10rem]',
    cellClass: 'image-table__actions-cell',
  },
]

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

      <section class="image-board workspace-directory-section">
        <header class="list-heading image-board__head">
          <div>
            <div class="workspace-overline">
              Images
            </div>
            <h2 class="list-heading__title image-section-title">
              镜像列表
            </h2>
          </div>
        </header>

        <WorkspaceDirectoryToolbar
          v-model="keyword"
          :total="filteredTotal"
          :selected-sort-label="sortConfig.label"
          :sort-options="sortOptions"
          search-placeholder="检索镜像名称、标签或说明..."
          total-suffix="个镜像"
          :reset-disabled="!hasActiveFilters"
          @select-sort="setSort"
          @reset-filters="resetFilters"
        >
          <template #filter-panel>
            <div class="image-filter-grid">
              <label class="image-filter-field">
                <span class="image-filter-label">构建状态</span>
                <select
                  v-model="statusFilter"
                  class="image-filter-select"
                >
                  <option value="">全部状态</option>
                  <option value="available">可用</option>
                  <option value="building">构建中</option>
                  <option value="pending">等待中</option>
                  <option value="failed">失败</option>
                </select>
              </label>
            </div>
          </template>
        </WorkspaceDirectoryToolbar>

        <div
          v-if="loading"
          class="workspace-directory-loading flex items-center justify-center py-12"
        >
          <div
            class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]"
          />
        </div>

        <template v-else>
          <div
            v-if="list.length === 0"
            class="admin-empty workspace-directory-empty"
          >
            当前还没有镜像。
          </div>

          <div
            v-else-if="filteredRows.length === 0"
            class="admin-empty workspace-directory-empty"
          >
            当前筛选条件下没有匹配镜像。
          </div>

          <WorkspaceDataTable
            v-else
            class="image-list workspace-directory-list"
            :columns="imageTableColumns"
            :rows="filteredRows"
            row-key="id"
            row-class="image-row"
          >
            <template #cell-name="{ row }">
              <span
                class="image-row__name"
                :title="(row as AdminImageListItem).name"
              >
                {{ (row as AdminImageListItem).name }}
              </span>
            </template>

            <template #cell-tag="{ row }">
              <span
                class="image-row__tag"
                :title="(row as AdminImageListItem).tag"
              >
                {{ (row as AdminImageListItem).tag }}
              </span>
            </template>

            <template #cell-description="{ row }">
              <p
                class="image-row__description"
                :title="(row as AdminImageListItem).description || '未填写镜像说明'"
              >
                {{ (row as AdminImageListItem).description || '未填写镜像说明' }}
              </p>
            </template>

            <template #cell-status="{ row }">
              <div class="image-row__status">
                <span
                  class="admin-status-chip"
                  :style="getStatusStyle((row as AdminImageListItem).status)"
                >
                  {{ getStatusLabel((row as AdminImageListItem).status) }}
                </span>
              </div>
            </template>

            <template #cell-created_at="{ row }">
              <span class="image-row__time">
                {{ formatDateTime((row as AdminImageListItem).created_at) }}
              </span>
            </template>

            <template #cell-actions="{ row }">
              <div class="image-row__actions">
                <button
                  class="ui-btn ui-btn--sm ui-btn--ghost"
                  @click="openDetail(row as AdminImageListItem)"
                >
                  详情
                </button>
                <button
                  class="ui-btn ui-btn--sm ui-btn--danger"
                  @click="handleDelete((row as AdminImageListItem).id)"
                >
                  删除
                </button>
              </div>
            </template>
          </WorkspaceDataTable>

          <div
            v-if="total > 0"
            class="admin-pagination workspace-directory-pagination"
          >
            <PlatformPaginationControls
              :page="page"
              :total-pages="Math.max(1, Math.ceil(total / pageSize))"
              :total="total"
              :total-label="`共 ${total} 条`"
              @change-page="void changePage($event)"
            />
          </div>
        </template>
      </section>
    </main>

    <AdminSurfaceModal
      class="image-detail-modal"
      :open="!!activeImage"
      :frosted="true"
      title="镜像详情"
      eyebrow="Image Registry"
      width="34rem"
      @close="closeDetail"
      @update:open="!$event && closeDetail()"
    >
      <section
        v-if="activeImage"
        class="image-detail"
      >
        <div class="image-detail__grid">
          <article class="image-detail__item">
            <div class="image-detail__head">
              <Database class="h-3.5 w-3.5" />
              <span class="image-detail__label">镜像名称</span>
            </div>
            <strong class="image-detail__value">{{ activeImage.name }}</strong>
          </article>

          <article class="image-detail__item">
            <div class="image-detail__head">
              <Tag class="h-3.5 w-3.5" />
              <span class="image-detail__label">标签版本</span>
            </div>
            <strong class="image-detail__value">{{ activeImage.tag }}</strong>
          </article>

          <article class="image-detail__item">
            <div class="image-detail__head">
              <Fingerprint class="h-3.5 w-3.5" />
              <span class="image-detail__label">镜像 ID</span>
            </div>
            <strong class="image-detail__value image-detail__value--mono">
              {{ activeImage.id }}
            </strong>
          </article>

          <article class="image-detail__item">
            <div class="image-detail__head">
              <Info class="h-3.5 w-3.5" />
              <span class="image-detail__label">状态</span>
            </div>
            <div class="image-detail__value">
              <span
                class="admin-status-chip"
                :style="getStatusStyle(activeImage.status)"
              >
                {{ getStatusLabel(activeImage.status) }}
              </span>
            </div>
          </article>

          <article class="image-detail__item">
            <div class="image-detail__head">
              <Maximize2 class="h-3.5 w-3.5" />
              <span class="image-detail__label">占用空间</span>
            </div>
            <strong class="image-detail__value">{{ formatSize(activeImage.size_bytes) }}</strong>
          </article>

          <article class="image-detail__item">
            <div class="image-detail__head">
              <Clock class="h-3.5 w-3.5" />
              <span class="image-detail__label">最后更新</span>
            </div>
            <strong class="image-detail__value">
              {{ formatDateTime(activeImage.updated_at || activeImage.created_at) }}
            </strong>
          </article>

          <article class="image-detail__item image-detail__item--wide">
            <div class="image-detail__head">
              <FileText class="h-3.5 w-3.5" />
              <span class="image-detail__label">描述信息</span>
            </div>
            <p class="image-detail__description">
              {{ activeImage.description || '未提供详细描述' }}
            </p>
          </article>
        </div>
      </section>
    </AdminSurfaceModal>

    <AdminSurfaceModal
      class="image-create-modal"
      :open="dialogVisible"
      :frosted="true"
      title="创建镜像"
      subtitle="填写镜像名称、标签和说明，提交后会进入镜像目录并参与构建状态跟踪。"
      eyebrow="Image Registry"
      width="31.25rem"
      @close="dialogVisible = false"
      @update:open="dialogVisible = $event"
    >
      <form
        class="image-create-form"
        @submit.prevent="handleCreate"
      >
        <label class="ui-field image-create-field">
          <span class="ui-field__label">
            镜像名称
            <span
              class="ui-field__required"
              aria-hidden="true"
            >*</span>
          </span>
          <span class="ui-control-wrap">
            <input
              v-model="form.name"
              type="text"
              class="ui-control"
              placeholder="例如：ubuntu"
            >
          </span>
        </label>

        <label class="ui-field image-create-field">
          <span class="ui-field__label">
            标签
            <span
              class="ui-field__required"
              aria-hidden="true"
            >*</span>
          </span>
          <span class="ui-control-wrap">
            <input
              v-model="form.tag"
              type="text"
              class="ui-control"
              placeholder="例如：22.04"
            >
          </span>
        </label>

        <label class="ui-field image-create-field">
          <span class="ui-field__label">描述</span>
          <span class="ui-control-wrap image-create-field__textarea">
            <textarea
              v-model="form.description"
              class="ui-control"
              rows="3"
              placeholder="镜像说明（可选）"
            />
          </span>
        </label>
      </form>
      <template #footer>
        <div class="image-create-dialog__footer">
          <button
            type="button"
            class="ui-btn ui-btn--secondary"
            @click="dialogVisible = false"
          >
            取消
          </button>
          <button
            type="button"
            :disabled="creating"
            class="ui-btn ui-btn--primary"
            @click="handleCreate"
          >
            {{ creating ? '创建中...' : '创建' }}
          </button>
        </div>
      </template>
    </AdminSurfaceModal>
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

.image-header__actions > .ui-btn,
.image-row__actions .ui-btn {
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
  --ui-btn-danger-border: color-mix(in srgb, var(--color-danger) 28%, transparent);
  --ui-btn-danger-background: color-mix(in srgb, var(--color-danger) 10%, var(--journal-surface));
  --ui-btn-danger-color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
  --ui-btn-danger-hover-border: color-mix(in srgb, var(--color-danger) 34%, transparent);
  --ui-btn-danger-hover-background: color-mix(in srgb, var(--color-danger) 14%, var(--journal-surface));
}

.admin-status-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 0.5rem;
  padding: var(--space-1) var(--space-2-5);
  font-size: var(--font-size-0-72);
  font-weight: 600;
}

.admin-empty {
  padding: var(--space-4) 0 0;
  font-size: var(--font-size-0-875);
  color: var(--journal-muted);
}

.image-create-form {
  display: grid;
  gap: var(--space-4);
}

.image-create-field {
  --ui-field-gap: var(--space-2);
}

.image-create-field__textarea {
  align-items: stretch;
}

.image-create-field__textarea .ui-control {
  min-height: 6.25rem;
  resize: vertical;
}

.image-create-dialog__footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-2);
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

.image-board {
  display: grid;
  gap: var(--space-4);
  padding-top: var(--space-1);
}

.image-section-title,
.image-row__time {
  color: var(--journal-ink);
}

.image-row__time {
  font-size: var(--font-size-0-82);
  line-height: 1.6;
  color: var(--journal-muted);
}

.image-filter-grid {
  display: grid;
  gap: var(--space-4);
}

.image-filter-field {
  display: grid;
  gap: var(--space-2);
}

.image-filter-label {
  font-size: var(--font-size-0-72);
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.image-filter-select {
  width: 100%;
  min-height: 2.5rem;
  padding: 0 var(--space-3);
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-radius: 0.75rem;
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
  font-size: var(--font-size-0-875);
  color: var(--journal-ink);
}

.image-list {
  border: 1px solid color-mix(in srgb, var(--journal-border) 82%, transparent);
  border-radius: 1.35rem;
  background: color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base));
  padding: 0.25rem 0.9rem 0.4rem;
}

.image-list :deep(.workspace-data-table__head-cell) {
  border-bottom-color: color-mix(in srgb, var(--journal-border) 82%, transparent);
}

.image-list :deep(.workspace-data-table__row) {
  border-bottom-color: color-mix(in srgb, var(--journal-border) 82%, transparent);
}

.image-list :deep(.workspace-data-table__body tr:last-child) {
  border-bottom-color: transparent;
}

.image-list :deep(.workspace-data-table__body-cell) {
  vertical-align: top;
}

.image-row__name,
.image-row__tag {
  display: block;
  min-width: 0;
  font-family: var(--font-family-mono);
}

.image-row__name {
  font-size: var(--font-size-1-00);
  font-weight: 700;
  color: var(--journal-ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.image-row__tag {
  color: var(--journal-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.image-row__description {
  display: -webkit-box;
  margin: 0;
  min-width: 0;
  font-size: var(--font-size-0-88);
  line-height: 1.65;
  color: var(--journal-muted);
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.image-row__status {
  display: flex;
  justify-content: center;
}

.image-row__time {
  display: block;
}

.image-row__actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.image-detail__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1.75rem 2rem;
  padding: 0.25rem;
}

.image-detail__item {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.image-detail__item--wide {
  grid-column: 1 / -1;
}

.image-detail__head {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--journal-muted);
}

.image-detail__label {
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.image-detail__value {
  font-size: 16px;
  font-weight: 800;
  line-height: 1.2;
  color: var(--journal-ink);
}

.image-detail__value--mono {
  font-family: var(--font-family-mono);
}

.image-detail__description {
  margin: 0;
  font-size: 14px;
  line-height: 1.7;
  color: var(--journal-muted);
}

@media (max-width: 1040px) {
  .image-list {
    min-width: 56rem;
  }
}

@media (max-width: 720px) {
  .image-status-strip {
    align-items: flex-start;
  }

  .image-status-strip__note {
    width: 100%;
  }

  .image-create-dialog__footer {
    flex-direction: column-reverse;
  }

  .image-create-dialog__footer > .ui-btn {
    width: 100%;
  }
}
</style>



