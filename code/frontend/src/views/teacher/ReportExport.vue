<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

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
import MetricCard from '@/components/common/MetricCard.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionCard from '@/components/common/SectionCard.vue'
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

onMounted(() => {
  if (normalizeClassName()) {
    void loadPreview()
  }
})
</script>

<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Teacher"
      title="报告导出"
      description="先查看当前班级报告预览，再决定是否创建导出任务下载 PDF 或 Excel 文件。"
    >
      <AppCard
        variant="action"
        accent="neutral"
      >
        当前账号：<span class="font-medium text-[var(--color-text-primary)]">{{
          authStore.user?.username || '-'
        }}</span>
      </AppCard>
    </PageHeader>

    <section class="grid gap-6 xl:grid-cols-[1.25fr_0.9fr]">
      <SectionCard
        title="创建导出任务"
        subtitle="预览确认无误后，再选择是否下载为 PDF 或 Excel 文件。"
      >
        <AppCard
          variant="hero"
          accent="primary"
          eyebrow="Export Task"
          title="班级报告生成器"
          subtitle="先确定班级和导出格式，再把任务交给后端。下载变为可选动作，不影响当前页面预览。"
        >
          <label class="block">
            <span class="mb-2 block text-sm font-medium text-[var(--color-text-primary)]">班级名称</span>
            <input
              v-model="form.className"
              type="text"
              :placeholder="classNamePlaceholder"
              class="w-full border border-l-2 border-[var(--color-border-default)] bg-transparent px-3 py-2.5 text-[var(--color-text-primary)] outline-none transition focus:border-[var(--color-primary)] focus:ring-0"
            >
          </label>

          <fieldset>
            <legend class="mb-2 text-sm font-medium text-[var(--color-text-primary)]">
              导出格式
            </legend>
            <div class="grid gap-3 sm:grid-cols-2">
              <label
                class="flex items-start gap-3 border border-l-2 px-3 py-3 transition"
                :class="
                  form.format === 'pdf'
                    ? 'border-[var(--color-primary)] bg-[var(--color-primary)]/6'
                    : 'border-[var(--color-border-default)] bg-transparent hover:border-[var(--color-primary)]/50'
                "
              >
                <input
                  v-model="form.format"
                  type="radio"
                  value="pdf"
                  class="mt-1"
                >
                <span>
                  <span class="block font-medium text-[var(--color-text-primary)]">PDF</span>
                  <span class="mt-1 block text-sm text-[var(--color-text-secondary)]">适合打印、归档和正式汇报。</span>
                </span>
              </label>

              <label
                class="flex items-start gap-3 border border-l-2 px-3 py-3 transition"
                :class="
                  form.format === 'excel'
                    ? 'border-[var(--color-primary)] bg-[var(--color-primary)]/6'
                    : 'border-[var(--color-border-default)] bg-transparent hover:border-[var(--color-primary)]/50'
                "
              >
                <input
                  v-model="form.format"
                  type="radio"
                  value="excel"
                  class="mt-1"
                >
                <span>
                  <span class="block font-medium text-[var(--color-text-primary)]">Excel</span>
                  <span class="mt-1 block text-sm text-[var(--color-text-secondary)]">适合继续分析、筛选和二次加工。</span>
                </span>
              </label>
            </div>
          </fieldset>

          <AppCard
            variant="action"
            accent="neutral"
          >
            如果当前账号已绑定班级，可直接留空使用默认班级；管理员也可手动输入其他班级名称。
          </AppCard>

          <div class="flex flex-wrap gap-3">
            <button
              type="button"
              :disabled="previewLoading"
              class="inline-flex items-center border border-[var(--color-border-default)] px-4 py-2.5 text-sm font-medium text-[var(--color-text-primary)] transition hover:border-[var(--color-primary)] hover:text-[var(--color-primary)] disabled:cursor-not-allowed disabled:opacity-60"
              @click="loadPreview"
            >
              {{ previewLoading ? '加载预览中...' : '查看当前预览' }}
            </button>

            <button
              type="button"
              :disabled="submitting"
              class="inline-flex items-center border border-[var(--color-primary)] bg-[var(--color-primary)] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[var(--color-primary-hover)] disabled:cursor-not-allowed disabled:opacity-60"
              @click="handleExport"
            >
              {{ submitting ? '提交中...' : '创建导出任务' }}
            </button>
          </div>
        </AppCard>
      </SectionCard>

      <div class="space-y-6">
        <SectionCard
          title="最近一次任务"
          subtitle="导出状态、下载信息和任务元数据都在这里收口。"
        >
          <AppEmpty
            v-if="!latestExport"
            title="还没有创建导出任务"
            description="先在左侧创建一次班级报告任务，这里会展示最近一次任务状态。"
            icon="FileChartColumnIncreasing"
          />

          <AppCard
            v-else
            variant="hero"
            :accent="latestExport.result.status === 'ready' ? 'success' : latestExport.result.status === 'failed' ? 'danger' : 'warning'"
            eyebrow="Latest Task"
            :title="String(latestExport.result.report_id)"
            :subtitle="derivedDownloadHint"
          >
            <template #header>
              <span
                class="rounded-full px-3 py-1 text-xs font-semibold"
                :class="
                  latestExport.result.status === 'ready'
                    ? 'bg-emerald-500/12 text-emerald-600'
                    : latestExport.result.status === 'failed'
                      ? 'bg-rose-500/12 text-rose-600'
                      : 'bg-amber-500/12 text-amber-600'
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
                variant="metric"
                accent="primary"
                eyebrow="班级"
                :title="latestExport.className"
              />
              <AppCard
                variant="metric"
                accent="primary"
                eyebrow="格式"
                :title="latestExport.format.toUpperCase()"
              />
              <AppCard
                variant="metric"
                accent="neutral"
                eyebrow="创建时间"
                :title="formatDate(latestExport.createdAt)"
              />
              <AppCard
                variant="metric"
                accent="neutral"
                eyebrow="过期时间"
                :title="latestExport.result.expires_at ? formatDate(latestExport.result.expires_at) : '待生成完成后返回'"
              />
            </div>

            <button
              type="button"
              :disabled="downloading || latestExport.result.status !== 'ready'"
              class="inline-flex items-center border border-[var(--color-border-default)] px-4 py-2.5 text-sm font-medium text-[var(--color-text-primary)] transition hover:border-[var(--color-primary)] hover:text-[var(--color-primary)] disabled:cursor-not-allowed disabled:opacity-60"
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
        </SectionCard>

        <SectionCard
          title="使用说明"
          subtitle="导出链路和当前后端能力边界。"
        >
          <ol class="space-y-3 text-sm leading-6 text-[var(--color-text-secondary)]">
            <li>1. 先点击“查看当前预览”，直接在页面内查看当前班级报告内容。</li>
            <li>2. 确认需要留档时，再创建后台导出任务。</li>
            <li>3. 若状态为“已就绪”，可直接下载；若为“生成中”，页面会自动轮询状态。</li>
          </ol>
        </SectionCard>
      </div>
    </section>

    <section class="space-y-6">
      <PageHeader
        eyebrow="Live Preview"
        title="当前报告预览"
        description="不下载也能直接查看当前班级的关键报告内容。"
      >
        <AppCard
          variant="action"
          accent="neutral"
        >
          预览班级：<span class="font-medium text-[var(--color-text-primary)]">{{
            previewClassName || '未选择'
          }}</span>
        </AppCard>
      </PageHeader>

      <div
        v-if="previewError"
        class="border-b border-l-2 border-amber-300 bg-amber-50/60 px-4 py-3 text-sm text-amber-700"
      >
        {{ previewError }}
      </div>

      <div
        v-if="previewLoading"
        class="grid gap-4 md:grid-cols-3"
      >
        <div
          v-for="index in 3"
          :key="index"
          class="h-24 animate-pulse border-b border-[var(--color-border-default)] bg-[var(--color-bg-surface)]"
        />
      </div>

      <template v-else-if="previewSummary">
        <section class="grid gap-4 md:grid-cols-3">
          <MetricCard
            label="班级人数"
            :value="previewSummary.student_count"
            hint="当前预览班级纳入统计的学生数"
            accent="success"
          />
          <MetricCard
            label="平均解题"
            :value="averageSolvedText"
            hint="当前班级学生的人均解题数"
            accent="warning"
          />
          <MetricCard
            label="近 7 天活跃率"
            :value="activeRateText"
            hint="近 7 天至少有一次训练动作的学生占比"
            accent="primary"
          />
        </section>

        <TeacherClassTrendPanel
          :trend="previewTrend"
          title="班级近 7 天训练趋势"
          subtitle="直接查看当前班级训练事件、成功解题和活跃学生走势。"
        />

        <TeacherClassReviewPanel
          :review="previewReview"
          :class-name="previewClassName"
        />

        <TeacherClassInsightsPanel
          :students="previewStudents"
          :class-name="previewClassName"
        />
      </template>

      <AppEmpty
        v-else
        title="还没有可用预览"
        description="先选择班级并加载一次预览，这里会展示当前报告内容。"
        icon="FileChartColumnIncreasing"
      />
    </section>
  </div>
</template>
