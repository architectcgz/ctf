<script setup lang="ts">
import { TeacherAWDReviewIndexWorkspace } from '@/widgets/teacher-awd-review'
import { useTeacherAwdReviewIndex } from '@/features/teacher-awd-review'

const {
  router,
  loading,
  error,
  contests,
  filters,
  hasContests,
  statusOptions,
  contestSummary,
  loadContests,
  openContest,
  contestStatusLabel,
} = useTeacherAwdReviewIndex()

function openDashboard(): void {
  router.push({ name: 'TeacherDashboard' })
}
</script>

<template>
  <TeacherAWDReviewIndexWorkspace
    :loading="loading"
    :error="error"
    :contests="contests"
    :has-contests="hasContests"
    :status-options="statusOptions"
    :contest-summary="contestSummary"
    :status-filter="filters.status"
    :keyword-filter="filters.keyword"
    :contest-status-label="contestStatusLabel"
    @open-dashboard="openDashboard"
    @refresh="loadContests"
    @reload="loadContests"
    @open-contest="openContest"
    @update-status-filter="filters.status = $event"
    @update-keyword-filter="filters.keyword = $event"
  />
</template>
