<script setup lang="ts">
import { AwdReviewIndexWorkspace } from '@/widgets/awd-review-workspace'
import { useAwdReviewIndexPage } from '@/features/awd-review-workspace'

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
  openHome,
  openContest,
  contestStatusLabel,
} = useAwdReviewIndexPage('teacher')
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
    @open-dashboard="openHome"
    @refresh="loadContests"
    @reload="loadContests"
    @open-contest="openContest"
    @change-page="changePage"
    @update-status-filter="filters.status = $event"
    @update-keyword-filter="filters.keyword = $event"
  />
</template>
