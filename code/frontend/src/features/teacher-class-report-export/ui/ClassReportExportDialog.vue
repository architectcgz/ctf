<script setup lang="ts">
import { computed, watch } from 'vue'

import AdminSurfaceModal from '@/components/common/modal-templates/AdminSurfaceModal.vue'
import { useClassReportExport } from '@/features/teacher-class-report-export'

import ClassReportExportContextSection from './ClassReportExportContextSection.vue'
import ClassReportExportPreviewSection from './ClassReportExportPreviewSection.vue'
import ClassReportExportTaskRail from './ClassReportExportTaskRail.vue'
import './classReportExportDialog.css'

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
} = useClassReportExport()

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
        <ClassReportExportContextSection
          :form="form"
          :preview-class-name="previewClassName"
          :normalized-class-name-text="normalizedClassNameText"
          :selected-window-label="selectedWindowLabel"
          :selected-window-error="selectedWindowError"
          :class-name-placeholder="classNamePlaceholder"
          :selected-format-label="selectedFormatLabel"
          :selected-format-hint="selectedFormatHint"
          :latest-export="latestExport"
          :latest-status-meta="latestStatusMeta"
          :derived-download-hint="derivedDownloadHint"
          :preview-loading="previewLoading"
          :submitting="submitting"
          :load-preview="loadPreview"
          :handle-export="handleExport"
        />

        <ClassReportExportPreviewSection
          :preview-error="previewError"
          :preview-loading="previewLoading"
          :preview-summary="previewSummary"
          :preview-trend="previewTrend"
          :preview-review="previewReview"
          :preview-students="previewStudents"
          :preview-class-name="previewClassName"
          :selected-window-label="selectedWindowLabel"
          :average-solved-text="averageSolvedText"
          :active-rate-text="activeRateText"
        />
      </main>

      <aside class="class-report-dialog__rail">
        <ClassReportExportTaskRail
          :latest-export="latestExport"
          :derived-download-hint="derivedDownloadHint"
          :latest-status-meta="latestStatusMeta"
          :latest-window-label="latestWindowLabel"
          :latest-expires-text="latestExpiresText"
          :polling="polling"
          :downloading="downloading"
          :handle-download="handleDownload"
        />
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
