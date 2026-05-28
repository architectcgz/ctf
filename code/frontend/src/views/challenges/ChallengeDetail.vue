<template>
  <section class="journal-shell journal-shell-user journal-hero workspace-shell min-h-full">
    <div
      v-if="loading"
      class="flex items-center justify-center py-12"
    >
      <div
        class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]"
      />
    </div>

    <div
      v-else-if="challengeLoadState"
      class="challenge-detail-state"
    >
      <AppEmpty
        :icon="challengeLoadState.icon"
        :title="challengeLoadState.title"
        :description="challengeLoadState.description"
      >
        <template #action>
          <div class="challenge-detail-state__actions">
            <button
              type="button"
              class="ui-btn ui-btn--secondary"
              @click="goBackToChallengeList"
            >
              返回题目列表
            </button>
            <button
              v-if="challengeLoadState.retryable"
              type="button"
              class="ui-btn ui-btn--primary"
              @click="retryChallengeLoad"
            >
              重新加载
            </button>
          </div>
        </template>
      </AppEmpty>
    </div>

    <ChallengeWorkspaceShell
      v-else-if="challenge"
      :workspace-tabs="workspaceTabs"
      :active-workspace-tab="activeWorkspaceTab"
      :set-tab-button-ref="setTabButtonRef"
      :handle-workspace-tab-keydown="handleWorkspaceTabKeydown"
      :challenge="challenge"
      :challenge-solved="Boolean(challenge.is_solved)"
      :sanitized-description="sanitizedDescription"
      :score-rail-probe-message="scoreRailProbeMessage"
      :is-hint-expanded="isHintExpanded"
      :solutions-loading="solutionsLoading"
      :recommended-solution-count="recommendedSolutions.length"
      :community-solution-count="communitySolutions.length"
      :active-solution-tab="activeSolutionTab"
      :displayed-solution-cards="displayedSolutionCards"
      :active-solution="activeSolution"
      :sanitized-active-solution-content="sanitizedActiveSolutionContent"
      :format-writeup-time="formatWriteupTime"
      :set-solution-tab-button-ref="setSolutionTabButtonRef"
      :handle-solution-tab-keydown="handleSolutionTabKeydown"
      :submission-records-loading="submissionRecordsLoading"
      :submission-records="submissionRecords"
      :paginated-submission-records="paginatedSubmissionRecords"
      :submission-record-page="submissionRecordPage"
      :submission-record-total="submissionRecordTotal"
      :submission-record-total-pages="submissionRecordTotalPages"
      :format-submission-time="formatSubmissionTime"
      :submission-record-message="submissionRecordMessage"
      :submission-status-text="submissionStatusText"
      :my-writeup="myWriteup"
      :submission-loading="submissionLoading"
      :submission-saving="submissionSaving"
      :writeup-title="writeupTitle"
      :writeup-content="writeupContent"
      :submission-status-label="submissionStatusLabel"
      :need-target="needTarget"
      :submit-panel-title="submitPanelTitle"
      :submit-panel-copy="submitPanelCopy"
      :submit-field-label="submitFieldLabel"
      :submit-input-class="submitInputClass"
      :submit-placeholder="submitPlaceholder"
      :submitting="submitting"
      :flag-input="flagInput"
      :submit-result="submitResult"
      :instance="instance"
      :instance-sharing="challenge.instance_sharing ?? 'per_user'"
      :instance-loading="instanceLoading"
      :instance-creating="instanceCreating"
      :instance-opening="instanceOpening"
      :instance-extending="instanceExtending"
      :instance-destroying="instanceDestroying"
      @select-tab="selectWorkspaceTab"
      @download-attachment="downloadAttachment"
      @toggle-hint="toggleHint"
      @score-rail-probe="handleScoreRailProbe"
      @select-solution-tab="selectSolutionTab"
      @update:selected-solution-id="selectedSolutionId = $event"
      @change-submission-record-page="changeSubmissionRecordPage"
      @update:writeup-title="writeupTitle = $event"
      @update:writeup-content="writeupContent = $event"
      @save-writeup="saveWriteup"
      @update:flag-input="flagInput = $event"
      @submit-flag="submitFlagHandler"
      @start-instance="startInstance"
      @open-instance="openInstance"
      @extend-instance="extendChallengeInstance"
      @destroy-instance="destroyChallengeInstance"
    />
  </section>
</template>

<script setup lang="ts">
import AppEmpty from '@/components/common/AppEmpty.vue'
import { ChallengeWorkspaceShell, useChallengeDetailPage } from '@/features/challenge-detail'
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
  handleWorkspaceTabKeydown,
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
  setTabButtonRef,
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
</script>

<style scoped>
@font-face {
  font-family: 'IBM Plex Sans';
  font-style: normal;
  font-weight: 500;
  font-display: swap;
  src: url('/fonts/ibm-plex-sans-500.woff2') format('woff2');
}

@font-face {
  font-family: 'IBM Plex Sans';
  font-style: normal;
  font-weight: 600;
  font-display: swap;
  src: url('/fonts/ibm-plex-sans-600.woff2') format('woff2');
}

@font-face {
  font-family: 'IBM Plex Sans';
  font-style: normal;
  font-weight: 700;
  font-display: swap;
  src: url('/fonts/ibm-plex-sans-700.woff2') format('woff2');
}

@font-face {
  font-family: 'IBM Plex Mono';
  font-style: normal;
  font-weight: 500;
  font-display: swap;
  src: url('/fonts/ibm-plex-mono-500.woff2') format('woff2');
}

@font-face {
  font-family: 'IBM Plex Mono';
  font-style: normal;
  font-weight: 600;
  font-display: swap;
  src: url('/fonts/ibm-plex-mono-600.woff2') format('woff2');
}

@font-face {
  font-family: 'IBM Plex Mono';
  font-style: normal;
  font-weight: 700;
  font-display: swap;
  src: url('/fonts/ibm-plex-mono-700.woff2') format('woff2');
}

.journal-shell {
  --bg-page: color-mix(in srgb, var(--color-bg-base) 94%, var(--color-bg-surface));
  --bg-shell: var(--journal-surface);
  --bg-panel: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  --bg-muted: color-mix(in srgb, var(--journal-surface-subtle) 90%, var(--color-bg-base));
  --line-soft: var(--journal-border);
  --line-strong: color-mix(in srgb, var(--journal-border) 92%, var(--color-border-default));
  --text-main: var(--journal-ink);
  --text-subtle: var(--journal-muted);
  --text-faint: color-mix(in srgb, var(--journal-muted) 84%, var(--color-bg-base));
  --brand: var(--journal-accent);
  --brand-soft: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  --brand-soft-strong: color-mix(in srgb, var(--brand) 14%, transparent);
  --brand-ink: var(--journal-accent-strong);
  --success: var(--color-success);
  --warning: var(--color-warning);
  --danger: var(--color-danger);
  --shadow-shell: var(--journal-shell-hero-shadow, 0 22px 50px var(--color-shadow-soft));
  --radius-xl: 28px;
  --radius-lg: 18px;
  --font-sans: var(--font-family-sans);
  --font-mono: var(--font-family-mono);
  --journal-faint: var(--text-faint);
  --journal-accent-soft: var(--brand-soft);
  --journal-line-soft: var(--line-soft);
  --journal-line-strong: var(--line-strong);
  --journal-shadow: var(--shadow-shell);
  --journal-success-ink: color-mix(in srgb, var(--color-success) 80%, var(--journal-ink));
  --journal-success-soft: color-mix(in srgb, var(--color-success) 12%, transparent);
  --journal-warning-ink: color-mix(in srgb, var(--color-warning) 88%, var(--journal-ink));
  --journal-warning-soft: color-mix(in srgb, var(--color-warning) 12%, transparent);
  --journal-danger-ink: color-mix(in srgb, var(--color-danger) 82%, var(--journal-ink));
  --journal-danger-soft: color-mix(in srgb, var(--color-danger) 10%, transparent);
}

.workspace-shell {
  --workspace-shell-border: var(--journal-line-soft);
  --workspace-shell-page: var(--bg-page);
  --workspace-shell-bg: var(--bg-shell);
  --workspace-brand: var(--brand);
  --workspace-brand-ink: var(--brand-ink);
  --workspace-brand-soft: var(--brand-soft);
  --workspace-faint: var(--text-faint);
  --workspace-shadow-shell: var(--journal-shadow);
  --workspace-radius-xl: var(--radius-xl);
  --workspace-font-sans: var(--font-sans);
  --workspace-tabs-panel-gap: var(--space-2);
  --workspace-panel-padding-top: var(--workspace-tabs-panel-gap);
  min-height: max(100%, calc(100vh - 5rem));
  flex: 1 1 auto;
  color: var(--text-main);
}

.workspace-shell,
.workspace-shell button,
.workspace-shell input,
.workspace-shell textarea {
  font-family: var(--font-sans);
}

.workspace-shell code,
.workspace-shell pre,
.workspace-shell .flag-input,
.workspace-shell .score-value {
  font-family: var(--font-mono) !important;
}

.challenge-detail-state {
  padding: var(--space-7);
}

.challenge-detail-state__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: var(--space-3);
}





.section-title:not(.workspace-tab-heading__title) {
  margin: var(--space-2-5) 0 0;
  font-size: var(--font-size-20);
  line-height: 1.2;
  color: var(--text-main);
}


























































@media (max-width: 767px) {
  .workspace-shell {
    min-height: 100%;
  }
}

@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation: none !important;
    transition-duration: 0.01ms !important;
  }
}

:global([data-theme='dark']) .workspace-shell {
  --workspace-shell-radial-strength: 14%;
  --workspace-shell-radial-size: 24rem;
  --workspace-shell-top-strength: 97%;
}

</style>
