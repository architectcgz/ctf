<script setup lang="ts">
import type { AuditLogItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryToolbar from '@/components/common/WorkspaceDirectoryToolbar.vue'
import PlatformPaginationControls from '@/components/platform/PlatformPaginationControls.vue'
import type { WorkspaceDirectorySortOption } from '@/entities/workspace-directory'

interface Props {
  rows: AuditLogItem[]
  total: number
  page: number
  totalPages: number
  loading: boolean
  error: string | null
  keyword: string
  hasActiveFilters: boolean
  selectedSortLabel: string
  sortOptions: WorkspaceDirectorySortOption[]
  actionFilter: string
  resourceTypeFilter: string
  actorUserIdFilter: string
  formatDate: (value: string) => string
  detailPreview: (detail: Record<string, unknown> | undefined) => string
  actorDisplayName: (item: AuditLogItem) => string
}

defineProps<Props>()

const emit = defineEmits<{
  'update:keyword': [value: string]
  'update:actionFilter': [value: string]
  'update:resourceTypeFilter': [value: string]
  'update:actorUserIdFilter': [value: string]
  'select-sort': [option: WorkspaceDirectorySortOption]
  'reset-filters': []
  retry: []
  'open-actor-detail': [row: AuditLogItem]
  'change-page': [page: number]
}>()

const auditTableColumns = [
  {
    key: 'created_at',
    label: '时间',
    widthClass: 'w-[18%] min-w-[11rem]',
    cellClass: 'audit-table__time-cell',
  },
  {
    key: 'action',
    label: '动作',
    widthClass: 'w-[12%] min-w-[7rem]',
    cellClass: 'audit-table__action-cell',
  },
  {
    key: 'resource',
    label: '资源',
    widthClass: 'w-[18%] min-w-[10rem]',
    cellClass: 'audit-table__resource-cell',
  },
  {
    key: 'actor',
    label: '执行人',
    widthClass: 'w-[18%] min-w-[10rem]',
    cellClass: 'audit-table__actor-cell',
  },
  {
    key: 'detail',
    label: '明细',
    widthClass: 'w-[34%] min-w-[16rem]',
    cellClass: 'audit-table__detail-cell',
  },
]

function updateActionFilter(event: Event): void {
  emit('update:actionFilter', (event.target as HTMLSelectElement).value)
}

function updateResourceTypeFilter(event: Event): void {
  emit('update:resourceTypeFilter', (event.target as HTMLInputElement).value)
}

function updateActorUserIdFilter(event: Event): void {
  emit('update:actorUserIdFilter', (event.target as HTMLInputElement).value)
}
</script>

<template>
  <section class="admin-board workspace-directory-section">
    <header class="list-heading admin-board__head">
      <div>
        <div class="workspace-overline">Operational Stream</div>
        <h2 class="list-heading__title">操作流水</h2>
      </div>
    </header>

    <WorkspaceDirectoryToolbar
      :model-value="keyword"
      :total="total"
      :selected-sort-label="selectedSortLabel"
      :sort-options="sortOptions"
      search-placeholder="检索动作、资源类型、执行人..."
      total-suffix="条日志"
      reset-label="重置筛选"
      :reset-disabled="!hasActiveFilters"
      @update:model-value="emit('update:keyword', $event)"
      @select-sort="emit('select-sort', $event)"
      @reset-filters="emit('reset-filters')"
    >
      <template #filter-panel>
        <div class="audit-filter-grid">
          <label class="audit-filter-field">
            <span class="audit-filter-label">动作</span>
            <select
              :value="actionFilter"
              class="workspace-directory-filter-control audit-filter-select"
              @change="updateActionFilter"
            >
              <option value="">全部动作</option>
              <option value="login">登录</option>
              <option value="logout">登出</option>
              <option value="create">创建</option>
              <option value="update">更新</option>
              <option value="delete">删除</option>
              <option value="submit">提交</option>
              <option value="admin_op">管理员操作</option>
            </select>
          </label>

          <label class="audit-filter-field">
            <span class="audit-filter-label">资源类型</span>
            <input
              :value="resourceTypeFilter"
              type="text"
              placeholder="资源类型，如 challenge"
              class="workspace-directory-filter-control audit-filter-select"
              @input="updateResourceTypeFilter"
            />
          </label>

          <label class="audit-filter-field">
            <span class="audit-filter-label">执行人</span>
            <input
              :value="actorUserIdFilter"
              type="number"
              min="1"
              placeholder="执行人 ID"
              class="workspace-directory-filter-control audit-filter-select"
              @input="updateActorUserIdFilter"
            />
          </label>
        </div>
      </template>
    </WorkspaceDirectoryToolbar>

    <div v-if="error" class="audit-error-banner">
      {{ error }}
      <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">重试</button>
    </div>

    <div
      v-else-if="loading && rows.length === 0"
      class="workspace-directory-loading flex justify-center py-12"
    >
      <AppLoading>正在同步审计数据...</AppLoading>
    </div>

    <template v-else>
      <AppEmpty
        v-if="rows.length === 0"
        icon="Inbox"
        title="当前筛选条件下没有日志记录"
        description="可以放宽动作、资源类型或执行人条件，再重新检索。"
        class="audit-empty-state workspace-directory-empty py-20"
      />

      <WorkspaceDataTable
        v-else
        class="audit-list workspace-directory-list"
        :columns="auditTableColumns"
        :rows="rows"
        row-key="id"
        row-class="audit-row"
      >
        <template #cell-created_at="{ row }">
          <span class="audit-row__time">
            {{ formatDate((row as AuditLogItem).created_at) }}
          </span>
        </template>

        <template #cell-action="{ row }">
          <span
            class="audit-chip"
            :class="['workspace-directory-status-pill', 'workspace-directory-status-pill--muted']"
            >{{ (row as AuditLogItem).action }}</span
          >
        </template>

        <template #cell-resource="{ row }">
          <div class="audit-row__resource">
            <span class="audit-row__resource-type">{{ (row as AuditLogItem).resource_type }}</span>
            <span v-if="(row as AuditLogItem).resource_id" class="audit-row__resource-id">
              #{{ (row as AuditLogItem).resource_id }}
            </span>
          </div>
        </template>

        <template #cell-actor="{ row }">
          <div class="audit-row__actor">
            <button
              type="button"
              class="audit-row__actor-link"
              @click="emit('open-actor-detail', row as AuditLogItem)"
            >
              {{ actorDisplayName(row as AuditLogItem) }}
            </button>
          </div>
        </template>

        <template #cell-detail="{ row }">
          <p class="audit-row__detail" :title="detailPreview((row as AuditLogItem).detail)">
            {{ detailPreview((row as AuditLogItem).detail) }}
          </p>
        </template>
      </WorkspaceDataTable>

      <div v-if="total > 0" class="admin-pagination workspace-directory-pagination">
        <PlatformPaginationControls
          :page="page"
          :total-pages="totalPages"
          :total="total"
          total-label="条记录"
          @change-page="emit('change-page', $event)"
        />
      </div>
    </template>
  </section>
</template>

<style scoped>
.admin-board {
  display: grid;
  gap: var(--space-4);
  padding-top: var(--space-1);
}

.admin-board__head {
  margin-bottom: 0;
}

.admin-board :deep(.workspace-directory-toolbar) {
  margin-bottom: 0;
}

.audit-filter-grid {
  display: grid;
  gap: var(--space-4);
}

.audit-filter-field {
  display: grid;
  gap: var(--space-2);
}

.audit-filter-label {
  font-size: var(--font-size-0-72);
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.audit-error-banner {
  border: 1px solid color-mix(in srgb, var(--color-danger) 20%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--color-danger) 8%, transparent);
  padding: var(--space-4);
  font-size: var(--font-size-0-875);
  color: var(--color-danger);
}

.audit-list {
  --audit-table-border: color-mix(in srgb, var(--journal-border) 74%, transparent);
  --audit-row-divider: color-mix(in srgb, var(--journal-border) 62%, transparent);
  border: 1px solid var(--audit-table-border);
  border-radius: 1rem;
  overflow: hidden;
}

.audit-list :deep(.workspace-data-table__row) {
  border-bottom-color: var(--audit-row-divider);
}

.audit-chip {
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.audit-row__time {
  display: block;
  font-size: var(--font-size-13);
  line-height: 1.6;
  color: var(--color-text-muted);
}

.audit-row__resource,
.audit-row__actor {
  display: grid;
  gap: 0.15rem;
  min-width: 0;
}

.audit-row__resource-type {
  font-size: var(--font-size-14);
  font-weight: 700;
  color: var(--color-text-primary);
}

.audit-row__resource-id {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-12);
  color: var(--color-text-muted);
}

.audit-row__actor-link {
  display: inline-flex;
  align-items: center;
  min-width: 0;
  border: 0;
  background: transparent;
  padding: 0;
  text-align: left;
  cursor: pointer;
  color: var(--color-primary);
  font-size: var(--font-size-14);
  font-weight: 700;
  line-height: 1.45;
  text-decoration: underline;
  text-decoration-thickness: 1px;
  text-underline-offset: 0.18em;
  transition: all 150ms ease;
}

.audit-row__actor-link:hover {
  color: var(--color-primary-hover);
}

.audit-row__detail {
  display: -webkit-box;
  margin: 0;
  min-width: 0;
  font-size: var(--font-size-13);
  line-height: 1.65;
  color: var(--color-text-secondary);
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
