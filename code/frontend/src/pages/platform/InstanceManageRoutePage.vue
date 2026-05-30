<template>
  <div class="workspace-shell journal-shell journal-shell-admin journal-hero admin-instance-manage-shell">
    <main class="content-pane">
      <InstanceManageHeroPanel
        :running-count="runningCount"
        :total="totalInstances"
        :warning-count="warningCount"
        :overview-route="overviewRoute"
        @refresh="void loadInstances()"
      />

      <InstanceManageWorkspacePanel
        :loading="loading"
        :has-instances="list.length > 0"
        :rows="pageRows"
        :keyword="keyword"
        :status-filter="statusFilter"
        :page="page"
        :total-pages="totalPages"
        :total="filteredTotal"
        :destroying-id="destroyingId"
        :error="error"
        @update:keyword="setKeyword"
        @change:status-filter="setStatusFilter"
        @reset-filters="resetFilters"
        @destroy-instance="requestDestroyById"
        @change-page="handlePageChange"
      />
    </main>
  </div>
</template>

<style scoped>
.admin-instance-manage-shell {
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
}
</style>

<script setup lang="ts">
import {
  InstanceManageHeroPanel,
  InstanceManageWorkspacePanel,
  usePlatformInstanceManagementPage,
} from '@/features/platform/instance-management'

const {
  list,
  page,
  loading,
  destroyingId,
  error,
  keyword,
  statusFilter,
  totalInstances,
  filteredTotal,
  overviewRoute,
  totalPages,
  pageRows,
  runningCount,
  warningCount,
  loadInstances,
  requestDestroyById,
  handlePageChange,
  setKeyword,
  setStatusFilter,
  resetFilters,
} = usePlatformInstanceManagementPage()
</script>
