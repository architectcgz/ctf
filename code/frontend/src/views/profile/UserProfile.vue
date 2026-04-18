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
const { start: startPolling, stop: stopPolling } = useReportStatusPolling()
const currentRole = computed(() => profile.value?.role ?? authStore.user?.role)
const canManagePersonalReport = computed(() => currentRole.value !== 'admin')
const currentProfile = computed(() => profile.value ?? authStore.user ?? null)
const pageCopy = computed(() =>
  canManagePersonalReport.value
    ? '查看账号信息、个人报告与最近导出状态。'
    : '查看账号信息与当前账号状态。'
)

function getRoleLabel(role: AuthUser['role'] | undefined): string {
  if (role === 'admin') return '管理员'
  if (role === 'teacher') return '教师'
  if (role === 'student') return '学生'
  return '未知'
}

const profileFields = computed(() => {
  const current = currentProfile.value
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

const profileSummaryItems = computed(() => {
  const current = currentProfile.value
  const className = current?.class_name || '未分配'
  const displayName = current?.name || '未填写'

  return [
    {
      key: 'status',
      label: '账号状态',
      value: '正常',
      helper: '当前账号可正常访问个人工作区',
      icon: ShieldCheck,
      techFont: false,
    },
    {
      key: 'role',
      label: '当前角色',
      value: getRoleLabel(current?.role),
      helper: '决定当前账号可访问的功能范围',
      icon: UserCircle2,
      techFont: false,
    },
    {
      key: canManagePersonalReport.value ? 'report' : 'name',
      label: canManagePersonalReport.value ? '报告状态' : '实名信息',
      value: canManagePersonalReport.value ? reportTaskMeta.value.label : displayName,
      helper: canManagePersonalReport.value
        ? latestReportCreatedAt.value
          ? `最近生成于 ${formatDate(latestReportCreatedAt.value)}`
          : '当前还没有生成过个人报告'
        : current?.name
          ? '用于账号展示与身份识别'
          : '当前未填写姓名信息',
      icon: canManagePersonalReport.value ? Activity : UserCircle2,
      techFont: false,
    },
    {
      key: 'class',
      label: '所属班级',
      value: className,
      helper: current?.class_name ? '当前归属的班级信息' : '当前账号还未绑定班级',
      icon: UserCircle2,
      techFont: false,
    },
  ]
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
    class="workspace-shell journal-shell journal-shell-user journal-hero flex min-h-full flex-1 flex-col"
  >
    <main class="content-pane">
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

      <div v-else class="profile-page flex flex-1 flex-col">
        <PageHeader
          class="profile-topbar"
          title="个人资料"
          :description="pageCopy"
          eyebrow="Profile"
        >
          <div class="profile-topbar-actions">
            <div class="profile-pill">
              <span class="status-dot status-dot-ready" />
              账号状态正常
            </div>
            <button type="button" class="journal-btn" @click="loadProfile">
              <RefreshCw class="h-4 w-4" />
              刷新
            </button>
          </div>
        </PageHeader>

      <section class="profile-summary" aria-label="账号概况">
        <div class="profile-summary-title">
          <ShieldCheck class="h-4 w-4" />
          <span>账号概况</span>
        </div>
        <div class="profile-summary-grid metric-panel-grid">
          <article
            v-for="item in profileSummaryItems"
            :key="item.key"
            class="profile-summary-item metric-panel-card"
          >
            <div class="profile-summary-icon">
              <component :is="item.icon" class="h-4 w-4" />
            </div>
            <div>
              <div class="journal-note-label metric-panel-label">{{ item.label }}</div>
              <div
                class="profile-summary-value metric-panel-value"
                :class="{ 'tech-font': item.techFont }"
              >
                {{ item.value }}
              </div>
              <div class="journal-note-helper metric-panel-helper">{{ item.helper }}</div>
            </div>
          </article>
        </div>
      </section>

      <div class="journal-divider profile-divider" />

      <div class="profile-layout" :class="{ 'profile-layout--single': !canManagePersonalReport }">
        <section class="profile-section">
          <div class="profile-section-head">
            <div>
              <div class="profile-section-kicker">Account</div>
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
              <div class="profile-section-kicker">Report</div>
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
    </main>
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-shell-font: var(--font-family-sans);
  --journal-shell-accent: var(--color-primary);
  --journal-shell-accent-strong: color-mix(
    in srgb,
    var(--color-primary-hover) 82%,
    var(--journal-ink)
  );
  --journal-shell-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-shell-surface-subtle: color-mix(
    in srgb,
    var(--color-bg-surface) 74%,
    var(--color-bg-base)
  );
  --journal-shell-hero-radial-strength: 8%;
  --journal-shell-hero-radial-size: 18rem;
  --journal-shell-hero-end: color-mix(
    in srgb,
    var(--journal-surface-subtle) 94%,
    var(--color-bg-base)
  );
  --journal-shell-hero-shadow: 0 18px 40px var(--color-shadow-soft);
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
  --journal-user-button-primary-color: color-mix(
    in srgb,
    var(--journal-accent) 88%,
    var(--journal-ink)
  );
  --journal-user-tech-font: var(--font-family-mono);
}

.profile-loading {
  display: grid;
  gap: 1rem;
}

.profile-page {
  min-height: 100%;
}

.profile-topbar-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: 0.75rem;
}

.profile-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--color-success) 22%, transparent);
  background: color-mix(in srgb, var(--color-success) 8%, transparent);
  padding: 0.55rem 0.95rem;
  font-size: var(--font-size-0-875);
  font-weight: 500;
  color: var(--journal-ink);
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

.profile-section-kicker {
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.profile-section-title {
  margin-top: 0.35rem;
  display: flex;
  align-items: center;
  gap: 0.6rem;
  font-size: var(--font-size-1-15);
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
  font-size: var(--font-size-1-00);
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
  font-size: var(--font-size-0-88);
  color: var(--journal-ink);
}

.profile-status__meta {
  margin-top: 0.5rem;
  font-size: var(--font-size-0-78);
  color: var(--journal-muted);
}

.profile-fieldset {
  margin-top: 1rem;
}

.profile-fieldset__legend {
  margin-bottom: 0.75rem;
  font-size: var(--font-size-0-88);
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
  font-size: var(--font-size-0-75);
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
  font-size: var(--font-size-0-90);
  font-weight: 600;
  color: var(--journal-ink);
}

.profile-error-copy {
  margin-top: 0.75rem;
  font-size: var(--font-size-0-84);
  color: var(--color-danger);
}

.profile-inline-notice {
  margin-bottom: 1rem;
  border-inline-start: 2px solid color-mix(in srgb, var(--color-warning) 60%, transparent);
  background: color-mix(in srgb, var(--color-warning) 8%, transparent);
  padding: 0.8rem 0.95rem;
  font-size: var(--font-size-0-875);
  color: color-mix(in srgb, var(--color-warning) 88%, var(--journal-ink));
}

.profile-divider {
  margin: 0;
}

.status-dot {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-dot-ready {
  background: var(--color-success);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-success) 20%, transparent);
}

.status-dot-warning {
  background: var(--color-warning);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-warning) 18%, transparent);
}

.status-dot-idle {
  background: var(--color-text-muted);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-text-muted) 18%, transparent);
}

.status-dot-danger {
  background: var(--color-danger);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-danger) 16%, transparent);
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
  .content-pane {
    padding-inline: 1rem;
  }

  .profile-topbar-actions {
    justify-content: flex-start;
  }

  .profile-field-list,
  .profile-format-grid,
  .profile-report-meta {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
