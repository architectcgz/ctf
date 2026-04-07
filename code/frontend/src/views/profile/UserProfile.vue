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
const { start: startPolling, stop: stopPolling } = useReportStatusPolling()
const currentRole = computed(() => profile.value?.role ?? authStore.user?.role)
const canManagePersonalReport = computed(() => currentRole.value !== 'admin')
const pageCopy = computed(() =>
  canManagePersonalReport.value
    ? '查看账号信息、个人报告与最近导出状态。'
    : '查看账号信息与当前账号状态。'
)

const profileFields = computed(() => {
  const current = profile.value
  if (!current) return []
  return [
    { label: 'Username', value: current.username },
    { label: 'Role', value: current.role },
    { label: 'Class', value: current.class_name || '未分配' },
    { label: 'Name', value: current.name || '未填写' },
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
  void loadProfile()
})

onUnmounted(() => {
  stopPolling()
})
</script>

<template>
  <section
    class="journal-shell journal-shell-user journal-eyebrow-text journal-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
  >
    <div v-if="error" class="profile-inline-notice">
      {{ error }}
    </div>

    <div v-if="loading" class="profile-loading">
      <div class="h-12 animate-pulse rounded-2xl bg-[var(--journal-surface)]/90"></div>
      <div class="grid gap-4 md:grid-cols-2">
        <div class="h-24 animate-pulse rounded-2xl bg-[var(--journal-surface)]"></div>
        <div class="h-24 animate-pulse rounded-2xl bg-[var(--journal-surface)]"></div>
      </div>
      <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_minmax(320px,0.98fr)]">
        <div class="h-72 animate-pulse rounded-[24px] bg-[var(--journal-surface)]"></div>
        <div class="h-72 animate-pulse rounded-[24px] bg-[var(--journal-surface)]"></div>
      </div>
    </div>

    <div v-else class="flex flex-1 flex-col">
      <header class="profile-header">
        <div class="profile-header__intro">
          <div class="journal-eyebrow">Profile</div>
          <h1 class="workspace-page-title">个人资料</h1>
          <p class="workspace-page-copy">{{ pageCopy }}</p>

          <div class="profile-header__actions">
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

        <div v-if="canManagePersonalReport" class="profile-summary-grid">
          <article class="profile-summary-item">
            <div class="profile-summary-icon">
              <ShieldCheck class="h-4 w-4" />
            </div>
            <div>
              <div class="journal-note-label">报告状态</div>
              <div class="profile-summary-value">{{ reportTaskMeta.label }}</div>
            </div>
          </article>
          <article class="profile-summary-item">
            <div class="profile-summary-icon">
              <Activity class="h-4 w-4" />
            </div>
            <div>
              <div class="journal-note-label">最近生成</div>
              <div class="profile-summary-value">
                {{ latestReportCreatedAt ? formatDate(latestReportCreatedAt) : '尚未生成' }}
              </div>
            </div>
          </article>
        </div>
      </header>

      <div class="journal-divider profile-divider" />

      <div class="profile-layout" :class="{ 'profile-layout--single': !canManagePersonalReport }">
        <section class="profile-section">
          <div class="profile-section-head">
            <div>
              <div class="journal-eyebrow journal-eyebrow-soft">Account</div>
              <h2 class="profile-section-title">
                <UserCircle2 class="h-5 w-5 text-[var(--journal-accent)]" />
                账号信息
              </h2>
            </div>
          </div>

          <div v-if="profile" class="profile-field-list">
            <article v-for="item in profileFields" :key="item.label" class="profile-field-item">
              <div class="journal-note-label">{{ item.label }}</div>
              <div class="profile-field-value tech-font">{{ item.value }}</div>
            </article>
          </div>

          <AppEmpty
            v-else
            title="暂无用户信息"
            description="当前没有可展示的用户信息。"
            icon="UsersRound"
          />
        </section>

        <section v-if="canManagePersonalReport" class="profile-section profile-section--report">
          <div class="profile-section-head">
            <div>
              <div class="journal-eyebrow journal-eyebrow-soft">Report</div>
              <h2 class="profile-section-title">个人报告</h2>
            </div>
            <span class="journal-chip" :class="reportTaskMeta.chipClass">
              {{ reportTaskMeta.label }}
            </span>
          </div>

          <div class="profile-status">
            <div class="profile-status__row">
              <div class="profile-status__label">
                <Activity class="h-4 w-4 text-[var(--journal-accent)]" />
                当前状态
              </div>
              <div class="profile-status__value">
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
            <div v-if="latestReport" class="profile-status__meta">
              报告编号：{{ latestReport.report_id }}
            </div>
          </div>

          <fieldset class="profile-fieldset">
            <legend class="profile-fieldset__legend">导出格式</legend>
            <div class="profile-format-grid">
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

          <p v-if="exportError" class="profile-error-copy">
            {{ exportError }}
          </p>

          <template v-if="latestReport">
            <div class="profile-report-meta">
              <div class="profile-report-meta__item">
                <div class="journal-note-label">格式</div>
                <div class="profile-report-meta__value tech-font">
                  {{ latestReportFormat.toUpperCase() }}
                </div>
              </div>
              <div class="profile-report-meta__item">
                <div class="journal-note-label">创建时间</div>
                <div class="profile-report-meta__value">
                  {{ latestReportCreatedAt ? formatDate(latestReportCreatedAt) : '—' }}
                </div>
              </div>
              <div class="profile-report-meta__item">
                <div class="journal-note-label">有效期</div>
                <div class="profile-report-meta__value">
                  {{
                    latestReport.expires_at ? formatDate(latestReport.expires_at) : '待完成后返回'
                  }}
                </div>
              </div>
            </div>

            <p v-if="latestReport.error_message" class="profile-error-copy">
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
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-shell-font: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
  --journal-shell-accent: var(--color-primary);
  --journal-shell-accent-strong: color-mix(in srgb, var(--color-primary-hover) 82%, var(--journal-ink));
  --journal-shell-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-shell-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  --journal-shell-hero-radial-strength: 8%;
  --journal-shell-hero-radial-size: 18rem;
  --journal-shell-hero-end: color-mix(in srgb, var(--journal-surface-subtle) 94%, var(--color-bg-base));
  --journal-shell-hero-shadow: 0 18px 40px rgba(15, 23, 42, 0.05);
  --journal-user-button-height: 2.7rem;
  --journal-user-button-radius: 999px;
  --journal-user-button-padding: 0.62rem 1rem;
  --journal-user-button-size: 0.875rem;
  --journal-user-button-weight: 600;
  --journal-user-button-hover-color: var(--journal-accent);
  --journal-user-button-primary-border: color-mix(in srgb, var(--journal-accent) 32%, transparent);
  --journal-user-button-primary-background: color-mix(
    in srgb,
    var(--journal-accent) 12%,
    var(--journal-surface)
  );
  --journal-user-button-primary-color: color-mix(in srgb, var(--journal-accent) 88%, var(--journal-ink));
  --journal-user-tech-font: 'JetBrains Mono', 'Fira Code', 'SFMono-Regular', monospace;
}

.profile-loading {
  display: grid;
  gap: 1rem;
}

.profile-header {
  display: grid;
  gap: 1rem;
}

.profile-header__actions {
  margin-top: 1rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
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

.profile-summary-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.profile-summary-item {
  display: flex;
  gap: 0.75rem;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 86%, transparent);
  padding-top: 0.85rem;
}

.profile-summary-icon {
  display: inline-flex;
  margin-top: 0.1rem;
  color: var(--journal-accent);
}

.profile-summary-value {
  margin-top: 0.35rem;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.profile-divider {
  margin: 1.2rem 0 0;
}

.profile-layout {
  display: grid;
  gap: 1.5rem;
  padding-top: 1.25rem;
  grid-template-columns: minmax(0, 1fr) minmax(320px, 0.98fr);
}

.profile-layout--single {
  grid-template-columns: minmax(0, 1fr);
}

.profile-section {
  min-width: 0;
}

.profile-section + .profile-section {
  border-left: 1px solid color-mix(in srgb, var(--journal-border) 86%, transparent);
  padding-left: 1.5rem;
}

.profile-section-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.8rem;
}

.profile-section-title {
  margin-top: 0.35rem;
  display: flex;
  align-items: center;
  gap: 0.6rem;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.profile-field-list {
  margin-top: 1rem;
  display: grid;
  gap: 0.85rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.profile-field-item,
.profile-report-meta__item {
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 84%, transparent);
  padding-bottom: 0.85rem;
}

.profile-field-value {
  margin-top: 0.45rem;
  font-size: 1rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.profile-status {
  margin-top: 1rem;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 84%, transparent);
  padding-top: 1rem;
}

.profile-status__row {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: minmax(0, 1fr) auto;
}

.profile-status__label,
.profile-status__value {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.88rem;
  color: var(--journal-ink);
}

.profile-status__meta {
  margin-top: 0.5rem;
  font-size: 0.78rem;
  color: var(--journal-muted);
}

.profile-fieldset {
  margin-top: 1rem;
}

.profile-fieldset__legend {
  margin-bottom: 0.75rem;
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.profile-format-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.journal-format-option {
  display: block;
  cursor: pointer;
  border-radius: 16px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 95%, var(--color-bg-base));
  padding: 0.8rem 0.95rem;
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

.journal-btn--download {
  border-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
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

.profile-report-meta {
  margin-top: 1rem;
  display: grid;
  gap: 0.85rem;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.profile-report-meta__value {
  margin-top: 0.45rem;
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.profile-error-copy {
  margin-top: 0.75rem;
  font-size: 0.84rem;
  color: var(--color-danger);
}

.profile-inline-notice {
  margin-bottom: 1rem;
  border-inline-start: 2px solid color-mix(in srgb, var(--color-warning) 60%, transparent);
  background: color-mix(in srgb, var(--color-warning) 8%, transparent);
  padding: 0.8rem 0.95rem;
  font-size: 0.875rem;
  color: color-mix(in srgb, var(--color-warning) 88%, var(--journal-ink));
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
}

.status-dot-warning {
  background: #f59e0b;
  box-shadow: 0 0 0 2px rgba(245, 158, 11, 0.18);
}

.status-dot-idle {
  background: #94a3b8;
  box-shadow: 0 0 0 2px rgba(148, 163, 184, 0.18);
}

.status-dot-danger {
  background: #ef4444;
  box-shadow: 0 0 0 2px rgba(239, 68, 68, 0.16);
}

@media (max-width: 1024px) {
  .profile-layout {
    grid-template-columns: minmax(0, 1fr);
  }

  .profile-section + .profile-section {
    border-left: 0;
    border-top: 1px solid color-mix(in srgb, var(--journal-border) 86%, transparent);
    padding-left: 0;
    padding-top: 1.25rem;
  }
}

@media (max-width: 720px) {
  .journal-shell {
    padding-inline: 1rem;
  }

  .profile-summary-grid,
  .profile-field-list,
  .profile-format-grid,
  .profile-report-meta {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
