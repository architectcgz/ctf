<script setup lang="ts">
import { ArrowDownWideNarrow, Calendar, UserRound } from 'lucide-vue-next'
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getAuditLogs } from '@/api/admin'
import type { AuditLogItem } from '@/api/contracts'
import AdminSurfaceModal from '@/components/common/modal-templates/AdminSurfaceModal.vue'
import PlatformPaginationControls from '@/components/platform/PlatformPaginationControls.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryToolbar, {
  type WorkspaceDirectorySortOption,
} from '@/components/common/WorkspaceDirectoryToolbar.vue'
import { useAbortController } from '@/composables/useAbortController'

type AuditSortKey = 'created_at' | 'action' | 'actor'
type AuditSortOption = WorkspaceDirectorySortOption & {
  key: AuditSortKey
  order: 'asc' | 'desc'
}

const route = useRoute()
const router = useRouter()

const filters = reactive({
  action: '',
  resource_type: '',
  actor_user_id: '',
})

const list = ref<AuditLogItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const error = ref<string | null>(null)
const keyword = ref('')
const activeActorLog = ref<AuditLogItem | null>(null)
let textFilterTimer: ReturnType<typeof setTimeout> | null = null
let suppressAutoApply = false
const autoApplyReady = ref(false)
const { createController, abort } = useAbortController()
let latestLogsRequestId = 0

const sortOptions: AuditSortOption[] = [
  { key: 'created_at', order: 'desc', label: '最近操作', icon: Calendar },
  { key: 'action', order: 'asc', label: '动作顺序', icon: ArrowDownWideNarrow },
  { key: 'actor', order: 'asc', label: '执行人顺序', icon: UserRound },
]
const sortConfig = ref<AuditSortOption>(sortOptions[0]!)
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

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const hasActiveFilters = computed(() =>
  Boolean(keyword.value.trim() || filters.action || filters.resource_type || filters.actor_user_id)
)
const filteredRows = computed<AuditLogItem[]>(() => {
  const normalizedKeyword = keyword.value.trim().toLowerCase()
  const nextRows = list.value.filter((item) => {
    if (!normalizedKeyword) {
      return true
    }

    const resourceLabel = resourceDisplayName(item)
    const actorLabel = `${actorDisplayName(item)} ${item.actor_user_id || ''}`
    const detailLabel = detailPreview(item.detail)

    return [item.action, resourceLabel, actorLabel, detailLabel, formatDate(item.created_at)].some(
      (value) => value.toLowerCase().includes(normalizedKeyword)
    )
  })

  const sortedRows = [...nextRows]
  sortedRows.sort((left, right) => {
    switch (sortConfig.value.key) {
      case 'action': {
        const delta = left.action.localeCompare(right.action, 'zh-CN')
        return sortConfig.value.order === 'asc' ? delta : -delta
      }
      case 'actor': {
        const delta = (left.actor_username || '').localeCompare(right.actor_username || '', 'zh-CN')
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

function formatDate(value: string): string {
  return new Date(value).toLocaleString('zh-CN')
}

function normalizeQueryValue(value: unknown): string {
  if (Array.isArray(value)) return typeof value[0] === 'string' ? value[0] : ''
  return typeof value === 'string' ? value : ''
}

function hydrateFromRoute(): void {
  filters.action = normalizeQueryValue(route.query.action)
  filters.resource_type = normalizeQueryValue(route.query.resource_type)
  filters.actor_user_id = normalizeQueryValue(route.query.actor_user_id)

  const nextPage = Number.parseInt(normalizeQueryValue(route.query.page), 10)
  page.value = Number.isFinite(nextPage) && nextPage > 0 ? nextPage : 1
}

async function syncRouteQuery(): Promise<void> {
  const query: Record<string, string> = {}

  if (filters.action) query.action = filters.action
  if (filters.resource_type) query.resource_type = filters.resource_type
  if (filters.actor_user_id) query.actor_user_id = filters.actor_user_id
  if (page.value > 1) query.page = String(page.value)

  await router.replace({ name: 'AuditLog', query })
}

function detailPreview(detail: Record<string, unknown> | undefined): string {
  if (!detail) return '-'
  return Object.entries(detail)
    .slice(0, 3)
    .map(([key, value]) => `${key}: ${String(value)}`)
    .join(' / ')
}

function actorDisplayName(item: AuditLogItem): string {
  return item.actor_username || '未知执行人'
}

function resourceDisplayName(item: AuditLogItem): string {
  return item.resource_id ? `${item.resource_type} #${item.resource_id}` : item.resource_type
}

function openActorDetail(item: AuditLogItem): void {
  activeActorLog.value = item
}

function closeActorDetail(): void {
  activeActorLog.value = null
}

async function loadLogs(): Promise<void> {
  const requestId = ++latestLogsRequestId
  const controller = createController()
  loading.value = true
  error.value = null
  try {
    const payload = await getAuditLogs(
      {
        page: page.value,
        page_size: pageSize.value,
        action: filters.action || undefined,
        resource_type: filters.resource_type || undefined,
        actor_user_id: filters.actor_user_id ? Number(filters.actor_user_id) : undefined,
      },
      {
        signal: controller.signal,
      }
    )
    if (requestId !== latestLogsRequestId) {
      return
    }
    list.value = payload.list
    total.value = payload.total
    page.value = payload.page
    pageSize.value = payload.page_size
  } catch (err) {
    if (requestId !== latestLogsRequestId) {
      return
    }
    if (
      err &&
      typeof err === 'object' &&
      'code' in err &&
      (err as { code?: unknown }).code === 'ERR_CANCELED'
    ) {
      return
    }
    console.error('加载审计日志失败:', err)
    error.value = '加载审计日志失败，请稍后重试'
  } finally {
    if (requestId !== latestLogsRequestId) {
      return
    }
    loading.value = false
  }
}

async function applyFilters(): Promise<void> {
  page.value = 1
  await syncRouteQuery()
  await loadLogs()
}

function clearTextFilterTimer(): void {
  if (textFilterTimer !== null) {
    clearTimeout(textFilterTimer)
    textFilterTimer = null
  }
}

function scheduleTextFilterApply(): void {
  if (!autoApplyReady.value || suppressAutoApply) {
    return
  }
  clearTextFilterTimer()
  textFilterTimer = setTimeout(() => {
    void applyFilters()
  }, 250)
}

async function changePage(next: number): Promise<void> {
  page.value = Math.max(1, Math.floor(next))
  await syncRouteQuery()
  await loadLogs()
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

async function resetFilters(): Promise<void> {
  clearTextFilterTimer()
  suppressAutoApply = true
  keyword.value = ''
  filters.action = ''
  filters.resource_type = ''
  filters.actor_user_id = ''
  suppressAutoApply = false
  page.value = 1
  await syncRouteQuery()
  await loadLogs()
}

onMounted(() => {
  hydrateFromRoute()
  autoApplyReady.value = true
  void loadLogs()
})

onBeforeUnmount(() => {
  clearTextFilterTimer()
  abort()
})

watch(
  () => filters.action,
  () => {
    if (!autoApplyReady.value || suppressAutoApply) {
      return
    }
    clearTextFilterTimer()
    void applyFilters()
  }
)

watch(
  () => [filters.resource_type, filters.actor_user_id] as const,
  () => {
    scheduleTextFilterApply()
  }
)
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero flex min-h-full flex-1 flex-col"
  >
    <main class="content-pane">
      <header class="admin-overview">
        <div class="admin-overview__intro">
          <div class="workspace-overline">Audit Log</div>
          <h1 class="admin-page-title">审计日志</h1>
        </div>

        <div
          class="admin-summary-grid progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"
        >
          <article class="journal-note progress-card metric-panel-card">
            <div class="journal-note-label progress-card-label metric-panel-label">当前页</div>
            <div class="journal-note-value progress-card-value metric-panel-value">
              {{ list.length }}
            </div>
            <div class="journal-note-helper progress-card-hint metric-panel-helper">
              本页已加载的日志条数
            </div>
          </article>
          <article class="journal-note progress-card metric-panel-card">
            <div class="journal-note-label progress-card-label metric-panel-label">总记录</div>
            <div class="journal-note-value progress-card-value metric-panel-value">{{ total }}</div>
            <div class="journal-note-helper progress-card-hint metric-panel-helper">
              符合条件的审计记录总量
            </div>
          </article>
          <article class="journal-note progress-card metric-panel-card">
            <div class="journal-note-label progress-card-label metric-panel-label">总页数</div>
            <div class="journal-note-value progress-card-value metric-panel-value">
              {{ totalPages }}
            </div>
            <div class="journal-note-helper progress-card-hint metric-panel-helper">
              当前筛选结果的分页范围
            </div>
          </article>
        </div>
      </header>

      <section class="admin-board workspace-directory-section">
        <header class="list-heading audit-board__head">
          <div>
            <div class="workspace-overline">Audit Trail</div>
            <h2 class="list-heading__title">操作流水</h2>
          </div>
          <div class="admin-caption">第 {{ page }} / {{ totalPages }} 页</div>
        </header>

        <WorkspaceDirectoryToolbar
          v-model="keyword"
          :total="filteredTotal"
          :selected-sort-label="sortConfig.label"
          :sort-options="sortOptions"
          search-placeholder="检索动作、资源类型、执行人..."
          total-suffix="条日志"
          reset-label="重置筛选"
          :reset-disabled="!hasActiveFilters"
          @select-sort="setSort"
          @reset-filters="void resetFilters()"
        >
          <template #filter-panel>
            <div class="audit-filter-grid">
              <label class="audit-filter-field">
                <span class="audit-filter-label">动作</span>
                <select v-model="filters.action" class="audit-filter-select">
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
                  v-model="filters.resource_type"
                  type="text"
                  placeholder="资源类型，如 challenge"
                  class="audit-filter-input"
                />
              </label>

              <label class="audit-filter-field">
                <span class="audit-filter-label">执行人</span>
                <input
                  v-model="filters.actor_user_id"
                  type="number"
                  min="1"
                  placeholder="执行人 ID"
                  class="audit-filter-input"
                />
              </label>
            </div>
          </template>
        </WorkspaceDirectoryToolbar>

        <div v-if="error" class="admin-error">
          {{ error }}
          <button type="button" class="ml-3 font-medium underline" @click="loadLogs">重试</button>
        </div>

        <div v-else-if="loading" class="space-y-3 workspace-directory-loading">
          <div
            v-for="index in 6"
            :key="index"
            class="h-14 animate-pulse rounded-2xl bg-[color-mix(in_srgb,var(--journal-surface)_90%,var(--color-bg-base))]"
          />
        </div>

        <div
          v-else-if="filteredRows.length === 0"
          class="audit-empty-state workspace-directory-empty"
        >
          <AppEmpty
            icon="Inbox"
            title="当前筛选条件下没有日志记录"
            description="可以放宽动作、资源类型或执行人条件，再重新检索。"
          />
        </div>

        <WorkspaceDataTable
          v-else
          class="audit-list workspace-directory-list"
          :columns="auditTableColumns"
          :rows="filteredRows"
          row-key="id"
          row-class="audit-row"
        >
          <template #cell-created_at="{ row }">
            <span class="audit-row__time">
              {{ formatDate((row as AuditLogItem).created_at) }}
            </span>
          </template>

          <template #cell-action="{ row }">
            <span class="audit-chip">{{ (row as AuditLogItem).action }}</span>
          </template>

          <template #cell-resource="{ row }">
            <div class="audit-row__resource">
              <span class="audit-row__resource-type">{{
                (row as AuditLogItem).resource_type
              }}</span>
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
                :aria-label="`查看 ${actorDisplayName(row as AuditLogItem)} 的执行人详情`"
                @click="openActorDetail(row as AuditLogItem)"
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

        <div v-if="!loading && total > 0" class="admin-pagination workspace-directory-pagination">
          <PlatformPaginationControls
            :page="page"
            :total-pages="totalPages"
            :total="total"
            :total-label="`共 ${total} 条记录`"
            @change-page="void changePage($event)"
          />
        </div>
      </section>
    </main>

    <AdminSurfaceModal
      class="audit-actor-modal"
      :open="!!activeActorLog"
      title="执行人详情"
      subtitle="查看当前审计记录中执行人的标识、动作和资源上下文。"
      eyebrow="Audit Trail"
      width="34rem"
      @close="closeActorDetail"
      @update:open="!$event && closeActorDetail()"
    >
      <section v-if="activeActorLog" class="audit-actor-detail">
        <div class="audit-actor-detail__grid">
          <article class="audit-actor-detail__item">
            <span class="audit-actor-detail__label">用户名</span>
            <strong class="audit-actor-detail__value">
              {{ actorDisplayName(activeActorLog) }}
            </strong>
          </article>
          <article class="audit-actor-detail__item">
            <span class="audit-actor-detail__label">用户 ID</span>
            <strong class="audit-actor-detail__value audit-actor-detail__value--mono">
              {{ activeActorLog.actor_user_id || '-' }}
            </strong>
          </article>
          <article class="audit-actor-detail__item">
            <span class="audit-actor-detail__label">动作</span>
            <strong class="audit-actor-detail__value">{{ activeActorLog.action }}</strong>
          </article>
          <article class="audit-actor-detail__item">
            <span class="audit-actor-detail__label">时间</span>
            <strong class="audit-actor-detail__value">
              {{ formatDate(activeActorLog.created_at) }}
            </strong>
          </article>
          <article class="audit-actor-detail__item">
            <span class="audit-actor-detail__label">资源</span>
            <strong class="audit-actor-detail__value">
              {{ resourceDisplayName(activeActorLog) }}
            </strong>
          </article>
          <article class="audit-actor-detail__item audit-actor-detail__item--wide">
            <span class="audit-actor-detail__label">明细</span>
            <p class="audit-actor-detail__detail">
              {{ detailPreview(activeActorLog.detail) }}
            </p>
          </article>
        </div>
      </section>
    </AdminSurfaceModal>
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-shell-surface-subtle: color-mix(
    in srgb,
    var(--color-bg-surface) 80%,
    var(--color-bg-base)
  );
  --admin-summary-grid-columns: repeat(3, minmax(0, 1fr));
  --admin-control-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --audit-table-border: color-mix(in srgb, var(--journal-border) 74%, transparent);
  --audit-row-divider: color-mix(in srgb, var(--journal-border) 62%, transparent);
  --journal-eyebrow-spacing: 0.18em;
  --journal-shell-hero-radial-strength: 10%;
  --journal-shell-hero-radial-size: 16rem;
  --journal-shell-hero-top-strength: 97%;
  --journal-shell-hero-end-strength: 95%;
}

.admin-overview {
  display: grid;
  gap: var(--space-4);
}

.admin-page-copy {
  max-width: 48rem;
}

.admin-board {
  display: grid;
  gap: var(--space-4);
  padding-top: var(--space-1);
}

.admin-board :deep(.workspace-directory-toolbar) {
  margin-bottom: 0;
}

.admin-caption {
  font-size: var(--font-size-0-82);
  line-height: 1.6;
  color: var(--journal-muted);
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
  color: var(--journal-muted);
}

.audit-filter-select,
.audit-filter-input {
  width: 100%;
  min-height: 2.75rem;
  border-radius: 0.95rem;
  border: 1px solid var(--admin-control-border);
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
  padding: 0 var(--space-4);
  font-size: var(--font-size-0-875);
  color: var(--journal-ink);
  outline: none;
  transition:
    border-color 150ms ease,
    background 150ms ease,
    box-shadow 150ms ease;
}

.audit-filter-input {
  padding-block: var(--space-3);
}

.audit-filter-select:focus,
.audit-filter-input:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 44%, transparent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 12%, transparent);
}

.admin-error {
  border: 1px solid color-mix(in srgb, var(--color-danger) 20%, var(--journal-border));
  border-radius: 18px;
  background: color-mix(in srgb, var(--color-danger) 8%, transparent);
  padding: var(--space-4) var(--space-4);
  font-size: var(--font-size-0-875);
  color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
}

.audit-empty-state {
  border: 1px solid var(--audit-table-border);
  border-radius: 20px;
  background: color-mix(in srgb, var(--journal-surface-subtle) 92%, var(--color-bg-base));
  padding: var(--space-1-5);
}

.audit-list {
  border: 1px solid var(--audit-table-border);
  border-radius: 1.35rem;
  background: color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base));
  padding: 0.25rem 0.9rem 0.4rem;
}

.audit-list :deep(.workspace-data-table__head-cell) {
  border-bottom-color: var(--audit-table-border);
}

.audit-list :deep(.workspace-data-table__row) {
  border-bottom-color: var(--audit-row-divider);
}

.audit-list :deep(.workspace-data-table__body tr:last-child) {
  border-bottom-color: transparent;
}

.audit-list :deep(.workspace-data-table__body-cell) {
  vertical-align: top;
}

.audit-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 18%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  padding: var(--space-1-5) var(--space-3);
  font-size: var(--font-size-0-72);
  font-weight: 700;
  color: color-mix(in srgb, var(--journal-accent) 84%, var(--journal-ink));
}

.audit-row__time {
  display: block;
  font-size: var(--font-size-0-82);
  line-height: 1.6;
  color: var(--journal-muted);
}

.audit-row__resource,
.audit-row__actor {
  display: grid;
  gap: 0.18rem;
  min-width: 0;
}

.audit-row__resource-type {
  font-size: var(--font-size-1-00);
  font-weight: 700;
  color: var(--journal-ink);
}

.audit-row__resource-id {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-0-78);
  color: var(--journal-muted);
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
  color: color-mix(in srgb, var(--journal-accent) 88%, var(--journal-ink));
  font-size: var(--font-size-1-00);
  font-weight: 700;
  line-height: 1.45;
  text-decoration: underline;
  text-decoration-thickness: 1px;
  text-underline-offset: 0.18em;
  transition:
    color 150ms ease,
    text-decoration-color 150ms ease;
}

.audit-row__actor-link:hover,
.audit-row__actor-link:focus-visible {
  color: color-mix(in srgb, var(--journal-accent) 100%, var(--journal-ink));
  text-decoration-color: currentColor;
}

.audit-row__actor-link:focus-visible {
  outline: none;
  box-shadow: 0 2px 0 0 color-mix(in srgb, var(--journal-accent) 26%, transparent);
}

.audit-row__detail {
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

.audit-actor-detail {
  display: grid;
}

.audit-actor-detail__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-3);
}

.audit-actor-detail__item {
  display: grid;
  gap: var(--space-1-5);
  border: 1px solid var(--audit-table-border);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--journal-surface-subtle) 92%, var(--color-bg-base));
  padding: var(--space-3);
}

.audit-actor-detail__item--wide {
  grid-column: 1 / -1;
}

.audit-actor-detail__label {
  font-size: var(--font-size-0-72);
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.audit-actor-detail__value {
  font-size: var(--font-size-0-94);
  line-height: 1.55;
  color: var(--journal-ink);
}

.audit-actor-detail__value--mono {
  font-family: var(--font-family-mono);
}

.audit-actor-detail__detail {
  margin: 0;
  font-size: var(--font-size-0-88);
  line-height: 1.7;
  color: var(--journal-muted);
}

:deep(.audit-actor-modal .modal-template-panel--classic) {
  --modal-template-classic-surface: #fff;
  --modal-template-classic-surface-muted: #fff;
  --modal-template-classic-line: color-mix(in srgb, var(--color-border-default) 88%, transparent);
  --modal-template-classic-text: color-mix(in srgb, var(--color-text-primary) 96%, transparent);
  --modal-template-classic-muted: color-mix(in srgb, var(--color-text-secondary) 94%, transparent);
  --modal-template-classic-faint: color-mix(in srgb, var(--color-text-muted) 92%, transparent);
}

:deep(.audit-actor-modal .modal-template-classic__header),
:deep(.audit-actor-modal .modal-template-classic__body),
:deep(.audit-actor-modal .modal-template-classic__footer) {
  background: #fff;
}

@media (max-width: 1080px) {
  .admin-summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 720px) {
  .admin-summary-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .audit-list {
    min-width: 56rem;
  }

  .audit-actor-detail__grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .journal-shell {
    padding-inline: var(--space-4);
  }
}
</style>
