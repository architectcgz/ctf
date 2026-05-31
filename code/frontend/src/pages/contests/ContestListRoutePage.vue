<template>
  <ContestListWorkspace
    :loading="loading"
    :visible-contests="visibleContests"
    :total="total"
    :page="page"
    :total-pages="totalPages"
    :summary-metrics="summaryMetrics"
    :load-error-message="loadErrorMessage"
    :status-filter="statusFilter"
    :mode-filter="modeFilter"
    :refresh="refresh"
    :change-page="changePage"
    :update-status-filter="updateStatusFilter"
    :update-mode-filter="updateModeFilter"
    :reset-filters="resetFilters"
    :format-time="formatTime"
    :build-contest-route="buildContestRoute"
  />
</template>

<script setup lang="ts">
import { useContestListPage } from '@/features/contest-detail'
import { ContestListWorkspace } from '@/widgets/contest-list-workspace'

const {
  loading,
  refresh,
  visibleContests,
  total,
  page,
  totalPages,
  changePage,
  summaryMetrics,
  loadErrorMessage,
  statusFilter,
  modeFilter,
  updateStatusFilter,
  updateModeFilter,
  resetFilters,
  formatTime,
  buildContestRoute,
} = useContestListPage()
</script>
