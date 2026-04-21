<script setup lang="ts">
import { computed, ref } from 'vue'
import {
  Activity,
  Box,
  CheckCircle,
  Clock,
  Plus,
  RefreshCw,
} from 'lucide-vue-next'

import type {
  AdminAwdServiceTemplateData,
  AdminAwdServiceTemplateImportPreview,
} from '@/api/contracts'
import ChallengePackageImportEntry from '@/components/platform/challenge/ChallengePackageImportEntry.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import WorkspaceDirectoryToolbar from '@/components/common/WorkspaceDirectoryToolbar.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import type { PlatformAwdServiceTemplateImportUploadResult } from '@/composables/usePlatformAwdServiceTemplates'

type AwdServiceTypeFilter = AdminAwdServiceTemplateData['service_type'] | ''
type AwdServiceStatusFilter = AdminAwdServiceTemplateData['status'] | ''
type LibraryTab = 'library' | 'import'

const props = defineProps<{
  list: AdminAwdServiceTemplateData[]
  total: number
  page: number
  pageSize: number
  loading: boolean
  keyword: string
  serviceTypeFilter: AwdServiceTypeFilter
  statusFilter: AwdServiceStatusFilter
  uploading: boolean
  queueLoading: boolean
  importQueue: AdminAwdServiceTemplateImportPreview[]
  uploadResults: PlatformAwdServiceTemplateImportUploadResult[]
  selectedFileName?: string
}>()

const emit = defineEmits<{
  refresh: []
  refreshImportQueue: []
  updateKeyword: [value: string]
  updateServiceTypeFilter: [value: AwdServiceTypeFilter]
  updateStatusFilter: [value: AwdServiceStatusFilter]
  selectImportPackages: [files: File[]]
  commitImport: [preview: AdminAwdServiceTemplateImportPreview]
  openCreateDialog: []
  openEditDialog: [template: AdminAwdServiceTemplateData]
  deleteTemplate: [template: AdminAwdServiceTemplateData]
  changePage: [page: number]
}>()

const activeTab = ref<LibraryTab>('library')

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const publishedCount = computed(() => props.list.filter((item) => item.status === 'published').length)
const webHttpCount = computed(() => props.list.filter((item) => item.service_type === 'web_http').length)
const pendingReadinessCount = computed(
  () => props.list.filter((item) => item.readiness_status === 'pending').length
)
const importQueueCount = computed(() => props.importQueue.length)
const hasActiveFilters = computed(() =>
  Boolean(props.keyword.trim() || props.serviceTypeFilter || props.statusFilter)
)

const templateTableColumns = [
  {
    key: 'name',
    label: '模板名称',
    widthClass: 'w-[28%] min-w-[16rem]',
    cellClass: 'awd-template-table__name-cell',
  },
  {
    key: 'service_type',
    label: '类型',
    align: 'center' as const,
    widthClass: 'w-[12%] min-w-[7rem]',
    cellClass: 'awd-template-table__compact-cell',
  },
  {
    key: 'deployment_mode',
    label: '部署方式',
    align: 'center' as const,
    widthClass: 'w-[12%] min-w-[7rem]',
    cellClass: 'awd-template-table__compact-cell',
  },
  {
    key: 'difficulty',
    label: '难度',
    align: 'center' as const,
    widthClass: 'w-[10%] min-w-[6rem]',
    cellClass: 'awd-template-table__compact-cell',
  },
  {
    key: 'readiness_status',
    label: '就绪度',
    align: 'center' as const,
    widthClass: 'w-[10%] min-w-[6rem]',
    cellClass: 'awd-template-table__compact-cell',
  },
  {
    key: 'status',
    label: '状态',
    align: 'center' as const,
    widthClass: 'w-[10%] min-w-[6rem]',
    cellClass: 'awd-template-table__compact-cell',
  },
  {
    key: 'actions',
    label: '操作',
    align: 'right' as const,
    widthClass: 'w-[10rem]',
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
  return value === 'topology' ? 'Topology' : 'Single'
}

function getDifficultyLabel(value: AdminAwdServiceTemplateData['difficulty']): string {
  switch (value) {
    case 'beginner': return '入门'
    case 'easy': return '简单'
    case 'medium': return '中等'
    case 'hard': return '困难'
    case 'insane': return '高强度'
    default: return value
  }
}

function getStatusLabel(value: AdminAwdServiceTemplateData['status']): string {
  switch (value) {
    case 'published': return '已发布'
    case 'archived': return '已归档'
    case 'draft':
    default: return '草稿'
  }
}

function getReadinessLabel(value: AdminAwdServiceTemplateData['readiness_status']): string {
  switch (value) {
    case 'passed': return '已通过'
    case 'failed': return '未通过'
    case 'pending':
    default: return '待验证'
  }
}

function getStatusClass(status: AdminAwdServiceTemplateData['status']): string {
  if (status === 'published') return 'awd-status-pill--success'
  if (status === 'archived') return 'awd-status-pill--muted'
  return 'awd-status-pill--primary'
}

function getReadinessClass(readiness: AdminAwdServiceTemplateData['readiness_status']): string {
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

function selectTab(tab: LibraryTab) {
  activeTab.value = tab
}
</script>

<template>
  <section class="journal-shell journal-shell-admin journal-notes-card journal-hero workspace-shell flex min-h-full flex-1 flex-col">
    <header class="admin-workbench-header">
      <div class="admin-workbench-header__top">
        <div class="admin-workbench-header__identity">
          <div class="workspace-overline">
            AWD Service Authoring
          </div>
          <h1 class="admin-workbench-header__title">
            AWD 服务模板
          </h1>
          <p class="admin-workbench-header__description">
            这里单独维护 AWD 题目的服务模板，不再和解题赛题目混在同一资源目录。
          </p>
        </div>

        <div class="admin-workbench-header__actions">
          <button
            type="button"
            class="ui-btn ui-btn--ghost ui-btn--sm"
            @click="emit('refresh')"
          >
            <RefreshCw class="h-3.5 w-3.5" />
            刷新列表
          </button>
          <button
            id="awd-template-open-create"
            type="button"
            class="ui-btn ui-btn--primary ui-btn--sm"
            @click="emit('openCreateDialog')"
          >
            <Plus class="h-3.5 w-3.5" />
            创建模板
          </button>
        </div>
      </div>

      <nav class="admin-workbench-header__nav">
        <button
          class="admin-workbench-nav-item"
          :class="{ 'is-active': activeTab === 'library' }"
          @click="selectTab('library')"
        >
          全部模板
        </button>
        <button
          class="admin-workbench-nav-item"
          :class="{ 'is-active': activeTab === 'import' }"
          @click="selectTab('import')"
        >
          题目包导入
          <span
            v-if="importQueueCount > 0"
            class="admin-workbench-nav-badge"
          >{{ importQueueCount }}</span>
        </button>
      </nav>
    </header>

    <main class="awd-library-content flex-1">
      <!-- Tab A: Library -->
      <div
        v-if="activeTab === 'library'"
        class="awd-library-pane space-y-6"
      >
        <div class="metric-panel-grid metric-panel-grid--premium cols-4">
          <article class="metric-panel-card metric-panel-card--premium">
            <div class="metric-panel-label">
              <span>模板总量</span>
              <Box class="h-4 w-4" />
            </div>
            <div class="metric-panel-value">
              {{ total.toString().padStart(2, '0') }}
            </div>
          </article>

          <article class="metric-panel-card metric-panel-card--premium">
            <div class="metric-panel-label">
              <span>已发布</span>
              <CheckCircle class="h-4 w-4" />
            </div>
            <div class="metric-panel-value">
              {{ publishedCount.toString().padStart(2, '0') }}
            </div>
          </article>

          <article class="metric-panel-card metric-panel-card--premium">
            <div class="metric-panel-label">
              <span>Web HTTP</span>
              <Activity class="h-4 w-4" />
            </div>
            <div class="metric-panel-value">
              {{ webHttpCount.toString().padStart(2, '0') }}
            </div>
          </article>

          <article class="metric-panel-card metric-panel-card--premium">
            <div class="metric-panel-label">
              <span>待验证</span>
              <Clock class="h-4 w-4" />
            </div>
            <div class="metric-panel-value">
              {{ pendingReadinessCount.toString().padStart(2, '0') }}
            </div>
          </article>
        </div>

        <section class="workspace-directory-section">
          <WorkspaceDirectoryToolbar
            :model-value="keyword"
            :total="total"
            selected-sort-label=""
            :sort-options="[]"
            search-placeholder="检索模板名称、Slug 或描述..."
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
                    class="awd-filter-control"
                    @change="handleServiceTypeFilterChange"
                  >
                    <option value="">全部类型</option>
                    <option value="web_http">Web HTTP</option>
                    <option value="binary_tcp">Binary TCP</option>
                    <option value="multi_container">Multi Container</option>
                  </select>
                </label>

                <label class="awd-template-library__filter-field">
                  <span class="awd-template-library__filter-label">发布状态</span>
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
            <AppLoading>正在同步模板数据...</AppLoading>
          </div>

          <template v-else>
            <AppEmpty
              v-if="list.length === 0"
              class="awd-template-library__empty"
              icon="Box"
              title="暂无服务模板"
              :description="hasActiveFilters ? '当前筛选条件下没有匹配模板。' : '还没有 AWD 模板，请先点击右上角创建。'"
            />

            <WorkspaceDataTable
              v-else
              class="awd-template-list workspace-directory-list"
              :columns="templateTableColumns"
              :rows="list"
              row-key="id"
              row-class="awd-template-table-row group"
            >
              <template #cell-name="{ row }">
                <div class="awd-template-table__name">
                  <div class="awd-template-table__title">
                    {{ (row as AdminAwdServiceTemplateData).name }}
                  </div>
                  <div class="awd-template-table__slug">
                    @{{ (row as AdminAwdServiceTemplateData).slug }}
                  </div>
                </div>
              </template>

              <template #cell-service_type="{ row }">
                <span class="awd-template-table__mono">{{ getServiceTypeLabel((row as AdminAwdServiceTemplateData).service_type) }}</span>
              </template>

              <template #cell-deployment_mode="{ row }">
                <span class="awd-template-table__compact-text">{{ getDeploymentModeLabel((row as AdminAwdServiceTemplateData).deployment_mode) }}</span>
              </template>

              <template #cell-difficulty="{ row }">
                <span class="awd-template-table__difficulty">{{ getDifficultyLabel((row as AdminAwdServiceTemplateData).difficulty) }}</span>
              </template>

              <template #cell-readiness_status="{ row }">
                <span
                  class="awd-status-pill"
                  :class="getReadinessClass((row as AdminAwdServiceTemplateData).readiness_status)"
                >
                  {{ getReadinessLabel((row as AdminAwdServiceTemplateData).readiness_status) }}
                </span>
              </template>

              <template #cell-status="{ row }">
                <span
                  class="awd-status-pill"
                  :class="getStatusClass((row as AdminAwdServiceTemplateData).status)"
                >
                  {{ getStatusLabel((row as AdminAwdServiceTemplateData).status) }}
                </span>
              </template>

              <template #cell-actions="{ row }">
                <div class="awd-template-table__actions">
                  <button
                    type="button"
                    class="awd-row-btn"
                    @click="emit('openEditDialog', row as AdminAwdServiceTemplateData)"
                  >
                    编辑
                  </button>
                  <button
                    type="button"
                    class="awd-row-btn awd-row-btn--danger"
                    @click="emit('deleteTemplate', row as AdminAwdServiceTemplateData)"
                  >
                    删除
                  </button>
                </div>
              </template>
            </WorkspaceDataTable>

            <div
              v-if="total > 0"
              class="admin-pagination workspace-directory-pagination mt-6"
            >
              <WorkspaceDirectoryPagination
                :page="page"
                :total-pages="totalPages"
                :total="total"
                :disabled="loading"
                total-label="个模板"
                @change-page="emit('changePage', $event)"
              />
            </div>
          </template>
        </section>
      </div>

      <!-- Tab B: Import -->
      <div
        v-if="activeTab === 'import'"
        class="awd-import-pane space-y-8"
      >
        <section class="awd-import-tool-section">
          <header class="list-heading awd-template-import__head">
            <div>
              <div class="workspace-overline">
                Ingestion
              </div>
              <h2 class="list-heading__title">
                导入 AWD 题目包
              </h2>
              <p class="workspace-page-copy awd-template-import__copy">
                教师按统一题目包规范写好 `challenge.yml` 后，从这里导入完整模板。
              </p>
            </div>
            <div class="awd-template-import__head-actions">
              <a
                class="ui-btn ui-btn--ghost"
                href="/downloads/awd-service-template-package-sample-v1.zip"
                download="awd-service-template-package-sample-v1.zip"
              >
                下载示例题包
              </a>
              <button
                type="button"
                class="ui-btn ui-btn--ghost"
                @click="emit('refreshImportQueue')"
              >
                <RefreshCw class="h-4 w-4" />
                刷新队列
              </button>
            </div>
          </header>

          <ChallengePackageImportEntry
            :hide-header="true"
            :uploading="uploading"
            :selected-file-name="selectedFileName"
            @select="handleSelectImportPackages"
          />

          <div
            v-if="uploadResults.length > 0"
            class="awd-template-import__uploads"
          >
            <article
              v-for="item in uploadResults"
              :key="item.id"
              class="awd-template-import__upload"
              :class="item.status === 'success' ? 'is-success' : 'is-error'"
            >
              <div class="awd-template-import__upload-head">
                <strong>{{ item.fileName }}</strong>
                <span>{{ item.status === 'success' ? '成功' : '失败' }}</span>
              </div>
              <p>{{ item.message }}</p>
            </article>
          </div>
        </section>

        <section class="awd-import-queue-section">
          <div class="awd-template-import__queue-head">
            <div class="workspace-overline">
              Review Queue
            </div>
            <span class="awd-template-import__queue-count">共 {{ importQueueCount }} 个待确认包</span>
          </div>

          <div
            v-if="queueLoading"
            class="awd-template-import__state"
          >
            正在同步导入队列...
          </div>
          <AppEmpty
            v-else-if="importQueue.length === 0"
            class="awd-template-import__empty"
            icon="Box"
            title="队列为空"
            description="上传题目包后，待确认的项将出现在此处。"
          />
          <div
            v-else
            class="awd-template-import__queue"
          >
            <article
              v-for="item in importQueue"
              :key="item.id"
              class="awd-template-import__card"
            >
              <div class="awd-template-import__card-head">
                <div>
                  <h3 class="awd-template-import__card-title">
                    {{ item.title }}
                  </h3>
                  <p class="awd-template-import__card-file">
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

              <div class="awd-template-import__chips">
                <span class="awd-status-pill awd-status-pill--primary">{{ item.service_type }}</span>
                <span class="awd-status-pill awd-status-pill--warning">{{ item.deployment_mode }}</span>
                <span class="awd-status-pill awd-status-pill--muted">{{ item.flag_mode || '未定义 flag_mode' }}</span>
                <span class="awd-status-pill awd-status-pill--success">{{ item.defense_entry_mode || '未定义入口' }}</span>
              </div>

              <div class="awd-template-import__grid">
                <pre class="awd-template-import__json">{{ formatStructuredJSON(item.access_config) }}</pre>
                <pre class="awd-template-import__json">{{ formatStructuredJSON(item.runtime_config) }}</pre>
              </div>
            </article>
          </div>
        </section>
      </div>
    </main>
  </section>
</template>

<style scoped>
.awd-library-content {
  padding: 2rem 3rem;
}

.awd-template-import__copy { margin: 0.5rem 0 0; max-width: 48rem; }
.awd-template-import__uploads { display: grid; gap: 0.75rem; margin-top: 1.5rem; }
.awd-template-import__upload { padding: 1.1rem; border-radius: 1rem; border: 1px solid var(--color-border-default); background: var(--color-bg-surface); }
.awd-template-import__upload.is-success { border-color: color-mix(in srgb, var(--color-success) 24%, transparent); }
.awd-template-import__upload-head { display: flex; align-items: center; justify-content: space-between; gap: 0.75rem; margin-bottom: 0.5rem; }
.awd-template-import__queue-head { display: flex; align-items: center; justify-content: space-between; gap: 0.75rem; margin-bottom: 1.5rem; }
.awd-template-import__queue-count { color: var(--color-text-muted); font-size: 13px; font-weight: 700; }
.awd-template-import__queue { display: grid; gap: 1.75rem; }
.awd-template-import__card { display: grid; gap: 1.5rem; padding: 1.75rem; border: 1px solid var(--color-border-default); border-radius: 1.25rem; background: var(--color-bg-surface); box-shadow: var(--color-shadow-soft); }
.awd-template-import__card-title { margin: 0; font-size: 1.15rem; font-weight: 900; color: var(--color-text-primary); }
.awd-template-import__card-file { margin: 0.35rem 0 0; color: var(--color-text-muted); font-family: var(--font-family-mono); font-size: 12px; }
.awd-template-import__chips { display: flex; flex-wrap: wrap; gap: 0.6rem; }
.awd-template-import__grid { display: grid; gap: 1.25rem; grid-template-columns: repeat(2, minmax(0, 1fr)); }
.awd-template-import__json { margin: 0; min-height: 10rem; padding: 1.25rem; border-radius: 1rem; background: var(--color-bg-elevated); border: 1px solid var(--color-border-subtle); color: var(--color-text-secondary); font-family: var(--font-family-mono); font-size: 12px; line-height: 1.6; white-space: pre-wrap; word-break: break-word; }

.awd-template-library__filter-grid { display: grid; gap: 1.25rem; }
.awd-template-library__filter-field { display: grid; gap: 0.6rem; }
.awd-template-library__filter-label { font-size: 11px; font-weight: 800; letter-spacing: 0.1em; text-transform: uppercase; color: var(--color-text-muted); }
.awd-filter-control { width: 100%; min-height: 2.85rem; padding: 0 1rem; font-size: 14px; font-weight: 600; border-radius: 0.9rem; border: 1px solid var(--color-border-default); background: var(--color-bg-surface); color: var(--color-text-primary); outline: none; transition: all 180ms ease; }
.awd-filter-control:focus { border-color: var(--color-primary); box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 14%, transparent); }

.awd-template-table__name { display: flex; flex-direction: column; gap: 0.25rem; }
.awd-template-table__title { font-size: 15px; font-weight: 700; color: var(--color-text-primary); }
.awd-template-table__slug { font-family: var(--font-family-mono); font-size: 12px; color: var(--color-text-muted); }
.awd-template-table__mono { font-family: var(--font-family-mono); font-size: 13px; font-weight: 700; color: var(--color-text-primary); }
.awd-template-table__difficulty { font-size: 13px; font-weight: 700; color: var(--color-text-muted); }
.awd-template-table__compact-text { font-size: 13px; color: var(--color-text-primary); }

.awd-status-pill { display: inline-flex; align-items: center; justify-content: center; min-height: 1.85rem; min-width: 4.8rem; padding: 0 0.75rem; border: 1px solid transparent; border-radius: 999px; font-size: 12px; font-weight: 800; }
.awd-status-pill--success { border-color: color-mix(in srgb, var(--color-success) 22%, transparent); background: color-mix(in srgb, var(--color-success) 8%, transparent); color: var(--color-success); }
.awd-status-pill--primary { border-color: color-mix(in srgb, var(--color-primary) 22%, transparent); background: color-mix(in srgb, var(--color-primary) 8%, transparent); color: var(--color-primary); }
.awd-status-pill--warning { border-color: color-mix(in srgb, var(--color-warning) 22%, transparent); background: color-mix(in srgb, var(--color-warning) 8%, transparent); color: var(--color-warning); }
.awd-status-pill--danger { border-color: color-mix(in srgb, var(--color-danger) 22%, transparent); background: color-mix(in srgb, var(--color-danger) 8%, transparent); color: var(--color-danger); }
.awd-status-pill--muted { border-color: var(--color-border-default); background: var(--color-bg-elevated); color: var(--color-text-muted); }

.awd-template-table__actions { display: flex; align-items: center; justify-content: flex-end; gap: 0.6rem; }
.awd-row-btn { display: inline-flex; align-items: center; justify-content: center; min-height: 1.9rem; padding: 0 0.95rem; border: 1px solid var(--color-border-default); border-radius: 9px; background: var(--color-bg-surface); font-size: 12px; font-weight: 800; color: var(--color-text-secondary); transition: all 0.2s ease; }
.awd-row-btn:hover { border-color: var(--color-primary); background: var(--color-primary-soft); color: var(--color-primary); transform: translateY(-1px); }
.awd-row-btn--danger:hover { border-color: var(--color-danger); background: color-mix(in srgb, var(--color-danger) 8%, var(--color-bg-surface)); color: var(--color-danger); }

@media (max-width: 1024px) {
  .awd-template-import__grid { grid-template-columns: 1fr; }
  .awd-template-table__actions { flex-direction: column; align-items: stretch; }
}
</style>