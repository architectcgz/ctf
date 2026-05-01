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
      v-else-if="challenge"
      class="detail-content"
    >
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
          @click="selectWorkspaceTab(tab.id)"
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
            :build-meta-pill-style="buildMetaPillStyle"
            :get-category-label="getCategoryLabel"
            :get-category-color="getCategoryColor"
            :get-difficulty-label="getDifficultyLabel"
            :get-difficulty-color="getDifficultyColor"
            :is-hint-expanded="isHintExpanded"
            @download-attachment="downloadAttachment"
            @toggle-hint="toggleHint"
            @score-rail-probe="handleScoreRailProbe"
          />

          <ChallengeSolutionsPanel
            v-else-if="activeWorkspaceTab === 'solution'"
            :challenge-solved="Boolean(challenge?.is_solved)"
            :recommended-solution-count="recommendedSolutions.length"
            :community-solution-count="communitySolutions.length"
            :active-solution-tab="activeSolutionTab"
            :displayed-solution-cards="displayedSolutionCards"
            :active-solution="activeSolution"
            :sanitized-active-solution-content="sanitizedActiveSolutionContent"
            :format-writeup-time="formatWriteupTime"
            :set-solution-tab-button-ref="setSolutionTabButtonRef"
            :handle-solution-tab-keydown="handleSolutionTabKeydown"
            @select-tab="selectSolutionTab"
            @select-solution="selectedSolutionId = $event"
          />

          <ChallengeSubmissionRecordsPanel
            v-else-if="activeWorkspaceTab === 'records'"
            :submission-records="submissionRecords"
            :paginated-submission-records="paginatedSubmissionRecords"
            :submission-record-page="submissionRecordPage"
            :submission-record-total="submissionRecordTotal"
            :submission-record-total-pages="submissionRecordTotalPages"
            :format-submission-time="formatSubmissionTime"
            :submission-record-message="submissionRecordMessage"
            :submission-status-text="submissionStatusText"
            @change-page="changeSubmissionRecordPage"
          />

          <ChallengeWriteupPanel
            v-else
            :challenge-solved="Boolean(challenge?.is_solved)"
            :my-writeup="myWriteup"
            :submission-loading="submissionLoading"
            :submission-saving="submissionSaving"
            :writeup-title="writeupTitle"
            :writeup-content="writeupContent"
            :format-writeup-time="formatWriteupTime"
            :submission-status-label="submissionStatusLabel"
            @update:writeup-title="writeupTitle = $event"
            @update:writeup-content="writeupContent = $event"
            @save="saveWriteup"
          />
        </main>

        <ChallengeActionAside
          v-if="activeWorkspaceTab === 'question'"
          :need-target="needTarget"
          :challenge-solved="Boolean(challenge?.is_solved)"
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
          @update:flag-input="flagInput = $event"
          @submit-flag="submitFlagHandler"
          @start-instance="startInstance"
          @open-instance="openInstance"
          @extend-instance="extendChallengeInstance"
          @destroy-instance="destroyChallengeInstance"
        />
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import ChallengeActionAside from '@/components/challenge/ChallengeActionAside.vue'
import ChallengeQuestionPanel from '@/components/challenge/ChallengeQuestionPanel.vue'
import ChallengeSolutionsPanel from '@/components/challenge/ChallengeSolutionsPanel.vue'
import ChallengeSubmissionRecordsPanel from '@/components/challenge/ChallengeSubmissionRecordsPanel.vue'
import ChallengeWriteupPanel from '@/components/challenge/ChallengeWriteupPanel.vue'
import { useChallengeDetailPage } from '@/features/challenge-detail'
const {
  activeSolution,
  activeSolutionTab,
  activeWorkspaceTab,
  buildMetaPillStyle,
  challenge,
  changeSubmissionRecordPage,
  displayedSolutionCards,
  destroyChallengeInstance,
  downloadAttachment,
  extendChallengeInstance,
  flagInput,
  formatSubmissionTime,
  formatWriteupTime,
  getCategoryColor,
  getCategoryLabel,
  getDifficultyColor,
  getDifficultyLabel,
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
  --challenge-tone-web: color-mix(in srgb, var(--color-primary) 82%, var(--journal-ink));
  --challenge-tone-pwn: color-mix(in srgb, var(--color-danger) 78%, var(--journal-ink));
  --challenge-tone-reverse: color-mix(in srgb, var(--journal-accent) 74%, var(--journal-ink));
  --challenge-tone-crypto: color-mix(in srgb, var(--color-warning) 82%, var(--journal-ink));
  --challenge-tone-misc: color-mix(in srgb, var(--color-success) 76%, var(--journal-ink));
  --challenge-tone-forensics: color-mix(in srgb, var(--color-primary) 62%, var(--journal-ink));
  --challenge-tone-beginner: var(--challenge-tone-misc);
  --challenge-tone-easy: var(--challenge-tone-web);
  --challenge-tone-medium: var(--challenge-tone-crypto);
  --challenge-tone-hard: var(--challenge-tone-pwn);
  --challenge-tone-insane: var(--challenge-tone-reverse);
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
  padding: var(--space-7);
}

.tool-pane {
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: var(--space-7);
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


















.section-title:not(.workspace-tab-heading__title) {
  margin: var(--space-2-5) 0 0;
  font-size: var(--font-size-20);
  line-height: 1.2;
  color: var(--text-main);
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
