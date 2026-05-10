<script setup lang="ts">
import InstanceManageHeroPanel from '@/components/platform/instance/InstanceManageHeroPanel.vue'
import InstanceManageWorkspacePanel from '@/components/platform/instance/InstanceManageWorkspacePanel.vue'
import { usePlatformInstanceManagementPage } from '@/features/platform-users'

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
  totalPages,
  pageRows,
  runningCount,
  warningCount,
  loadInstances,
  openOverview,
  openStudent,
  requestDestroyById,
  handlePageChange,
  setKeyword,
  setStatusFilter,
  resetFilters,
} = usePlatformInstanceManagementPage()
</script>

<template>
  <div class="workspace-shell journal-shell journal-shell-admin journal-hero admin-instance-manage-shell">
    <main class="content-pane">
        <InstanceManageHeroPanel
          :running-count="runningCount"
          :total="totalInstances"
          :warning-count="warningCount"
          @back="openOverview"
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
          @open-student="openStudent"
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
