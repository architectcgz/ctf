<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { Activity, FileDown, Loader2, RefreshCw, ShieldCheck, UserCircle2 } from 'lucide-vue-next'

import { downloadReport, exportPersonalReport } from '@/api/assessment'
import { getProfile } from '@/api/auth'
import type { AuthUser, ReportExportData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
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

onUnmounted(() => {
  stopPolling()
})
</script>

<template>
  <section
    class="journal-shell journal-hero flex min-h-full flex-1 flex-col space-y-6 rounded-[30px] border px-6 py-6 md:px-8"
  >
    <div
      v-if="error"
      class="rounded-[20px] border border-[var(--color-warning)]/30 bg-[var(--color-warning)]/8 px-5 py-4 text-sm text-[var(--color-warning)]"
    >
      {{ error }}
    </div>

    <div v-if="loading" class="space-y-6">
      <div class="space-y-6">
        <div class="h-12 animate-pulse rounded-2xl bg-[var(--journal-surface)]/90"></div>
        <div class="grid gap-6 xl:grid-cols-[minmax(0,1.02fr)_minmax(320px,0.98fr)]">
          <div class="h-72 animate-pulse rounded-[24px] bg-[var(--journal-surface)]"></div>
          <div class="h-72 animate-pulse rounded-[24px] bg-[var(--journal-surface)]"></div>
        </div>
      </div>
    </div>

    <div v-else class="flex flex-1 flex-col">
      <div class="grid gap-6 xl:grid-cols-[minmax(0,1.04fr)_minmax(300px,0.96fr)] xl:items-start">
        <div>
          <div class="journal-eyebrow">My Account</div>
          <h1
            class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
          >
            个人资料
          </h1>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            查看账号信息，并在这里生成个人报告。
          </p>

          <div class="mt-6 flex flex-wrap gap-3">
            <div class="profile-pill">
              <span class="status-dot status-dot-ready" />
              账号状态正常
            </div>
            <button type="button" class="journal-btn" @click="loadProfile">
              <RefreshCw class="h-4 w-4" />
              刷新
            </button>
          </div>
        </div>

        <aside class="journal-brief rounded-[24px] border px-5 py-5">
          <div class="flex items-center gap-3 text-sm font-medium text-[var(--journal-ink)]">
            <ShieldCheck class="h-5 w-5 text-[var(--journal-accent)]" />
            当前概况
          </div>
          <div class="mt-5 space-y-3">
            <div class="journal-note">
              <div class="journal-note-label">报告状态</div>
              <div class="mt-2 flex items-center gap-2 text-sm font-semibold text-[var(--journal-ink)]">
                <span
                  class="status-dot"
                  :class="{
                    'status-dot-ready': reportTaskMeta.status === 'ready',
                    'status-dot-warning': reportTaskMeta.status === 'processing',
                    'status-dot-idle': reportTaskMeta.status === 'idle',
                    'status-dot-danger': reportTaskMeta.status === 'failed',
                  }"
                />
                {{ reportTaskMeta.label }}
              </div>
            </div>
            <div class="journal-note">
              <div class="journal-note-label">最近生成</div>
              <div class="journal-note-value">
                {{ latestReportCreatedAt ? formatDate(latestReportCreatedAt) : '尚未生成' }}
              </div>
            </div>
          </div>
        </aside>
      </div>

      <div class="profile-board mt-6 px-1 pt-5 md:px-2 md:pt-6">
        <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_minmax(320px,0.98fr)]">
          <section class="profile-section">
            <div class="profile-section-head">
              <div class="journal-eyebrow journal-eyebrow-soft">Account Info</div>
              <h2 class="mt-3 flex items-center gap-3 text-xl font-semibold text-[var(--journal-ink)]">
                <UserCircle2 class="h-5 w-5 text-[var(--journal-accent)]" />
                账号信息
              </h2>
            </div>

            <div v-if="profile" class="profile-field-list mt-5">
              <article v-for="item in profileFields" :key="item.label" class="profile-field-item">
                <div class="journal-note-label">{{ item.label }}</div>
                <div class="mt-2 text-base font-semibold text-[var(--journal-ink)] tech-font">
                  {{ item.value }}
                </div>
              </article>
            </div>

            <AppEmpty
              v-else
              title="暂无用户信息"
              description="当前没有可展示的用户信息。"
              icon="UsersRound"
            />
          </section>

          <section class="profile-section profile-section--report">
            <div class="profile-section-head">
              <div class="journal-eyebrow journal-eyebrow-soft">Personal Export</div>
              <div class="flex flex-wrap items-start justify-between gap-3">
                <h2 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">个人报告</h2>
                <span class="journal-chip" :class="reportTaskMeta.chipClass">
                  {{ reportTaskMeta.label }}
                </span>
              </div>
            </div>

            <div class="profile-status mt-5">
              <div class="flex items-center justify-between gap-3">
                <div class="flex items-center gap-3 text-sm font-medium text-[var(--journal-ink)]">
                  <Activity class="h-4 w-4 text-[var(--journal-accent)]" />
                  当前状态
                </div>
                <div class="flex items-center gap-2 text-sm font-semibold text-[var(--journal-ink)]">
                  <span
                    class="status-dot"
                    :class="{
                      'status-dot-ready': reportTaskMeta.status === 'ready',
                      'status-dot-warning': reportTaskMeta.status === 'processing',
                      'status-dot-idle': reportTaskMeta.status === 'idle',
                      'status-dot-danger': reportTaskMeta.status === 'failed',
                    }"
                  />
                  {{ reportTaskMeta.label }}
                </div>
              </div>
              <div v-if="latestReport" class="mt-2 text-xs text-[var(--journal-muted)]">
                报告编号：{{ latestReport.report_id }}
              </div>
            </div>

            <fieldset class="mt-5">
              <legend class="mb-3 text-sm font-medium text-[var(--journal-ink)]">导出格式</legend>
              <div class="grid gap-3 sm:grid-cols-2">
                <label
                  class="journal-format-option"
                  :class="{ 'journal-format-option--active': reportFormat === 'pdf' }"
                >
                  <input v-model="reportFormat" type="radio" value="pdf" class="sr-only" />
                  <div class="text-sm font-semibold text-[var(--journal-ink)]">PDF 报告</div>
                  <div class="mt-1 text-xs text-[var(--journal-muted)]">适合阅读和保存</div>
                </label>
                <label
                  class="journal-format-option"
                  :class="{ 'journal-format-option--active': reportFormat === 'excel' }"
                >
                  <input v-model="reportFormat" type="radio" value="excel" class="sr-only" />
                  <div class="text-sm font-semibold text-[var(--journal-ink)]">Excel 报告</div>
                  <div class="mt-1 text-xs text-[var(--journal-muted)]">适合筛选和整理数据</div>
                </label>
              </div>
            </fieldset>

            <button
              type="button"
              class="journal-btn journal-btn--primary mt-4 w-full justify-center"
              :disabled="exportLoading"
              @click="createReport"
            >
              <Loader2 v-if="exportLoading" class="h-4 w-4 animate-spin" />
              {{ exportLoading ? '创建中…' : '生成个人报告' }}
            </button>

            <p v-if="exportError" class="mt-3 text-sm text-[var(--color-danger)]">
              {{ exportError }}
            </p>

            <template v-if="latestReport">
              <div class="profile-report-meta mt-5">
                <div class="profile-report-meta__item">
                  <div class="journal-note-label">格式</div>
                  <div class="mt-2 text-base font-semibold text-[var(--journal-ink)] tech-font">
                    {{ latestReportFormat.toUpperCase() }}
                  </div>
                </div>
                <div class="profile-report-meta__item">
                  <div class="journal-note-label">创建时间</div>
                  <div class="mt-2 text-sm font-semibold text-[var(--journal-ink)]">
                    {{ latestReportCreatedAt ? formatDate(latestReportCreatedAt) : '—' }}
                  </div>
                </div>
                <div class="profile-report-meta__item">
                  <div class="journal-note-label">有效期</div>
                  <div class="mt-2 text-sm font-semibold text-[var(--journal-ink)]">
                    {{ latestReport.expires_at ? formatDate(latestReport.expires_at) : '待完成后返回' }}
                  </div>
                </div>
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
    </div>
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: #4f46e5;
  --journal-accent-strong: #4338ca;
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(37, 99, 235, 0.08), transparent 18rem),
    linear-gradient(180deg, color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 96%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle, var(--color-bg-elevated)) 94%, var(--color-bg-base)));
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.05);
}

.journal-eyebrow {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid rgba(99, 102, 241, 0.22);
  background: rgba(99, 102, 241, 0.07);
  padding: 0.2rem 0.75rem;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.journal-eyebrow-soft {
  color: var(--journal-muted);
  border-color: rgba(148, 163, 184, 0.28);
  background: color-mix(in srgb, var(--journal-border, var(--color-border-default)) 34%, transparent);
}

.journal-note {
  border-radius: 16px;
  border: 1px solid color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
  background: linear-gradient(180deg, color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base)));
  padding: 0.875rem 1rem;
}

.journal-note-label {
  font-size: 0.68rem;
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.journal-note-value {
  margin-top: 0.35rem;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.profile-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  border-radius: 999px;
  border: 1px solid rgba(16, 185, 129, 0.22);
  background: rgba(16, 185, 129, 0.08);
  padding: 0.55rem 0.95rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--journal-ink);
}

.journal-brief {
  border-color: var(--journal-border);
  background: var(--journal-surface-subtle);
}

.journal-format-option {
  display: block;
  cursor: pointer;
  border-radius: 16px;
  border: 1px solid color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
  background: linear-gradient(180deg, color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base)));
  padding: 0.75rem 1rem;
  transition:
    border-color 0.2s,
    background 0.2s;
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
  transition:
    border-color 0.2s,
    color 0.2s;
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

.journal-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 0.4rem 0.8rem;
  font-size: 0.75rem;
  font-weight: 600;
}

.chip--primary {
  color: var(--journal-accent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
}

.chip--success {
  color: var(--color-success);
  background: color-mix(in srgb, var(--color-success) 12%, transparent);
}

.chip--warning {
  color: var(--color-warning);
  background: color-mix(in srgb, var(--color-warning) 12%, transparent);
}

.chip--danger {
  color: var(--color-danger);
  background: color-mix(in srgb, var(--color-danger) 12%, transparent);
}

.profile-board {
  border-top: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
}

.profile-section {
  min-width: 0;
}

.profile-section--report {
  position: relative;
}

.profile-section-head {
  min-height: 5rem;
}

.profile-field-list,
.profile-report-meta {
  border-radius: 22px;
  border: 1px solid color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
}

.profile-field-item,
.profile-report-meta__item {
  padding: 1rem 1.1rem;
}

.profile-field-item + .profile-field-item,
.profile-report-meta__item + .profile-report-meta__item {
  border-top: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
}

.profile-status {
  border-radius: 20px;
  border: 1px solid color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
  padding: 1rem 1.1rem;
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
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.tech-font {
  font-family: 'JetBrains Mono', 'Fira Code', 'SFMono-Regular', monospace;
}

@media (min-width: 1280px) {
  .profile-section--report {
    padding-left: 1.5rem;
  }

  .profile-section--report::before {
    content: '';
    position: absolute;
    left: -0.75rem;
    top: 0;
    bottom: 0;
    border-left: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
  }
}

:global([data-theme='dark']) .journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
}

:global([data-theme='dark']) .journal-hero {
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.18), transparent 20rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.95), rgba(2, 6, 23, 0.98));
}

:global([data-theme='dark']) .journal-note,
:global([data-theme='dark']) .journal-format-option,
:global([data-theme='dark']) .profile-status,
:global([data-theme='dark']) .profile-field-list,
:global([data-theme='dark']) .profile-report-meta {
  background: rgba(15, 23, 42, 0.42);
}
</style>
