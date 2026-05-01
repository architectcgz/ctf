<script setup lang="ts">
import { computed } from 'vue'
import {
  Activity,
  Box,
  CheckCircle,
  Clock,
  RefreshCw,
  Upload,
} from 'lucide-vue-next'

import type {
  AdminAwdChallengeData,
  AdminAwdChallengeImportPreview,
} from '@/api/contracts'
import ChallengePackageImportEntry from '@/components/platform/challenge/ChallengePackageImportEntry.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import WorkspaceDirectoryToolbar from '@/components/common/WorkspaceDirectoryToolbar.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { ChallengeDifficultyText } from '@/entities/challenge'
import type { PlatformAwdChallengeImportUploadResult } from '@/features/platform-awd-challenges'

type AwdServiceTypeFilter = AdminAwdChallengeData['service_type'] | ''
type AwdServiceStatusFilter = AdminAwdChallengeData['status'] | ''

const props = withDefaults(defineProps<{
  mode?: 'library' | 'import'
  list: AdminAwdChallengeData[]
  total: number
  page: number
  pageSize: number
  loading: boolean
  keyword: string
  serviceTypeFilter: AwdServiceTypeFilter
  statusFilter: AwdServiceStatusFilter
  uploading: boolean
  queueLoading: boolean
  importQueue: AdminAwdChallengeImportPreview[]
  uploadResults: PlatformAwdChallengeImportUploadResult[]
  selectedFileName?: string
}>(), {
  mode: 'library',
})

const emit = defineEmits<{
  refresh: []
  refreshImportQueue: []
  updateKeyword: [value: string]
  updateServiceTypeFilter: [value: AwdServiceTypeFilter]
  updateStatusFilter: [value: AwdServiceStatusFilter]
  selectImportPackages: [files: File[]]
  commitImport: [preview: AdminAwdChallengeImportPreview]
  openImportPage: []
  openEditDialog: [challenge: AdminAwdChallengeData]
  deleteChallenge: [challenge: AdminAwdChallengeData]
  changePage: [page: number]
}>()

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const publishedCount = computed(() => props.list.filter((item) => item.status === 'published').length)
const webHttpCount = computed(() => props.list.filter((item) => item.service_type === 'web_http').length)
const pendingReadinessCount = computed(
  () => props.list.filter((item) => item.readiness_status === 'pending').length
)
const importQueueCount = computed(() => props.importQueue.length)
const heroTitle = computed(() => props.mode === 'import' ? '导入 AWD 题目包' : 'AWD 题目库')
const heroSummary = computed(() =>
  props.mode === 'import'
    ? '上传符合规范的 AWD 题目包，确认后生成可用于编排的 AWD 题目。'
    : '管理 AWD 赛事使用的题目。'
)
const hasActiveFilters = computed(() =>
  Boolean(props.keyword.trim() || props.serviceTypeFilter || props.statusFilter)
)

const awdChallengeTableColumns = [
  {
    key: 'name',
    label: '题目名称',
    widthClass: 'w-[28%] min-w-[16rem]',
    cellClass: 'awd-challenge-table__name-cell',
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
    case 'published': return '已发布'
    case 'archived': return '已归档'
    case 'draft':
    default: return '草稿'
  }
}

function getReadinessLabel(value: AdminAwdChallengeData['readiness_status']): string {
  switch (value) {
    case 'passed': return '已通过'
    case 'failed': return '未通过'
    case 'pending':
    default: return '待验证'
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
  emit('updateServiceTypeFilter', target instanceof HTMLSelectElement ? target.value as AwdServiceTypeFilter : '')
}

function handleStatusFilterChange(event: Event): void {
  const target = event.target
  emit('updateStatusFilter', target instanceof HTMLSelectElement ? target.value as AwdServiceStatusFilter : '')
}

function handleSelectImportPackages(files: File[]) {
  emit('selectImportPackages', files)
}

function formatStructuredJSON(value?: Record<string, unknown>): string {
  if (!value || Object.keys(value).length === 0) {
    return '{}'
  }
  return JSON.stringify(value, null, 2)
}

</script>

<template>
  <div class="workspace-shell journal-shell journal-shell-admin journal-hero awd-challenge-library-shell">
    <div class="workspace-grid">
      <main class="content-pane awd-challenge-library-content">
        <section class="workspace-hero">
          <div class="workspace-tab-heading__main">
            <div class="workspace-overline">
              AWD Service Authoring
            </div>
            <h1 class="hero-title">
              {{ heroTitle }}
            </h1>
            <p class="hero-summary">
              {{ heroSummary }}
            </p>

            <div
              v-if="mode === 'import'"
              class="awd-import-page-note"
            >
              上传题目包并确认导入后，系统会生成可用于 AWD 编排的题目。
            </div>
          </div>

          <div class="awd-library-hero-actions">
            <div class="quick-actions">
              <button
                v-if="mode === 'library'"
                type="button"
                class="ui-btn ui-btn--ghost"
                @click="emit('refresh')"
              >
                <RefreshCw class="h-4 w-4" />
                刷新列表
              </button>
              <button
                v-if="mode === 'library'"
                id="awd-challenge-open-import"
                type="button"
                class="ui-btn ui-btn--primary"
                @click="emit('openImportPage')"
              >
                <Upload class="h-4 w-4" />
                导入题目包
              </button>
              <button
                v-if="mode === 'import'"
                type="button"
                class="ui-btn ui-btn--ghost"
                @click="emit('refreshImportQueue')"
              >
                <RefreshCw class="h-4 w-4" />
                刷新队列
              </button>
            </div>
          </div>
        </section>

        <div>
          <div
            v-if="mode === 'library'"
            class="awd-library-pane"
          >
            <div class="admin-summary-grid awd-challenge-summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface">
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
                        class="awd-filter-control"
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
                        class="awd-filter-control"
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

              <div
                v-if="loading && list.length === 0"
                class="flex justify-center py-12"
              >
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
                  class="awd-challenge-list workspace-directory-list"
                  :columns="awdChallengeTableColumns"
                  :rows="list"
                  row-key="id"
                  row-class="awd-challenge-table-row group"
                >
                  <template #cell-name="{ row }">
                    <div class="awd-challenge-table__name">
                      <div class="awd-challenge-table__title">
                        {{ (row as AdminAwdChallengeData).name }}
                      </div>
                      <div class="awd-challenge-table__slug">
                        {{ (row as AdminAwdChallengeData).slug }}
                      </div>
                    </div>
                  </template>

                  <template #cell-service_type="{ row }">
                    <span class="awd-challenge-table__mono">{{ getServiceTypeLabel((row as AdminAwdChallengeData).service_type) }}</span>
                  </template>

                  <template #cell-deployment_mode="{ row }">
                    <span class="awd-challenge-table__compact-text">{{ getDeploymentModeLabel((row as AdminAwdChallengeData).deployment_mode) }}</span>
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
                      class="awd-status-pill"
                      :class="getReadinessClass((row as AdminAwdChallengeData).readiness_status)"
                    >
                      {{ getReadinessLabel((row as AdminAwdChallengeData).readiness_status) }}
                    </span>
                  </template>

                  <template #cell-status="{ row }">
                    <span
                      class="awd-status-pill"
                      :class="getStatusClass((row as AdminAwdChallengeData).status)"
                    >
                      {{ getStatusLabel((row as AdminAwdChallengeData).status) }}
                    </span>
                  </template>

                  <template #cell-actions="{ row }">
                    <div class="awd-challenge-table__actions">
                      <button
                        type="button"
                        class="awd-row-btn"
                        @click="emit('openEditDialog', row as AdminAwdChallengeData)"
                      >
                        编辑
                      </button>
                      <button
                        type="button"
                        class="awd-row-btn awd-row-btn--danger"
                        @click="emit('deleteChallenge', row as AdminAwdChallengeData)"
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
          </div>

          <div
            v-if="mode === 'import'"
            class="awd-import-pane"
          >
            <section class="workspace-directory-section awd-import-tool-section">
              <header class="list-heading awd-challenge-import__head">
                <div>
                  <div class="workspace-overline">
                    Ingestion
                  </div>
                  <h2 class="list-heading__title">
                    导入 AWD 题目包
                  </h2>
                  <p class="hero-summary awd-challenge-import__copy">
                    教师按统一题目包规范写好 `challenge.yml` 后，从这里导入 AWD 题目。
                  </p>
                </div>
                <div class="awd-challenge-import__head-actions">
                  <div class="quick-actions">
                    <a
                      class="ui-btn ui-btn--ghost"
                      href="/downloads/awd-challenge-package-sample-v1.zip"
                      download="awd-challenge-package-sample-v1.zip"
                    >
                      下载示例题包
                    </a>
                  </div>
                </div>
              </header>

              <div class="awd-challenge-import__entry">
                <ChallengePackageImportEntry
                  :hide-header="true"
                  :uploading="uploading"
                  :selected-file-name="selectedFileName"
                  @select="handleSelectImportPackages"
                />
              </div>

              <div
                v-if="uploadResults.length > 0"
                class="awd-challenge-import__uploads"
              >
                <article
                  v-for="item in uploadResults"
                  :key="item.id"
                  class="awd-challenge-import__upload"
                  :class="item.status === 'success' ? 'is-success' : 'is-error'"
                >
                  <div class="awd-challenge-import__upload-head">
                    <strong>{{ item.fileName }}</strong>
                    <span>{{ item.status === 'success' ? '成功' : '失败' }}</span>
                  </div>
                  <p>{{ item.message }}</p>
                </article>
              </div>
            </section>

            <section class="workspace-directory-section awd-import-queue-section">
              <header class="list-heading awd-challenge-import__queue-head">
                <div>
                  <div class="workspace-overline">
                    Review Queue
                  </div>
                  <h2 class="list-heading__title">
                    待确认题目包
                  </h2>
                </div>
                <span class="awd-challenge-import__queue-count">共 {{ importQueueCount }} 个待确认包</span>
              </header>

              <div
                v-if="queueLoading"
                class="awd-challenge-import__state"
              >
                正在同步导入队列...
              </div>
              <AppEmpty
                v-else-if="importQueue.length === 0"
                class="awd-challenge-import__empty"
                icon="Box"
                title="队列为空"
                description="上传题目包后，待确认的项将出现在此处。"
              />
              <div
                v-else
                class="workspace-directory-list awd-challenge-import__queue"
              >
                <article
                  v-for="item in importQueue"
                  :key="item.id"
                  class="awd-challenge-import__card"
                >
                  <div class="awd-challenge-import__card-head">
                    <div>
                      <h3 class="awd-challenge-import__card-title">
                        {{ item.title }}
                      </h3>
                      <p class="awd-challenge-import__card-file">
                        {{ item.file_name }}
                      </p>
                    </div>
                    <button
                      type="button"
                      class="ui-btn ui-btn--primary"
                      @click="emit('commitImport', item)"
                    >
                      确认导入
                    </button>
                  </div>

                  <div class="awd-challenge-import__chips">
                    <span class="awd-status-pill awd-status-pill--primary">{{ item.service_type }}</span>
                    <span class="awd-status-pill awd-status-pill--warning">{{ item.deployment_mode }}</span>
                    <span class="awd-status-pill awd-status-pill--muted">{{ item.flag_mode || '未定义 flag_mode' }}</span>
                    <span class="awd-status-pill awd-status-pill--success">{{ item.defense_entry_mode || '未定义入口' }}</span>
                  </div>

                  <div class="awd-challenge-import__grid">
                    <pre class="awd-challenge-import__json">{{ formatStructuredJSON(item.access_config) }}</pre>
                    <pre class="awd-challenge-import__json">{{ formatStructuredJSON(item.runtime_config) }}</pre>
                  </div>
                </article>
              </div>
            </section>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.awd-challenge-library-content,
.awd-library-pane,
.awd-import-pane {
  display: flex;
  flex-direction: column;
  gap: var(--workspace-directory-page-block-gap, var(--space-5));
}

.workspace-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-7);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid var(--workspace-line-soft);
}

.hero-title {
  margin: 0.5rem 0 0;
  font-size: var(--workspace-page-title-font-size);
  line-height: var(--workspace-page-title-line-height);
  letter-spacing: var(--workspace-page-title-letter-spacing);
  color: var(--journal-ink);
}

.hero-summary {
  max-width: 760px;
  margin-top: var(--space-3-5);
  font-size: var(--font-size-15);
  line-height: 1.9;
  color: var(--journal-muted);
}

.awd-library-hero-actions {
  display: flex;
  align-items: flex-end;
  padding-bottom: 0.5rem;
}

.quick-actions {
  display: flex;
  gap: var(--space-3);
}

.awd-import-page-note {
  max-width: 46rem;
  margin-top: var(--space-4);
  font-size: var(--font-size-13);
  line-height: 1.7;
  color: var(--journal-muted);
}

.awd-challenge-summary {
  --admin-summary-grid-columns: repeat(4, minmax(0, 1fr));
  --metric-panel-columns: repeat(4, minmax(0, 1fr));
}

.awd-challenge-import__uploads { display: grid; gap: var(--space-3); margin-top: var(--space-4); }
.awd-challenge-import__upload { padding: var(--space-4); border-radius: 1rem; border: 1px solid var(--color-border-default); background: var(--color-bg-surface); }
.awd-challenge-import__upload.is-success { border-color: color-mix(in srgb, var(--color-success) 24%, transparent); }
.awd-challenge-import__upload-head { display: flex; align-items: center; justify-content: space-between; gap: var(--space-3); margin-bottom: var(--space-2); }
.awd-challenge-import__entry { margin-top: var(--workspace-directory-gap-top); }
.awd-challenge-import__queue-head { margin-bottom: 0; }
.awd-challenge-import__queue-count { color: var(--color-text-muted); font-size: var(--font-size-13); font-weight: 700; }
.awd-challenge-import__queue { display: grid; gap: 0; padding: 0; }
.awd-challenge-import__card { display: grid; gap: var(--space-4); padding: var(--space-4-5) var(--space-5); border-bottom: 1px solid var(--workspace-directory-row-divider); }
.awd-challenge-import__card:last-child { border-bottom: 0; }
.awd-challenge-import__card-title { margin: 0; font-size: var(--font-size-17); font-weight: 800; color: var(--color-text-primary); }
.awd-challenge-import__card-file { margin: var(--space-1) 0 0; color: var(--color-text-muted); font-family: var(--font-family-mono); font-size: var(--font-size-12); }
.awd-challenge-import__chips { display: flex; flex-wrap: wrap; gap: var(--space-2); }
.awd-challenge-import__grid { display: grid; gap: var(--space-3); grid-template-columns: repeat(2, minmax(0, 1fr)); }
.awd-challenge-import__json { margin: 0; min-height: 10rem; padding: var(--space-4); border-radius: 1rem; background: var(--color-bg-elevated); border: 1px solid var(--color-border-subtle); color: var(--color-text-secondary); font-family: var(--font-family-mono); font-size: var(--font-size-12); line-height: 1.6; white-space: pre-wrap; word-break: break-word; }

.awd-challenge-library__filter-grid { display: grid; gap: var(--space-4); }
.awd-challenge-library__filter-field { display: grid; gap: var(--space-2); }
.awd-challenge-library__filter-label { font-size: var(--font-size-11); font-weight: 800; letter-spacing: 0.1em; text-transform: uppercase; color: var(--color-text-muted); }
.awd-filter-control { width: 100%; min-height: 2.85rem; padding: 0 var(--space-4); font-size: var(--font-size-14); font-weight: 600; border-radius: 0.9rem; border: 1px solid var(--color-border-default); background: var(--color-bg-surface); color: var(--color-text-primary); outline: none; transition: all 180ms ease; }
.awd-filter-control:focus { border-color: var(--color-primary); box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 14%, transparent); }

.awd-challenge-table__name { display: flex; flex-direction: column; gap: 0.25rem; }
.awd-challenge-table__title { font-size: var(--font-size-15); font-weight: 700; color: var(--color-text-primary); }
.awd-challenge-table__slug { font-family: var(--font-family-mono); font-size: var(--font-size-12); color: var(--color-text-muted); }
.awd-challenge-table__mono { font-family: var(--font-family-mono); font-size: var(--font-size-13); font-weight: 700; color: var(--color-text-primary); }
.awd-challenge-table__difficulty { font-size: var(--font-size-13); font-weight: 700; color: var(--color-text-muted); }
.awd-challenge-table__compact-text { font-size: var(--font-size-13); color: var(--color-text-primary); }

.awd-status-pill { display: inline-flex; align-items: center; justify-content: center; min-height: 1.85rem; min-width: 4.8rem; padding: 0 var(--space-3); border: 1px solid transparent; border-radius: 999px; font-size: var(--font-size-12); font-weight: 800; }
.awd-status-pill--success { border-color: color-mix(in srgb, var(--color-success) 22%, transparent); background: color-mix(in srgb, var(--color-success) 8%, transparent); color: var(--color-success); }
.awd-status-pill--primary { border-color: color-mix(in srgb, var(--color-primary) 22%, transparent); background: color-mix(in srgb, var(--color-primary) 8%, transparent); color: var(--color-primary); }
.awd-status-pill--warning { border-color: color-mix(in srgb, var(--color-warning) 22%, transparent); background: color-mix(in srgb, var(--color-warning) 8%, transparent); color: var(--color-warning); }
.awd-status-pill--danger { border-color: color-mix(in srgb, var(--color-danger) 22%, transparent); background: color-mix(in srgb, var(--color-danger) 8%, transparent); color: var(--color-danger); }
.awd-status-pill--muted { border-color: var(--color-border-default); background: var(--color-bg-elevated); color: var(--color-text-muted); }

.awd-challenge-table__actions { display: flex; align-items: center; justify-content: flex-end; gap: var(--space-2); }
.awd-row-btn { display: inline-flex; align-items: center; justify-content: center; min-height: 1.9rem; padding: 0 var(--space-3); border: 1px solid var(--color-border-default); border-radius: 9px; background: var(--color-bg-surface); font-size: var(--font-size-12); font-weight: 800; color: var(--color-text-secondary); transition: all 0.2s ease; }
.awd-row-btn:hover { border-color: var(--color-primary); background: var(--color-primary-soft); color: var(--color-primary); transform: translateY(-1px); }
.awd-row-btn--danger:hover { border-color: var(--color-danger); background: color-mix(in srgb, var(--color-danger) 8%, var(--color-bg-surface)); color: var(--color-danger); }

@media (max-width: 1024px) {
  .awd-challenge-import__grid { grid-template-columns: 1fr; }
  .awd-challenge-table__actions { flex-direction: column; align-items: stretch; }
  .workspace-hero { grid-template-columns: 1fr; }
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
