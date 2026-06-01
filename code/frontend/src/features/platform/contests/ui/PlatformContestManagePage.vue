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
      :active-panel="activePanel"
      :build-edit-route="buildContestEditRoute"
      :build-workbench-route="buildContestOperationsRoute"
      @refresh="refresh"
      @prepare-create-contest="prepareCreateContest"
      @save-create-contest="handleCreateContestSave"
      @switch-panel="switchPanel"
      @update-status-filter="updateStatusFilter"
      @open-edit-dialog="openEditDialog"
      @announce="openAnnouncementDrawer"
      @change-page="changePage"
    />

    <ContestAnnouncementManageDrawer
      :open="announcementDrawerOpen"
      :contest="activeAnnouncementContest"
      :full-page-route="
        activeAnnouncementContest
          ? buildContestAnnouncementsRoute(activeAnnouncementContest.id)
          : null
      "
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

<script setup lang="ts">
import { ContestAnnouncementManageDrawer } from '@/features/contest-announcements'
import { AWDReadinessOverrideDialog } from '@/features/awd-readiness'
import { useContestManagePage } from '../model'
import ContestOrchestrationPage from './ContestOrchestrationPage.vue'
import PlatformContestFormDialog from './PlatformContestFormDialog.vue'

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
  activePanel,
  announcementDrawerOpen,
  activeAnnouncementContest,
  updateStatusFilter,
  switchPanel,
  handleDialogOpenChange,
  handleAwdStartOverrideDialogOpenChange,
  openAnnouncementDrawer,
  closeAnnouncementDrawer,
  handleCreateContestSave,
  buildContestEditRoute,
  buildContestOperationsRoute,
  buildContestAnnouncementsRoute,
} = useContestManagePage()
</script>
