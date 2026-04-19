<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { downloadReport } from '@/api/assessment'
import { exportContestArchive } from '@/api/admin'
import type { ContestDetailData } from '@/api/contracts'
import PlatformContestFormDialog from '@/components/platform/contest/PlatformContestFormDialog.vue'
import AWDReadinessOverrideDialog from '@/components/platform/contest/AWDReadinessOverrideDialog.vue'
import ContestOrchestrationPage from '@/components/platform/contest/ContestOrchestrationPage.vue'
import { useReportStatusPolling } from '@/composables/useReportStatusPolling'
import { useToast } from '@/composables/useToast'
import { usePlatformContests } from '@/composables/usePlatformContests'

const toast = useToast()
const { start: startPolling, stop: stopPolling } = useReportStatusPolling()

const {
  list,
  total,
  page,
  pageSize,
  loading,
  refresh,
  changePage,
  statusFilter,
  dialogOpen,
  mode,
  saving,
  formDraft,
  fieldLocks,
  statusOptions,
  awdStartOverrideDialogState,
  prepareCreateContest,
  openEditDialog,
  closeDialog,
  closeAWDStartOverrideDialog,
  confirmAWDStartOverride,
  saveContest,
} = usePlatformContests()

const awdContests = computed(() => list.value.filter((item) => item.mode === 'awd'))
const exportingContestId = ref<string | null>(null)
const downloadingContestReport = ref(false)
const pendingContestReportId = ref<string | null>(null)
const requestedPanel = ref<'overview' | 'list' | 'create' | null>(null)
const requestedPanelVersion = ref(0)

onMounted(() => {
  void refresh()
})

function updateStatusFilter(value: typeof statusFilter.value) {
  statusFilter.value = value
}

function requestContestPanel(panel: 'overview' | 'list' | 'create') {
  requestedPanel.value = panel
  requestedPanelVersion.value += 1
}

function handleDialogOpenChange(value: boolean) {
  if (!value) {
    closeDialog()
  }
}

function handleAwdStartOverrideDialogOpenChange(value: boolean) {
  if (!value) {
    closeAWDStartOverrideDialog()
  }
}

async function downloadGeneratedReport(reportId: string): Promise<void> {
  downloadingContestReport.value = true
  try {
    const { blob, filename } = await downloadReport(reportId)
    const objectUrl = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = objectUrl
    link.download = filename
    document.body.appendChild(link)
    link.click()
    link.remove()
    URL.revokeObjectURL(objectUrl)
  } finally {
    downloadingContestReport.value = false
  }
}

async function handleExportContest(contest: ContestDetailData): Promise<void> {
  exportingContestId.value = contest.id
  try {
    const result = await exportContestArchive(contest.id, { format: 'json' })

    if (result.status === 'ready') {
      stopPolling()
      await downloadGeneratedReport(result.report_id)
      toast.success(`赛事结果已导出：${contest.title}`)
      return
    }

    if (result.status === 'failed') {
      stopPolling()
      toast.error(result.error_message || '赛事结果导出失败')
      return
    }

    pendingContestReportId.value = result.report_id
    startPolling(result.report_id, (next) => {
      if (next.report_id !== pendingContestReportId.value) return
      if (next.status === 'ready') {
        pendingContestReportId.value = null
        void downloadGeneratedReport(next.report_id)
        toast.success(`赛事结果已导出：${contest.title}`)
        return
      }
      if (next.status === 'failed') {
        pendingContestReportId.value = null
        toast.error(next.error_message || '赛事结果导出失败')
      }
    })
    toast.info(`已开始导出赛事结果：${contest.title}`)
  } finally {
    exportingContestId.value = null
  }
}

async function handleCreateContestSave(draft: Parameters<typeof saveContest>[0]): Promise<void> {
  const result = await saveContest(draft)
  if (result === 'create') {
    requestContestPanel('list')
  }
}
</script>

<template>
  <div class="space-y-6">
    <ContestOrchestrationPage
      :list="list"
      :total="total"
      :page="page"
      :page-size="pageSize"
      :loading="loading"
      :status-filter="statusFilter"
      :awd-contests="awdContests"
      :create-draft="formDraft"
      :create-saving="saving"
      :create-field-locks="fieldLocks"
      :requested-panel="requestedPanel"
      :requested-panel-version="requestedPanelVersion"
      @refresh="refresh"
      @prepare-create-contest="prepareCreateContest"
      @save-create-contest="handleCreateContestSave"
      @update-status-filter="updateStatusFilter"
      @open-edit-dialog="openEditDialog"
      @export-contest="handleExportContest"
      @change-page="changePage"
    />

    <PlatformContestFormDialog
      :open="dialogOpen"
      :mode="mode"
      :draft="formDraft"
      :saving="saving"
      :status-options="statusOptions"
      :field-locks="fieldLocks"
      @update:open="handleDialogOpenChange"
      @save="saveContest"
    />

    <AWDReadinessOverrideDialog
      :open="awdStartOverrideDialogState.open"
      :title="awdStartOverrideDialogState.title"
      :readiness="awdStartOverrideDialogState.readiness"
      :confirm-loading="awdStartOverrideDialogState.confirmLoading"
      @update:open="handleAwdStartOverrideDialogOpenChange"
      @confirm="confirmAWDStartOverride"
    />
  </div>
</template>
