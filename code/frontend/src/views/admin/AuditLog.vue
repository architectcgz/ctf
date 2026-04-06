<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getAuditLogs } from '@/api/admin'
import type { AuditLogItem } from '@/api/contracts'
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

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const activeFilterCount = computed(
  () => [filters.action, filters.resource_type, filters.actor_user_id].filter(Boolean).length
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

async function changePage(next: number): Promise<void> {
  page.value = Math.max(1, Math.floor(next))
  await syncRouteQuery()
  await loadLogs()
}

async function resetFilters(): Promise<void> {
  filters.action = ''
  filters.resource_type = ''
  filters.actor_user_id = ''
  page.value = 1
  await syncRouteQuery()
  await loadLogs()
}

onMounted(() => {
  hydrateFromRoute()
  void loadLogs()
})
</script>

<template>
  <section
    class="journal-shell journal-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
  >
    <header class="admin-overview">
      <div class="admin-overview__intro">
        <div class="journal-eyebrow">Audit Log</div>
        <h1 class="admin-page-title">审计日志</h1>
        <p class="admin-page-copy">按动作、资源类型和执行人快速检索关键操作记录。</p>
      </div>

      <div class="admin-summary-grid">
        <article class="journal-note">
          <div class="journal-note-label">激活筛选</div>
          <div class="journal-note-value">{{ activeFilterCount }}</div>
          <div class="journal-note-helper">当前生效的筛选项数量</div>
        </article>
        <article class="journal-note">
          <div class="journal-note-label">当前页</div>
          <div class="journal-note-value">{{ list.length }}</div>
          <div class="journal-note-helper">本页已加载的日志条数</div>
        </article>
        <article class="journal-note">
          <div class="journal-note-label">总记录</div>
          <div class="journal-note-value">{{ total }}</div>
          <div class="journal-note-helper">符合条件的审计记录总量</div>
        </article>
      </div>
    </header>

    <section class="admin-toolbar">
      <div class="admin-section-head">
        <div>
          <div class="journal-note-label">Filters</div>
          <h2 class="admin-section-title">筛选条件</h2>
        </div>
        <p class="admin-section-copy">支持按动作、资源类型与执行人组合筛选。</p>
      </div>

      <div class="admin-filter-grid">
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

        <input
          v-model="filters.resource_type"
          type="text"
          placeholder="资源类型，如 challenge"
          class="admin-input"
        />

        <input
          v-model="filters.actor_user_id"
          type="number"
          min="1"
          placeholder="执行人 ID"
          class="admin-input"
        />

        <button type="button" class="admin-btn admin-btn-primary" @click="applyFilters">
          应用筛选
        </button>

        <button type="button" class="admin-btn admin-btn-ghost" @click="resetFilters">重置</button>
      </div>
    </section>

    <div class="journal-divider admin-divider" />

    <section class="admin-board">
      <div class="admin-section-head">
        <div>
          <div class="journal-note-label">Results</div>
          <h2 class="admin-section-title">操作流水</h2>
        </div>
        <div class="admin-caption">第 {{ page }} / {{ totalPages }} 页</div>
      </div>

      <div v-if="error" class="admin-error">
        {{ error }}
        <button type="button" class="ml-3 font-medium underline" @click="loadLogs">重试</button>
      </div>

      <div v-else-if="loading" class="space-y-3">
        <div
          v-for="index in 6"
          :key="index"
          class="h-14 animate-pulse rounded-2xl bg-[color-mix(in_srgb,var(--journal-surface)_90%,var(--color-bg-base))]"
        />
      </div>

      <div v-else-if="list.length === 0" class="audit-empty-state">
        <AppEmpty
          icon="Inbox"
          title="当前筛选条件下没有日志记录"
          description="可以放宽动作、资源类型或执行人条件，再重新检索。"
        />
      </div>

      <div
        v-else
        class="audit-table-shell overflow-hidden rounded-[20px] border border-[var(--audit-table-border)]"
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

      <div v-if="!loading && total > 0" class="admin-pagination">
        <span>共 {{ total }} 条记录</span>
        <div class="flex items-center gap-2">
          <button
            type="button"
            :disabled="page === 1"
            class="admin-btn admin-btn-ghost admin-btn-compact disabled:cursor-not-allowed disabled:opacity-50"
            @click="changePage(page - 1)"
          >
            上一页
          </button>
          <span class="text-sm text-[var(--color-text-secondary)]"
            >{{ page }} / {{ totalPages }}</span
          >
          <button
            type="button"
            :disabled="page >= totalPages"
            class="admin-btn admin-btn-ghost admin-btn-compact disabled:cursor-not-allowed disabled:opacity-50"
            @click="changePage(page + 1)"
          >
            下一页
          </button>
        </div>
      </div>
    </section>
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: var(--color-primary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 80%, var(--color-bg-base));
  --admin-control-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --audit-table-border: color-mix(in srgb, var(--journal-border) 74%, transparent);
  --audit-row-divider: color-mix(in srgb, var(--journal-border) 62%, transparent);
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--journal-accent) 10%, transparent),
      transparent 16rem
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 97%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 95%, var(--color-bg-base))
    );
  box-shadow: 0 18px 40px var(--color-shadow-soft);
}

.journal-eyebrow,
.journal-note-label {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.journal-note {
  border: 1px solid var(--journal-border);
  border-radius: 18px;
  background: color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base));
  padding: 0.95rem 1rem;
}

.journal-note-value {
  margin-top: 0.35rem;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.journal-note-helper {
  margin-top: 0.45rem;
  font-size: 0.78rem;
  line-height: 1.55;
  color: var(--journal-muted);
}

.journal-divider {
  border-top: 1px dashed color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.admin-overview {
  display: grid;
  gap: 1rem;
}

.admin-page-title {
  margin-top: 0.85rem;
  font-size: clamp(1.9rem, 2vw, 2.4rem);
  font-weight: 700;
  line-height: 1.08;
  color: var(--journal-ink);
}

.admin-page-copy {
  margin-top: 0.7rem;
  max-width: 48rem;
  font-size: 0.92rem;
  line-height: 1.7;
  color: var(--journal-muted);
}

.admin-summary-grid {
  display: grid;
  gap: 0.85rem;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.admin-toolbar,
.admin-board {
  padding: 0;
}

.admin-divider {
  margin: 1.2rem 0;
}

.admin-section-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 0.75rem;
}

.admin-section-title {
  margin-top: 0.35rem;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.admin-section-copy,
.admin-caption {
  font-size: 0.82rem;
  line-height: 1.6;
  color: var(--journal-muted);
}

.admin-filter-grid {
  margin-top: 1rem;
  display: grid;
  gap: 0.85rem;
  grid-template-columns: repeat(3, minmax(0, 1fr)) auto auto;
}

.admin-input {
  width: 100%;
  min-height: 2.75rem;
  border-radius: 1rem;
  border: 1px solid var(--admin-control-border);
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  padding: 0.7rem 1rem;
  font-size: 0.875rem;
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
  padding: 0.65rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  transition:
    border-color 150ms ease,
    background 150ms ease,
    color 150ms ease;
}

.admin-btn-compact {
  min-height: 2.3rem;
  padding: 0.48rem 0.82rem;
}

.admin-btn-primary {
  border-color: color-mix(in srgb, var(--journal-accent) 24%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 12%, var(--journal-surface));
  color: color-mix(in srgb, var(--journal-accent) 86%, var(--journal-ink));
}

.admin-btn-ghost {
  border: 1px solid var(--admin-control-border);
  background: color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base));
  color: var(--journal-ink);
}

.admin-error {
  margin-top: 1rem;
  border: 1px solid color-mix(in srgb, var(--color-danger) 20%, var(--journal-border));
  border-radius: 18px;
  background: color-mix(in srgb, var(--color-danger) 8%, transparent);
  padding: 0.95rem 1rem;
  font-size: 0.875rem;
  color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
}

.audit-empty-state {
  margin-top: 1rem;
  border: 1px solid var(--audit-table-border);
  border-radius: 20px;
  background: color-mix(in srgb, var(--journal-surface-subtle) 92%, var(--color-bg-base));
  padding: 0.4rem;
}

.audit-table-shell {
  margin-top: 1rem;
  background: color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base));
}

.audit-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 18%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  padding: 0.34rem 0.72rem;
  font-size: 0.72rem;
  font-weight: 700;
  color: color-mix(in srgb, var(--journal-accent) 84%, var(--journal-ink));
}

.admin-pagination {
  margin-top: 1rem;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border-top: 1px dashed color-mix(in srgb, var(--journal-border) 84%, transparent);
  padding-top: 1rem;
  font-size: 0.875rem;
  color: var(--journal-muted);
}

@media (max-width: 1080px) {
  .admin-summary-grid,
  .admin-filter-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 720px) {
  .admin-summary-grid,
  .admin-filter-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .journal-shell {
    padding-inline: 1rem;
  }
}
</style>
