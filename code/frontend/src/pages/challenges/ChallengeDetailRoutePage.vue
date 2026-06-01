<template>
  <ChallengeDetailWorkspace
    :active-solution="activeSolution"
    :active-solution-tab="activeSolutionTab"
    :active-workspace-tab="activeWorkspaceTab"
    :challenge="challenge"
    :challenge-load-state="challengeLoadState"
    :displayed-solution-cards="displayedSolutionCards"
    :flag-input="flagInput"
    :format-submission-time="formatSubmissionTime"
    :format-writeup-time="formatWriteupTime"
    :go-back-to-challenge-list="goBackToChallengeList"
    :handle-score-rail-probe="handleScoreRailProbe"
    :handle-solution-tab-keydown="handleSolutionTabKeydown"
    :instance="instance"
    :instance-creating="instanceCreating"
    :instance-destroying="instanceDestroying"
    :instance-extending="instanceExtending"
    :instance-loading="instanceLoading"
    :instance-opening="instanceOpening"
    :is-hint-expanded="isHintExpanded"
    :loading="loading"
    :my-writeup="myWriteup"
    :need-target="needTarget"
    :open-instance="openInstance"
    :community-solutions="communitySolutions"
    :paginated-submission-records="paginatedSubmissionRecords"
    :recommended-solutions="recommendedSolutions"
    :retry-challenge-load="retryChallengeLoad"
    :sanitized-active-solution-content="sanitizedActiveSolutionContent"
    :sanitized-description="sanitizedDescription"
    :score-rail-probe-message="scoreRailProbeMessage"
    :select-solution-tab="selectSolutionTab"
    :select-workspace-tab="selectWorkspaceTab"
    :set-selected-solution-id="setSelectedSolutionId"
    :set-solution-tab-button-ref="setSolutionTabButtonRef"
    :start-instance="startInstance"
    :submission-loading="submissionLoading"
    :submission-record-message="submissionRecordMessage"
    :submission-record-page="submissionRecordPage"
    :submission-records-loading="submissionRecordsLoading"
    :submission-record-total="submissionRecordTotal"
    :submission-record-total-pages="submissionRecordTotalPages"
    :submission-records="submissionRecords"
    :submission-saving="submissionSaving"
    :submission-status-label="submissionStatusLabel"
    :submission-status-text="submissionStatusText"
    :submit-field-label="submitFieldLabel"
    :submit-flag-handler="submitFlagHandler"
    :set-flag-input="setFlagInput"
    :submit-input-class="submitInputClass"
    :submit-panel-copy="submitPanelCopy"
    :submit-panel-title="submitPanelTitle"
    :submit-placeholder="submitPlaceholder"
    :submit-result="submitResult"
    :submitting="submitting"
    :solutions-loading="solutionsLoading"
    :toggle-hint="toggleHint"
    :workspace-tabs="workspaceTabs"
    :writeup-content="writeupContent"
    :set-writeup-content="setWriteupContent"
    :writeup-title="writeupTitle"
    :set-writeup-title="setWriteupTitle"
    :save-writeup="saveWriteup"
    :destroy-challenge-instance="destroyChallengeInstance"
    :extend-challenge-instance="extendChallengeInstance"
    :download-attachment="downloadAttachment"
    :change-submission-record-page="changeSubmissionRecordPage"
  />
</template>

<script setup lang="ts">
import { useChallengeDetailPage } from '@/features/challenge-detail'
import { ChallengeDetailWorkspace } from '@/widgets/challenge-detail-workspace'

const {
  activeSolution,
  activeSolutionTab,
  activeWorkspaceTab,
  challenge,
  challengeLoadState,
  changeSubmissionRecordPage,
  displayedSolutionCards,
  destroyChallengeInstance,
  downloadAttachment,
  extendChallengeInstance,
  flagInput,
  formatSubmissionTime,
  formatWriteupTime,
  goBackToChallengeList,
  handleScoreRailProbe,
  handleSolutionTabKeydown,
  instance,
  instanceCreating,
  instanceDestroying,
  instanceExtending,
  instanceLoading,
  instanceOpening,
  isHintExpanded,
  loading,
  myWriteup,
  needTarget,
  openInstance,
  communitySolutions,
  paginatedSubmissionRecords,
  recommendedSolutions,
  retryChallengeLoad,
  sanitizedActiveSolutionContent,
  sanitizedDescription,
  scoreRailProbeMessage,
  selectSolutionTab,
  selectWorkspaceTab,
  selectedSolutionId,
  setSolutionTabButtonRef,
  startInstance,
  submissionLoading,
  submissionRecordMessage,
  submissionRecordPage,
  submissionRecordsLoading,
  submissionRecordTotal,
  submissionRecordTotalPages,
  submissionRecords,
  submissionSaving,
  submissionStatusLabel,
  submissionStatusText,
  submitFieldLabel,
  submitFlagHandler,
  submitInputClass,
  submitPanelCopy,
  submitPanelTitle,
  submitPlaceholder,
  submitResult,
  submitting,
  solutionsLoading,
  toggleHint,
  workspaceTabs,
  writeupContent,
  writeupTitle,
  saveWriteup,
} = useChallengeDetailPage()

function setSelectedSolutionId(value: string | null): void {
  selectedSolutionId.value = value
}

function setFlagInput(value: string): void {
  flagInput.value = value
}

function setWriteupContent(value: string): void {
  writeupContent.value = value
}

function setWriteupTitle(value: string): void {
  writeupTitle.value = value
}
</script>
