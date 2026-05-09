<script setup lang="ts">
import PlatformContestFormDialog from '@/components/platform/contest/PlatformContestFormDialog.vue'
import AWDReadinessOverrideDialog from '@/components/platform/contest/AWDReadinessOverrideDialog.vue'
import ContestAnnouncementManageDrawer from '@/components/platform/contest/ContestAnnouncementManageDrawer.vue'
import ContestOrchestrationPage from '@/components/platform/contest/ContestOrchestrationPage.vue'
import { useContestManagePage } from '@/features/platform-contests'

const {
  list,
  total,
  summary,
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
  confirmAWDStartOverride,
  saveContest,
  awdContests,
  requestedPanel,
  requestedPanelVersion,
  announcementDrawerOpen,
  activeAnnouncementContest,
  updateStatusFilter,
  handleDialogOpenChange,
  handleAwdStartOverrideDialogOpenChange,
  openAnnouncementDrawer,
  closeAnnouncementDrawer,
  handleCreateContestSave,
} = useContestManagePage()
</script>

<template>
  <div class="space-y-6">
    <ContestOrchestrationPage
      :list="list"
      :total="total"
      :summary="summary"
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
