<script setup lang="ts">
import { computed, watch } from 'vue'
import { FileDown, RefreshCcw } from 'lucide-vue-next'

import AppEmpty from '@/components/common/AppEmpty.vue'
import AdminSurfaceModal from '@/components/common/modal-templates/AdminSurfaceModal.vue'
import TeacherClassInsightsPanel from '@/components/teacher/TeacherClassInsightsPanel.vue'
import TeacherClassReviewPanel from '@/components/teacher/TeacherClassReviewPanel.vue'
import TeacherClassTrendPanel from '@/components/teacher/TeacherClassTrendPanel.vue'
import { useTeacherClassReportExport } from '@/features/teacher-class-report-export'
import { formatDate } from '@/utils/format'

const props = defineProps<{
  modelValue: boolean
  defaultClassName?: string
  defaultFromDate?: string
  defaultToDate?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const {
  polling,
  form,
  submitting,
  downloading,
  latestExport,
  previewLoading,
  previewError,
  previewClassName,
  previewStudents,
  previewReview,
  previewSummary,
  previewTrend,
  classNamePlaceholder,
  normalizedClassNameText,
  selectedWindowLabel,
  selectedWindowError,
  selectedFormatLabel,
  selectedFormatHint,
  derivedDownloadHint,
  averageSolvedText,
  activeRateText,
  latestStatusMeta,
  latestExpiresText,
  latestWindowLabel,
  syncContext,
  loadPreview,
  handleExport,
  handleDownload,
} = useTeacherClassReportExport()

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit('update:modelValue', value),
})

watch(
  () => props.modelValue,
  (isOpen) => {
    if (!isOpen) return
    syncContext({
      className: props.defaultClassName,
      fromDate: props.defaultFromDate,
      toDate: props.defaultToDate,
    })
    void loadPreview()
  },
  { immediate: true }
)

watch(
  () => [props.defaultClassName, props.defaultFromDate, props.defaultToDate] as const,
  (nextValue, previousValue) => {
    if (!props.modelValue || nextValue === previousValue) return
    syncContext({
      className: nextValue[0],
      fromDate: nextValue[1],
      toDate: nextValue[2],
    })
    void loadPreview()
  }
)

function closeDialog(): void {
  dialogVisible.value = false
}
</script>

<template>
  <AdminSurfaceModal
    :open="dialogVisible"
    title="班级报告导出"
    subtitle="在当前教师上下文内查看班级训练预览，并创建可下载的报告任务。"
    eyebrow="Class Report"
    width="74rem"
    @close="closeDialog"
    @update:open="dialogVisible = $event"
  >
    <div class="class-report-dialog__shell">
      <main class="class-report-dialog__main">
        <section class="class-report-section class-report-section--context">
          <div class="class-report-section__head">
            <div>
              <div class="journal-eyebrow">
                Current Context
              </div>
              <h4 class="class-report-section__title">
                当前教师上下文
              </h4>
              <p class="class-report-section__copy">
                导出任务会优先使用当前教师绑定班级，并沿用当前页面的训练时间段。
              </p>
            </div>
            <div class="class-report-context-chips">
              <span class="teacher-surface-chip">
                当前班级：{{ previewClassName || normalizedClassNameText }}
              </span>
              <span class="teacher-surface-chip">
                当前窗口：{{ selectedWindowLabel }}
              </span>
            </div>
          </div>
        </section>

        <section class="class-report-section class-report-section--controls">
          <div class="class-report-section__head">
            <div>
              <div class="journal-eyebrow">
                Context Action
              </div>
              <h4 class="class-report-section__title">
                导出设置
              </h4>
              <p class="class-report-section__copy">
                默认沿用当前教师或页面上下文班级，也可以在这里临时调整导出窗口与班级范围。
              </p>
            </div>
          </div>

          <div class="class-report-form-grid">
            <label class="ui-field class-report-field">
              <span class="ui-field__label">班级名称</span>
              <span class="ui-control-wrap">
                <input
                  v-model="form.className"
                  type="text"
                  :placeholder="classNamePlaceholder"
                  class="ui-control class-report-field__control"
                >
              </span>
            </label>

            <div class="class-report-range-grid">
              <label class="ui-field class-report-field">
                <span class="ui-field__label">开始日期</span>
                <span class="ui-control-wrap">
                  <input
                    v-model="form.fromDate"
                    type="date"
                    class="ui-control class-report-field__control"
                  >
                </span>
              </label>

              <label class="ui-field class-report-field">
                <span class="ui-field__label">结束日期</span>
                <span class="ui-control-wrap">
                  <input
                    v-model="form.toDate"
                    type="date"
                    class="ui-control class-report-field__control"
                  >
                </span>
              </label>
            </div>

            <fieldset class="class-report-format-group">
              <legend class="ui-field__label class-report-format-group__label">
                导出格式
              </legend>
              <div class="class-report-format-grid">
                <label
                  class="class-report-format-option"
                  :class="{ 'class-report-format-option--active': form.format === 'pdf' }"
                >
                  <input
                    v-model="form.format"
                    type="radio"
                    value="pdf"
                  >
                  <span>
                    <span class="class-report-format-option__title">PDF</span>
                    <span class="class-report-format-option__copy">适合打印、归档和正式汇报。</span>
                  </span>
                </label>

                <label
                  class="class-report-format-option"
                  :class="{ 'class-report-format-option--active': form.format === 'excel' }"
                >
                  <input
                    v-model="form.format"
                    type="radio"
                    value="excel"
                  >
                  <span>
                    <span class="class-report-format-option__title">Excel</span>
                    <span class="class-report-format-option__copy">适合继续分析、筛选和二次加工。</span>
                  </span>
                </label>
              </div>
            </fieldset>
          </div>

          <p
            v-if="selectedWindowError"
            class="class-report-section__warning"
            role="alert"
          >
            {{ selectedWindowError }}
          </p>

          <div
            class="class-report-section__actions"
            role="group"
            aria-label="班级报告操作"
          >
            <button
              type="button"
              class="ui-btn ui-btn--secondary"
              :disabled="previewLoading"
              @click="loadPreview"
            >
              <RefreshCcw class="h-4 w-4" />
              {{ previewLoading ? '加载预览中...' : '重新加载预览' }}
            </button>
            <button
              type="button"
              class="ui-btn ui-btn--primary"
              :disabled="submitting"
              @click="handleExport"
            >
              <FileDown class="h-4 w-4" />
              {{ submitting ? '提交中...' : '创建导出任务' }}
            </button>
          </div>

          <section class="class-report-preview-summary metric-panel-default-surface">
            <div class="class-report-preview-summary__title">
              Preview Snapshot
            </div>
            <div class="class-report-preview-summary__grid progress-strip metric-panel-grid metric-panel-default-surface">
              <article class="progress-card metric-panel-card">
                <div class="progress-card-label metric-panel-label">
                  目标班级
                </div>
                <div class="progress-card-value metric-panel-value">
                  {{ normalizedClassNameText }}
                </div>
                <div class="progress-card-hint metric-panel-helper">
                  预览与导出都会优先使用这个班级
                </div>
              </article>
              <article class="progress-card metric-panel-card">
                <div class="progress-card-label metric-panel-label">
                  时间窗口
                </div>
                <div class="progress-card-value metric-panel-value">
                  {{ selectedWindowLabel }}
                </div>
                <div class="progress-card-hint metric-panel-helper">
                  预览与导出共用这一段训练时间
                </div>
              </article>
              <article class="progress-card metric-panel-card">
                <div class="progress-card-label metric-panel-label">
                  导出格式
                </div>
                <div class="progress-card-value metric-panel-value">
                  {{ selectedFormatLabel }}
                </div>
                <div class="progress-card-hint metric-panel-helper">
                  {{ selectedFormatHint }}
                </div>
              </article>
              <article class="progress-card metric-panel-card">
                <div class="progress-card-label metric-panel-label">
                  任务状态
                </div>
                <div class="progress-card-value metric-panel-value">
                  {{ latestStatusMeta.label }}
                </div>
                <div class="progress-card-hint metric-panel-helper">
                  {{
                    latestExport
                      ? derivedDownloadHint
                      : '创建一次导出任务后，这里会展示最近一次任务状态。'
                  }}
                </div>
              </article>
            </div>
          </section>
        </section>

        <section class="class-report-section class-report-section--preview">
          <div class="class-report-section__head">
            <div>
              <div class="journal-eyebrow">
                Live Preview
              </div>
              <h4 class="class-report-section__title">
                当前班级报告预览
              </h4>
              <p class="class-report-section__copy">
                不下载也能先看当前时间段内的班级趋势、教学复盘结论和学生洞察。
              </p>
            </div>
          </div>

          <div
            v-if="previewError"
            class="teacher-surface-error class-report-preview-error"
          >
            {{ previewError }}
          </div>

          <div
            v-else-if="previewLoading"
            class="class-report-preview-skeletons"
          >
            <div
              v-for="index in 3"
              :key="index"
              class="class-report-preview-skeleton"
            />
          </div>

          <template v-else-if="previewSummary">
            <section class="metric-panel-grid metric-panel-workspace-surface class-report-kpi-grid">
              <article class="progress-card metric-panel-card">
                <div class="progress-card-label metric-panel-label">
                  班级人数
                </div>
                <div class="progress-card-value metric-panel-value">
                  {{ previewSummary.student_count }}
                </div>
                <div class="progress-card-hint metric-panel-helper">
                  当前班级纳入统计的学生数
                </div>
              </article>
              <article class="progress-card metric-panel-card">
                <div class="progress-card-label metric-panel-label">
                  平均解题
                </div>
                <div class="progress-card-value metric-panel-value">
                  {{ averageSolvedText }}
                </div>
                <div class="progress-card-hint metric-panel-helper">
                  当前班级学生的人均解题数
                </div>
              </article>
              <article class="progress-card metric-panel-card">
                <div class="progress-card-label metric-panel-label">
                  当前窗口活跃率
                </div>
                <div class="progress-card-value metric-panel-value">
                  {{ activeRateText }}
                </div>
                <div class="progress-card-hint metric-panel-helper">
                  当前时间段至少有一次训练动作的学生占比
                </div>
              </article>
            </section>

            <div class="class-report-preview-stack">
              <TeacherClassTrendPanel
                :trend="previewTrend"
                title="班级训练趋势"
                :subtitle="`当前窗口：${selectedWindowLabel}`"
              />

              <TeacherClassReviewPanel
                :review="previewReview"
                :class-name="previewClassName"
              />

              <TeacherClassInsightsPanel
                :students="previewStudents"
                :class-name="previewClassName"
              />
            </div>
          </template>

          <AppEmpty
            v-else
            title="还没有可用预览"
            description="先填写班级名称并加载一次预览，这里会展示当前报告内容。"
            icon="FileChartColumnIncreasing"
          />
        </section>
      </main>

      <aside class="class-report-dialog__rail">
        <section class="class-report-section class-report-section--aside">
          <div class="class-report-section__head">
            <div>
              <div class="journal-eyebrow">
                Latest Task
              </div>
              <h4 class="class-report-section__title">
                最近一次任务
              </h4>
            </div>
          </div>

          <AppEmpty
            v-if="!latestExport"
            class="class-report-empty"
            title="还没有创建导出任务"
            description="先创建一次班级报告任务，这里会展示最近一次任务状态。"
            icon="FileChartColumnIncreasing"
          />

          <div
            v-else
            class="class-report-task-stack"
          >
            <div class="class-report-task-banner">
              <div>
                <div class="class-report-task-label">
                  任务编号
                </div>
                <div class="class-report-task-value">
                  #{{ latestExport.result.report_id }}
                </div>
                <div class="class-report-task-copy">
                  {{ derivedDownloadHint }}
                </div>
              </div>
              <span
                class="class-report-task-chip"
                :class="latestStatusMeta.chipClass"
              >
                {{ latestStatusMeta.label }}
              </span>
            </div>

            <dl class="class-report-task-details">
              <div>
                <dt>班级</dt>
                <dd>{{ latestExport.className }}</dd>
              </div>
              <div>
                <dt>时间窗口</dt>
                <dd>{{ latestWindowLabel }}</dd>
              </div>
              <div>
                <dt>格式</dt>
                <dd>{{ latestExport.format.toUpperCase() }}</dd>
              </div>
              <div>
                <dt>创建时间</dt>
                <dd>{{ formatDate(latestExport.createdAt) }}</dd>
              </div>
              <div>
                <dt>过期时间</dt>
                <dd>{{ latestExpiresText }}</dd>
              </div>
              <div>
                <dt>轮询状态</dt>
                <dd>{{ polling ? '自动更新中' : '空闲' }}</dd>
              </div>
            </dl>

            <button
              type="button"
              class="ui-btn ui-btn--primary class-report-task-download"
              :disabled="downloading || latestExport.result.status !== 'ready'"
              @click="handleDownload"
            >
              {{
                downloading
                  ? '下载中...'
                  : latestExport.result.status === 'ready'
                    ? '下载报告'
                    : polling
                      ? '等待生成完成'
                      : '等待任务完成'
              }}
            </button>
          </div>
        </section>

        <section class="class-report-section class-report-section--aside">
          <div class="class-report-section__head">
            <div>
              <div class="journal-eyebrow">
                Guide
              </div>
              <h4 class="class-report-section__title">
                使用说明
              </h4>
            </div>
          </div>

          <ul class="class-report-guide-list">
            <li>先看预览，再决定是否需要正式导出。</li>
            <li>导出任务在后端异步生成，状态会自动轮询刷新。</li>
            <li>只有任务变为“已就绪”后才可以下载报告文件。</li>
          </ul>
        </section>
      </aside>
    </div>

    <template #footer>
      <div class="class-report-dialog__footer">
        <button
          type="button"
          class="ui-btn ui-btn--secondary"
          @click="closeDialog"
        >
          取消
        </button>
        <button
          type="button"
          class="ui-btn ui-btn--primary"
          :disabled="submitting"
          @click="handleExport"
        >
          {{ submitting ? '提交中...' : '创建导出任务' }}
        </button>
      </div>
    </template>
  </AdminSurfaceModal>
</template>

<style scoped>
.class-report-dialog__shell {
  display: grid;
  grid-template-columns: minmax(0, 1.7fr) minmax(18rem, 0.95fr);
  gap: var(--space-5);
}

.class-report-dialog__main,
.class-report-dialog__rail {
  display: grid;
  gap: var(--space-5);
  align-content: start;
}

.class-report-section {
  border: 1px solid color-mix(in srgb, var(--journal-border) 78%, transparent);
  border-radius: 22px;
  background: color-mix(in srgb, var(--journal-surface-subtle) 78%, var(--color-bg-base));
  padding: var(--space-5);
}

.class-report-section__head {
  display: flex;
  justify-content: space-between;
  gap: var(--space-4);
}

.class-report-context-chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.class-report-section__title {
  margin: var(--space-3) 0 0;
  font-size: var(--font-size-1-08);
  line-height: 1.25;
  color: var(--journal-ink);
}

.class-report-section__copy {
  margin-top: var(--space-2);
  font-size: var(--font-size-0-82);
  line-height: 1.7;
  color: var(--journal-muted);
}

.class-report-form-grid {
  display: grid;
  gap: var(--space-4);
  margin-top: var(--space-5);
}

.class-report-range-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-3);
}

.class-report-field {
  --ui-field-gap: var(--space-2);
}

.class-report-field__control {
  width: 100%;
}

.class-report-format-group {
  min-width: 0;
  border: 0;
  margin: 0;
  padding: 0;
}

.class-report-format-group__label {
  display: block;
  margin-bottom: var(--space-2);
}

.class-report-format-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-3);
}

.class-report-section__warning {
  margin: var(--space-4) 0 0;
  font-size: var(--font-size-0-82);
  color: var(--color-danger);
}

.class-report-format-option {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: var(--space-3);
  align-items: start;
  border: 1px solid color-mix(in srgb, var(--journal-border) 84%, transparent);
  border-radius: 16px;
  background: color-mix(in srgb, var(--journal-surface) 78%, var(--color-bg-base));
  padding: 1rem 1rem 1.05rem;
  cursor: pointer;
  transition:
    border-color 160ms ease,
    background 160ms ease,
    transform 160ms ease;
}

.class-report-format-option:hover,
.class-report-format-option--active {
  border-color: color-mix(in srgb, var(--journal-accent) 44%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, var(--journal-surface));
}

.class-report-format-option input {
  margin-top: 0.15rem;
}

.class-report-format-option__title {
  display: block;
  font-size: var(--font-size-0-90);
  font-weight: 700;
  color: var(--journal-ink);
}

.class-report-format-option__copy {
  display: block;
  margin-top: var(--space-1-5);
  font-size: var(--font-size-0-78);
  line-height: 1.7;
  color: var(--journal-muted);
}

.class-report-section__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  margin-top: var(--space-5);
}

.class-report-preview-summary {
  margin-top: var(--space-5);
  padding: var(--space-4-5) 0 0;
  border-top: 1px dashed color-mix(in srgb, var(--journal-border) 84%, transparent);
}

.class-report-preview-summary__title {
  font-size: var(--font-size-0-78);
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.class-report-preview-summary__grid,
.class-report-kpi-grid {
  margin-top: var(--space-4);
}

.class-report-preview-error {
  margin-top: var(--space-5);
}

.class-report-preview-skeletons {
  display: grid;
  gap: var(--space-3);
  margin-top: var(--space-5);
}

.class-report-preview-skeleton {
  height: 7.25rem;
  border-radius: 20px;
  background: color-mix(in srgb, var(--journal-surface-subtle) 92%, transparent);
  animation: class-report-pulse 1.3s ease-in-out infinite;
}

.class-report-preview-stack {
  display: grid;
  gap: var(--space-5);
  margin-top: var(--space-5);
}

.class-report-empty {
  margin-top: var(--space-4);
}

.class-report-task-stack {
  display: grid;
  gap: var(--space-4);
  margin-top: var(--space-4);
}

.class-report-task-banner {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-3);
  align-items: start;
  border: 1px solid color-mix(in srgb, var(--journal-border) 78%, transparent);
  border-radius: 18px;
  background: color-mix(in srgb, var(--journal-surface) 80%, var(--color-bg-base));
  padding: 1rem;
}

.class-report-task-label {
  font-size: var(--font-size-0-76);
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.class-report-task-value {
  margin-top: var(--space-2);
  font-size: var(--font-size-1-20);
  font-weight: 700;
  color: var(--journal-ink);
}

.class-report-task-copy {
  margin-top: var(--space-2);
  font-size: var(--font-size-0-80);
  line-height: 1.7;
  color: var(--journal-muted);
}

.class-report-task-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2rem;
  padding: 0.2rem 0.85rem;
  border-radius: 999px;
  font-size: var(--font-size-0-76);
  font-weight: 700;
}

.class-report-task-chip--idle,
.class-report-task-chip--pending {
  border: 1px solid color-mix(in srgb, var(--journal-accent) 22%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent-strong);
}

.class-report-task-chip--ready {
  border: 1px solid color-mix(in srgb, var(--color-success) 28%, transparent);
  background: color-mix(in srgb, var(--color-success) 12%, transparent);
  color: color-mix(in srgb, var(--color-success) 86%, var(--journal-ink));
}

.class-report-task-chip--failed {
  border: 1px solid color-mix(in srgb, var(--color-danger) 28%, transparent);
  background: color-mix(in srgb, var(--color-danger) 10%, transparent);
  color: color-mix(in srgb, var(--color-danger) 86%, var(--journal-ink));
}

.class-report-task-details {
  display: grid;
  gap: var(--space-3);
}

.class-report-task-details div {
  display: grid;
  gap: var(--space-1);
}

.class-report-task-details dt {
  font-size: var(--font-size-0-76);
  color: var(--journal-muted);
}

.class-report-task-details dd {
  margin: 0;
  font-size: var(--font-size-0-86);
  color: var(--journal-ink);
}

.class-report-task-download {
  width: 100%;
  justify-content: center;
}

.class-report-guide-list {
  display: grid;
  gap: var(--space-3);
  margin: var(--space-4) 0 0;
  padding-left: 1.1rem;
  color: var(--journal-muted);
}

.class-report-guide-list li {
  line-height: 1.7;
}

.class-report-dialog__footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-2);
}

@keyframes class-report-pulse {
  0%,
  100% {
    opacity: 0.72;
  }

  50% {
    opacity: 1;
  }
}

@media (max-width: 1180px) {
  .class-report-dialog__shell {
    grid-template-columns: minmax(0, 1fr);
  }
}

@media (max-width: 720px) {
  .class-report-range-grid,
  .class-report-format-grid,
  .class-report-preview-summary__grid,
  .class-report-kpi-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .class-report-section__head {
    flex-direction: column;
    align-items: flex-start;
  }

  .class-report-task-banner {
    grid-template-columns: minmax(0, 1fr);
  }

  .class-report-dialog__footer {
    flex-direction: column-reverse;
  }

  .class-report-dialog__footer > .ui-btn {
    width: 100%;
  }
}
</style>
