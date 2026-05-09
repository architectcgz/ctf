<script setup lang="ts">
import PlatformPaginationControls from '@/components/platform/PlatformPaginationControls.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryToolbar from '@/components/common/WorkspaceDirectoryToolbar.vue'
import type { AdminImageListItem, ImageStatus } from '@/api/contracts'
import type { WorkspaceDirectorySortOption } from '@/entities/workspace-directory'

interface Props {
  list: AdminImageListItem[]
  rows: AdminImageListItem[]
  total: number
  filteredTotal: number
  page: number
  totalPages: number
  loading: boolean
  keyword: string
  statusFilter: ImageStatus | ''
  hasActiveFilters: boolean
  sortOptions: WorkspaceDirectorySortOption[]
  selectedSortLabel: string
  getStatusLabel: (status: ImageStatus) => string
  getStatusStyle: (status: ImageStatus) => Record<string, string>
  formatDateTime: (value: string) => string
}

defineProps<Props>()

const emit = defineEmits<{
  'update:keyword': [value: string]
  'update:statusFilter': [value: ImageStatus | '']
  'select-sort': [option: WorkspaceDirectorySortOption]
  'reset-filters': []
  'open-detail': [row: AdminImageListItem]
  'delete-image': [id: string]
  'change-page': [page: number]
}>()

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
    key: 'source_type',
    label: '来源',
    align: 'center' as const,
    widthClass: 'w-[10%] min-w-[7rem]',
    cellClass: 'image-table__source-cell',
  },
  {
    key: 'digest',
    label: '摘要',
    widthClass: 'w-[22%] min-w-[12rem]',
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
    key: 'verified_at',
    label: '验证时间',
    widthClass: 'w-[14%] min-w-[9rem]',
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

function updateStatusFilter(event: Event): void {
  emit('update:statusFilter', (event.target as HTMLSelectElement).value as ImageStatus | '')
}

function getImageSourceLabel(value?: AdminImageListItem['source_type']): string {
  switch (value) {
    case 'platform_build':
      return '平台构建'
    case 'external_ref':
      return '外部镜像'
    case 'manual':
      return '手工登记'
    default:
      return '未标记'
  }
}
</script>

<template>
  <section class="image-board workspace-directory-section">
    <section class="image-directory-shell workspace-directory-list">
      <header class="list-heading image-board__head">
        <div>
          <div class="workspace-overline">Images</div>
          <h2 class="list-heading__title image-section-title">镜像列表</h2>
        </div>
      </header>

      <WorkspaceDirectoryToolbar
        :model-value="keyword"
        :total="filteredTotal"
        :selected-sort-label="selectedSortLabel"
        :sort-options="sortOptions"
        search-placeholder="检索镜像名称、标签或说明..."
        total-suffix="个镜像"
        :reset-disabled="!hasActiveFilters"
        @update:model-value="emit('update:keyword', $event)"
        @select-sort="emit('select-sort', $event)"
        @reset-filters="emit('reset-filters')"
      >
        <template #filter-panel>
          <div class="image-filter-grid">
            <label class="image-filter-field">
              <span class="image-filter-label">构建状态</span>
              <select
                :value="statusFilter"
                class="image-filter-select"
                @change="updateStatusFilter"
              >
                <option value="">全部状态</option>
                <option value="available">可用</option>
                <option value="building">构建中</option>
                <option value="pushed">已推送</option>
                <option value="verifying">校验中</option>
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
        <div v-if="list.length === 0" class="admin-empty workspace-directory-empty">
          当前还没有镜像。
        </div>

        <div v-else-if="rows.length === 0" class="admin-empty workspace-directory-empty">
          当前筛选条件下没有匹配镜像。
        </div>

        <WorkspaceDataTable
          v-else
          class="image-list"
          :columns="imageTableColumns"
          :rows="rows"
          row-key="id"
          row-class="image-row"
        >
          <template #cell-name="{ row }">
            <span class="image-row__name" :title="(row as AdminImageListItem).name">
              {{ (row as AdminImageListItem).name }}
            </span>
          </template>

          <template #cell-tag="{ row }">
            <span class="image-row__tag" :title="(row as AdminImageListItem).tag">
              {{ (row as AdminImageListItem).tag }}
            </span>
          </template>

          <template #cell-source_type="{ row }">
            <span class="image-row__source">
              {{ getImageSourceLabel((row as AdminImageListItem).source_type) }}
            </span>
          </template>

          <template #cell-digest="{ row }">
            <p
              class="image-row__description"
              :title="
                (row as AdminImageListItem).last_error ||
                (row as AdminImageListItem).digest ||
                (row as AdminImageListItem).description ||
                '未生成摘要'
              "
            >
              {{
                (row as AdminImageListItem).last_error ||
                (row as AdminImageListItem).digest ||
                (row as AdminImageListItem).description ||
                '未生成摘要'
              }}
            </p>
          </template>

          <template #cell-status="{ row }">
            <div class="image-row__status">
              <span
                class="admin-status-chip"
                :class="'workspace-directory-status-pill'"
                :style="getStatusStyle((row as AdminImageListItem).status)"
              >
                {{ getStatusLabel((row as AdminImageListItem).status) }}
              </span>
            </div>
          </template>

          <template #cell-verified_at="{ row }">
            <span class="image-row__time">
              {{
                (row as AdminImageListItem).verified_at
                  ? formatDateTime((row as AdminImageListItem).verified_at as string)
                  : '未验证'
              }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <div class="workspace-directory-row-actions image-row__actions">
              <button
                type="button"
                class="ui-btn ui-btn--sm ui-btn--primary"
                @click="emit('open-detail', row as AdminImageListItem)"
              >
                详情
              </button>
              <button
                type="button"
                class="ui-btn ui-btn--sm ui-btn--danger"
                @click="emit('delete-image', (row as AdminImageListItem).id)"
              >
                删除
              </button>
            </div>
          </template>
        </WorkspaceDataTable>

        <div v-if="total > 0" class="admin-pagination workspace-directory-pagination">
          <PlatformPaginationControls
            :page="page"
            :total-pages="totalPages"
            :total="total"
            :total-label="`共 ${total} 条`"
            @change-page="emit('change-page', $event)"
          />
        </div>
      </template>
    </section>
  </section>
</template>

<style scoped>
.admin-status-chip {
}

.admin-empty {
  padding: var(--space-4) 0 0;
  font-size: var(--font-size-0-875);
  color: var(--journal-muted);
}

.image-board {
  display: grid;
  gap: var(--space-4);
  padding-top: var(--space-1);
}

.image-directory-shell {
  --workspace-directory-shell-padding: var(--space-5);
  --workspace-directory-shell-radius: var(--radius-2xl);
  --workspace-directory-shell-border: color-mix(in srgb, var(--journal-border) 84%, transparent);
  --workspace-directory-shell-background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--color-primary) 6%, transparent),
      transparent 38%
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 74%, var(--color-bg-base))
    );
  display: grid;
  gap: var(--space-4);
  box-shadow: 0 calc(var(--space-4) + var(--space-0-5)) calc(var(--space-8) + var(--space-0-5))
    color-mix(in srgb, var(--color-shadow-soft) 20%, transparent);
}

.image-board__head {
  margin-bottom: 0;
}

.list-heading__title {
  font-size: clamp(1.2rem, 1rem + 0.5vw, 1.45rem);
}

.image-board :deep(.workspace-directory-toolbar) {
  margin-bottom: 0;
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

.image-row__source {
  font-size: var(--font-size-0-82);
  color: var(--journal-muted);
}

.image-row__time {
  display: block;
}

.image-row__actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

@media (max-width: 1040px) {
  .image-list {
    min-width: 56rem;
  }
}
</style>
