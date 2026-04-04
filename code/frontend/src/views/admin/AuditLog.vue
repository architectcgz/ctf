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
  loadLogs()
})
</script>

<template>
  <section class="journal-shell journal-hero flex min-h-full flex-col rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="journal-eyebrow">Audit Trail</div>
          <h1 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]">审计日志</h1>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            按动作、资源类型和执行人快速检索关键记录。
          </p>
        </div>

        <article class="journal-brief rounded-[24px] border px-5 py-5">
          <div class="journal-note-label">筛选视角</div>
          <div class="mt-5 grid gap-3 sm:grid-cols-2">
            <div class="journal-note">
              <div class="journal-note-label">动作</div>
              <div class="journal-note-value">{{ filters.action || '全部' }}</div>
              <div class="journal-note-helper">当前检索动作</div>
            </div>
            <div class="journal-note">
              <div class="journal-note-label">资源</div>
              <div class="journal-note-value">{{ filters.resource_type || '全部' }}</div>
              <div class="journal-note-helper">当前资源类型</div>
            </div>
          </div>
        </article>
      </div>
      <div class="journal-divider" />

      <div class="space-y-3">
      <div class="admin-section-head">
        <div>
          <div class="journal-note-label">Filters</div>
          <h2 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">筛选条件</h2>
        </div>
      </div>

      <div class="mt-5 grid gap-3 md:grid-cols-[repeat(3,minmax(0,1fr))_auto_auto]">
        <select
          v-model="filters.action"
          class="admin-input"
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

        <button
          type="button"
          class="admin-btn admin-btn-primary"
          @click="applyFilters"
        >
          应用筛选
        </button>

        <button
          type="button"
          class="admin-btn admin-btn-ghost"
          @click="resetFilters"
        >
          重置
        </button>
      </div>

      <div class="journal-divider" />

      <div class="space-y-3">
        <div class="admin-section-head">
          <div>
            <div class="journal-note-label">Logs</div>
            <h2 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">操作流水</h2>
          </div>
        </div>

      <div v-if="error" class="rounded-xl border border-[var(--color-danger)]/20 bg-[var(--color-danger)]/10 px-4 py-4 text-sm text-[var(--color-danger)]">
        {{ error }}
        <button type="button" class="ml-3 font-medium underline" @click="loadLogs">重试</button>
      </div>

      <div v-else-if="loading" class="space-y-3">
        <div v-for="index in 6" :key="index" class="h-14 animate-pulse rounded-xl bg-[var(--color-bg-base)]"></div>
      </div>

      <AppEmpty
        v-else-if="list.length === 0"
        icon="Inbox"
        title="当前筛选条件下没有日志记录"
        description="可以放宽动作、资源类型或执行人条件，再重新检索。"
      />

      <div v-else class="overflow-hidden rounded-[18px] border border-[var(--journal-border)]">
        <table class="min-w-full divide-y divide-[var(--color-border-default)] text-sm">
          <thead class="bg-[var(--journal-surface-subtle)]">
            <tr>
              <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">时间</th>
              <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">动作</th>
              <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">资源</th>
              <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">执行人</th>
              <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">明细</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[var(--journal-border)] bg-[var(--journal-surface)]">
            <tr v-for="item in list" :key="item.id">
              <td class="px-4 py-3 text-[var(--color-text-secondary)]">{{ formatDate(item.created_at) }}</td>
              <td class="px-4 py-3">
                <span class="rounded-full bg-[var(--color-primary)]/10 px-3 py-1 text-xs font-medium text-[var(--color-primary)]">{{ item.action }}</span>
              </td>
              <td class="px-4 py-3 text-[var(--color-text-primary)]">
                {{ item.resource_type }}<span v-if="item.resource_id">#{{ item.resource_id }}</span>
              </td>
              <td class="px-4 py-3 text-[var(--color-text-primary)]">
                {{ item.actor_username }}<span v-if="item.actor_user_id" class="text-[var(--color-text-secondary)]"> ({{ item.actor_user_id }})</span>
              </td>
              <td class="px-4 py-3 text-[var(--color-text-secondary)]">{{ detailPreview(item.detail) }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="!loading && total > 0" class="admin-pagination mt-4">
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
          <span class="text-sm text-[var(--color-text-secondary)]">{{ page }} / {{ totalPages }}</span>
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
      </div>
      </div>
    </section>
</template>

<style scoped>
.journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: var(--color-primary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 78%, var(--color-bg-base));
}

.journal-hero,
.journal-panel {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 12%, transparent), transparent 18rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 94%, var(--color-bg-base))
    );
  border-radius: 16px !important;
  box-shadow: 0 18px 40px var(--color-shadow-soft);
}

.journal-brief {
  background: var(--journal-surface-subtle);
  border-color: var(--journal-border);
  border-radius: 16px !important;
}

.journal-eyebrow,
.journal-note-label {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.journal-note {
  border-radius: 14px;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.75rem 0.875rem;
}

.journal-note-value {
  margin-top: 0.35rem;
  font-size: 1rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.journal-note-helper {
  margin-top: 0.55rem;
  font-size: 0.78rem;
  line-height: 1.5;
  color: var(--journal-muted);
}

.journal-divider {
  margin-block: 1rem;
  border-top: 1px dashed rgba(148, 163, 184, 0.7);
}

.admin-section-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.admin-input {
  width: 100%;
  min-height: 2.75rem;
  border-radius: 1rem;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.7rem 1rem;
  font-size: 0.875rem;
  color: var(--journal-ink);
  outline: none;
}

.admin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.75rem;
  border-radius: 1rem;
  padding: 0.65rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
}

.admin-btn-compact {
  min-height: 2.35rem;
  padding: 0.5rem 0.85rem;
}

.admin-btn-primary {
  background: var(--journal-accent);
  color: #fff;
}

.admin-btn-ghost {
  border: 1px solid var(--journal-border);
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  color: var(--journal-ink);
}

.admin-pagination {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border-top: 1px dashed rgba(148, 163, 184, 0.72);
  padding-top: 1rem;
  font-size: 0.875rem;
  color: var(--journal-muted);
}

:global([data-theme='dark']) .journal-shell {
  --journal-ink: color-mix(in srgb, var(--color-text-primary) 88%, var(--color-text-secondary));
  --journal-muted: var(--color-text-secondary);
  --journal-accent: #60a5fa;
  --journal-border: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 76%, var(--color-bg-base));
}

:global([data-theme='dark']) .journal-hero,
:global([data-theme='dark']) .journal-panel {
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 16%, transparent), transparent 18rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 97%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 95%, var(--color-bg-base))
    );
}
</style>
