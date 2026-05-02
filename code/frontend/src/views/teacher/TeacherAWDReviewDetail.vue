<script setup lang="ts">
import { TeacherAWDReviewWorkspace } from '@/widgets/teacher-awd-review'
import { useTeacherAwdReviewDetail } from '@/features/teacher-awd-review'

const {
  router,
  polling,
  loading,
  error,
  review,
  exporting,
  activeContestTitle,
  activeSummaryTitle,
  summaryStats,
  timelineRounds,
  selectedRoundNumber,
  selectedRound,
  selectedTeam,
  selectedTeamServices,
  selectedTeamAttacks,
  selectedTeamTraffic,
  canExportReport,
  loadReview,
  setRound,
  openTeam,
  closeTeam,
  contestStatusLabel,
  formatServiceRef,
  exportArchive,
  exportReport,
} = useTeacherAwdReviewDetail()

function openReviewIndex(): void {
  router.push({ name: 'TeacherAWDReviewIndex' })
}
</script>

<template>
  <TeacherAWDReviewWorkspace
    :polling="polling"
    :loading="loading"
    :error="error"
    :review="review"
    :exporting="exporting"
    :active-contest-title="activeContestTitle"
    :active-summary-title="activeSummaryTitle"
    :summary-stats="summaryStats"
    :timeline-rounds="timelineRounds"
    :selected-round-number="selectedRoundNumber"
    :selected-round="selectedRound"
    :selected-team="selectedTeam"
    :selected-team-services="selectedTeamServices"
    :selected-team-attacks="selectedTeamAttacks"
    :selected-team-traffic="selectedTeamTraffic"
    :can-export-report="canExportReport"
    :contest-status-label="contestStatusLabel"
    :format-service-ref="formatServiceRef"
    @open-index="openReviewIndex"
    @export-archive="exportArchive"
    @export-report="exportReport"
    @load-review="loadReview"
    @set-round="setRound"
    @open-team="openTeam"
    @close-team="closeTeam"
  />
</template>
