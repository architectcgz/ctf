<script setup lang="ts">
import { TeacherAWDReviewIndexWorkspace } from '@/widgets/teacher-awd-review'
import { useTeacherAwdReviewIndex } from '@/features/teacher-awd-review'

const {
  loading,
  error,
  contests,
  total,
  page,
  totalPages,
  filters,
  hasContests,
  statusOptions,
  contestSummary,
  loadContests,
  changePage,
  openDashboard,
  openContest,
  contestStatusLabel,
} = useTeacherAwdReviewIndex()
</script>

<template>
  <TeacherAWDReviewIndexWorkspace
    :loading="loading"
    :error="error"
    :contests="contests"
    :total="total"
    :page="page"
    :total-pages="totalPages"
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
    @change-page="changePage"
    @update-status-filter="filters.status = $event"
    @update-keyword-filter="filters.keyword = $event"
  />
</template>
