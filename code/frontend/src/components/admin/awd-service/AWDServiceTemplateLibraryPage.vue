<script setup lang="ts">
import { computed } from 'vue'
import { Plus, RefreshCw } from 'lucide-vue-next'

import type { AdminAwdServiceTemplateData } from '@/api/contracts'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import WorkspaceDirectoryToolbar from '@/components/common/WorkspaceDirectoryToolbar.vue'

type AwdServiceTypeFilter = AdminAwdServiceTemplateData['service_type'] | ''
type AwdServiceStatusFilter = AdminAwdServiceTemplateData['status'] | ''

const props = defineProps<{
  list: AdminAwdServiceTemplateData[]
  total: number
  page: number
  pageSize: number
  loading: boolean
  keyword: string
  serviceTypeFilter: AwdServiceTypeFilter
  statusFilter: AwdServiceStatusFilter
}>()

const emit = defineEmits<{
  refresh: []
  updateKeyword: [value: string]
  updateServiceTypeFilter: [value: AwdServiceTypeFilter]
  updateStatusFilter: [value: AwdServiceStatusFilter]
  openCreateDialog: []
  openEditDialog: [template: AdminAwdServiceTemplateData]
  deleteTemplate: [template: AdminAwdServiceTemplateData]
  changePage: [page: number]
}>()

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const publishedCount = computed(() => props.list.filter((item) => item.status === 'published').length)
const webHttpCount = computed(() => props.list.filter((item) => item.service_type === 'web_http').length)
const pendingReadinessCount = computed(
  () => props.list.filter((item) => item.readiness_status === 'pending').length
)
const hasActiveFilters = computed(() =>
  Boolean(props.keyword.trim() || props.serviceTypeFilter || props.statusFilter)
)

const templateTableColumns = [
  {
    key: 'name',
    label: '模板',
    widthClass: 'w-[28%] min-w-[16rem]',
    cellClass: 'awd-template-table__name-cell',
  },
  {
    key: 'service_type',
    label: '服务类型',
    widthClass: 'w-[12%] min-w-[7rem]',
    cellClass: 'awd-template-table__compact-cell',
  },
  {
    key: 'deployment_mode',
    label: '部署方式',
    widthClass: 'w-[12%] min-w-[7rem]',
    cellClass: 'awd-template-table__compact-cell',
  },
  {
    key: 'difficulty',
    label: '难度',
    widthClass: 'w-[10%] min-w-[6rem]',
    cellClass: 'awd-template-table__compact-cell',
  },
  {
    key: 'readiness_status',
    label: '就绪度',
    widthClass: 'w-[10%] min-w-[6rem]',
    cellClass: 'awd-template-table__compact-cell',
  },
  {
    key: 'status',
    label: '状态',
    widthClass: 'w-[10%] min-w-[6rem]',
    cellClass: 'awd-template-table__compact-cell',
  },
  {
    key: 'updated_at',
    label: '更新时间',
    widthClass: 'w-[10%] min-w-[8rem]',
    cellClass: 'awd-template-table__time-cell',
  },
  {
    key: 'actions',
    label: '操作',
    align: 'right' as const,
    widthClass: 'w-[12rem]',
    cellClass: 'awd-template-table__actions-cell',
  },
]

function getServiceTypeLabel(value: AdminAwdServiceTemplateData['service_type']): string {
  switch (value) {
    case 'binary_tcp':
      return 'Binary TCP'
    case 'multi_container':
      return 'Multi Container'
    case 'web_http':
    default:
      return 'Web HTTP'
  }
}

function getDeploymentModeLabel(value: AdminAwdServiceTemplateData['deployment_mode']): string {
  return value === 'topology' ? 'Topology' : 'Single Container'
}

function getDifficultyLabel(value: AdminAwdServiceTemplateData['difficulty']): string {
  switch (value) {
    case 'beginner':
      return '入门'
    case 'easy':
      return '简单'
    case 'medium':
      return '中等'
    case 'hard':
      return '困难'
    case 'insane':
      return '高强度'
    default:
      return value
  }
}

function getStatusLabel(value: AdminAwdServiceTemplateData['status']): string {
  switch (value) {
    case 'published':
      return '已发布'
    case 'archived':
      return '已归档'
    case 'draft':
    default:
      return '草稿'
  }
}

function getReadinessLabel(value: AdminAwdServiceTemplateData['readiness_status']): string {
  switch (value) {
    case 'passed':
      return '已通过'
    case 'failed':
      return '未通过'
    case 'pending':
    default:
      return '待验证'
  }
}

function getStatusStyle(status: AdminAwdServiceTemplateData['status']): Record<string, string> {
  const accent =
    status === 'published'
      ? 'var(--color-success)'
      : status === 'archived'
        ? 'var(--journal-muted)'
        : 'var(--color-primary)'

  return {
    color: accent,
    borderColor: `color-mix(in srgb, ${accent} 18%, transparent)`,
    backgroundColor: `color-mix(in srgb, ${accent} 10%, var(--journal-surface))`,
  }
}

function getReadinessStyle(
  readiness: AdminAwdServiceTemplateData['readiness_status']
): Record<string, string> {
  const accent =
    readiness === 'passed'
      ? 'var(--color-success)'
      : readiness === 'failed'
        ? 'var(--color-danger)'
        : 'var(--color-warning)'

  return {
    color: accent,
    borderColor: `color-mix(in srgb, ${accent} 18%, transparent)`,
    backgroundColor: `color-mix(in srgb, ${accent} 10%, var(--journal-surface))`,
  }
}

function formatDateTime(value: string): string {
  return new Date(value).toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function resetFilters(): void {
  emit('updateKeyword', '')
  emit('updateServiceTypeFilter', '')
  emit('updateStatusFilter', '')
}

function handleServiceTypeFilterChange(event: Event): void {
  const target = event.target
  emit(
    'updateServiceTypeFilter',
    target instanceof HTMLSelectElement ? (target.value as AwdServiceTypeFilter) : ''
  )
}

function handleStatusFilterChange(event: Event): void {
  const target = event.target
  emit(
    'updateStatusFilter',
    target instanceof HTMLSelectElement ? (target.value as AwdServiceStatusFilter) : ''
  )
}
</script>

<template>
  <section
    class="journal-shell journal-shell-admin journal-notes-card journal-hero workspace-shell flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
  >
    <header class="list-heading awd-template-library__header">
      <div class="workspace-tab-heading__main">
        <div class="workspace-overline">AWD Service Library</div>
        <h1 class="workspace-page-title">AWD 服务模板库</h1>
        <p class="workspace-page-copy">
          这里单独维护 AWD 题目的服务模板，不再和 Jeopardy 题目混在同一资源目录里。
        </p>
      </div>

      <div class="awd-template-library__actions">
        <button type="button" class="admin-btn admin-btn-ghost" @click="emit('refresh')">
          <RefreshCw class="h-4 w-4" />
          刷新列表
        </button>
        <button
          id="awd-template-open-create"
          type="button"
          class="admin-btn admin-btn-primary"
          @click="emit('openCreateDialog')"
        >
          <Plus class="h-4 w-4" />
          创建模板
        </button>
      </div>
    </header>

    <div class="admin-summary-grid awd-template-library__summary progress-strip metric-panel-grid metric-panel-default-surface">
      <article class="journal-note progress-card metric-panel-card">
        <div class="journal-note-label progress-card-label metric-panel-label">模板总量</div>
        <div class="journal-note-value progress-card-value metric-panel-value">{{ total }}</div>
        <div class="journal-note-helper progress-card-hint metric-panel-helper">
          当前筛选条件下的模板总数
        </div>
      </article>
      <article class="journal-note progress-card metric-panel-card">
        <div class="journal-note-label progress-card-label metric-panel-label">已发布</div>
        <div class="journal-note-value progress-card-value metric-panel-value">{{ publishedCount }}</div>
        <div class="journal-note-helper progress-card-hint metric-panel-helper">
          当前页已发布模板数量
        </div>
      </article>
      <article class="journal-note progress-card metric-panel-card">
        <div class="journal-note-label progress-card-label metric-panel-label">Web HTTP</div>
        <div class="journal-note-value progress-card-value metric-panel-value">{{ webHttpCount }}</div>
        <div class="journal-note-helper progress-card-hint metric-panel-helper">
          当前页 HTTP 服务模板数量
        </div>
      </article>
      <article class="journal-note progress-card metric-panel-card">
        <div class="journal-note-label progress-card-label metric-panel-label">待验证</div>
        <div class="journal-note-value progress-card-value metric-panel-value">
          {{ pendingReadinessCount }}
        </div>
        <div class="journal-note-helper progress-card-hint metric-panel-helper">
          当前页尚未完成就绪检查
        </div>
      </article>
    </div>

    <section class="workspace-directory-section awd-template-library__directory">
      <header class="list-heading awd-template-library__directory-head">
        <div>
          <div class="journal-note-label">Service Templates</div>
          <h2 class="list-heading__title">模板目录</h2>
        </div>
        <div class="awd-template-library__directory-meta">当前页 {{ list.length }} 个模板</div>
      </header>

      <WorkspaceDirectoryToolbar
        :model-value="keyword"
        :total="total"
        selected-sort-label=""
        :sort-options="[]"
        search-placeholder="检索模板名称、slug 或分类"
        filter-panel-title="AWD 模板筛选"
        total-suffix="个模板"
        reset-label="重置筛选"
        :reset-disabled="!hasActiveFilters"
        @update:model-value="emit('updateKeyword', $event)"
        @reset-filters="resetFilters"
      >
        <template #filter-panel>
          <div class="awd-template-library__filter-grid">
            <label class="awd-template-library__filter-field">
              <span class="awd-template-library__filter-label">服务类型</span>
              <select
                :value="serviceTypeFilter"
                class="admin-input awd-template-library__filter-control"
                @change="handleServiceTypeFilterChange"
              >
                <option value="">全部类型</option>
                <option value="web_http">web_http</option>
                <option value="binary_tcp">binary_tcp</option>
                <option value="multi_container">multi_container</option>
              </select>
            </label>

            <label class="awd-template-library__filter-field">
              <span class="awd-template-library__filter-label">模板状态</span>
              <select
                :value="statusFilter"
                class="admin-input awd-template-library__filter-control"
                @change="handleStatusFilterChange"
              >
                <option value="">全部状态</option>
                <option value="draft">draft</option>
                <option value="published">published</option>
                <option value="archived">archived</option>
              </select>
            </label>
          </div>
        </template>
      </WorkspaceDirectoryToolbar>

      <div v-if="loading" class="workspace-directory-loading awd-template-library__loading">
        <div class="awd-template-library__spinner"></div>
      </div>

      <template v-else>
        <div
          v-if="list.length === 0"
          class="admin-empty workspace-directory-empty awd-template-library__empty"
        >
          {{ hasActiveFilters ? '当前筛选条件下没有匹配模板。' : '当前还没有 AWD 服务模板，请先创建模板。' }}
        </div>

        <WorkspaceDataTable
          v-else
          :columns="templateTableColumns"
          :rows="list"
          row-key="id"
          row-class="awd-template-table__row"
        >
          <template #cell-name="{ row }">
            <div class="awd-template-table__name">
              <div class="awd-template-table__title">{{ (row as AdminAwdServiceTemplateData).name }}</div>
              <div class="awd-template-table__slug">{{ (row as AdminAwdServiceTemplateData).slug }}</div>
            </div>
          </template>

          <template #cell-service_type="{ row }">
            <span class="awd-template-table__mono">
              {{ getServiceTypeLabel((row as AdminAwdServiceTemplateData).service_type) }}
            </span>
          </template>

          <template #cell-deployment_mode="{ row }">
            <span class="awd-template-table__compact-text">
              {{ getDeploymentModeLabel((row as AdminAwdServiceTemplateData).deployment_mode) }}
            </span>
          </template>

          <template #cell-difficulty="{ row }">
            <span class="awd-template-table__compact-text">
              {{ getDifficultyLabel((row as AdminAwdServiceTemplateData).difficulty) }}
            </span>
          </template>

          <template #cell-readiness_status="{ row }">
            <span
              class="admin-status-chip"
              :style="getReadinessStyle((row as AdminAwdServiceTemplateData).readiness_status)"
            >
              {{ getReadinessLabel((row as AdminAwdServiceTemplateData).readiness_status) }}
            </span>
          </template>

          <template #cell-status="{ row }">
            <span
              class="admin-status-chip"
              :style="getStatusStyle((row as AdminAwdServiceTemplateData).status)"
            >
              {{ getStatusLabel((row as AdminAwdServiceTemplateData).status) }}
            </span>
          </template>

          <template #cell-updated_at="{ row }">
            <span class="awd-template-table__time">
              {{ formatDateTime((row as AdminAwdServiceTemplateData).updated_at) }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <div class="awd-template-table__actions">
              <button
                type="button"
                class="ui-btn ui-btn--sm ui-btn--secondary"
                @click="emit('openEditDialog', row as AdminAwdServiceTemplateData)"
              >
                编辑
              </button>
              <button
                type="button"
                class="ui-btn ui-btn--sm ui-btn--danger"
                @click="emit('deleteTemplate', row as AdminAwdServiceTemplateData)"
              >
                删除
              </button>
            </div>
          </template>
        </WorkspaceDataTable>

        <WorkspaceDirectoryPagination
          :page="page"
          :total-pages="totalPages"
          :total="total"
          :disabled="loading"
          total-label="个模板"
          @change-page="emit('changePage', $event)"
        />
      </template>
    </section>
  </section>
</template>

<style scoped>
.awd-template-library__header,
.awd-template-library__directory-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.awd-template-library__actions,
.awd-template-table__actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.awd-template-library__summary {
  --metric-panel-columns: repeat(4, minmax(0, 1fr));
  margin-top: 1.5rem;
}

.awd-template-library__directory {
  margin-top: 1.5rem;
}

.awd-template-library__directory-meta {
  color: var(--journal-muted);
  font-size: 0.875rem;
}

.awd-template-library__filter-grid {
  display: grid;
  gap: 1rem;
}

.awd-template-library__filter-field {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.awd-template-library__filter-label {
  color: var(--journal-muted);
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.awd-template-library__filter-control {
  width: 100%;
}

.awd-template-library__loading {
  display: flex;
  justify-content: center;
  padding: 3rem 0;
}

.awd-template-library__spinner {
  width: 2rem;
  height: 2rem;
  border: 4px solid var(--journal-border);
  border-top-color: var(--journal-accent);
  border-radius: 999px;
  animation: awd-template-library-spin 0.8s linear infinite;
}

.awd-template-library__empty {
  padding: 2.5rem 1rem;
}

.awd-template-table__name {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.awd-template-table__title {
  color: var(--journal-ink);
  font-weight: 700;
}

.awd-template-table__slug,
.awd-template-table__mono,
.awd-template-table__time {
  color: var(--journal-muted);
  font-family: var(--font-family-mono);
  font-size: 0.8rem;
}

.awd-template-table__compact-text {
  color: var(--journal-ink);
  font-size: 0.875rem;
}

.admin-status-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 4.75rem;
  padding: 0.32rem 0.7rem;
  border: 1px solid transparent;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.04em;
}

@keyframes awd-template-library-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 960px) {
  .awd-template-library__header,
  .awd-template-library__directory-head {
    flex-direction: column;
  }

  .awd-template-library__actions {
    width: 100%;
    flex-wrap: wrap;
  }

  .awd-template-library__summary {
    --metric-panel-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .awd-template-library__summary {
    --metric-panel-columns: 1fr;
  }

  .awd-template-table__actions {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
