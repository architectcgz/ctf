<script setup lang="ts">
import { FileDown, RefreshCcw } from 'lucide-vue-next'

import type { ReportExportData } from '@/api/contracts'

type ReportFormat = 'pdf' | 'excel'

interface ClassReportExportFormState {
  className: string
  format: ReportFormat
  fromDate: string
  toDate: string
}

interface ExportRecord {
  className: string
  format: ReportFormat
  fromDate: string
  toDate: string
  createdAt: string
  result: ReportExportData
}

defineProps<{
  form: ClassReportExportFormState
  previewClassName: string
  normalizedClassNameText: string
  selectedWindowLabel: string
  selectedWindowError: string | null
  classNamePlaceholder: string
  selectedFormatLabel: string
  selectedFormatHint: string
  latestExport: ExportRecord | null
  latestStatusMeta: {
    label: string
    chipClass: string
  }
  derivedDownloadHint: string
  previewLoading: boolean
  submitting: boolean
  loadPreview: () => void | Promise<void>
  handleExport: () => void | Promise<void>
}>()
</script>

<template>
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
</template>
