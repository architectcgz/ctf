<script setup lang="ts">
import { computed, ref } from 'vue'

import { downloadReport } from '@/api/assessment'
import {
  exportClassReport,
  getClassReview,
  getClassStudents,
  getClassSummary,
  getClassTrend,
} from '@/api/teacher'
import type {
  ReportExportData,
  TeacherClassReviewData,
  TeacherClassSummaryData,
  TeacherClassTrendData,
  TeacherStudentItem,
} from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import TeacherClassInsightsPanel from '@/components/teacher/TeacherClassInsightsPanel.vue'
import TeacherClassReviewPanel from '@/components/teacher/TeacherClassReviewPanel.vue'
import TeacherClassTrendPanel from '@/components/teacher/TeacherClassTrendPanel.vue'
import { useReportStatusPolling } from '@/composables/useReportStatusPolling'
import { useToast } from '@/composables/useToast'
import { useAuthStore } from '@/stores/auth'
import { formatDate } from '@/utils/format'

type ReportFormat = 'pdf' | 'excel'

interface ExportRecord {
  className: string
  format: ReportFormat
  createdAt: string
  result: ReportExportData
}

const authStore = useAuthStore()
const toast = useToast()
const { polling, start: startPolling, stop: stopPolling } = useReportStatusPolling()

const form = ref({
  className: authStore.user?.class_name ?? '',
  format: 'pdf' as ReportFormat,
})

const submitting = ref(false)
const downloading = ref(false)
const latestExport = ref<ExportRecord | null>(null)
const previewDialogVisible = ref(false)
const previewLoading = ref(false)
const previewError = ref<string | null>(null)
const previewClassName = ref('')
const previewStudents = ref<TeacherStudentItem[]>([])
const previewReview = ref<TeacherClassReviewData | null>(null)
const previewSummary = ref<TeacherClassSummaryData | null>(null)
const previewTrend = ref<TeacherClassTrendData | null>(null)

const classNamePlaceholder = computed(() =>
  authStore.user?.class_name ? `默认班级：${authStore.user.class_name}` : '请输入要导出的班级名称'
)

const normalizedClassNameText = computed(() => normalizeClassName() || '未选择')
const selectedFormatLabel = computed(() => (form.value.format === 'pdf' ? 'PDF' : 'Excel'))

const selectedFormatHint = computed(() =>
  form.value.format === 'pdf' ? '适合打印、归档和正式汇报。' : '适合继续分析、筛选和二次加工。'
)

const derivedDownloadHint = computed(() => {
  if (!latestExport.value) return ''
  if (latestExport.value.result.status === 'ready') {
    return '报告已生成，可直接下载。'
  }
  if (latestExport.value.result.status === 'failed') {
    return latestExport.value.result.error_message || '报告生成失败，请重新发起导出任务。'
  }
  return '正在轮询导出状态，生成完成后会自动更新为可下载。'
})

const averageSolvedText = computed(() => {
  if (!previewSummary.value) return '--'
  return previewSummary.value.average_solved.toFixed(1)
})

const activeRateText = computed(() => {
  if (!previewSummary.value) return '--'
  return `${Math.round(previewSummary.value.active_rate)}%`
})

const latestStatusMeta = computed(() => {
  if (!latestExport.value) {
    return {
      label: '未创建',
      chipClass: 'report-status-chip--idle',
    }
  }

  switch (latestExport.value.result.status) {
    case 'ready':
      return {
        label: '已就绪',
        chipClass: 'report-status-chip--ready',
      }
    case 'failed':
      return {
        label: '失败',
        chipClass: 'report-status-chip--failed',
      }
    default:
      return {
        label: polling.value ? '生成中' : '等待更新',
        chipClass: 'report-status-chip--pending',
      }
  }
})

const latestExpiresText = computed(() => {
  if (!latestExport.value) return '--'
  return latestExport.value.result.expires_at
    ? formatDate(latestExport.value.result.expires_at)
    : '待生成完成后返回'
})

function normalizeClassName(): string {
  return form.value.className.trim() || authStore.user?.class_name?.trim() || ''
}

async function loadPreview(): Promise<void> {
  const className = normalizeClassName()
  if (!className) {
    previewClassName.value = ''
    previewStudents.value = []
    previewReview.value = null
    previewSummary.value = null
    previewTrend.value = null
    previewError.value = '请先填写班级名称'
    return
  }

  previewLoading.value = true
  previewError.value = null
  previewClassName.value = className

  try {
    const [students, review, summary, trend] = await Promise.all([
      getClassStudents(className),
      getClassReview(className),
      getClassSummary(className),
      getClassTrend(className),
    ])
    previewStudents.value = students
    previewReview.value = review
    previewSummary.value = summary
    previewTrend.value = trend
  } catch (err) {
    console.error('加载报告预览失败:', err)
    previewStudents.value = []
    previewReview.value = null
    previewSummary.value = null
    previewTrend.value = null
    previewError.value = '加载当前班级预览失败，请稍后重试'
  } finally {
    previewLoading.value = false
  }
}

async function openPreviewDialog(): Promise<void> {
  previewDialogVisible.value = true
  await loadPreview()
}

async function handleExport(): Promise<void> {
  const className = normalizeClassName()
  if (!className) {
    toast.warning('请先填写班级名称')
    return
  }

  submitting.value = true
  try {
    const result = await exportClassReport({
      class_name: className,
      format: form.value.format,
    })

    latestExport.value = {
      className,
      format: form.value.format,
      createdAt: new Date().toISOString(),
      result,
    }

    if (result.status === 'ready') {
      stopPolling()
      toast.success('报告已生成，可立即下载')
    } else if (result.status === 'failed') {
      stopPolling()
      toast.error(result.error_message || '报告生成失败')
    } else {
      startPolling(String(result.report_id), (next) => {
        if (!latestExport.value) return
        latestExport.value = {
          ...latestExport.value,
          result: next,
        }
      })
      toast.info('报告开始生成，系统会自动刷新任务状态')
    }
  } finally {
    submitting.value = false
  }
}

async function handleDownload(): Promise<void> {
  if (!latestExport.value) return

  downloading.value = true
  try {
    const { blob, filename } = await downloadReport(latestExport.value.result.report_id)
    const objectUrl = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = objectUrl
    link.download = filename
    document.body.appendChild(link)
    link.click()
    link.remove()
    URL.revokeObjectURL(objectUrl)
    toast.success('下载已开始')
  } finally {
    downloading.value = false
  }
}
</script>

<template>
  <section
    class="report-shell report-hero journal-shell teacher-management-shell teacher-surface flex min-h-full flex-1 flex-col rounded-[30px] px-6 py-6 md:px-8"
  >
    <div class="report-page">
      <header class="report-topbar">
        <div class="report-header__intro">
          <div class="teacher-surface-eyebrow journal-eyebrow">Teacher Export</div>
          <h1 class="report-title">报告导出</h1>
          <p class="report-copy">选择班级并创建导出任务，生成完成后下载报告文件。</p>
        </div>
      </header>

      <section class="report-summary">
        <div class="report-summary-title">Export Snapshot</div>
        <div class="report-summary-grid">
          <div class="report-summary-item">
            <div class="report-summary-label">当前账号</div>
            <div class="report-summary-value">{{ authStore.user?.username || '-' }}</div>
            <div class="report-summary-helper">用于发起当前导出任务的教师账号</div>
          </div>
          <div class="report-summary-item">
            <div class="report-summary-label">默认班级</div>
            <div class="report-summary-value">{{ authStore.user?.class_name || '未绑定' }}</div>
            <div class="report-summary-helper">留空时将优先使用当前账号绑定班级</div>
          </div>
          <div class="report-summary-item">
            <div class="report-summary-label">当前格式</div>
            <div class="report-summary-value">{{ selectedFormatLabel }}</div>
            <div class="report-summary-helper">{{ selectedFormatHint }}</div>
          </div>
          <div class="report-summary-item">
            <div class="report-summary-label">最近状态</div>
            <div class="report-summary-value">{{ latestStatusMeta.label }}</div>
            <div class="report-summary-helper">
              {{
                latestExport ? derivedDownloadHint : '创建一次导出任务后，这里会同步展示最新状态。'
              }}
            </div>
          </div>
        </div>
      </section>

      <div class="report-hero-divider" />

      <div class="report-workspace">
        <div class="report-main">
          <section class="report-flat-section">
            <div class="report-section-head">
              <div>
                <div class="teacher-surface-eyebrow journal-eyebrow">Export Task</div>
                <h2 class="report-section-title">创建导出任务</h2>
              </div>
            </div>

            <div class="report-command-grid">
              <div class="space-y-4">
                <label class="space-y-2">
                  <span class="text-sm text-[var(--journal-muted)]">班级名称</span>
                  <input
                    v-model="form.className"
                    type="text"
                    :placeholder="classNamePlaceholder"
                    class="report-field w-full px-4 py-3 text-sm outline-none transition"
                  />
                </label>

                <fieldset class="space-y-2">
                  <legend class="text-sm text-[var(--journal-muted)]">导出格式</legend>
                  <div class="report-format-grid">
                    <label
                      class="report-format-option"
                      :class="{ 'report-format-option--active': form.format === 'pdf' }"
                    >
                      <input v-model="form.format" type="radio" value="pdf" class="mt-1" />
                      <span>
                        <span class="block font-medium text-[var(--journal-ink)]">PDF</span>
                        <span class="mt-1 block text-sm text-[var(--journal-muted)]">
                          适合打印、归档和正式汇报。
                        </span>
                      </span>
                    </label>

                    <label
                      class="report-format-option"
                      :class="{ 'report-format-option--active': form.format === 'excel' }"
                    >
                      <input v-model="form.format" type="radio" value="excel" class="mt-1" />
                      <span>
                        <span class="block font-medium text-[var(--journal-ink)]">Excel</span>
                        <span class="mt-1 block text-sm text-[var(--journal-muted)]">
                          适合继续分析、筛选和二次加工。
                        </span>
                      </span>
                    </label>
                  </div>
                </fieldset>

                <div class="flex flex-wrap gap-3">
                  <button
                    type="button"
                    :disabled="previewLoading"
                    class="teacher-btn"
                    @click="openPreviewDialog"
                  >
                    {{ previewLoading ? '加载预览中...' : '打开报告预览' }}
                  </button>
                  <button
                    type="button"
                    :disabled="submitting"
                    class="teacher-btn teacher-btn--primary"
                    @click="handleExport"
                  >
                    {{ submitting ? '提交中...' : '创建导出任务' }}
                  </button>
                </div>
              </div>

              <aside class="report-side-notes">
                <article class="report-note">
                  <div class="report-note-label">当前目标班级</div>
                  <div class="report-note-value">{{ normalizedClassNameText }}</div>
                  <div class="report-note-helper">导出与预览都将以这个班级名称作为优先目标。</div>
                </article>
                <article class="report-note">
                  <div class="report-note-label">当前导出格式</div>
                  <div class="report-note-value">{{ selectedFormatLabel }}</div>
                  <div class="report-note-helper">{{ selectedFormatHint }}</div>
                </article>
                <article class="report-note">
                  <div class="report-note-label">建议流程</div>
                  <ol class="report-guide-list">
                    <li>1. 先打开报告预览，确认当前班级数据内容。</li>
                    <li>2. 需要留档时，再创建后台导出任务。</li>
                    <li>3. 任务完成后下载文件，生成中会自动轮询状态。</li>
                  </ol>
                </article>
              </aside>
            </div>
          </section>

          <section class="report-flat-section">
            <div class="report-section-head">
              <div>
                <div class="teacher-surface-eyebrow journal-eyebrow">Latest Task</div>
                <h2 class="report-section-title">最近一次任务</h2>
              </div>
            </div>

            <AppEmpty
              v-if="!latestExport"
              class="mt-5"
              title="还没有创建导出任务"
              description="先创建一次班级报告任务，这里会展示最近一次任务状态。"
              icon="FileChartColumnIncreasing"
            />

            <div v-else class="mt-5 space-y-4">
              <div class="report-status-banner">
                <div>
                  <div class="report-note-label">任务编号</div>
                  <div class="report-note-value">#{{ latestExport.result.report_id }}</div>
                  <div class="report-note-helper">{{ derivedDownloadHint }}</div>
                </div>
                <span class="report-status-chip" :class="latestStatusMeta.chipClass">
                  {{ latestStatusMeta.label }}
                </span>
              </div>

              <div class="report-kpi-grid report-kpi-grid--task">
                <article class="journal-brief journal-metric report-kpi-card">
                  <div class="report-kpi-label">班级</div>
                  <div class="report-kpi-value">{{ latestExport.className }}</div>
                  <div class="report-kpi-hint">本次导出的目标班级</div>
                </article>
                <article class="journal-brief journal-metric report-kpi-card">
                  <div class="report-kpi-label">格式</div>
                  <div class="report-kpi-value">{{ latestExport.format.toUpperCase() }}</div>
                  <div class="report-kpi-hint">本次任务选择的导出文件格式</div>
                </article>
                <article class="journal-brief journal-metric report-kpi-card">
                  <div class="report-kpi-label">创建时间</div>
                  <div class="report-kpi-value">{{ formatDate(latestExport.createdAt) }}</div>
                  <div class="report-kpi-hint">最近一次任务创建时间</div>
                </article>
                <article class="journal-brief journal-metric report-kpi-card">
                  <div class="report-kpi-label">过期时间</div>
                  <div class="report-kpi-value">{{ latestExpiresText }}</div>
                  <div class="report-kpi-hint">文件生成完成后，后端会返回有效期截止时间</div>
                </article>
              </div>

              <div class="flex flex-wrap items-center gap-3">
                <button
                  type="button"
                  :disabled="downloading || latestExport.result.status !== 'ready'"
                  class="teacher-btn"
                  @click="handleDownload"
                >
                  {{
                    downloading
                      ? '下载中...'
                      : latestExport.result.status === 'ready'
                        ? '下载报告'
                        : polling
                          ? '正在等待完成...'
                          : '等待生成完成'
                  }}
                </button>
                <p class="text-sm leading-6 text-[var(--journal-muted)]">
                  {{ derivedDownloadHint }}
                </p>
              </div>
            </div>
          </section>
        </div>

        <aside class="report-aside">
          <section class="report-flat-section report-flat-section--aside">
            <div class="report-section-head">
              <div>
                <div class="teacher-surface-eyebrow journal-eyebrow">Guide</div>
                <h2 class="report-section-title">使用说明</h2>
              </div>
            </div>

            <div class="report-guide-stack">
              <article class="report-note">
                <div class="report-note-label">预览优先</div>
                <div class="report-note-helper">报告内容可以直接在页面内查看，不必先发起下载。</div>
              </article>
              <article class="report-note">
                <div class="report-note-label">任务异步</div>
                <div class="report-note-helper">
                  导出任务提交后由后端异步生成，页面会自动刷新任务状态。
                </div>
              </article>
              <article class="report-note">
                <div class="report-note-label">下载触发</div>
                <div class="report-note-helper">
                  只有状态变为“已就绪”后才可下载，失败时会保留错误信息便于重试。
                </div>
              </article>
            </div>
          </section>
        </aside>
      </div>
    </div>

    <ElDialog
      v-model="previewDialogVisible"
      width="min(1180px, calc(100vw - 32px))"
      top="4vh"
      destroy-on-close
      class="report-preview-dialog teacher-surface-dialog"
    >
      <template #header>
        <div class="report-dialog-header">
          <div>
            <div class="teacher-surface-eyebrow journal-eyebrow">Live Preview</div>
            <h3 class="mt-3 text-2xl font-semibold tracking-tight text-[var(--journal-ink)]">
              当前报告预览
            </h3>
            <p class="mt-2 text-sm leading-7 text-[var(--journal-muted)]">
              不下载也能直接查看当前班级的关键报告内容。
            </p>
          </div>
          <div class="teacher-surface-chip">
            预览班级：{{ previewClassName || normalizedClassNameText }}
          </div>
        </div>
      </template>

      <div v-if="previewError" class="teacher-surface-error report-preview-error">
        {{ previewError }}
      </div>

      <div v-if="previewLoading" class="grid gap-4 md:grid-cols-3">
        <div v-for="index in 3" :key="index" class="report-preview-skeleton" />
      </div>

      <template v-else-if="previewSummary">
        <section class="report-kpi-grid report-kpi-grid--dialog">
          <article class="journal-brief journal-metric report-kpi-card">
            <div class="report-kpi-label">班级人数</div>
            <div class="report-kpi-value">{{ previewSummary.student_count }}</div>
            <div class="report-kpi-hint">当前预览班级纳入统计的学生数</div>
          </article>
          <article class="journal-brief journal-metric report-kpi-card">
            <div class="report-kpi-label">平均解题</div>
            <div class="report-kpi-value">{{ averageSolvedText }}</div>
            <div class="report-kpi-hint">当前班级学生的人均解题数</div>
          </article>
          <article class="journal-brief journal-metric report-kpi-card">
            <div class="report-kpi-label">近 7 天活跃率</div>
            <div class="report-kpi-value">{{ activeRateText }}</div>
            <div class="report-kpi-hint">近 7 天至少有一次训练动作的学生占比</div>
          </article>
        </section>

        <div class="mt-5 space-y-5">
          <TeacherClassTrendPanel
            :trend="previewTrend"
            title="班级近 7 天训练趋势"
            subtitle="直接查看当前班级训练事件、成功解题和活跃学生走势。"
          />

          <TeacherClassReviewPanel :review="previewReview" :class-name="previewClassName" />

          <TeacherClassInsightsPanel :students="previewStudents" :class-name="previewClassName" />
        </div>
      </template>

      <AppEmpty
        class="mt-5"
        v-else
        title="还没有可用预览"
        description="先选择班级并加载一次预览，这里会展示当前报告内容。"
        icon="FileChartColumnIncreasing"
      />
    </ElDialog>
  </section>
</template>

<style scoped>
.teacher-management-shell {
  --report-shell-page: color-mix(in srgb, var(--color-bg-base) 94%, var(--color-bg-surface));
  --report-shell-bg: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --report-shell-brand: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
  --teacher-card-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --teacher-control-border: color-mix(in srgb, var(--journal-border) 78%, transparent);
  --teacher-divider: color-mix(in srgb, var(--journal-border) 86%, transparent);
  --report-card-border: var(--teacher-card-border);
  --report-control-border: var(--teacher-control-border);
  --report-divider: var(--teacher-divider);
}

.teacher-management-shell.report-shell {
  gap: 0;
  background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--report-shell-brand) 6%, transparent),
      transparent 26rem
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--report-shell-bg) 96%, var(--report-shell-page)),
      var(--report-shell-bg)
    );
}

.journal-brief,
.journal-metric {
  border-radius: 18px;
}

.report-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.report-topbar {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 1.5rem;
  padding-bottom: 1.5rem;
}

.report-copy {
  max-width: 52rem;
}

.report-summary-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.75rem;
}

.report-hero-divider {
  margin-top: 1.5rem;
  border-top: 1px dashed var(--report-divider);
}

.report-workspace {
  display: grid;
  gap: 1.5rem;
  grid-template-columns: minmax(0, 1.28fr) minmax(280px, 0.72fr);
  padding-top: 1.5rem;
}

.report-main,
.report-aside {
  display: grid;
  gap: 1.5rem;
}

.report-flat-section {
  display: grid;
  gap: 1rem;
}

.report-flat-section--aside {
  align-content: start;
  gap: 1rem;
  border-left: 1px solid var(--report-divider);
  padding-left: 1rem;
}

.report-section-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 0.85rem;
}

.report-section-title {
  margin-top: 0.35rem;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.report-section-copy {
  max-width: 30rem;
  font-size: 0.82rem;
  line-height: 1.6;
  color: var(--journal-muted);
}

.report-command-grid {
  margin-top: 1rem;
  display: grid;
  gap: 1rem;
  grid-template-columns: minmax(0, 1.08fr) minmax(260px, 0.92fr);
}

.report-side-notes {
  display: grid;
  align-content: start;
  gap: 0.85rem;
  border-left: 1px solid var(--report-divider);
  padding-left: 1rem;
}

.report-field {
  border-radius: 1rem;
  border: 1px solid var(--report-control-border);
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  color: var(--journal-ink);
}

.report-field::placeholder {
  color: color-mix(in srgb, var(--journal-muted) 74%, transparent);
}

.report-field:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 52%, transparent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 12%, transparent);
}

.report-format-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.report-format-option {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  border-radius: 1rem;
  border: 1px solid var(--report-control-border);
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  padding: 0.95rem 1rem;
  transition:
    border-color 0.18s ease,
    background 0.18s ease,
    box-shadow 0.18s ease;
}

.report-format-option--active {
  border-color: color-mix(in srgb, var(--journal-accent) 34%, var(--report-control-border));
  background: color-mix(in srgb, var(--journal-accent) 8%, var(--journal-surface));
  box-shadow: 0 10px 24px color-mix(in srgb, var(--journal-accent) 10%, transparent);
}

.report-note {
  border: 1px solid var(--report-card-border);
  border-radius: 18px;
  background: color-mix(in srgb, var(--journal-surface) 95%, var(--color-bg-base));
  padding: 0.95rem 1rem;
}

.report-note-label {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.report-note-value {
  margin-top: 0.45rem;
  font-size: 1.02rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.report-note-helper {
  margin-top: 0.45rem;
  font-size: 0.82rem;
  line-height: 1.6;
  color: var(--journal-muted);
}

.report-guide-list {
  margin-top: 0.6rem;
  display: grid;
  gap: 0.55rem;
  list-style: none;
  padding-left: 0;
  color: var(--journal-muted);
}

.report-guide-stack {
  margin-top: 1rem;
  display: grid;
  gap: 0.85rem;
}

.report-side-notes .report-note,
.report-guide-stack .report-note {
  border: 0;
  border-radius: 0;
  background: transparent;
  border-bottom: 1px solid var(--report-divider);
  padding: 0 0 0.85rem;
}

.report-side-notes .report-note:last-child,
.report-guide-stack .report-note:last-child {
  border-bottom: 0;
  padding-bottom: 0;
}

.report-kpi-grid {
  display: grid;
  gap: 0.85rem;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.report-kpi-grid--task,
.report-kpi-grid--dialog {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.report-kpi-grid--dialog {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.report-kpi-card {
  border: 1px solid var(--report-card-border);
  background: color-mix(in srgb, var(--journal-surface-subtle) 88%, var(--color-bg-base));
  padding: 0.95rem 1rem;
  box-shadow: 0 10px 24px var(--color-shadow-soft);
}

.report-kpi-label {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.report-kpi-value {
  margin-top: 0.42rem;
  font-size: 1.08rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.report-kpi-hint {
  margin-top: 0.45rem;
  font-size: 0.8rem;
  line-height: 1.55;
  color: var(--journal-muted);
}

.report-status-banner {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  border: 1px solid var(--report-card-border);
  border-radius: 18px;
  background: color-mix(in srgb, var(--journal-surface) 95%, var(--color-bg-base));
  padding: 1rem;
}

.report-status-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid transparent;
  padding: 0.35rem 0.8rem;
  font-size: 0.78rem;
  font-weight: 700;
}

.report-status-chip--idle {
  border-color: color-mix(in srgb, var(--journal-border) 80%, transparent);
  background: color-mix(in srgb, var(--journal-border) 24%, transparent);
  color: var(--journal-muted);
}

.report-status-chip--ready {
  border-color: color-mix(in srgb, var(--color-success) 30%, transparent);
  background: color-mix(in srgb, var(--color-success) 12%, transparent);
  color: var(--color-success);
}

.report-status-chip--failed {
  border-color: color-mix(in srgb, var(--color-danger) 30%, transparent);
  background: color-mix(in srgb, var(--color-danger) 12%, transparent);
  color: var(--color-danger);
}

.report-status-chip--pending {
  border-color: color-mix(in srgb, var(--color-warning) 30%, transparent);
  background: color-mix(in srgb, var(--color-warning) 12%, transparent);
  color: var(--color-warning);
}

.teacher-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  min-height: 2.6rem;
  border-radius: 999px;
  border: 1px solid var(--report-control-border);
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  padding: 0.58rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--journal-ink);
  transition:
    border-color 0.18s ease,
    background 0.18s ease,
    color 0.18s ease;
}

.teacher-btn:hover:not(:disabled) {
  border-color: color-mix(in srgb, var(--journal-accent) 36%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, var(--journal-surface));
  color: var(--journal-accent-strong);
}

.teacher-btn--primary {
  border-color: color-mix(in srgb, var(--journal-accent) 26%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 12%, var(--journal-surface));
  color: color-mix(in srgb, var(--journal-accent) 88%, var(--journal-ink));
}

.teacher-btn:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.report-dialog-header {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding-right: 2rem;
}

.report-preview-error {
  margin-bottom: 1rem;
}

.report-preview-skeleton {
  height: 6rem;
  border-radius: 1rem;
  border: 1px solid var(--report-card-border);
  background: linear-gradient(
    90deg,
    color-mix(in srgb, var(--journal-surface-subtle) 88%, var(--color-bg-base)),
    color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base))
  );
  animation: reportPulse 1.2s ease-in-out infinite;
}

@keyframes reportPulse {
  0%,
  100% {
    opacity: 0.72;
  }

  50% {
    opacity: 1;
  }
}

:deep(.page-header) {
  border: 1px solid var(--report-card-border);
  border-radius: 18px;
  background: color-mix(in srgb, var(--journal-surface-subtle) 90%, var(--color-bg-base));
}

:deep(.report-preview-dialog .el-dialog) {
  border: 1px solid var(--report-card-border);
  border-radius: 20px;
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base)),
    color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base))
  );
  box-shadow: 0 24px 60px var(--color-shadow-soft);
}

:deep(.report-preview-dialog .el-dialog__header) {
  margin-right: 0;
  padding: 1.25rem 1.25rem 0;
}

:deep(.report-preview-dialog .el-dialog__body) {
  padding: 1rem 1.25rem 1.25rem;
}

@media (max-width: 1180px) {
  .report-workspace {
    grid-template-columns: minmax(0, 1fr);
  }

  .report-summary-grid,
  .report-command-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .report-flat-section--aside,
  .report-side-notes {
    border-left: 0;
    border-top: 1px solid var(--report-divider);
    padding-left: 0;
    padding-top: 1rem;
  }

  .report-kpi-grid,
  .report-kpi-grid--task {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 720px) {
  .report-summary-grid,
  .report-kpi-grid,
  .report-kpi-grid--task,
  .report-kpi-grid--dialog,
  .report-format-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .report-shell {
    padding-inline: 1rem;
  }

  .report-dialog-header {
    padding-right: 0;
  }
}
</style>
