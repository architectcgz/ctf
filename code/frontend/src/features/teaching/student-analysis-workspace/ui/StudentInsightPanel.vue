<script setup lang="ts">
import { computed } from 'vue'

import AppCard from '@/components/common/AppCard.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import type {
  AttackSessionQuery,
  ManualReviewSubmissionDetailData,
  ManualReviewSubmissionItemData,
  MyProgressData,
  RecommendationItem,
  SkillProfileData,
  AttackSessionResponseData,
  StudentEvidenceData,
  WriteupSubmissionItemData,
  StudentDirectoryItem,
  TimelineEvent,
} from '@/api/contracts'
import TrainingTimelinePanel from '@/components/training/TrainingTimelinePanel.vue'
import StudentInsightAttackSessionsSection from './StudentInsightAttackSessionsSection.vue'
import StudentInsightManualReviewSection from './StudentInsightManualReviewSection.vue'
import StudentInsightOverviewSection from './StudentInsightOverviewSection.vue'
import StudentInsightRecommendationsSection from './StudentInsightRecommendationsSection.vue'
import StudentInsightWriteupsSection from './StudentInsightWriteupsSection.vue'
import type { StudentInsightSection } from './studentInsightShared'

const props = defineProps<{
  student: StudentDirectoryItem | null
  progress: MyProgressData | null
  profile: SkillProfileData | null
  recommendations: RecommendationItem[]
  timeline: TimelineEvent[]
  evidence: StudentEvidenceData | null
  attackSessions: AttackSessionResponseData | null
  reviewChallengeOptions: Array<{ value: string; label: string }>
  reviewWorkspaceLoading: boolean
  reviewWorkspaceQuery: AttackSessionQuery
  writeupSubmissions: WriteupSubmissionItemData[]
  writeupPage: number
  writeupTotal: number
  writeupTotalPages: number
  writeupPaginationLoading: boolean
  manualReviewSubmissions: ManualReviewSubmissionItemData[]
  activeManualReview: ManualReviewSubmissionDetailData | null
  manualReviewLoading: boolean
  manualReviewSaving: boolean
  loading: boolean
  emptyText?: string
  activeSection?: StudentInsightSection
}>()

const emit = defineEmits<{
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

const showManualReviewSection = computed(() => props.activeSection === 'manual-review')

function isSectionVisible(section: Exclude<StudentInsightSection, 'all'>): boolean {
  return !props.activeSection || props.activeSection === 'all' || props.activeSection === section
}
</script>

<template>
  <div class="student-insight-shell teacher-surface space-y-6">
    <AppEmpty
      v-if="!student && !loading"
      title="尚未选择学员"
      :description="emptyText || '请先选择学员。'"
      icon="GraduationCap"
    />

    <template v-else>
      <div v-if="loading" class="grid gap-6 lg:grid-cols-[1.15fr_0.85fr]">
        <AppCard variant="panel" accent="neutral">
          <div class="insight-skeleton-line h-6 w-36 animate-pulse rounded" />
          <div class="mt-6 space-y-3">
            <div class="insight-skeleton-block h-16 animate-pulse rounded-xl" />
            <div class="insight-skeleton-block h-16 animate-pulse rounded-xl" />
          </div>
        </AppCard>
        <AppCard variant="panel" accent="neutral">
          <div class="insight-skeleton-block h-[280px] animate-pulse rounded-2xl" />
        </AppCard>
      </div>

      <template v-else-if="student">
        <StudentInsightOverviewSection
          v-if="isSectionVisible('overview')"
          :profile="profile"
        />

        <StudentInsightRecommendationsSection
          v-if="isSectionVisible('recommendations')"
          :recommendations="recommendations"
          @open-challenge="emit('openChallenge', $event)"
        />

        <StudentInsightWriteupsSection
          v-if="isSectionVisible('writeups')"
          :writeup-submissions="writeupSubmissions"
          :writeup-page="writeupPage"
          :writeup-total="writeupTotal"
          :writeup-total-pages="writeupTotalPages"
          :writeup-pagination-loading="writeupPaginationLoading"
          :manual-review-submissions="manualReviewSubmissions"
          :active-manual-review="activeManualReview"
          :manual-review-loading="manualReviewLoading"
          :manual-review-saving="manualReviewSaving"
          @open-challenge="emit('openChallenge', $event)"
          @open-manual-review="emit('openManualReview', $event)"
          @moderate-writeup="emit('moderateWriteup', $event)"
          @review-manual-review="emit('reviewManualReview', $event)"
          @change-writeup-page="emit('changeWriteupPage', $event)"
        />

        <StudentInsightManualReviewSection
          v-if="showManualReviewSection"
          :manual-review-submissions="manualReviewSubmissions"
          :active-manual-review="activeManualReview"
          :manual-review-loading="manualReviewLoading"
          :manual-review-saving="manualReviewSaving"
          @open-manual-review="emit('openManualReview', $event)"
          @review-manual-review="emit('reviewManualReview', $event)"
        />

        <StudentInsightAttackSessionsSection
          v-if="isSectionVisible('evidence')"
          :attack-sessions="attackSessions"
          :evidence="evidence"
          :review-challenge-options="reviewChallengeOptions"
          :review-workspace-loading="reviewWorkspaceLoading"
          :review-workspace-query="reviewWorkspaceQuery"
          @update-review-workspace-filters="emit('updateReviewWorkspaceFilters', $event)"
        />

        <TrainingTimelinePanel v-if="isSectionVisible('timeline')" :timeline="timeline" />
      </template>
    </template>
  </div>
</template>

<style scoped>
.student-insight-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: var(--color-primary);
  --journal-accent-strong: color-mix(in srgb, var(--color-primary-hover) 82%, var(--journal-ink));
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  --teacher-card-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --teacher-divider: color-mix(in srgb, var(--journal-border) 86%, transparent);
}

.insight-skeleton-line,
.insight-skeleton-block {
  background: linear-gradient(
    90deg,
    color-mix(in srgb, var(--journal-border) 78%, transparent),
    color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base))
  );
}

:deep(.section-card) {
  padding: var(--space-4) var(--space-1) var(--space-3);
  border: 0;
  border-top: 1px solid color-mix(in srgb, var(--teacher-divider) 88%, transparent);
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

:deep(.section-card__header) {
  margin-bottom: var(--space-4);
  padding-bottom: var(--space-3);
  border-bottom: 1px dashed color-mix(in srgb, var(--teacher-divider) 86%, transparent);
}

:deep(.section-card__body) {
  padding-left: 0;
}

.insight-tab-section-card :deep(.section-card__header) {
  border-bottom: 0;
}

.insight-tab-section-card.section-card {
  border-top: 0;
}
</style>
