<script setup lang="ts">
import ContestEditTopbarPanel from '@/components/platform/contest/ContestEditTopbarPanel.vue'
import ContestEditWorkspacePanel from '@/components/platform/contest/ContestEditWorkspacePanel.vue'
import ContestWorkbenchStageTabs from '@/components/platform/contest/ContestWorkbenchStageTabs.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import { useContestEditPage } from '@/features/platform-contests'

const {
  loading,
  loadError,
  saving,
  contest,
  formDraft,
  fieldLocks,
  statusOptions,
  pageTitle,
  activeStage,
  selectTab,
  workbench,
  awdChallengeLinks,
  awdChallengePoolCreateRequestKey,
  awdPreflightLoadError,
  awdReadiness,
  loadingAwdStageData,
  refreshAwdWorkbenchData,
  handleDraftChange,
  goBackToContestList,
  goToContestAnnouncements,
  handleWorkspaceStageNavigation,
  openAwdConfigPage,
  handleNavigateAwdChallengeFromPreflight,
  handleSave,
  getModeLabel,
  getStatusLabel,
} = useContestEditPage()
</script>

<template>
  <div class="workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero contest-studio-shell">
    <div
      v-if="loading"
      class="studio-loading-overlay"
    >
      <AppLoading>正在同步竞赛工作台...</AppLoading>
    </div>

    <main class="studio-content">
      <ContestEditTopbarPanel
        v-if="contest"
        :page-title="pageTitle"
        :contest-mode="contest.mode"
        :contest-status="contest.status"
        :contest-mode-label="getModeLabel(contest.mode)"
        :contest-status-label="getStatusLabel(contest.status)"
        :active-stage="activeStage"
        :saving="saving"
        @back="goBackToContestList"
        @open-announcements="goToContestAnnouncements"
        @save="formDraft && void handleSave(formDraft)"
      />

      <ContestWorkbenchStageTabs
        v-if="contest"
        :stages="workbench.visibleStages"
        :active-stage="activeStage"
        :select-stage="selectTab"
      />

      <ContestEditWorkspacePanel
        :load-error="loadError"
        :form-draft="formDraft"
        :contest="contest"
        :active-stage="activeStage"
        :saving="saving"
        :status-options="statusOptions"
        :field-locks="fieldLocks"
        :loading-awd-stage-data="loadingAwdStageData"
        :awd-challenge-links="awdChallengeLinks"
        :awd-challenge-pool-create-request-key="awdChallengePoolCreateRequestKey"
        :awd-preflight-load-error="awdPreflightLoadError"
        :awd-readiness="awdReadiness"
        @go-back="goBackToContestList"
        @update:draft="handleDraftChange"
        @save="handleSave"
        @refresh-awd-workbench="contest && void refreshAwdWorkbenchData(contest.id)"
        @edit:awd-challenge="openAwdConfigPage"
        @retry:preflight="contest && void refreshAwdWorkbenchData(contest.id)"
        @navigate:awd-challenge-from-preflight="handleNavigateAwdChallengeFromPreflight"
        @navigate:stage="handleWorkspaceStageNavigation"
      />
    </main>
  </div>
</template>

<style scoped>
.contest-studio-shell {
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
  display: flex;
  flex-direction: column;
  height: calc(100vh - 64px);
  width: 100%;
  overflow: hidden;
  background: var(--color-bg-base);
}

.studio-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  min-width: 0;
}

.studio-loading-overlay {
  position: absolute;
  inset: 0;
  z-index: 100;
  background: color-mix(in srgb, var(--color-bg-base) 80%, transparent);
  backdrop-filter: blur(12px);
  display: flex;
  align-items: center;
  justify-content: center;
}

</style>
