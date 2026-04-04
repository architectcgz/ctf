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
      metricClass: 'teacher-metric-card--soft',
    }
  }

  switch (latestExport.value.result.status) {
    case 'ready':
      return {
        label: '已就绪',
        chipClass: 'report-status-chip--ready',
        metricClass: 'teacher-metric-card--calm',
      }
    case 'failed':
      return {
        label: '失败',
        chipClass: 'report-status-chip--failed',
        metricClass: 'teacher-metric-card--danger',
      }
    default:
      return {
        label: polling.value ? '生成中' : '等待更新',
        chipClass: 'report-status-chip--pending',
        metricClass: 'teacher-metric-card--accent',
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
  <div class="report-export-shell teacher-management-shell teacher-surface space-y-6">
    <section class="report-hero teacher-hero teacher-surface-hero px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="teacher-surface-eyebrow">Teacher Export</div>
          <h2
            class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
          >
            报告导出
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            先在页面内查看班级报告预览，再决定是否创建导出任务下载 PDF 或 Excel 文件。
          </p>
        </div>

        <article class="report-brief report-panel report-panel--brief teacher-brief teacher-surface-brief px-5 py-5">
          <div class="text-sm font-medium text-[var(--journal-ink)]">当前导出概况</div>
          <div class="teacher-metric-grid mt-5 grid gap-3 sm:grid-cols-2">
            <article
              class="teacher-surface-metric teacher-metric-card teacher-metric-card--accent px-4 py-4"
            >
              <div class="teacher-metric-label">当前账号</div>
              <div class="teacher-metric-value">{{ authStore.user?.username || '-' }}</div>
              <div class="teacher-metric-hint">用于发起当前导出任务的教师账号</div>
            </article>
            <article
              class="teacher-surface-metric teacher-metric-card teacher-metric-card--calm px-4 py-4"
            >
              <div class="teacher-metric-label">默认班级</div>
              <div class="teacher-metric-value">{{ authStore.user?.class_name || '未绑定' }}</div>
              <div class="teacher-metric-hint">留空时将优先使用当前账号绑定班级</div>
            </article>
            <article
              class="teacher-surface-metric teacher-metric-card teacher-metric-card--soft px-4 py-4"
            >
              <div class="teacher-metric-label">当前格式</div>
              <div class="teacher-metric-value">{{ selectedFormatLabel }}</div>
              <div class="teacher-metric-hint">{{ selectedFormatHint }}</div>
            </article>
            <article
              class="teacher-surface-metric teacher-metric-card px-4 py-4"
              :class="latestStatusMeta.metricClass"
            >
              <div class="teacher-metric-label">最近状态</div>
              <div class="teacher-metric-value">{{ latestStatusMeta.label }}</div>
              <div class="teacher-metric-hint">
                {{ latestExport ? derivedDownloadHint : '创建一次导出任务后，这里会同步展示最新状态。' }}
              </div>
            </article>
          </div>
        </article>
      </div>

      <div class="teacher-surface-board mt-6">
        <section class="report-panel report-panel--section teacher-surface-section teacher-surface-filter">
          <div>
            <div class="teacher-surface-eyebrow">Export Task</div>
            <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">创建导出任务</h3>
            <p class="mt-2 max-w-3xl text-sm leading-7 text-[var(--journal-muted)]">
              先确认班级和导出格式，再将任务交给后端处理。预览与下载彼此独立，不会影响当前页面浏览。
            </p>
          </div>

          <div class="mt-5 grid gap-4 xl:grid-cols-[minmax(0,1.08fr)_minmax(280px,0.92fr)]">
            <div class="space-y-4">
              <label class="space-y-2">
                <span class="text-sm text-text-secondary">班级名称</span>
                <input
                  v-model="form.className"
                  type="text"
                  :placeholder="classNamePlaceholder"
                  class="report-field w-full px-4 py-3 text-sm outline-none transition"
                />
              </label>

              <fieldset class="space-y-2">
                <legend class="text-sm text-text-secondary">导出格式</legend>
                <div class="grid gap-3 sm:grid-cols-2">
                  <label
                    class="report-format-option"
                    :class="{ 'report-format-option--active': form.format === 'pdf' }"
                  >
                    <input v-model="form.format" type="radio" value="pdf" class="mt-1" />
                    <span>
                      <span class="block font-medium text-text-primary">PDF</span>
                      <span class="mt-1 block text-sm text-text-secondary">
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
                      <span class="block font-medium text-text-primary">Excel</span>
                      <span class="mt-1 block text-sm text-text-secondary">
                        适合继续分析、筛选和二次加工。
                      </span>
                    </span>
                  </label>
                </div>
              </fieldset>

              <div class="report-inline-card">
                <div class="report-inline-label">填写提示</div>
                <div class="report-inline-hint">
                  如果当前账号已绑定班级，可直接留空使用默认班级；管理员也可手动输入其他班级名称。
                </div>
              </div>

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

            <aside class="report-aside space-y-3">
              <article class="report-inline-card">
                <div class="report-inline-label">当前目标班级</div>
                <div class="report-inline-value">{{ normalizedClassNameText }}</div>
                <div class="report-inline-hint">导出与预览都将以这个班级名称作为优先目标。</div>
              </article>
              <article class="report-inline-card">
                <div class="report-inline-label">当前导出格式</div>
                <div class="report-inline-value">{{ selectedFormatLabel }}</div>
                <div class="report-inline-hint">{{ selectedFormatHint }}</div>
              </article>
              <article class="report-inline-card">
                <div class="report-inline-label">建议流程</div>
                <ol class="report-guide-list">
                  <li>1. 先打开报告预览，确认当前班级数据内容。</li>
                  <li>2. 需要留档时，再创建后台导出任务。</li>
                  <li>3. 任务完成后下载文件，生成中会自动轮询状态。</li>
                </ol>
              </article>
            </aside>
          </div>
        </section>

        <section class="report-panel report-panel--section teacher-surface-section">
          <div>
            <div class="teacher-surface-eyebrow">Latest Task</div>
            <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">最近一次任务</h3>
            <p class="mt-2 max-w-3xl text-sm leading-7 text-[var(--journal-muted)]">
              导出状态、下载入口和任务元数据统一收在这里。
            </p>
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
                <div class="report-inline-label">任务编号</div>
                <div class="report-inline-value">#{{ latestExport.result.report_id }}</div>
                <div class="report-inline-hint">{{ derivedDownloadHint }}</div>
              </div>
              <span class="report-status-chip" :class="latestStatusMeta.chipClass">
                {{ latestStatusMeta.label }}
              </span>
            </div>

            <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
              <article
                class="teacher-surface-metric teacher-metric-card teacher-metric-card--accent px-4 py-4"
              >
                <div class="teacher-metric-label">班级</div>
                <div class="teacher-metric-value">{{ latestExport.className }}</div>
                <div class="teacher-metric-hint">本次导出的目标班级</div>
              </article>
              <article
                class="teacher-surface-metric teacher-metric-card teacher-metric-card--soft px-4 py-4"
              >
                <div class="teacher-metric-label">格式</div>
                <div class="teacher-metric-value">{{ latestExport.format.toUpperCase() }}</div>
                <div class="teacher-metric-hint">本次任务选择的导出文件格式</div>
              </article>
              <article
                class="teacher-surface-metric teacher-metric-card teacher-metric-card--calm px-4 py-4"
              >
                <div class="teacher-metric-label">创建时间</div>
                <div class="teacher-metric-value">{{ formatDate(latestExport.createdAt) }}</div>
                <div class="teacher-metric-hint">最近一次任务创建时间</div>
              </article>
              <article
                class="teacher-surface-metric teacher-metric-card px-4 py-4"
                :class="latestStatusMeta.metricClass"
              >
                <div class="teacher-metric-label">过期时间</div>
                <div class="teacher-metric-value">{{ latestExpiresText }}</div>
                <div class="teacher-metric-hint">文件生成完成后，后端会返回有效期截止时间</div>
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

        <section class="report-panel report-panel--section teacher-surface-section">
          <div>
            <div class="teacher-surface-eyebrow">Guide</div>
            <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">使用说明</h3>
            <p class="mt-2 max-w-3xl text-sm leading-7 text-[var(--journal-muted)]">
              当前链路下推荐的导出节奏，以及后端生成完成前后的页面行为。
            </p>
          </div>

          <div class="mt-5 grid gap-3 md:grid-cols-3">
            <article class="report-guide-card">
              <div class="report-inline-label">预览优先</div>
              <div class="report-inline-hint">报告内容可以直接在页面内查看，不必先发起下载。</div>
            </article>
            <article class="report-guide-card">
              <div class="report-inline-label">任务异步</div>
              <div class="report-inline-hint">
                导出任务提交后由后端异步生成，页面会在生成中自动更新状态。
              </div>
            </article>
            <article class="report-guide-card">
              <div class="report-inline-label">下载触发</div>
              <div class="report-inline-hint">
                只有状态变为“已就绪”后才可下载，失败时会保留错误信息便于重试。
              </div>
            </article>
          </div>
        </section>
      </div>
    </section>
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
            <div class="teacher-surface-eyebrow">Live Preview</div>
            <h3 class="mt-3 text-2xl font-semibold tracking-tight text-[var(--journal-ink)]">
              当前报告预览
            </h3>
            <p class="mt-2 text-sm leading-7 text-[var(--journal-muted)]">
              不下载也能直接查看当前班级的关键报告内容。
            </p>
          </div>
          <div class="teacher-surface-chip">预览班级：{{ previewClassName || normalizedClassNameText }}</div>
        </div>
      </template>

      <div v-if="previewError" class="teacher-surface-error report-preview-error">
        {{ previewError }}
      </div>

      <div v-if="previewLoading" class="grid gap-4 md:grid-cols-3">
        <div
          v-for="index in 3"
          :key="index"
          class="report-preview-skeleton"
        />
      </div>

      <template v-else-if="previewSummary">
        <section class="teacher-metric-grid grid gap-3 md:grid-cols-3">
          <article
            class="teacher-surface-metric teacher-metric-card teacher-metric-card--calm px-4 py-4"
          >
            <div class="teacher-metric-label">班级人数</div>
            <div class="teacher-metric-value">{{ previewSummary.student_count }}</div>
            <div class="teacher-metric-hint">当前预览班级纳入统计的学生数</div>
          </article>
          <article
            class="teacher-surface-metric teacher-metric-card teacher-metric-card--soft px-4 py-4"
          >
            <div class="teacher-metric-label">平均解题</div>
            <div class="teacher-metric-value">{{ averageSolvedText }}</div>
            <div class="teacher-metric-hint">当前班级学生的人均解题数</div>
          </article>
          <article
            class="teacher-surface-metric teacher-metric-card teacher-metric-card--accent px-4 py-4"
          >
            <div class="teacher-metric-label">近 7 天活跃率</div>
            <div class="teacher-metric-value">{{ activeRateText }}</div>
            <div class="teacher-metric-hint">近 7 天至少有一次训练动作的学生占比</div>
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
  </div>
</template>

<style scoped>
.teacher-management-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: #4f46e5;
  --journal-accent-strong: #4338ca;
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  --color-primary: #4f46e5;
  --color-primary-hover: #4338ca;
  --color-primary-soft: rgba(79, 70, 229, 0.08);
  padding: 0.25rem 0 2rem;
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
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
  border-color: var(--journal-border);
}

.report-brief {
  border-color: var(--journal-border);
}

.report-panel {
  border: 1px solid var(--journal-border);
  border-radius: 16px;
  box-shadow: 0 8px 18px var(--color-shadow-soft);
}

.report-panel--brief {
  background: var(--journal-surface-subtle);
  overflow: hidden;
}

.report-panel--section {
  background: var(--journal-surface-subtle);
}

.report-brief-title {
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.report-note {
  border-radius: 16px;
  border: 1px solid var(--journal-border);
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

.report-field {
  border-radius: 0.9rem;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  color: var(--journal-ink);
}

.report-field::placeholder {
  color: color-mix(in srgb, var(--journal-muted) 74%, transparent);
}

.report-field:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 60%, transparent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 14%, transparent);
}

.report-format-option {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  border-radius: 1rem;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.95rem 1rem;
  transition:
    border-color 0.18s ease,
    background 0.18s ease,
    box-shadow 0.18s ease;
}

.report-format-option--active {
  border-color: color-mix(in srgb, var(--journal-accent) 42%, var(--journal-border));
  background: color-mix(in srgb, var(--journal-accent) 8%, var(--journal-surface));
  box-shadow: 0 10px 24px color-mix(in srgb, var(--journal-accent) 12%, transparent);
}

.report-inline-card,
.report-guide-card,
.report-status-banner {
  border: 1px solid var(--journal-border);
  border-radius: 1rem;
  background: var(--journal-surface);
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.05);
}

.report-inline-card,
.report-guide-card {
  padding: 1rem 1.05rem;
}

.report-inline-label {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.report-inline-value {
  margin-top: 0.5rem;
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.report-inline-hint {
  margin-top: 0.45rem;
  font-size: 0.82rem;
  line-height: 1.6;
  color: var(--journal-muted);
}

.report-guide-list {
  margin-top: 0.6rem;
  display: grid;
  gap: 0.55rem;
  padding-left: 0;
  list-style: none;
  color: var(--journal-muted);
}

.report-status-banner {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 1rem 1.05rem;
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

.report-preview-error {
  margin-bottom: 1rem;
}

.report-preview-skeleton {
  height: 6rem;
  border-radius: 1rem;
  border: 1px solid var(--journal-border);
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

.report-hero-divider {
  margin-top: 1.5rem;
  margin-bottom: 1.5rem;
  border-top: 1px dashed rgba(148, 163, 184, 0.58);
}

.report-section-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.report-guide {
  border-top: 1px dashed rgba(148, 163, 184, 0.72);
  padding-top: 1.25rem;
}

:deep(.report-preview-dialog .el-dialog) {
  border-radius: 24px;
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.08), transparent 20rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 96%, var(--color-bg-base)),
      color-mix(
        in srgb,
        var(--journal-surface-subtle, var(--color-bg-elevated)) 94%,
        var(--color-bg-base)
      )
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

.report-card {
  border: 1px solid var(--journal-border);
  border-radius: 16px;
  background: var(--journal-surface);
  box-shadow: 0 8px 18px var(--color-shadow-soft);
}

.report-card--hero {
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.06), transparent 18rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 94%, var(--color-bg-base))
    );
}

.report-card--action {
  background: var(--journal-surface-subtle);
}

.report-card--metric {
  background: var(--journal-surface-subtle);
  box-shadow: none;
}

.teacher-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  min-height: 2.5rem;
  border-radius: 0.9rem;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.55rem 1.1rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--journal-ink);
  cursor: pointer;
  transition:
    border-color 0.18s ease,
    background 0.18s ease,
    color 0.18s ease;
}

.teacher-btn:hover {
  border-color: var(--journal-accent);
  background: color-mix(in srgb, var(--journal-accent) 8%, var(--journal-surface));
  color: var(--journal-accent-strong);
}

.teacher-btn--primary {
  border-color: transparent;
  background: var(--journal-accent);
  color: #fff;
  box-shadow: 0 12px 24px var(--color-shadow-soft);
}

.teacher-btn--primary:hover {
  background: var(--journal-accent-strong);
  border-color: transparent;
  color: #fff;
}

.journal-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  border-radius: 0.9rem;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.55rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--journal-ink);
  transition:
    border-color 0.2s,
    color 0.2s,
    background 0.2s;
  cursor: pointer;
}

.journal-btn:hover:not(:disabled) {
  border-color: var(--journal-accent);
  background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
  color: var(--journal-accent);
}

.journal-btn--primary {
  border-color: color-mix(in srgb, var(--journal-accent) 50%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  color: var(--journal-accent);
}

.journal-btn--primary:hover:not(:disabled) {
  background: color-mix(in srgb, var(--journal-accent) 14%, transparent);
}

.journal-btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.report-kpi-grid {
  align-items: stretch;
}

.report-kpi-card {
  border: 1px solid var(--journal-border);
  border-radius: 16px;
  background: var(--journal-surface-subtle);
  padding: 0.95rem 1rem;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.035);
}

.report-kpi-card--primary {
  border-top: 3px solid rgba(79, 70, 229, 0.42);
}

.report-kpi-card--success {
  border-top: 3px solid rgba(16, 185, 129, 0.36);
}

.report-kpi-card--warning {
  border-top: 3px solid rgba(245, 158, 11, 0.38);
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


:global([data-theme='dark']) .teacher-management-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
}

:global([data-theme='dark']) .teacher-hero,
:global([data-theme='dark']) :deep(.report-preview-dialog .el-dialog) {
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.18), transparent 20rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.95), rgba(2, 6, 23, 0.98));
}

:global([data-theme='dark']) .report-brief,
:global([data-theme='dark']) .report-panel--brief,
:global([data-theme='dark']) .report-panel--section,
:global([data-theme='dark']) .report-field,
:global([data-theme='dark']) .report-format-option,
:global([data-theme='dark']) .report-inline-card,
:global([data-theme='dark']) .report-guide-card,
:global([data-theme='dark']) .report-status-banner {
  background: rgba(15, 23, 42, 0.42);
}

.report-kpi-hint {
  margin-top: 0.45rem;
  font-size: 0.8rem;
  line-height: 1.55;
  color: var(--journal-muted);
}
</style>
