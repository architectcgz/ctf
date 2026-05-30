<script setup lang="ts">
import { computed } from 'vue'

import type { ReportExportData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { formatDate } from '@/utils/format'

type ReportFormat = 'pdf' | 'excel'

interface ExportRecord {
  className: string
  format: ReportFormat
  fromDate: string
  toDate: string
  createdAt: string
  result: ReportExportData
}

const props = defineProps<{
  latestExport: ExportRecord | null
  derivedDownloadHint: string
  latestStatusMeta: {
    label: string
    chipClass: string
  }
  latestWindowLabel: string
  latestExpiresText: string
  polling: boolean
  downloading: boolean
  handleDownload: () => void | Promise<void>
}>()

const downloadButtonLabel = computed(() => {
  if (props.downloading) {
    return '下载中...'
  }
  if (props.latestExport?.result.status === 'ready') {
    return '下载报告'
  }
  return props.polling ? '等待生成完成' : '等待任务完成'
})

const latestCreatedAtText = computed(() => {
  if (!props.latestExport) return '--'
  return formatDate(props.latestExport.createdAt)
})
</script>

<template>
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
          <dd>{{ latestCreatedAtText }}</dd>
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
        {{ downloadButtonLabel }}
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
</template>
