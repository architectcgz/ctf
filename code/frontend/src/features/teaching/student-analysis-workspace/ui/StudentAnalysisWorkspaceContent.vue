<template>
  <section
    :id="activeWorkspaceTabMeta.panelId"
    class="tab-panel section student-analysis-workspace-content"
    :class="{
      active: true,
      'student-analysis-overview-panel': activeWorkspaceTabMeta.key === 'overview',
    }"
    role="tabpanel"
    :aria-labelledby="activeWorkspaceTabMeta.buttonId"
    aria-hidden="false"
  >
    <template v-if="activeWorkspaceTabMeta.key === 'overview'">
      <StudentAnalysisOverviewHeroPanel
        :selected-student="selectedStudent"
        :progress="progress"
        :solved-rate="solvedRate"
        :weak-dimensions="weakDimensions"
        @open-class-students="emit('openClassStudents')"
        @open-report-export="emit('openReportExport')"
        @open-review-archive="emit('openReviewArchive')"
        @export-review-archive="emit('exportReviewArchive')"
      />
    </template>

    <StudentInsightPanel
      :active-section="activeWorkspaceTab"
      :student="selectedStudent"
      :progress="progress"
      :profile="skillProfile"
      :recommendations="recommendations"
      :timeline="timeline"
      :evidence="evidence"
      :attack-sessions="attackSessions"
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
      :loading="loadingDetails"
      empty-text="请先选择一名学生。"
      @open-challenge="emit('openChallenge', $event)"
      @open-manual-review="emit('openManualReview', $event)"
      @moderate-writeup="emit('moderateWriteup', $event)"
      @review-manual-review="emit('reviewManualReview', $event)"
      @change-writeup-page="emit('changeWriteupPage', $event)"
      @update-review-workspace-filters="emit('updateReviewWorkspaceFilters', $event)"
    />
  </section>
</template>

<style scoped>
:deep(.section-card) {
  padding: var(--space-3-5) var(--space-1) var(--space-3);
  border: 0;
  border-top: 1px solid color-mix(in srgb, var(--teacher-divider) 90%, transparent);
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

:deep(.section-card__header) {
  margin-bottom: var(--space-4);
  border-bottom: 1px dashed color-mix(in srgb, var(--teacher-divider) 86%, transparent);
  padding-bottom: var(--space-3);
}

:deep(.section-card__body) {
  padding-left: 0;
}
</style>

<script setup lang="ts">
import { computed } from 'vue'

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
import StudentAnalysisOverviewHeroPanel from './StudentAnalysisOverviewHeroPanel.vue'
import StudentInsightPanel from './StudentInsightPanel.vue'
import { findStudentAnalysisWorkspaceTab } from './studentAnalysisWorkspaceTabs'

const props = defineProps<{
  selectedStudent: StudentDirectoryItem | null
  loadingDetails: boolean
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
  openClassStudents: []
  openReportExport: []
  openReviewArchive: []
  exportReviewArchive: []
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

const activeWorkspaceTabMeta = computed(() =>
  findStudentAnalysisWorkspaceTab(props.activeWorkspaceTab)
)
</script>
