<template>
  <div class="workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero contest-studio-shell">
    <AppRouteRedirect :to="saveSuccessRedirectRoute" />

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
        :announcements-route="contestAnnouncementsRoute"
        :active-stage="activeStage"
        :saving="saving"
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
        :back-route="backToContestListRoute"
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
        :build-awd-config-route="buildAwdConfigRoute"
        :resolve-awd-config-route-from-preflight="resolveAwdConfigRouteFromPreflight"
        @update:draft="handleDraftChange"
        @save="handleSave"
        @refresh-awd-workbench="contest && void refreshAwdWorkbenchData(contest.id)"
        @retry:preflight="contest && void refreshAwdWorkbenchData(contest.id)"
      />
    </main>
  </div>
</template>

<script setup lang="ts">
import { toRef } from 'vue'

import AppLoading from '@/shared/ui/common/AppLoading.vue'
import AppRouteRedirect from '@/shared/ui/navigation/AppRouteRedirect.vue'
import { ContestWorkbenchStageTabs } from '@/features/contest-workbench'
import {
  ContestEditTopbarPanel,
  ContestEditWorkspacePanel,
  useContestEditPage,
} from '@/features/platform/contests'

const props = defineProps<{
  contestId: string
}>()

const {
  loading,
  loadError,
  saving,
  saveSuccessRedirectRoute,
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
  backToContestListRoute,
  contestAnnouncementsRoute,
  buildAwdConfigRoute,
  resolveAwdConfigRouteFromPreflight,
  handleSave,
  getModeLabel,
  getStatusLabel,
} = useContestEditPage(toRef(props, 'contestId'))
</script>

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
