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
import AppCard from '@/components/common/AppCard.vue'
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
    class="report-shell teacher-surface report-hero teacher-surface-hero flex min-h-full flex-col space-y-6 rounded-[30px] border px-6 py-6 md:px-8"
  >
    <div>
      <div class="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
        <div>
          <div class="journal-eyebrow report-eyebrow">Teacher Export</div>
          <h2
            class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
          >
            报告导出
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            先查看当前班级报告预览，再决定是否创建导出任务下载 PDF 或 Excel 文件。
          </p>
        </div>

        <article class="report-brief teacher-surface-brief journal-brief rounded-[24px] border px-5 py-5">
          <div class="report-brief-title">当前导出概况</div>
          <div class="mt-5 space-y-3">
            <div class="report-note">
              <div class="report-note-label">当前账号</div>
              <div class="report-note-value">{{ authStore.user?.username || '-' }}</div>
              <div class="report-note-helper">用于发起当前导出任务的账号</div>
            </div>
            <div class="report-note">
              <div class="report-note-label">默认班级</div>
              <div class="report-note-value">{{ authStore.user?.class_name || '未绑定' }}</div>
              <div class="report-note-helper">留空时将优先使用当前账号绑定班级</div>
            </div>
          </div>
        </article>
      </div>

      <div class="report-hero-divider" />

      <div class="grid gap-6 xl:grid-cols-[1.25fr_0.9fr]">
        <div>
          <div class="report-section-head">
            <div>
              <div class="journal-eyebrow report-eyebrow report-eyebrow--soft">Export Task</div>
              <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">创建导出任务</h3>
              <p class="mt-2 max-w-3xl text-sm leading-7 text-[var(--journal-muted)]">
                预览确认无误后，再选择是否下载为 PDF 或 Excel 文件。
              </p>
            </div>
          </div>

          <AppCard
            class="report-card journal-brief mt-5"
            variant="hero"
            accent="primary"
            eyebrow="Export Task"
            title="班级报告生成器"
            subtitle="先确定班级和导出格式，再把任务交给后端。下载变为可选动作，不影响当前页面预览。"
          >
            <label class="block">
              <span class="mb-2 block text-sm font-medium text-[var(--color-text-primary)]"
                >班级名称</span
              >
              <input
                v-model="form.className"
                type="text"
                :placeholder="classNamePlaceholder"
                class="teacher-surface-filter w-full border border-l-2 border-[var(--color-border-default)] bg-transparent px-3 py-2.5 text-[var(--color-text-primary)] outline-none transition focus:border-[var(--color-primary)] focus:ring-0"
              />
            </label>

            <fieldset>
              <legend class="mb-2 text-sm font-medium text-[var(--color-text-primary)]">
                导出格式
              </legend>
              <div class="grid gap-3 sm:grid-cols-2">
                <label
                  class="teacher-surface-filter flex items-start gap-3 border border-l-2 px-3 py-3 transition"
                  :class="
                    form.format === 'pdf'
                      ? 'border-[var(--color-primary)] bg-[var(--color-primary)]/6'
                      : 'border-[var(--color-border-default)] bg-transparent hover:border-[var(--color-primary)]/50'
                  "
                >
                  <input v-model="form.format" type="radio" value="pdf" class="mt-1" />
                  <span>
                    <span class="block font-medium text-[var(--color-text-primary)]">PDF</span>
                    <span class="mt-1 block text-sm text-[var(--color-text-secondary)]"
                      >适合打印、归档和正式汇报。</span
                    >
                  </span>
                </label>

                <label
                  class="teacher-surface-filter flex items-start gap-3 border border-l-2 px-3 py-3 transition"
                  :class="
                    form.format === 'excel'
                      ? 'border-[var(--color-primary)] bg-[var(--color-primary)]/6'
                      : 'border-[var(--color-border-default)] bg-transparent hover:border-[var(--color-primary)]/50'
                  "
                >
                  <input v-model="form.format" type="radio" value="excel" class="mt-1" />
                  <span>
                    <span class="block font-medium text-[var(--color-text-primary)]">Excel</span>
                    <span class="mt-1 block text-sm text-[var(--color-text-secondary)]"
                      >适合继续分析、筛选和二次加工。</span
                    >
                  </span>
                </label>
              </div>
            </fieldset>

            <AppCard variant="action" accent="neutral" class="report-card journal-brief">
              如果当前账号已绑定班级，可直接留空使用默认班级；管理员也可手动输入其他班级名称。
            </AppCard>

            <div class="flex flex-wrap gap-3">
              <button
                type="button"
                :disabled="previewLoading"
                class="teacher-surface-btn"
                @click="openPreviewDialog"
              >
                {{ previewLoading ? '加载预览中...' : '打开报告预览' }}
              </button>

              <button
                type="button"
                :disabled="submitting"
                class="teacher-surface-btn teacher-surface-btn--primary"
                @click="handleExport"
              >
                {{ submitting ? '提交中...' : '创建导出任务' }}
              </button>
            </div>
          </AppCard>
        </div>

        <div class="space-y-6">
          <div class="report-section-head">
            <div>
              <div class="journal-eyebrow report-eyebrow report-eyebrow--soft">Latest Task</div>
              <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">最近一次任务</h3>
              <p class="mt-2 max-w-3xl text-sm leading-7 text-[var(--journal-muted)]">
                导出状态、下载信息和任务元数据都在这里收口。
              </p>
            </div>
          </div>

          <AppEmpty
            v-if="!latestExport"
            class="teacher-surface-empty mt-5"
            title="还没有创建导出任务"
            description="先在左侧创建一次班级报告任务，这里会展示最近一次任务状态。"
            icon="FileChartColumnIncreasing"
          />

          <AppCard
            v-else
            class="report-card journal-brief"
            variant="hero"
            :accent="
              latestExport.result.status === 'ready'
                ? 'success'
                : latestExport.result.status === 'failed'
                  ? 'danger'
                  : 'warning'
            "
            eyebrow="Latest Task"
            :title="String(latestExport.result.report_id)"
            :subtitle="derivedDownloadHint"
          >
            <template #header>
              <span
                class="rounded-full px-3 py-1 text-xs font-semibold"
                :class="
                  latestExport.result.status === 'ready'
                    ? 'bg-[var(--color-success)]/12 text-[var(--color-success)]'
                    : latestExport.result.status === 'failed'
                      ? 'bg-[var(--color-danger)]/12 text-[var(--color-danger)]'
                      : 'bg-[var(--color-warning)]/12 text-[var(--color-warning)]'
                "
              >
                {{
                  latestExport.result.status === 'ready'
                    ? '已就绪'
                    : latestExport.result.status === 'failed'
                      ? '失败'
                      : '生成中'
                }}
              </span>
            </template>

            <div class="grid grid-cols-2 gap-3 text-sm">
              <AppCard
                class="report-card journal-metric"
                variant="metric"
                accent="primary"
                eyebrow="班级"
                :title="latestExport.className"
              />
              <AppCard
                class="report-card journal-metric"
                variant="metric"
                accent="primary"
                eyebrow="格式"
                :title="latestExport.format.toUpperCase()"
              />
              <AppCard
                class="report-card journal-metric"
                variant="metric"
                accent="neutral"
                eyebrow="创建时间"
                :title="formatDate(latestExport.createdAt)"
              />
              <AppCard
                class="report-card journal-metric"
                variant="metric"
                accent="neutral"
                eyebrow="过期时间"
                :title="
                  latestExport.result.expires_at
                    ? formatDate(latestExport.result.expires_at)
                    : '待生成完成后返回'
                "
              />
            </div>

            <button
              type="button"
              :disabled="downloading || latestExport.result.status !== 'ready'"
              class="teacher-surface-btn"
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
          </AppCard>

          <div class="report-section-head">
            <div>
              <div class="journal-eyebrow report-eyebrow report-eyebrow--soft">Guide</div>
              <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">使用说明</h3>
              <p class="mt-2 max-w-3xl text-sm leading-7 text-[var(--journal-muted)]">
                导出链路和当前后端能力边界。
              </p>
            </div>
          </div>

          <div class="report-guide teacher-surface-empty mt-5">
            <ol class="space-y-3 text-sm leading-6 text-[var(--color-text-secondary)]">
              <li>1. 先点击“查看当前预览”，直接在页面内查看当前班级报告内容。</li>
              <li>2. 确认需要留档时，再创建后台导出任务。</li>
              <li>3. 若状态为“已就绪”，可直接下载；若为“生成中”，页面会自动轮询状态。</li>
            </ol>
          </div>
        </div>
      </div>
    </div>

    <ElDialog
      v-model="previewDialogVisible"
      width="min(1180px, calc(100vw - 32px))"
      top="4vh"
      destroy-on-close
      class="report-preview-dialog"
    >
      <template #header>
        <div class="report-dialog-header">
          <div>
            <div class="journal-eyebrow report-eyebrow">Live Preview</div>
            <h3 class="mt-3 text-2xl font-semibold tracking-tight text-[var(--journal-ink)]">
              当前报告预览
            </h3>
            <p class="mt-2 text-sm leading-7 text-[var(--journal-muted)]">
              不下载也能直接查看当前班级的关键报告内容。
            </p>
          </div>
          <div class="report-dialog-chip">
            预览班级：{{ previewClassName || normalizeClassName() || '未选择' }}
          </div>
        </div>
      </template>

      <div
        v-if="previewError"
        class="border-b border-l-2 border-[var(--color-warning)] bg-[var(--color-warning)]/8 px-4 py-3 text-sm text-[var(--color-warning)]"
      >
        {{ previewError }}
      </div>

      <div v-if="previewLoading" class="grid gap-4 md:grid-cols-3">
        <div
          v-for="index in 3"
          :key="index"
          class="h-24 animate-pulse border-b border-[var(--color-border-default)] bg-[var(--color-bg-surface)]"
        />
      </div>

      <template v-else-if="previewSummary">
        <section class="report-kpi-grid grid gap-3 md:grid-cols-3">
          <article class="report-kpi-card journal-metric">
            <div class="report-kpi-label">班级人数</div>
            <div class="report-kpi-value">{{ previewSummary.student_count }}</div>
            <div class="report-kpi-hint">当前预览班级纳入统计的学生数</div>
          </article>
          <article class="report-kpi-card journal-metric">
            <div class="report-kpi-label">平均解题</div>
            <div class="report-kpi-value">{{ averageSolvedText }}</div>
            <div class="report-kpi-hint">当前班级学生的人均解题数</div>
          </article>
          <article class="report-kpi-card journal-metric">
            <div class="report-kpi-label">近 7 天活跃率</div>
            <div class="report-kpi-value">{{ activeRateText }}</div>
            <div class="report-kpi-hint">近 7 天至少有一次训练动作的学生占比</div>
          </article>
        </section>

        <TeacherClassTrendPanel
          :trend="previewTrend"
          title="班级近 7 天训练趋势"
          subtitle="直接查看当前班级训练事件、成功解题和活跃学生走势。"
        />

        <TeacherClassReviewPanel :review="previewReview" :class-name="previewClassName" />

        <TeacherClassInsightsPanel :students="previewStudents" :class-name="previewClassName" />
      </template>

      <AppEmpty
        v-else
        class="teacher-surface-empty"
        title="还没有可用预览"
        description="先选择班级并加载一次预览，这里会展示当前报告内容。"
        icon="FileChartColumnIncreasing"
      />
    </ElDialog>
  </section>
</template>

<style scoped>
.report-shell {
  --report-card-border: color-mix(in srgb, var(--journal-border, var(--color-border-default)) 74%, transparent);
  --report-divider: color-mix(in srgb, var(--journal-border, var(--color-border-default)) 56%, transparent);
}

:deep(.page-header) {
  border: 1px solid var(--report-card-border);
  border-radius: 16px;
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.08), transparent 18rem),
    linear-gradient(180deg, color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 96%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle, var(--color-bg-elevated)) 94%, var(--color-bg-base)));
  box-shadow: 0 18px 40px var(--color-shadow-soft);
}

:deep(.page-header__eyebrow) {
  border: 1px solid rgba(99, 102, 241, 0.18);
  border-left: 1px solid rgba(99, 102, 241, 0.18) !important;
  border-radius: 999px;
  background: rgba(99, 102, 241, 0.06);
  padding: 0.2rem 0.72rem;
  padding-left: 0.72rem !important;
  letter-spacing: 0.2em;
  color: var(--journal-accent);
}

.report-eyebrow {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.report-eyebrow--soft {
  opacity: 0.88;
}

.report-hero {
  border-color: var(--report-card-border);
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.08), transparent 18rem),
    linear-gradient(180deg, color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 96%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle, var(--color-bg-elevated)) 94%, var(--color-bg-base)));
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 18px 40px var(--color-shadow-soft);
}

.report-brief {
  border-color: var(--report-card-border);
  background: var(--journal-surface-subtle);
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 8px 18px var(--color-shadow-soft);
}

.report-brief-title {
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.report-note {
  border-radius: 16px;
  border: 1px solid var(--report-card-border);
  background: var(--journal-surface);
  padding: 0.85rem 0.95rem;
}

.report-note-label {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.report-note-value {
  margin-top: 0.45rem;
  font-size: 1rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.report-note-helper {
  margin-top: 0.45rem;
  font-size: 0.8rem;
  line-height: 1.55;
  color: var(--journal-muted);
}

.report-hero-divider {
  margin-top: 1.5rem;
  margin-bottom: 1.5rem;
  border-top: 1px dashed var(--report-divider);
}

.report-section-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.report-guide {
  padding: 1.25rem 1.1rem 1.1rem;
}

:deep(.report-preview-dialog .el-dialog) {
  border: 1px solid var(--report-card-border);
  border-radius: 24px;
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.08), transparent 20rem),
    linear-gradient(180deg, color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 96%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle, var(--color-bg-elevated)) 94%, var(--color-bg-base)));
  box-shadow: 0 24px 60px rgba(15, 23, 42, 0.16);
}

:deep(.report-preview-dialog .el-dialog__header) {
  margin-right: 0;
  padding: 1.25rem 1.25rem 0;
}

:deep(.report-preview-dialog .el-dialog__body) {
  padding: 1rem 1.25rem 1.25rem;
}

.report-dialog-header {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding-right: 2rem;
}

.report-dialog-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid rgba(99, 102, 241, 0.16);
  background: rgba(99, 102, 241, 0.06);
  padding: 0.3rem 0.75rem;
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--journal-accent-strong);
}

.report-kpi-grid {
  align-items: stretch;
}

.report-kpi-card {
  border: 1px solid var(--report-card-border);
  border-radius: 16px;
  background: var(--journal-surface);
  padding: 0.95rem 1rem;
  box-shadow: 0 10px 24px var(--color-shadow-soft);
}

.report-kpi-label {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.report-kpi-value {
  margin-top: 0.45rem;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.report-kpi-hint {
  margin-top: 0.45rem;
  font-size: 0.8rem;
  line-height: 1.55;
  color: var(--journal-muted);
}
</style>
