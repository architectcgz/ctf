<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { Activity, FileDown, Loader2, RefreshCw, ShieldCheck, UserCircle2 } from 'lucide-vue-next'

import { downloadReport, exportPersonalReport } from '@/api/assessment'
import { getProfile } from '@/api/auth'
import type { AuthUser, ReportExportData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useReportStatusPolling } from '@/composables/useReportStatusPolling'
import { useAuthStore } from '@/stores/auth'
import { formatDate } from '@/utils/format'

const authStore = useAuthStore()

const loading = ref(false)
const error = ref<string | null>(null)
const profile = ref<AuthUser | null>(null)
const exportLoading = ref(false)
const exportError = ref<string | null>(null)
const reportFormat = ref<'pdf' | 'excel'>('pdf')
const latestReport = ref<ReportExportData | null>(null)
const latestReportFormat = ref<'pdf' | 'excel'>('pdf')
const latestReportCreatedAt = ref<string | null>(null)
const { polling, start: startPolling, stop: stopPolling } = useReportStatusPolling()

const profileFields = computed(() => {
  const current = profile.value
  if (!current) return []
  return [
    { label: 'Username', value: current.username, icon: 'user' },
    { label: 'Role', value: current.role, icon: 'role' },
    { label: 'Class', value: current.class_name || '未分配', icon: 'class' },
    { label: 'Name', value: current.name || '未填写', icon: 'name' },
  ]
})

const reportTaskMeta = computed(() => {
  if (latestReport.value?.status === 'ready') {
    return { label: '可下载', status: 'ready', chipClass: 'chip--success' }
  }
  if (latestReport.value?.status === 'failed') {
    return { label: '生成失败', status: 'failed', chipClass: 'chip--danger' }
  }
  if (latestReport.value?.status === 'processing') {
    return { label: '生成中', status: 'processing', chipClass: 'chip--warning' }
  }
  return { label: '待创建', status: 'idle', chipClass: 'chip--primary' }
})

async function loadProfile(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    profile.value = await getProfile()
  } catch (err) {
    console.error('加载个人资料失败:', err)
    profile.value = authStore.user ? { ...authStore.user } : null
    error.value = '加载个人资料失败，以下展示的是本地缓存信息'
  } finally {
    loading.value = false
  }
}

async function createReport(): Promise<void> {
  exportLoading.value = true
  exportError.value = null
  try {
    latestReportFormat.value = reportFormat.value
    latestReportCreatedAt.value = new Date().toISOString()
    latestReport.value = await exportPersonalReport({ format: reportFormat.value })
    if (latestReport.value.status === 'processing') {
      startPolling(String(latestReport.value.report_id), (next) => {
        latestReport.value = next
      })
    } else {
      stopPolling()
      if (latestReport.value.status === 'failed') {
        exportError.value = latestReport.value.error_message || '个人报告生成失败，请稍后重试'
      }
    }
  } catch (err) {
    console.error('导出个人报告失败:', err)
    exportError.value = '创建个人报告失败，请稍后重试'
  } finally {
    exportLoading.value = false
  }
}

async function handleDownload(): Promise<void> {
  if (!latestReport.value) return
  const file = await downloadReport(latestReport.value.report_id)
  const url = window.URL.createObjectURL(file.blob)
  const link = document.createElement('a')
  link.href = url
  link.download = file.filename
  link.click()
  window.URL.revokeObjectURL(url)
}

onMounted(() => {
  loadProfile()
})
</script>

<template>
  <div class="journal-shell space-y-6">
    <PageHeader
      eyebrow="My Account"
      title="个人资料"
      description="查看当前账号信息，并生成你的个人训练报告用于复盘和归档。"
    >
      <button type="button" class="journal-btn" @click="loadProfile">
        <RefreshCw class="h-4 w-4" />
        刷新
      </button>
    </PageHeader>

    <!-- 错误提示 -->
    <div
      v-if="error"
      class="rounded-[20px] border border-[var(--color-warning)]/30 bg-[var(--color-warning)]/8 px-5 py-4 text-sm text-[var(--color-warning)]"
    >
      {{ error }}
    </div>

    <!-- 加载骨架 -->
    <div v-if="loading" class="grid gap-6 xl:grid-cols-[1fr_0.9fr]">
      <div class="journal-hero rounded-[30px] border px-6 py-6">
        <div class="h-64 animate-pulse rounded-2xl bg-[var(--journal-surface)]"></div>
      </div>
      <div class="journal-hero rounded-[30px] border px-6 py-6">
        <div class="h-64 animate-pulse rounded-2xl bg-[var(--journal-surface)]"></div>
      </div>
    </div>

    <div v-else class="grid gap-6 xl:grid-cols-[1fr_0.9fr]">
      <!-- 账号信息 -->
      <section class="journal-hero rounded-[30px] border px-6 py-6 md:px-8">
        <div class="journal-eyebrow">Account Info</div>
        <h2 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]">
          {{ profile?.name || profile?.username || '—' }}
        </h2>
        <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
          聚合你的账号身份、班级归属和基础资料，方便快速确认当前训练账号的信息状态。
        </p>

        <!-- 账号状态栏 -->
        <div class="mt-5 rounded-[18px] border border-[var(--journal-border)] bg-[var(--journal-surface-subtle)] px-4 py-3">
          <div class="flex items-center gap-2 text-sm text-[var(--journal-muted)]">
            <span class="status-dot status-dot-ready" />
            账号状态正常
          </div>
          <div class="mt-1 tech-font text-xs text-[var(--journal-muted)]">
            uid://{{ profile?.id ?? '—' }}
          </div>
        </div>

        <div v-if="profile" class="mt-5 grid gap-3 sm:grid-cols-2">
          <article
            v-for="item in profileFields"
            :key="item.label"
            class="journal-metric rounded-[20px] border px-4 py-4"
          >
            <div class="text-[11px] font-semibold uppercase tracking-[0.2em] text-[var(--journal-muted)]">{{ item.label }}</div>
            <div class="mt-2 text-lg font-semibold text-[var(--journal-ink)] tech-font">{{ item.value }}</div>
          </article>
        </div>

        <AppEmpty
          v-else
          title="暂无用户信息"
          description="当前没有可展示的用户信息。"
          icon="UsersRound"
        />
      </section>

      <!-- 个人报告导出 -->
      <section class="journal-hero rounded-[30px] border px-6 py-6 md:px-8">
        <div class="flex items-start justify-between gap-4">
          <div>
            <div class="journal-eyebrow">Personal Export</div>
            <h2 class="mt-3 text-2xl font-semibold tracking-tight text-[var(--journal-ink)]">
              个人报告生成器
            </h2>
            <p class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">
              选择导出格式后创建任务，系统将自动跟踪生成状态并在就绪后提供下载。
            </p>
          </div>
          <span class="shrink-0 rounded-full px-3 py-1 text-xs font-semibold" :class="reportTaskMeta.chipClass">
            {{ reportTaskMeta.label }}
          </span>
        </div>

        <!-- 报告状态面板 -->
        <div class="mt-5 rounded-[18px] border border-[var(--journal-border)] bg-[var(--journal-surface-subtle)] px-4 py-3">
          <div class="flex items-center justify-between gap-3">
            <div class="flex items-center gap-3">
              <Activity class="h-4 w-4 text-[var(--journal-accent)]" />
              <div class="text-sm font-medium text-[var(--journal-ink)]">报告任务状态</div>
            </div>
            <div class="flex items-center gap-2">
              <span
                class="status-dot"
                :class="{
                  'status-dot-ready': reportTaskMeta.status === 'ready',
                  'status-dot-warning': reportTaskMeta.status === 'processing',
                  'status-dot-idle': reportTaskMeta.status === 'idle',
                  'status-dot-danger': reportTaskMeta.status === 'failed',
                }"
              />
              <span class="tech-font text-sm font-medium text-[var(--journal-ink)]">{{ reportTaskMeta.label }}</span>
            </div>
          </div>
          <div v-if="latestReport" class="mt-1 tech-font text-xs text-[var(--journal-muted)]">
            report://{{ latestReport.report_id }}
          </div>
        </div>

        <!-- 格式选择 -->
        <fieldset class="mt-5">
          <legend class="mb-3 text-sm font-medium text-[var(--journal-ink)]">导出格式</legend>
          <div class="grid gap-3 sm:grid-cols-2">
            <label
              class="journal-format-option"
              :class="{ 'journal-format-option--active': reportFormat === 'pdf' }"
            >
              <input v-model="reportFormat" type="radio" value="pdf" class="sr-only" />
              <div class="text-sm font-semibold text-[var(--journal-ink)]">PDF 报告</div>
              <div class="mt-1 text-xs text-[var(--journal-muted)]">适合打印或归档留存</div>
            </label>
            <label
              class="journal-format-option"
              :class="{ 'journal-format-option--active': reportFormat === 'excel' }"
            >
              <input v-model="reportFormat" type="radio" value="excel" class="sr-only" />
              <div class="text-sm font-semibold text-[var(--journal-ink)]">Excel 报告</div>
              <div class="mt-1 text-xs text-[var(--journal-muted)]">适合数据筛选与二次分析</div>
            </label>
          </div>
        </fieldset>

        <!-- 创建按钮 -->
        <button
          type="button"
          class="journal-btn journal-btn--primary mt-4 w-full justify-center"
          :disabled="exportLoading"
          @click="createReport"
        >
          <Loader2 v-if="exportLoading" class="h-4 w-4 animate-spin" />
          {{ exportLoading ? '创建中…' : '创建导出任务' }}
        </button>

        <p v-if="exportError" class="mt-3 text-sm text-[var(--color-danger)]">
          {{ exportError }}
        </p>

        <!-- 最近报告状态 -->
        <template v-if="latestReport">
          <div class="mt-5 grid grid-cols-2 gap-3">
            <article class="journal-metric rounded-[18px] border px-4 py-4">
              <div class="text-[11px] font-semibold uppercase tracking-[0.2em] text-[var(--journal-muted)]">格式</div>
              <div class="mt-2 tech-font text-lg font-semibold text-[var(--journal-ink)]">{{ latestReportFormat.toUpperCase() }}</div>
            </article>
            <article class="journal-metric rounded-[18px] border px-4 py-4">
              <div class="text-[11px] font-semibold uppercase tracking-[0.2em] text-[var(--journal-muted)]">创建时间</div>
              <div class="mt-2 text-sm font-semibold text-[var(--journal-ink)]">
                {{ latestReportCreatedAt ? formatDate(latestReportCreatedAt) : '—' }}
              </div>
            </article>
            <article class="journal-metric rounded-[18px] border px-4 py-4">
              <div class="text-[11px] font-semibold uppercase tracking-[0.2em] text-[var(--journal-muted)]">状态</div>
              <div class="mt-2 text-lg font-semibold text-[var(--journal-ink)]">{{ reportTaskMeta.label }}</div>
            </article>
            <article class="journal-metric rounded-[18px] border px-4 py-4">
              <div class="text-[11px] font-semibold uppercase tracking-[0.2em] text-[var(--journal-muted)]">有效期</div>
              <div class="mt-2 text-sm font-semibold text-[var(--journal-ink)]">
                {{ latestReport.expires_at ? formatDate(latestReport.expires_at) : '待完成后返回' }}
              </div>
            </article>
          </div>

          <p v-if="latestReport.error_message" class="mt-3 text-sm text-[var(--color-danger)]">
            {{ latestReport.error_message }}
          </p>

          <button
            type="button"
            class="journal-btn journal-btn--download mt-4 w-full justify-center"
            :disabled="latestReport.status !== 'ready'"
            @click="handleDownload"
          >
            <FileDown class="h-4 w-4" />
            下载最近报告
          </button>
        </template>
      </section>
    </div>
  </div>
</template>

<style scoped>
.journal-shell {
  --journal-ink: #0f172a;
  --journal-muted: #64748b;
  --journal-accent: #4f46e5;
  --journal-accent-strong: #4338ca;
  --journal-border: rgba(226, 232, 240, 0.8);
  --journal-surface: rgba(248, 250, 252, 0.9);
  --journal-surface-subtle: rgba(241, 245, 249, 0.7);
  font-family: "Inter", "Noto Sans SC", system-ui, sans-serif;
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.08), transparent 18rem),
    linear-gradient(180deg, #ffffff, #f8fafc);
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
}

.journal-eyebrow {
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
  padding: 0.625rem 0.875rem;
}

.journal-note-label {
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.journal-note-value {
  margin-top: 0.35rem;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.journal-brief {
  border-color: var(--journal-border);
  background: var(--journal-surface-subtle);
}

.journal-metric {
  border-color: var(--journal-border);
  background: var(--journal-surface);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.04);
  transition: border-color 0.2s, box-shadow 0.2s;
}

.journal-format-option {
  display: block;
  cursor: pointer;
  border-radius: 16px;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.75rem 1rem;
  transition: border-color 0.2s, background 0.2s;
}

.journal-format-option:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 40%, transparent);
}

.journal-format-option--active {
  border-color: color-mix(in srgb, var(--journal-accent) 60%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 6%, transparent);
}

.journal-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  border-radius: 12px;
  border: 1px solid var(--color-border-default);
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-primary);
  background: transparent;
  transition: border-color 0.2s, color 0.2s;
  cursor: pointer;
}

.journal-btn:hover {
  border-color: var(--journal-accent);
  color: var(--journal-accent);
}

.journal-btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.journal-btn--primary {
  border-color: color-mix(in srgb, var(--journal-accent) 50%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  color: var(--journal-accent);
}

.journal-btn--primary:hover:not(:disabled) {
  background: color-mix(in srgb, var(--journal-accent) 14%, transparent);
}

.status-dot {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-dot-ready {
  background: #10b981;
  box-shadow: 0 0 0 2px rgba(16, 185, 129, 0.2);
  animation: pulse-dot 2s infinite;
}

.status-dot-warning {
  background: #f59e0b;
  box-shadow: 0 0 0 2px rgba(245, 158, 11, 0.2);
  animation: pulse-dot 2s infinite;
}

.status-dot-idle {
  background: #94a3b8;
}

.status-dot-danger {
  background: #ef4444;
  box-shadow: 0 0 0 2px rgba(239, 68, 68, 0.2);
}

@keyframes pulse-dot {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.tech-font {
  font-family: "JetBrains Mono", "Fira Code", "SFMono-Regular", monospace;
}

:global([data-theme='dark']) .journal-shell {
  --journal-ink: #f1f5f9;
  --journal-muted: #94a3b8;
  --journal-border: rgba(51, 65, 85, 0.72);
  --journal-surface: rgba(15, 23, 42, 0.85);
  --journal-surface-subtle: rgba(30, 41, 59, 0.6);
}

:global([data-theme='dark']) .journal-hero {
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.18), transparent 20rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.95), rgba(2, 6, 23, 0.98));
}
</style>