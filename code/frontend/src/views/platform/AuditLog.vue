<script setup lang="ts">
import {
  Activity,
  ArrowDownWideNarrow,
  Calendar,
  Clock,
  FileJson,
  Fingerprint,
  Layers,
  Package,
  RefreshCw,
  Trophy,
  User,
  UserRound,
} from 'lucide-vue-next'
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getAuditLogs } from '@/api/admin'
import type { AuditLogItem } from '@/api/contracts'
import AdminSurfaceModal from '@/components/common/modal-templates/AdminSurfaceModal.vue'
import AuditLogDirectoryPanel from '@/components/platform/audit/AuditLogDirectoryPanel.vue'
import {
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

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const hasActiveFilters = computed(() =>
  Boolean(keyword.value.trim() || filters.action || filters.resource_type || filters.actor_user_id)
)
const filteredRows = computed<AuditLogItem[]>(() => {
  const nextRows = [...list.value]

  nextRows.sort((left, right) => {
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

  return nextRows
})

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
  <div class="workspace-shell journal-shell journal-shell-admin journal-hero">
    <div class="workspace-grid">
      <main class="content-pane">
        <section class="workspace-hero">
          <div class="workspace-tab-heading__main">
            <div class="workspace-overline">Audit Log</div>
            <h1 class="hero-title">
              审计日志
            </h1>
            <p class="hero-summary">
              追踪全站资源变更、用户行为与系统关键操作，确保平台安全与合规。
            </p>
          </div>

          <div class="awd-library-hero-actions">
            <div class="quick-actions">
              <button
                type="button"
                class="ui-btn ui-btn--primary"
                @click="loadLogs"
              >
                <RefreshCw class="h-4 w-4" />
                同步日志
              </button>
            </div>
          </div>
        </section>

        <div class="audit-log-body mt-10 space-y-10">
          <div class="admin-summary-grid progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface">
            <article class="journal-note progress-card metric-panel-card">
              <div class="journal-note-label progress-card-label metric-panel-label">
                <span>当前页加载</span>
                <Activity class="h-4 w-4" />
              </div>
              <div class="journal-note-value progress-card-value metric-panel-value">
                {{ list.length.toString().padStart(2, '0') }}
              </div>
              <div class="journal-note-helper progress-card-hint metric-panel-helper">
                本页已加载的日志条数
              </div>
            </article>

            <article class="journal-note progress-card metric-panel-card">
              <div class="journal-note-label progress-card-label metric-panel-label">
                <span>全站总记录</span>
                <Trophy class="h-4 w-4" />
              </div>
              <div class="journal-note-value progress-card-value metric-panel-value">
                {{ total.toString().padStart(2, '0') }}
              </div>
              <div class="journal-note-helper progress-card-hint metric-panel-helper">
                审计数据库中的累计总量
              </div>
            </article>

            <article class="journal-note progress-card metric-panel-card">
              <div class="journal-note-label progress-card-label metric-panel-label">
                <span>总分页范围</span>
                <Layers class="h-4 w-4" />
              </div>
              <div class="journal-note-value progress-card-value metric-panel-value">
                {{ totalPages.toString().padStart(2, '0') }}
              </div>
              <div class="journal-note-helper progress-card-hint metric-panel-helper">
                当前条件下的分页总数
              </div>
            </article>
          </div>

          <AuditLogDirectoryPanel
            :rows="filteredRows"
            :total="total"
            :page="page"
            :total-pages="totalPages"
            :loading="loading"
            :error="error"
            :keyword="keyword"
            :has-active-filters="hasActiveFilters"
            :selected-sort-label="sortConfig.label"
            :sort-options="sortOptions"
            :action-filter="filters.action"
            :resource-type-filter="filters.resource_type"
            :actor-user-id-filter="filters.actor_user_id"
            :format-date="formatDate"
            :detail-preview="detailPreview"
            :actor-display-name="actorDisplayName"
            @update:keyword="keyword = $event"
            @update:action-filter="filters.action = $event"
            @update:resource-type-filter="filters.resource_type = $event"
            @update:actor-user-id-filter="filters.actor_user_id = $event"
            @select-sort="setSort"
            @reset-filters="void resetFilters()"
            @retry="loadLogs"
            @open-actor-detail="openActorDetail"
            @change-page="void changePage($event)"
          />
        </div>
      </main>
    </div>

    <AdminSurfaceModal
      class="audit-actor-modal"
      :open="!!activeActorLog"
      title="执行人详情"
      eyebrow="Audit Log"
      width="34rem"
      @close="closeActorDetail"
      @update:open="!$event && closeActorDetail()"
    >
      <section
        v-if="activeActorLog"
        class="audit-actor-detail"
      >
        <div class="audit-actor-detail__grid">
          <article class="audit-actor-detail__item">
            <div class="audit-actor-detail__head">
              <User class="h-3.5 w-3.5" />
              <span class="audit-actor-detail__label">用户名</span>
            </div>
            <strong class="audit-actor-detail__value">
              {{ actorDisplayName(activeActorLog) }}
            </strong>
          </article>

          <article class="audit-actor-detail__item">
            <div class="audit-actor-detail__head">
              <Fingerprint class="h-3.5 w-3.5" />
              <span class="audit-actor-detail__label">用户 ID</span>
            </div>
            <strong class="audit-actor-detail__value audit-actor-detail__value--mono">
              {{ activeActorLog.actor_user_id || '-' }}
            </strong>
          </article>

          <article class="audit-actor-detail__item">
            <div class="audit-actor-detail__head">
              <Activity class="h-3.5 w-3.5" />
              <span class="audit-actor-detail__label">动作</span>
            </div>
            <strong class="audit-actor-detail__value">{{ activeActorLog.action }}</strong>
          </article>

          <article class="audit-actor-detail__item">
            <div class="audit-actor-detail__head">
              <Clock class="h-3.5 w-3.5" />
              <span class="audit-actor-detail__label">发生时间</span>
            </div>
            <strong class="audit-actor-detail__value">
              {{ formatDate(activeActorLog.created_at) }}
            </strong>
          </article>

          <article class="audit-actor-detail__item">
            <div class="audit-actor-detail__head">
              <Package class="h-3.5 w-3.5" />
              <span class="audit-actor-detail__label">目标资源</span>
            </div>
            <strong class="audit-actor-detail__value">
              {{ resourceDisplayName(activeActorLog) }}
            </strong>
          </article>

          <article class="audit-actor-detail__item audit-actor-detail__item--wide">
            <div class="audit-actor-detail__head">
              <FileJson class="h-3.5 w-3.5" />
              <span class="audit-actor-detail__label">明细上下文</span>
            </div>
            <p class="audit-actor-detail__detail">
              {{ detailPreview(activeActorLog.detail) }}
            </p>
          </article>
        </div>
      </section>
    </AdminSurfaceModal>
  </div>
</template>

<style scoped>
.workspace-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-7);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid var(--color-border-subtle);
}

.hero-title {
  margin: 0.5rem 0 0;
  font-size: var(--workspace-page-title-font-size);
  line-height: var(--workspace-page-title-line-height);
  letter-spacing: var(--workspace-page-title-letter-spacing);
  color: var(--color-text-primary);
}

.hero-summary {
  max-width: 760px;
  margin-top: var(--space-3-5);
  font-size: var(--font-size-15);
  line-height: 1.9;
  color: var(--color-text-secondary);
}

.quick-actions {
  display: flex;
  gap: 0.75rem;
  align-items: flex-end;
  height: 100%;
  padding-bottom: 0.5rem;
}

.audit-actor-detail__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1.75rem 2rem;
  padding: 0.25rem;
}

.audit-actor-detail__item {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.audit-actor-detail__head {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--color-text-muted);
}

.audit-actor-detail__label {
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.audit-actor-detail__value {
  font-size: var(--font-size-16);
  font-weight: 800;
  line-height: 1.2;
  color: var(--color-text-primary);
}

.audit-actor-detail__item--wide {
  grid-column: 1 / -1;
}

.audit-actor-detail__value--mono {
  font-family: var(--font-family-mono);
}

.audit-actor-detail__detail {
  margin: 0;
  font-size: var(--font-size-13);
  line-height: 1.7;
  color: var(--color-text-secondary);
}

@media (max-width: 720px) {
  .audit-actor-detail__grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
