<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import type { ContestDetailData } from '@/api/contracts'
import PlatformContestFormDialog from '@/components/platform/contest/PlatformContestFormDialog.vue'
import AWDReadinessOverrideDialog from '@/components/platform/contest/AWDReadinessOverrideDialog.vue'
import ContestAnnouncementManageDrawer from '@/components/platform/contest/ContestAnnouncementManageDrawer.vue'
import ContestOrchestrationPage from '@/components/platform/contest/ContestOrchestrationPage.vue'
import { useContestExportFlow } from '@/composables/useContestExportFlow'
import { usePlatformContests } from '@/composables/usePlatformContests'

const { handleExportContest } = useContestExportFlow()

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
