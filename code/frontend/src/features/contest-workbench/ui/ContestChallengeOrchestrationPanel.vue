<script setup lang="ts">
import { toRef } from 'vue'

import type {
  AdminContestChallengeViewData,
  ContestDetailData,
} from '@/api/contracts'
import { useContestChallengeOrchestration } from '../model'

import ContestChallengeDirectorySection from './ContestChallengeDirectorySection.vue'
import ContestChallengeEditorDialog from './ContestChallengeEditorDialog.vue'
import ContestChallengeFilterStrip from './ContestChallengeFilterStrip.vue'
import ContestChallengeOrchestrationHeader from './ContestChallengeOrchestrationHeader.vue'
import ContestChallengeSummaryStrip from './ContestChallengeSummaryStrip.vue'

const props = defineProps<{
  contestId: string
  contestMode: ContestDetailData['mode']
  challengeLinks?: AdminContestChallengeViewData[]
  loadingExternal?: boolean
  loadErrorExternal?: string
  createDialogRequestKey?: number
}>()

const emit = defineEmits<{
  updated: []
}>()

const {
  visibleItems,
  summaryItems,
  filterItems,
  activeFilter,
  isAwdContest,
  setFilter,
  showAwdChallengeFilters,
  showChallengeOverflowMenu,
  panelCopy,
  panelLoading,
  panelLoadError,
  currentChallengeLinks,
  emptyState,
  existingChallengeIds,
  awdChallengeFilters,
  awdChallengeCatalog,
  awdChallengePage,
  awdChallengePageSize,
  awdChallengeTotal,
  loadingAwdChallengeCatalog,
  awdChallengeLoadError,
  refreshAwdChallengeCatalog,
  changeAwdChallengePage,
  setAwdChallengeKeyword,
  setAwdChallengeServiceType,
  setAwdChallengeDeploymentMode,
  setAwdChallengeReadiness,
  dialogChallengeOptions,
  dialogOpen,
  dialogMode,
  editingChallenge,
  loadingChallengeCatalog,
  saving,
  removingChallengeId,
  openActionMenuId,
  refresh,
  handleCreateAction,
  openEditDialog,
  handleSave,
  handleRemove,
} = useContestChallengeOrchestration({
  contestId: toRef(props, 'contestId'),
  contestMode: toRef(props, 'contestMode'),
  challengeLinks: toRef(props, 'challengeLinks'),
  loadingExternal: toRef(props, 'loadingExternal'),
  loadErrorExternal: toRef(props, 'loadErrorExternal'),
  createDialogRequestKey: toRef(props, 'createDialogRequestKey'),
  onUpdated: () => emit('updated'),
})
</script>

<template>
  <section class="studio-orchestration">
    <ContestChallengeOrchestrationHeader
      :is-awd-contest="isAwdContest"
      :panel-copy="panelCopy"
      :loading="panelLoading"
      @refresh="refresh"
      @create="handleCreateAction"
    />

    <ContestChallengeSummaryStrip
      v-if="!isAwdContest && summaryItems.length > 0"
      :summary-items="summaryItems"
    />

    <ContestChallengeFilterStrip
      v-if="showAwdChallengeFilters && isAwdContest && filterItems.length > 0"
      :filter-items="filterItems"
      :active-filter="activeFilter"
      @select="setFilter"
    />

    <ContestChallengeDirectorySection
      :items="visibleItems"
      :loading="panelLoading"
      :load-error="panelLoadError"
      :empty-state="emptyState"
      :show-challenge-overflow-menu="showChallengeOverflowMenu"
      :open-action-menu-id="openActionMenuId"
      :removing-challenge-id="removingChallengeId"
      @refresh="refresh"
      @edit="openEditDialog"
      @remove="handleRemove"
      @update:open-action-menu-id="openActionMenuId = $event"
    />

    <ContestChallengeEditorDialog
      :key="`${dialogMode}:${existingChallengeIds.join(',')}`"
      :open="dialogOpen"
      :mode="dialogMode"
      :contest-mode="contestMode"
      :challenge-options="dialogChallengeOptions"
      :awd-challenge-options="awdChallengeCatalog"
      :awd-challenge-page="awdChallengePage"
      :awd-challenge-page-size="awdChallengePageSize"
      :awd-challenge-total="awdChallengeTotal"
      :awd-challenge-keyword="awdChallengeFilters.keyword"
      :awd-challenge-service-type="awdChallengeFilters.serviceType"
      :awd-challenge-deployment-mode="awdChallengeFilters.deploymentMode"
      :awd-challenge-readiness="awdChallengeFilters.readinessStatus"
      :awd-challenge-load-error="awdChallengeLoadError"
      :existing-challenge-ids="existingChallengeIds"
      :draft="editingChallenge"
      :loading-challenge-catalog="loadingChallengeCatalog"
      :loading-awd-challenge-catalog="loadingAwdChallengeCatalog"
      :saving="saving"
      @update:open="dialogOpen = $event"
      @update-awd-challenge-keyword="setAwdChallengeKeyword"
      @update-awd-challenge-service-type="setAwdChallengeServiceType"
      @update-awd-challenge-deployment-mode="setAwdChallengeDeploymentMode"
      @update-awd-challenge-readiness="setAwdChallengeReadiness"
      @change-awd-challenge-page="changeAwdChallengePage"
      @refresh-awd-challenge-catalog="refreshAwdChallengeCatalog"
      @save="handleSave"
    />
  </section>
</template>

<style scoped>
.studio-orchestration {
  display: flex;
  flex-direction: column;
  gap: var(--space-section-gap);
  background: transparent;
  padding: var(--space-6) var(--space-8);
}
</style>
