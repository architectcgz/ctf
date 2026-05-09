<script setup lang="ts">
import {
  ArrowLeft,
  Download,
  FileDown,
} from 'lucide-vue-next'

type ExportKind = 'archive' | 'report'

defineProps<{
  loading: boolean
  hasReview: boolean
  exporting: ExportKind | null
  canExportReport: boolean
}>()

const emit = defineEmits<{
  openIndex: []
  exportArchive: []
  exportReport: []
}>()
</script>

<template>
  <button
    type="button"
    class="header-btn header-btn--ghost"
    @click="emit('openIndex')"
  >
    <ArrowLeft class="h-4 w-4" />
    返回列表
  </button>
  <button
    data-testid="awd-review-export-archive"
    type="button"
    class="header-btn header-btn--ghost"
    :disabled="loading || !hasReview || exporting === 'archive'"
    @click="emit('exportArchive')"
  >
    <Download class="h-4 w-4" />
    归档导出
  </button>
  <button
    data-testid="awd-review-export-report"
    type="button"
    class="header-btn header-btn--primary"
    :disabled="loading || !hasReview || exporting === 'report' || !canExportReport"
    @click="emit('exportReport')"
  >
    <FileDown class="h-4 w-4" />
    生成评估报告
  </button>
</template>
