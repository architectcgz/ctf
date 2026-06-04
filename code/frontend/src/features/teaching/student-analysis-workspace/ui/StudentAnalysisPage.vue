<template>
  <div
    class="workspace-shell workspace-shell--plain teacher-management-shell teacher-surface student-analysis-shell journal-eyebrow-text flex min-h-full flex-1 flex-col"
  >
    <StudentAnalysisWorkspaceTabs
      :active-workspace-tab="props.activeWorkspaceTab"
      @select-workspace-tab="emit('selectWorkspaceTab', $event)"
    />

    <main class="content-pane">
      <div v-if="error" class="workspace-alert" role="alert" aria-live="polite">
        <div class="workspace-alert-title">
          学员分析加载失败
        </div>
        <div class="workspace-alert-copy">
          {{ error }}
        </div>
        <div class="workspace-alert-actions">
          <button
            type="button"
            class="quick-action quick-action--compact"
            @click="emit('retry')"
          >
            <span>重试加载</span>
            <span>→</span>
          </button>
        </div>
      </div>

      <StudentAnalysisWorkspaceContent
        :selected-student="selectedStudent"
        :loading-details="loadingDetails"
        :progress="progress"
        :skill-profile="skillProfile"
        :recommendations="recommendations"
        :timeline="timeline"
        :evidence="evidence"
        :attack-sessions="attackSessions"
        :review-challenge-options="reviewChallengeOptions"
        :review-workspace-loading="reviewWorkspaceLoading"
        :review-workspace-query="reviewWorkspaceQuery"
        :active-workspace-tab="activeWorkspaceTab"
        :writeup-submissions="writeupSubmissions"
        :writeup-page="writeupPage"
        :writeup-total="writeupTotal"
        :writeup-total-pages="writeupTotalPages"
        :writeup-pagination-loading="writeupPaginationLoading"
        :manual-review-submissions="manualReviewSubmissions"
        :active-manual-review="activeManualReview"
        :manual-review-loading="manualReviewLoading"
        :manual-review-saving="manualReviewSaving"
        :solved-rate="solvedRate"
        :weak-dimensions="weakDimensions"
        @open-class-students="emit('openClassStudents')"
        @open-report-export="emit('openReportExport')"
        @open-review-archive="emit('openReviewArchive')"
        @export-review-archive="emit('exportReviewArchive')"
        @open-challenge="emit('openChallenge', $event)"
        @open-manual-review="emit('openManualReview', $event)"
        @moderate-writeup="emit('moderateWriteup', $event)"
        @review-manual-review="emit('reviewManualReview', $event)"
        @change-writeup-page="emit('changeWriteupPage', $event)"
        @update-review-workspace-filters="emit('updateReviewWorkspaceFilters', $event)"
      />
    </main>
  </div>
</template>

<style scoped>
.student-analysis-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  --journal-accent: var(--color-primary);
  --journal-accent-strong: color-mix(in srgb, var(--color-primary-hover) 82%, var(--journal-ink));
  --journal-eyebrow-spacing: 0.15em;
  --journal-eyebrow-color: var(--journal-accent-strong);
  --teacher-card-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --teacher-control-border: color-mix(in srgb, var(--journal-border) 78%, transparent);
  --header-control-border: var(--teacher-control-border);
  --teacher-divider: color-mix(in srgb, var(--journal-border) 86%, transparent);
  --workspace-page: color-mix(in srgb, var(--color-bg-base) 94%, var(--color-bg-surface));
  --workspace-shell-bg: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --workspace-panel: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
  --workspace-faint: color-mix(in srgb, var(--color-text-secondary) 88%, var(--color-bg-base));
  --workspace-brand: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
  --workspace-brand-ink: color-mix(in srgb, var(--color-primary) 74%, var(--journal-ink));
  --workspace-brand-soft: color-mix(in srgb, var(--color-primary) 10%, transparent);
  --workspace-shadow-shell: 0 24px 84px
    color-mix(in srgb, var(--color-shadow-soft) 58%, transparent);
  --workspace-shadow-panel: 0 14px 34px
    color-mix(in srgb, var(--color-shadow-soft) 42%, transparent);
  --workspace-radius-xl: 28px;
  --workspace-radius-lg: 18px;
}

.content-pane {
  display: grid;
  flex: 1 1 auto;
  align-content: start;
  gap: var(--space-section-gap-compact, var(--space-4));
}

.workspace-alert {
  border: 1px solid var(--workspace-line-soft);
  border-radius: var(--workspace-radius-lg);
  background: color-mix(in srgb, var(--workspace-panel) 88%, transparent);
  box-shadow: var(--workspace-shadow-panel);
  padding: var(--space-4) var(--space-4-5);
}

.workspace-alert-title {
  font-size: var(--font-size-0-92);
  font-weight: 700;
  color: var(--journal-ink);
}

.workspace-alert-copy {
  margin-top: var(--space-2);
  font-size: var(--font-size-0-84);
  line-height: 1.65;
  color: var(--journal-muted);
}

.workspace-alert-actions {
  margin-top: var(--space-3);
}

.quick-action {
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2-5);
  min-height: 52px;
  padding: var(--space-3) var(--space-1);
  border: 0;
  border-bottom: 1px solid color-mix(in srgb, var(--teacher-divider) 88%, transparent);
  border-radius: 0;
  background: transparent;
  color: var(--journal-ink);
  cursor: pointer;
  transition:
    border-color 160ms ease,
    background 160ms ease,
    color 160ms ease;
}

.quick-action:hover,
.quick-action:focus-visible {
  border-bottom-color: color-mix(in srgb, var(--workspace-brand) 34%, transparent);
  background: color-mix(in srgb, var(--workspace-brand) 6%, transparent);
  color: var(--workspace-brand-ink);
  outline: none;
}

.quick-action--compact {
  min-height: 42px;
}

@media (max-width: 767px) {
  .top-tabs,
  .content-pane {
    padding-left: var(--space-5);
    padding-right: var(--space-5);
  }

  .top-tabs {
    gap: var(--space-4-5);
  }
}
</style>

<script setup lang="ts">
import type {
  AttackSessionQuery,
  AttackSessionResponseData,
  ManualReviewSubmissionDetailData,
  ManualReviewSubmissionItemData,
  MyProgressData,
  RecommendationItem,
  SkillProfileData,
  StudentDirectoryItem,
  StudentEvidenceData,
  TimelineEvent,
  WriteupSubmissionItemData,
} from '@/api/contracts'
import type { StudentAnalysisWorkspaceTab } from '../model/useStudentAnalysisPage'
import StudentAnalysisWorkspaceContent from './StudentAnalysisWorkspaceContent.vue'
import StudentAnalysisWorkspaceTabs from './StudentAnalysisWorkspaceTabs.vue'

const props = defineProps<{
  selectedStudent: StudentDirectoryItem | null
  loadingDetails: boolean
  error: string | null
  progress: MyProgressData | null
  skillProfile: SkillProfileData | null
  recommendations: RecommendationItem[]
  timeline: TimelineEvent[]
  evidence: StudentEvidenceData | null
  attackSessions: AttackSessionResponseData | null
  reviewChallengeOptions: Array<{ value: string; label: string }>
  reviewWorkspaceLoading: boolean
  reviewWorkspaceQuery: AttackSessionQuery
  activeWorkspaceTab: StudentAnalysisWorkspaceTab
  writeupSubmissions: WriteupSubmissionItemData[]
  writeupPage: number
  writeupTotal: number
  writeupTotalPages: number
  writeupPaginationLoading: boolean
  manualReviewSubmissions: ManualReviewSubmissionItemData[]
  activeManualReview: ManualReviewSubmissionDetailData | null
  manualReviewLoading: boolean
  manualReviewSaving: boolean
  solvedRate: number
  weakDimensions: string[]
}>()

const emit = defineEmits<{
  retry: []
  openClassStudents: []
  openReportExport: []
  openReviewArchive: []
  exportReviewArchive: []
  selectWorkspaceTab: [tab: StudentAnalysisWorkspaceTab]
  openChallenge: [challengeId: string]
  openManualReview: [submissionId: string]
  moderateWriteup: [
    payload: { submissionId: string; action: 'recommend' | 'unrecommend' | 'hide' | 'restore' },
  ]
  reviewManualReview: [
    payload: {
      submissionId: string
      reviewStatus: 'approved' | 'rejected'
      reviewComment?: string
    },
  ]
  changeWriteupPage: [page: number]
  updateReviewWorkspaceFilters: [payload: Partial<AttackSessionQuery>]
}>()
</script>
