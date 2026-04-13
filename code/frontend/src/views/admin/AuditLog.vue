<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getAuditLogs } from '@/api/admin'
import type { AuditLogItem } from '@/api/contracts'
import AdminPaginationControls from '@/components/admin/AdminPaginationControls.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'

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
let textFilterTimer: ReturnType<typeof setTimeout> | null = null
let suppressAutoApply = false
const autoApplyReady = ref(false)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const hasActiveFilters = computed(() =>
  Boolean(filters.action || filters.resource_type || filters.actor_user_id)
)

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

async function loadLogs(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    const payload = await getAuditLogs({
      page: page.value,
      page_size: pageSize.value,
      action: filters.action || undefined,
      resource_type: filters.resource_type || undefined,
      actor_user_id: filters.actor_user_id ? Number(filters.actor_user_id) : undefined,
    })
    list.value = payload.list
    total.value = payload.total
    page.value = payload.page
    pageSize.value = payload.page_size
  } catch (err) {
    console.error('加载审计日志失败:', err)
    error.value = '加载审计日志失败，请稍后重试'
  } finally {
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

async function resetFilters(): Promise<void> {
  clearTextFilterTimer()
  suppressAutoApply = true
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
    class="journal-shell journal-shell-admin journal-notes-card journal-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
  >
    <header class="admin-overview">
      <div class="admin-overview__intro">
        <div class="journal-eyebrow">Audit Log</div>
        <h1 class="admin-page-title">审计日志</h1>
      </div>

      <div class="admin-summary-grid progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface">
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
          <div class="journal-note-value progress-card-value metric-panel-value">{{ totalPages }}</div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">
            当前筛选结果的分页范围
          </div>
        </article>
      </div>
    </header>

    <section class="admin-board workspace-directory-section">
      <header class="list-heading">
        <div>
          <div class="journal-note-label">Audit Trail</div>
          <h2 class="list-heading__title">操作流水</h2>
        </div>
        <div class="admin-caption">第 {{ page }} / {{ totalPages }} 页</div>
      </header>

      <section class="audit-filter-strip" aria-label="日志筛选">
        <div class="audit-filter-grid">
          <label class="audit-filter-field">
            <span class="audit-filter-label">动作</span>
            <select v-model="filters.action" class="admin-input">
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
              class="admin-input"
            />
          </label>

          <label class="audit-filter-field">
            <span class="audit-filter-label">执行人</span>
            <input
              v-model="filters.actor_user_id"
              type="number"
              min="1"
              placeholder="执行人 ID"
              class="admin-input"
            />
          </label>

          <div class="audit-filter-actions">
            <span class="audit-filter-label audit-filter-label--ghost" aria-hidden="true">操作</span>
            <div class="audit-filter-action-row">
              <button
                type="button"
                class="admin-btn admin-btn-ghost audit-filter-reset"
                :disabled="!hasActiveFilters"
                @click="resetFilters"
              >
                重置筛选
              </button>
            </div>
          </div>
        </div>
      </section>

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

      <div v-else-if="list.length === 0" class="audit-empty-state workspace-directory-empty">
        <AppEmpty
          icon="Inbox"
          title="当前筛选条件下没有日志记录"
          description="可以放宽动作、资源类型或执行人条件，再重新检索。"
        />
      </div>

      <div
        v-else
        class="audit-table-shell workspace-directory-list overflow-hidden rounded-[20px] border border-[var(--audit-table-border)]"
      >
        <table class="min-w-full text-sm">
          <thead class="bg-[var(--journal-surface-subtle)]">
            <tr>
              <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">
                时间
              </th>
              <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">
                动作
              </th>
              <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">
                资源
              </th>
              <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">
                执行人
              </th>
              <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">
                明细
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[var(--audit-row-divider)] bg-[var(--journal-surface)]">
            <tr v-for="item in list" :key="item.id">
              <td class="px-4 py-3 text-[var(--color-text-secondary)]">
                {{ formatDate(item.created_at) }}
              </td>
              <td class="px-4 py-3">
                <span class="audit-chip">{{ item.action }}</span>
              </td>
              <td class="px-4 py-3 text-[var(--color-text-primary)]">
                {{ item.resource_type }}<span v-if="item.resource_id">#{{ item.resource_id }}</span>
              </td>
              <td class="px-4 py-3 text-[var(--color-text-primary)]">
                {{ item.actor_username }}
                <span v-if="item.actor_user_id" class="text-[var(--color-text-secondary)]">
                  ({{ item.actor_user_id }})
                </span>
              </td>
              <td class="px-4 py-3 text-[var(--color-text-secondary)]">
                {{ detailPreview(item.detail) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="!loading && total > 0" class="admin-pagination workspace-directory-pagination">
        <AdminPaginationControls
          :page="page"
          :total-pages="totalPages"
          :total="total"
          :total-label="`共 ${total} 条记录`"
          @change-page="void changePage($event)"
        />
      </div>
    </section>
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-shell-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 80%, var(--color-bg-base));
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
  padding: 0;
}

.list-heading {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-3);
}

.list-heading__title {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-1-20);
  font-weight: 700;
  color: var(--journal-ink);
}

.admin-caption {
  font-size: var(--font-size-0-82);
  line-height: 1.6;
  color: var(--journal-muted);
}

.audit-filter-strip {
  display: grid;
  gap: var(--space-4);
  padding: var(--space-5) 0;
}

.audit-filter-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: repeat(3, minmax(14rem, 16rem)) auto;
}

.audit-filter-field,
.audit-filter-actions {
  display: grid;
  gap: var(--space-2);
}

.audit-filter-label {
  font-size: var(--font-size-0-78);
  font-weight: 700;
  color: var(--journal-muted);
}

.audit-filter-label--ghost {
  opacity: 0;
  pointer-events: none;
}

.audit-filter-actions {
  justify-items: end;
}

.audit-filter-action-row {
  display: flex;
  gap: var(--space-2-5);
  justify-content: flex-end;
}

.admin-input {
  width: 100%;
  min-height: 2.75rem;
  border-radius: 1rem;
  border: 1px solid var(--admin-control-border);
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  padding: var(--space-3) var(--space-4);
  font-size: var(--font-size-0-875);
  color: var(--journal-ink);
  outline: none;
  transition:
    border-color 150ms ease,
    background 150ms ease,
    box-shadow 150ms ease;
}

.admin-input:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 44%, transparent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 12%, transparent);
}

.admin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid transparent;
  min-height: 2.75rem;
  border-radius: 999px;
  padding: var(--space-2-5) var(--space-4);
  font-size: var(--font-size-0-875);
  font-weight: 600;
  transition:
    border-color 150ms ease,
    background 150ms ease,
    color 150ms ease;
}

.admin-btn-compact {
  min-height: 2.3rem;
  padding: var(--space-2) var(--space-3);
}

.admin-btn-ghost {
  border: 1px solid var(--admin-control-border);
  background: color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base));
  color: var(--journal-ink);
}

.audit-filter-reset:disabled {
  cursor: not-allowed;
  opacity: 0.48;
}

.admin-error {
  margin-top: var(--space-4);
  border: 1px solid color-mix(in srgb, var(--color-danger) 20%, var(--journal-border));
  border-radius: 18px;
  background: color-mix(in srgb, var(--color-danger) 8%, transparent);
  padding: var(--space-4) var(--space-4);
  font-size: var(--font-size-0-875);
  color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
}

.audit-empty-state {
  margin-top: var(--space-4);
  border: 1px solid var(--audit-table-border);
  border-radius: 20px;
  background: color-mix(in srgb, var(--journal-surface-subtle) 92%, var(--color-bg-base));
  padding: var(--space-1-5);
}

.audit-table-shell {
  background: color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base));
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

@media (max-width: 1080px) {
  .admin-summary-grid,
  .audit-filter-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 720px) {
  .list-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .admin-summary-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .audit-filter-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .audit-filter-actions {
    justify-items: stretch;
  }

  .audit-filter-action-row {
    width: 100%;
  }

  .audit-filter-action-row > * {
    flex: 1 1 0;
  }

  .journal-shell {
    padding-inline: var(--space-4);
  }
}
</style>
