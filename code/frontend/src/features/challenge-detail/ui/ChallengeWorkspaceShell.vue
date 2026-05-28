<template>
  <div class="detail-content">
    <div
      class="workspace-tabbar top-tabs"
      role="tablist"
      aria-label="题目页面主切换"
    >
      <button
        v-for="(tab, index) in workspaceTabs"
        :id="`challenge-workspace-tab-${tab.id}`"
        :key="tab.id"
        :ref="(element) => setTabButtonRef(tab.id, element as HTMLButtonElement | null)"
        type="button"
        role="tab"
        class="workspace-tab top-tab"
        :class="{ active: activeWorkspaceTab === tab.id }"
        :aria-selected="activeWorkspaceTab === tab.id"
        :aria-controls="`challenge-workspace-panel-${tab.id}`"
        :tabindex="activeWorkspaceTab === tab.id ? 0 : -1"
        @click="emit('select-tab', tab.id)"
        @keydown="handleWorkspaceTabKeydown($event, index)"
      >
        {{ tab.label }}
      </button>
    </div>

    <div
      class="detail-grid detail-grid--workspace workspace-grid"
      :class="{ 'workspace-grid--single': activeWorkspaceTab !== 'question' }"
    >
      <main class="detail-main content-pane">
        <ChallengeQuestionPanel
          v-if="activeWorkspaceTab === 'question'"
          :challenge="challenge"
          :sanitized-description="sanitizedDescription"
          :score-rail-probe-message="scoreRailProbeMessage"
          :is-hint-expanded="isHintExpanded"
          @download-attachment="emit('download-attachment')"
          @toggle-hint="emit('toggle-hint', $event)"
          @score-rail-probe="emit('score-rail-probe')"
        />

        <ChallengeSolutionsPanel
          v-else-if="activeWorkspaceTab === 'solution'"
          :challenge-solved="challengeSolved"
          :solutions-loading="solutionsLoading"
          :recommended-solution-count="recommendedSolutionCount"
          :community-solution-count="communitySolutionCount"
          :active-solution-tab="activeSolutionTab"
          :displayed-solution-cards="displayedSolutionCards"
          :active-solution="activeSolution"
          :sanitized-active-solution-content="sanitizedActiveSolutionContent"
          :format-writeup-time="formatWriteupTime"
          :set-solution-tab-button-ref="setSolutionTabButtonRef"
          :handle-solution-tab-keydown="handleSolutionTabKeydown"
          @select-tab="emit('select-solution-tab', $event)"
          @select-solution="emit('update:selectedSolutionId', $event)"
        />

        <ChallengeSubmissionRecordsPanel
          v-else-if="activeWorkspaceTab === 'records'"
          :submission-records-loading="submissionRecordsLoading"
          :submission-records="submissionRecords"
          :paginated-submission-records="paginatedSubmissionRecords"
          :submission-record-page="submissionRecordPage"
          :submission-record-total="submissionRecordTotal"
          :submission-record-total-pages="submissionRecordTotalPages"
          :format-submission-time="formatSubmissionTime"
          :submission-record-message="submissionRecordMessage"
          :submission-status-text="submissionStatusText"
          @change-page="emit('change-submission-record-page', $event)"
        />

        <ChallengeWriteupPanel
          v-else
          :challenge-solved="challengeSolved"
          :my-writeup="myWriteup"
          :submission-loading="submissionLoading"
          :submission-saving="submissionSaving"
          :writeup-title="writeupTitle"
          :writeup-content="writeupContent"
          :format-writeup-time="formatWriteupTime"
          :submission-status-label="submissionStatusLabel"
          @update:writeup-title="emit('update:writeupTitle', $event)"
          @update:writeup-content="emit('update:writeupContent', $event)"
          @save="emit('save-writeup', $event)"
        />
      </main>

      <ChallengeActionAside
        v-if="activeWorkspaceTab === 'question'"
        :need-target="needTarget"
        :challenge-solved="challengeSolved"
        :submit-panel-title="submitPanelTitle"
        :submit-panel-copy="submitPanelCopy"
        :submit-field-label="submitFieldLabel"
        :submit-input-class="submitInputClass"
        :submit-placeholder="submitPlaceholder"
        :submitting="submitting"
        :flag-input="flagInput"
        :submit-result="submitResult"
        :instance="instance"
        :instance-sharing="instanceSharing"
        :instance-loading="instanceLoading"
        :instance-creating="instanceCreating"
        :instance-opening="instanceOpening"
        :instance-extending="instanceExtending"
        :instance-destroying="instanceDestroying"
        @update:flag-input="emit('update:flagInput', $event)"
        @submit-flag="emit('submit-flag')"
        @start-instance="emit('start-instance')"
        @open-instance="emit('open-instance')"
        @extend-instance="emit('extend-instance')"
        @destroy-instance="emit('destroy-instance')"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import type {
  ChallengeDetailData,
  InstanceData,
  InstanceSharing,
  SubmissionWriteupData,
  SubmissionWriteupStatus,
} from '@/api/contracts'
import type {
  ChallengeSolutionCard,
  ChallengeSolutionTab,
  ChallengeSubmissionRecordStatus,
} from '../model'
import ChallengeActionAside from '@/components/challenge/ChallengeActionAside.vue'
import ChallengeQuestionPanel from '@/components/challenge/ChallengeQuestionPanel.vue'
import ChallengeWriteupPanel from '@/components/challenge/ChallengeWriteupPanel.vue'
import ChallengeSolutionsPanel from './ChallengeSolutionsPanel.vue'
import ChallengeSubmissionRecordsPanel from './ChallengeSubmissionRecordsPanel.vue'

type ChallengeWorkspaceTab = 'question' | 'solution' | 'records' | 'writeup'

interface SubmitResultState {
  variant: 'success' | 'error' | 'pending'
  message: string
}

interface SubmissionRecordItem {
  id: string
  answer?: string
  status: ChallengeSubmissionRecordStatus
  submittedAt?: string
}

interface Props {
  workspaceTabs: ReadonlyArray<{ id: ChallengeWorkspaceTab; label: string }>
  activeWorkspaceTab: ChallengeWorkspaceTab
  setTabButtonRef: (key: ChallengeWorkspaceTab, element: HTMLButtonElement | null) => void
  handleWorkspaceTabKeydown: (event: KeyboardEvent, index: number) => void
  challenge: ChallengeDetailData
  challengeSolved: boolean
  sanitizedDescription: string
  scoreRailProbeMessage: string
  isHintExpanded: (level: number) => boolean
  solutionsLoading: boolean
  recommendedSolutionCount: number
  communitySolutionCount: number
  activeSolutionTab: ChallengeSolutionTab
  displayedSolutionCards: ChallengeSolutionCard[]
  activeSolution: ChallengeSolutionCard | null
  sanitizedActiveSolutionContent: string
  formatWriteupTime: (value?: string) => string
  setSolutionTabButtonRef: (
    tab: ChallengeSolutionTab,
    element: HTMLButtonElement | null,
  ) => void
  handleSolutionTabKeydown: (event: KeyboardEvent, index: number) => void
  submissionRecordsLoading: boolean
  submissionRecords: SubmissionRecordItem[]
  paginatedSubmissionRecords: SubmissionRecordItem[]
  submissionRecordPage: number
  submissionRecordTotal: number
  submissionRecordTotalPages: number
  formatSubmissionTime: (value?: string) => string
  submissionRecordMessage: (status: ChallengeSubmissionRecordStatus) => string
  submissionStatusText: (status: ChallengeSubmissionRecordStatus) => string
  myWriteup: SubmissionWriteupData | null
  submissionLoading: boolean
  submissionSaving: 'draft' | 'published' | null
  writeupTitle: string
  writeupContent: string
  submissionStatusLabel: (status?: SubmissionWriteupStatus) => string
  needTarget: boolean
  submitPanelTitle: string
  submitPanelCopy: string
  submitFieldLabel: string
  submitInputClass: string
  submitPlaceholder: string
  submitting: boolean
  flagInput: string
  submitResult: SubmitResultState | null
  instance: InstanceData | null
  instanceSharing: InstanceSharing
  instanceLoading: boolean
  instanceCreating: boolean
  instanceOpening: boolean
  instanceExtending: boolean
  instanceDestroying: boolean
}

defineProps<Props>()

const emit = defineEmits<{
  'select-tab': [tab: ChallengeWorkspaceTab]
  'download-attachment': []
  'toggle-hint': [level: number]
  'score-rail-probe': []
  'select-solution-tab': [tab: ChallengeSolutionTab]
  'update:selectedSolutionId': [solutionId: string]
  'change-submission-record-page': [page: number]
  'update:writeupTitle': [value: string]
  'update:writeupContent': [value: string]
  'save-writeup': [status: 'draft' | 'published']
  'update:flagInput': [value: string]
  'submit-flag': []
  'start-instance': []
  'open-instance': []
  'extend-instance': []
  'destroy-instance': []
}>()
</script>

<style scoped>
.detail-content {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  min-height: 0;
}

.detail-grid,
.workspace-grid {
  display: grid;
  flex: 1 1 auto;
  min-height: 0;
  grid-template-columns: minmax(0, 1.34fr) minmax(320px, 0.66fr);
  align-items: stretch;
}

.workspace-grid--single {
  grid-template-columns: minmax(0, 1fr);
}

.detail-main,
.detail-aside,
.content-pane,
.tool-pane {
  min-width: 0;
}

.detail-main,
.content-pane {
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: 0 var(--space-workspace-content-padding, var(--space-7))
    var(--space-workspace-content-padding, var(--space-7));
}

.tool-pane {
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: var(--workspace-tabs-panel-gap, var(--space-2))
    var(--space-workspace-content-padding, var(--space-7))
    var(--space-workspace-content-padding, var(--space-7));
  border-left: 1px solid var(--line-soft);
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--bg-panel) 92%, var(--color-bg-base)),
    color-mix(in srgb, var(--bg-shell) 88%, var(--color-bg-base))
  );
}

.workspace-panel,
.panel {
  display: block;
  min-height: 100%;
  animation: rise 280ms cubic-bezier(0.22, 1, 0.36, 1);
}

@keyframes rise {
  from {
    opacity: 0;
    transform: translateY(16px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 1080px) {
  .detail-grid,
  .workspace-grid {
    flex: initial;
    grid-template-columns: minmax(0, 1fr);
  }

  .tool-pane {
    border-left: 0;
    border-top: 1px solid var(--journal-line-soft);
  }
}

@media (max-width: 760px) {
  .detail-content > .workspace-tabbar,
  .content-pane,
  .tool-pane {
    padding-left: var(--space-4-5);
    padding-right: var(--space-4-5);
  }

  .detail-content > .workspace-tabbar {
    gap: var(--space-5-5);
  }
}
</style>
