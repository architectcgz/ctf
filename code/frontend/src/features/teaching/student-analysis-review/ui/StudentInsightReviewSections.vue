<template>
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
    :loading="writeupPaginationLoading"
    @open-challenge="emit('openChallenge', $event)"
    @open-manual-review="emit('openManualReview', $event)"
    @moderate-writeup="emit('moderateWriteup', $event)"
    @review-manual-review="emit('reviewManualReview', $event)"
    @change-writeup-page="emit('changeWriteupPage', $event)"
  />

  <StudentInsightManualReviewSection
    v-if="activeSection === 'manual-review'"
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
</template>

<script setup lang="ts">
import type {
  AttackSessionQuery,
  AttackSessionResponseData,
  ManualReviewSubmissionDetailData,
  ManualReviewSubmissionItemData,
  StudentEvidenceData,
  WriteupSubmissionItemData,
} from '@/api/contracts'
import type { StudentInsightSection } from './studentInsightShared'
import StudentInsightAttackSessionsSection from './StudentInsightAttackSessionsSection.vue'
import StudentInsightManualReviewSection from './StudentInsightManualReviewSection.vue'
import StudentInsightWriteupsSection from './StudentInsightWriteupsSection.vue'

const props = defineProps<{
  activeSection?: StudentInsightSection
  attackSessions: AttackSessionResponseData | null
  evidence: StudentEvidenceData | null
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

function isSectionVisible(section: Exclude<StudentInsightSection, 'all' | 'overview' | 'recommendations' | 'training-records'>): boolean {
  return !props.activeSection || props.activeSection === 'all' || props.activeSection === section
}
</script>
