<template>
  <div class="awd-library-pane">
    <div
      class="admin-summary-grid awd-challenge-summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"
    >
      <article class="journal-note progress-card metric-panel-card">
        <div class="journal-note-label progress-card-label metric-panel-label">
          <span>题目总量</span>
          <Box class="h-4 w-4" />
        </div>
        <div class="journal-note-value progress-card-value metric-panel-value">
          {{ total.toString().padStart(2, '0') }}
        </div>
        <div class="journal-note-helper progress-card-hint metric-panel-helper">
          当前筛选条件下可管理的题目
        </div>
      </article>

      <article class="journal-note progress-card metric-panel-card">
        <div class="journal-note-label progress-card-label metric-panel-label">
          <span>已发布</span>
          <CheckCircle class="h-4 w-4" />
        </div>
        <div class="journal-note-value progress-card-value metric-panel-value">
          {{ publishedCount.toString().padStart(2, '0') }}
        </div>
        <div class="journal-note-helper progress-card-hint metric-panel-helper">
          已开放给 AWD 编排使用的题目
        </div>
      </article>

      <article class="journal-note progress-card metric-panel-card">
        <div class="journal-note-label progress-card-label metric-panel-label">
          <span>Web HTTP</span>
          <Activity class="h-4 w-4" />
        </div>
        <div class="journal-note-value progress-card-value metric-panel-value">
          {{ webHttpCount.toString().padStart(2, '0') }}
        </div>
        <div class="journal-note-helper progress-card-hint metric-panel-helper">
          使用 HTTP 探测与 Web 服务模式的题目
        </div>
      </article>

      <article class="journal-note progress-card metric-panel-card">
        <div class="journal-note-label progress-card-label metric-panel-label">
          <span>待验证</span>
          <Clock class="h-4 w-4" />
        </div>
        <div class="journal-note-value progress-card-value metric-panel-value">
          {{ pendingReadinessCount.toString().padStart(2, '0') }}
        </div>
        <div class="journal-note-helper progress-card-hint metric-panel-helper">
          仍需完成 Checker 验证的题目
        </div>
      </article>
    </div>

    <section class="workspace-directory-section">
      <section
        class="awd-challenge-directory-shell workspace-directory-list workspace-directory-list--catalog"
      >
        <header class="list-heading">
          <div>
            <div class="workspace-overline">AWD Challenge Directory</div>
            <h2 class="list-heading__title">AWD 题目目录</h2>
          </div>
        </header>

        <WorkspaceDirectoryToolbar
          :model-value="keyword"
          :total="total"
          selected-sort-label=""
          :sort-options="[]"
          search-placeholder="检索题目名称、Slug 或描述..."
          filter-panel-title="AWD 题目筛选"
          total-suffix="个题目"
          reset-label="重置筛选"
          :reset-disabled="!hasActiveFilters"
          @update:model-value="emit('updateKeyword', $event)"
          @reset-filters="resetFilters"
        >
          <template #filter-panel>
            <div class="awd-challenge-library__filter-grid">
              <label class="awd-challenge-library__filter-field">
                <span class="awd-challenge-library__filter-label">服务类型</span>
                <select
                  :value="serviceTypeFilter"
                  class="workspace-directory-filter-control awd-filter-control"
                  @change="handleServiceTypeFilterChange"
                >
                  <option value="">全部类型</option>
                  <option value="web_http">Web HTTP</option>
                  <option value="binary_tcp">Binary TCP</option>
                  <option value="multi_container">Multi Container</option>
                </select>
              </label>

              <label class="awd-challenge-library__filter-field">
                <span class="awd-challenge-library__filter-label">发布状态</span>
                <select
                  :value="statusFilter"
                  class="workspace-directory-filter-control awd-filter-control"
                  @change="handleStatusFilterChange"
                >
                  <option value="">全部状态</option>
                  <option value="draft">草稿</option>
                  <option value="published">已发布</option>
                  <option value="archived">已归档</option>
                </select>
              </label>
            </div>
          </template>
        </WorkspaceDirectoryToolbar>

        <div v-if="loading && list.length === 0" class="flex justify-center py-12">
          <AppLoading>正在同步题目数据...</AppLoading>
        </div>

        <template v-else>
          <AppEmpty
            v-if="list.length === 0"
            class="awd-challenge-library__empty"
            icon="Box"
            title="暂无 AWD 题目"
            :description="hasActiveFilters ? '当前筛选条件下没有匹配题目。' : '还没有 AWD 题目。'"
          />

          <WorkspaceDataTable
            v-else
            class="awd-challenge-list"
            :columns="awdChallengeTableColumns"
            :rows="list"
            row-key="id"
            row-class="awd-challenge-table-row group"
          >
            <template #cell-name="{ row }">
              <div class="awd-challenge-table__name">
                <div class="workspace-directory-row-title awd-challenge-table__title">
                  {{ (row as AdminAwdChallengeData).name }}
                </div>
              </div>
            </template>

            <template #cell-slug="{ row }">
              <span class="workspace-directory-row-subtitle awd-challenge-table__slug">
                {{ (row as AdminAwdChallengeData).slug }}
              </span>
            </template>

            <template #cell-service_type="{ row }">
              <span class="workspace-directory-mono awd-challenge-table__mono">{{
                getServiceTypeLabel((row as AdminAwdChallengeData).service_type)
              }}</span>
            </template>

            <template #cell-deployment_mode="{ row }">
              <span class="workspace-directory-compact-text awd-challenge-table__compact-text">{{
                getDeploymentModeLabel((row as AdminAwdChallengeData).deployment_mode)
              }}</span>
            </template>

            <template #cell-difficulty="{ row }">
              <ChallengeDifficultyText
                class="awd-challenge-table__difficulty"
                :difficulty="(row as AdminAwdChallengeData).difficulty"
                :label-overrides="{ insane: '高强度' }"
              />
            </template>

            <template #cell-readiness_status="{ row }">
              <span
                class="awd-status-pill workspace-directory-status-pill"
                :class="getReadinessClass((row as AdminAwdChallengeData).readiness_status)"
              >
                {{ getReadinessLabel((row as AdminAwdChallengeData).readiness_status) }}
              </span>
            </template>

            <template #cell-status="{ row }">
              <span
                class="awd-status-pill workspace-directory-status-pill"
                :class="getStatusClass((row as AdminAwdChallengeData).status)"
              >
                {{ getStatusLabel((row as AdminAwdChallengeData).status) }}
              </span>
            </template>

            <template #cell-actions="{ row }">
              <div class="workspace-directory-row-actions awd-challenge-table__actions">
                <button
                  type="button"
                  class="workspace-directory-row-btn"
                  @click="emit('openEditDialog', row as AdminAwdChallengeData)"
                >
                  编辑
                </button>
                <button
                  type="button"
                  class="workspace-directory-row-btn workspace-directory-row-btn--danger"
                  @click="emit('deleteChallenge', row as AdminAwdChallengeData)"
                >
                  删除
                </button>
              </div>
            </template>
          </WorkspaceDataTable>

          <div v-if="total > 0" class="admin-pagination workspace-directory-pagination">
            <WorkspaceDirectoryPagination
              :page="page"
              :total-pages="totalPages"
              :total="total"
              :disabled="loading"
              total-label="个题目"
              @change-page="emit('changePage', $event)"
            />
          </div>
        </template>
      </section>
    </section>
  </div>
</template>

<style scoped>
.awd-library-pane {
  display: flex;
  flex-direction: column;
  gap: var(--workspace-directory-page-block-gap, var(--space-5));
}

.awd-challenge-summary {
  --admin-summary-grid-columns: repeat(4, minmax(0, 1fr));
  --metric-panel-columns: repeat(4, minmax(0, 1fr));
}

.awd-challenge-library__filter-grid {
  display: grid;
  gap: var(--space-4);
}

.awd-challenge-library__filter-field {
  display: grid;
  gap: var(--space-2);
}

.awd-challenge-library__filter-label {
  font-size: var(--font-size-11);
  font-weight: 800;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.awd-challenge-table__name {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.awd-challenge-table__slug {
  font-family: var(--font-family-mono);
}

.awd-challenge-table__difficulty {
  font-size: var(--font-size-13);
  font-weight: 700;
  color: var(--color-text-muted);
}

.awd-challenge-table__compact-text {
  color: var(--color-text-primary);
}

.awd-status-pill {
  --workspace-directory-pill-min-width: 4.8rem;
}

@media (max-width: 1024px) {
  .awd-challenge-table__actions {
    flex-direction: column;
    align-items: stretch;
  }
}

@media (max-width: 860px) {
  .awd-challenge-summary {
    --admin-summary-grid-columns: repeat(2, minmax(0, 1fr));
    --metric-panel-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 560px) {
  .awd-challenge-summary {
    --admin-summary-grid-columns: 1fr;
    --metric-panel-columns: 1fr;
  }
}
</style>

<script setup lang="ts">
import { computed } from 'vue'
import { Activity, Box, CheckCircle, Clock } from 'lucide-vue-next'

import type { AdminAwdChallengeData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import WorkspaceDirectoryToolbar from '@/components/common/WorkspaceDirectoryToolbar.vue'
import { ChallengeDifficultyText } from '@/entities/challenge'

type AwdServiceTypeFilter = AdminAwdChallengeData['service_type'] | ''
type AwdServiceStatusFilter = AdminAwdChallengeData['status'] | ''

const props = defineProps<{
  list: AdminAwdChallengeData[]
  total: number
  page: number
  pageSize: number
  loading: boolean
  keyword: string
  serviceTypeFilter: AwdServiceTypeFilter
  statusFilter: AwdServiceStatusFilter
}>()

const emit = defineEmits<{
  updateKeyword: [value: string]
  updateServiceTypeFilter: [value: AwdServiceTypeFilter]
  updateStatusFilter: [value: AwdServiceStatusFilter]
  openEditDialog: [challenge: AdminAwdChallengeData]
  deleteChallenge: [challenge: AdminAwdChallengeData]
  changePage: [page: number]
}>()

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const publishedCount = computed(
  () => props.list.filter((item) => item.status === 'published').length
)
const webHttpCount = computed(
  () => props.list.filter((item) => item.service_type === 'web_http').length
)
const pendingReadinessCount = computed(
  () => props.list.filter((item) => item.readiness_status === 'pending').length
)
const hasActiveFilters = computed(() =>
  Boolean(props.keyword.trim() || props.serviceTypeFilter || props.statusFilter)
)

const awdChallengeTableColumns = [
  {
    key: 'name',
    label: '题目名称',
    widthClass: 'w-[22%] min-w-[14rem]',
    cellClass: 'awd-challenge-table__name-cell',
  },
  {
    key: 'slug',
    label: '标识',
    widthClass: 'w-[12%] min-w-[8rem]',
    cellClass: 'awd-challenge-table__compact-cell',
  },
  {
    key: 'service_type',
    label: '类型',
    align: 'center' as const,
    widthClass: 'w-[12%] min-w-[7rem]',
    cellClass: 'awd-challenge-table__compact-cell',
  },
  {
    key: 'deployment_mode',
    label: '部署方式',
    align: 'center' as const,
    widthClass: 'w-[12%] min-w-[7rem]',
    cellClass: 'awd-challenge-table__compact-cell',
  },
  {
    key: 'difficulty',
    label: '难度',
    align: 'center' as const,
    widthClass: 'w-[10%] min-w-[6rem]',
    cellClass: 'awd-challenge-table__compact-cell',
  },
  {
    key: 'readiness_status',
    label: '就绪度',
    align: 'center' as const,
    widthClass: 'w-[10%] min-w-[6rem]',
    cellClass: 'awd-challenge-table__compact-cell',
  },
  {
    key: 'status',
    label: '状态',
    align: 'center' as const,
    widthClass: 'w-[10%] min-w-[6rem]',
    cellClass: 'awd-challenge-table__compact-cell',
  },
  {
    key: 'actions',
    label: '操作',
    align: 'right' as const,
    widthClass: 'w-[10rem]',
    cellClass: 'awd-challenge-table__actions-cell',
  },
]

function getServiceTypeLabel(value: AdminAwdChallengeData['service_type']): string {
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

function getDeploymentModeLabel(value: AdminAwdChallengeData['deployment_mode']): string {
  return value === 'topology' ? 'Topology' : 'Single'
}

function getStatusLabel(value: AdminAwdChallengeData['status']): string {
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

function getReadinessLabel(value: AdminAwdChallengeData['readiness_status']): string {
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

function getStatusClass(status: AdminAwdChallengeData['status']): string {
  if (status === 'published') return 'awd-status-pill--success'
  if (status === 'archived') return 'awd-status-pill--muted'
  return 'awd-status-pill--primary'
}

function getReadinessClass(readiness: AdminAwdChallengeData['readiness_status']): string {
  if (readiness === 'passed') return 'awd-status-pill--success'
  if (readiness === 'failed') return 'awd-status-pill--danger'
  return 'awd-status-pill--warning'
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
