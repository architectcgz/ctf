<template>
  <div class="student-insight-shell teacher-surface space-y-6">
    <AppEmpty
      v-if="!student && !loading"
      title="尚未选择学员"
      :description="emptyText || '请先选择学员。'"
      icon="GraduationCap"
    />

    <template v-else>
      <StudentInsightPrimarySections
        :profile="profile"
        :recommendations="recommendations"
        :recommendations-loading="loading"
        :timeline="timeline"
        :loading="loading"
        :active-section="activeSection"
        @open-challenge="emit('openChallenge', $event)"
      />

      <StudentInsightReviewSections
        v-if="student"
        :active-section="activeSection"
        :attack-sessions="attackSessions"
        :evidence="evidence"
        :review-challenge-options="reviewChallengeOptions"
        :review-workspace-loading="reviewWorkspaceLoading"
        :review-workspace-query="reviewWorkspaceQuery"
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
        @update-review-workspace-filters="emit('updateReviewWorkspaceFilters', $event)"
      />
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
import {
  StudentInsightReviewSections,
  type StudentInsightSection,
} from '@/features/teaching/student-analysis-review'
import AppEmpty from '@/shared/ui/common/AppEmpty.vue'
import StudentInsightPrimarySections from './StudentInsightPrimarySections.vue'

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
</script>
