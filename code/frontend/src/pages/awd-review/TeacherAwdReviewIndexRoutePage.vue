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
  homeRoute,
  buildContestRoute,
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
    :dashboard-route="homeRoute"
    :build-contest-route="buildContestRoute"
    :contest-summary="contestSummary"
    :status-filter="filters.status"
    :keyword-filter="filters.keyword"
    :contest-status-label="contestStatusLabel"
    @refresh="loadContests"
    @reload="loadContests"
    @change-page="changePage"
    @update-status-filter="filters.status = $event"
    @update-keyword-filter="filters.keyword = $event"
  />
</template>
