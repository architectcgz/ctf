<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getAuditLogs } from '@/api/admin'
import type { AuditLogItem } from '@/api/contracts'

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
  <div class="space-y-6">
    <section class="rounded-[28px] border border-[var(--color-border-default)] bg-[linear-gradient(135deg,rgba(15,23,42,0.08),rgba(8,145,178,0.12))] p-7 shadow-sm">
      <p class="text-xs font-semibold uppercase tracking-[0.28em] text-[var(--color-primary)]/85">Audit Trail</p>
      <h1 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--color-text-primary)]">审计日志</h1>
      <p class="mt-2 max-w-3xl text-sm leading-6 text-[var(--color-text-secondary)]">
        按动作、资源类型和执行人快速检索关键管理操作与提交流水。
      </p>
    </section>

    <section class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6 shadow-sm">
      <div class="grid gap-3 md:grid-cols-[repeat(3,minmax(0,1fr))_auto_auto]">
        <select
          v-model="filters.action"
          class="rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-base)] px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-[var(--color-primary)]"
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
          class="rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-base)] px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-[var(--color-primary)]"
        />

        <input
          v-model="filters.actor_user_id"
          type="number"
          min="1"
          placeholder="执行人 ID"
          class="rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-base)] px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-[var(--color-primary)]"
        />

        <button
          type="button"
          class="rounded-xl bg-[var(--color-primary)] px-4 py-3 text-sm font-medium text-white transition hover:bg-[var(--color-primary-hover)]"
          @click="applyFilters"
        >
          应用筛选
        </button>

        <button
          type="button"
          class="rounded-xl border border-[var(--color-border-default)] px-4 py-3 text-sm font-medium text-[var(--color-text-primary)] transition hover:border-[var(--color-primary)]"
          @click="resetFilters"
        >
          重置
        </button>
      </div>

      <div v-if="error" class="mt-6 rounded-xl border border-red-200 bg-red-50 px-4 py-4 text-sm text-red-600">
        {{ error }}
        <button type="button" class="ml-3 font-medium underline" @click="loadLogs">重试</button>
      </div>

      <div v-else-if="loading" class="mt-6 space-y-3">
        <div v-for="index in 6" :key="index" class="h-14 animate-pulse rounded-xl bg-[var(--color-bg-base)]"></div>
      </div>

      <div v-else-if="list.length === 0" class="mt-6 rounded-xl border border-dashed border-[var(--color-border-default)] px-4 py-10 text-center text-sm text-[var(--color-text-secondary)]">
        当前筛选条件下没有日志记录。
      </div>

      <div v-else class="mt-6 overflow-hidden rounded-xl border border-[var(--color-border-default)]">
        <table class="min-w-full divide-y divide-[var(--color-border-default)] text-sm">
          <thead class="bg-[var(--color-bg-base)]">
            <tr>
              <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">时间</th>
              <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">动作</th>
              <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">资源</th>
              <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">执行人</th>
              <th class="px-4 py-3 text-left font-medium text-[var(--color-text-secondary)]">明细</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[var(--color-border-default)] bg-[var(--color-bg-surface)]">
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

      <div v-if="!loading && total > 0" class="mt-5 flex items-center justify-between">
        <span class="text-sm text-[var(--color-text-secondary)]">共 {{ total }} 条记录</span>
        <div class="flex items-center gap-2">
          <button
            type="button"
            :disabled="page === 1"
            class="rounded-lg border border-[var(--color-border-default)] px-3 py-1.5 text-sm text-[var(--color-text-primary)] transition hover:border-[var(--color-primary)] disabled:cursor-not-allowed disabled:opacity-50"
            @click="changePage(page - 1)"
          >
            上一页
          </button>
          <span class="text-sm text-[var(--color-text-secondary)]">{{ page }} / {{ totalPages }}</span>
          <button
            type="button"
            :disabled="page >= totalPages"
            class="rounded-lg border border-[var(--color-border-default)] px-3 py-1.5 text-sm text-[var(--color-text-primary)] transition hover:border-[var(--color-primary)] disabled:cursor-not-allowed disabled:opacity-50"
            @click="changePage(page + 1)"
          >
            下一页
          </button>
        </div>
      </div>
    </section>
  </div>
</template>
