<script setup lang="ts">
import { AwdReviewIndexWorkspace } from '@/widgets/awd-review-workspace'
import { useAwdReviewIndex } from '@/features/awd-review-workspace'

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
} = useAwdReviewIndex()
</script>

<template>
  <AwdReviewIndexWorkspace
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
