<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import { downloadReport } from '@/api/assessment'
import { exportContestArchive } from '@/api/admin'
import type { ContestDetailData } from '@/api/contracts'
import { ApiError } from '@/api/request'
import PlatformContestFormDialog from '@/components/platform/contest/PlatformContestFormDialog.vue'
import AWDReadinessOverrideDialog from '@/components/platform/contest/AWDReadinessOverrideDialog.vue'
import ContestAnnouncementManageDrawer from '@/components/platform/contest/ContestAnnouncementManageDrawer.vue'
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
const announcementDrawerOpen = ref(false)
const activeAnnouncementContest = ref<ContestDetailData | null>(null)

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

function openAnnouncementDrawer(contest: ContestDetailData): void {
  activeAnnouncementContest.value = contest
  announcementDrawerOpen.value = true
}

function closeAnnouncementDrawer(): void {
  announcementDrawerOpen.value = false
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

function notifyContestExportError(error: unknown, fallback: string): void {
  console.error(fallback, error)
  if (error instanceof ApiError) {
    return
  }
  const message = error instanceof Error && error.message.trim() ? error.message : fallback
  toast.error(message)
}

async function downloadContestReport(reportId: string, contestTitle: string): Promise<void> {
  try {
    await downloadGeneratedReport(reportId)
    toast.success(`赛事结果已导出：${contestTitle}`)
  } catch (error) {
    notifyContestExportError(error, `赛事结果下载失败：${contestTitle}`)
  }
}

async function handleExportContest(contest: ContestDetailData): Promise<void> {
  exportingContestId.value = contest.id
  try {
    const result = await exportContestArchive(contest.id, { format: 'json' })

    if (result.status === 'ready') {
      pendingContestReportId.value = null
      stopPolling()
      await downloadContestReport(result.report_id, contest.title)
      return
    }

    if (result.status === 'failed') {
      pendingContestReportId.value = null
      stopPolling()
      toast.error(result.error_message || '赛事结果导出失败')
      return
    }

    pendingContestReportId.value = result.report_id
    startPolling(result.report_id, (next) => {
      if (next.report_id !== pendingContestReportId.value) return
      if (next.status === 'ready') {
        pendingContestReportId.value = null
        stopPolling()
        void downloadContestReport(next.report_id, contest.title)
        return
      }
      if (next.status === 'failed') {
        pendingContestReportId.value = null
        stopPolling()
        toast.error(next.error_message || '赛事结果导出失败')
      }
    }, (error) => {
      pendingContestReportId.value = null
      notifyContestExportError(error, `赛事结果生成状态同步失败：${contest.title}`)
    })
    toast.info(`已开始导出赛事结果：${contest.title}`)
  } catch (error) {
    pendingContestReportId.value = null
    stopPolling()
    notifyContestExportError(error, `赛事结果导出失败：${contest.title}`)
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
      @announce="openAnnouncementDrawer"
      @export-contest="handleExportContest"
      @change-page="changePage"
    />

    <ContestAnnouncementManageDrawer
      :open="announcementDrawerOpen"
      :contest="activeAnnouncementContest"
      @close="closeAnnouncementDrawer"
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
