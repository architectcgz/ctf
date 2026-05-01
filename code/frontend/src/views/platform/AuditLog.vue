<script setup lang="ts">
import AuditActorDetailModal from '@/components/platform/audit/AuditActorDetailModal.vue'
import AuditLogHeroPanel from '@/components/platform/audit/AuditLogHeroPanel.vue'
import AuditLogDirectoryPanel from '@/components/platform/audit/AuditLogDirectoryPanel.vue'
import { useAuditLogPage } from '@/features/audit-log'

const {
  activeActorLog,
  actorDisplayName,
  changePage,
  closeActorDetail,
  detailPreview,
  error,
  filteredRows,
  filters,
  formatDate,
  hasActiveFilters,
  keyword,
  list,
  loadLogs,
  loading,
  openActorDetail,
  page,
  resetFilters,
  resourceDisplayName,
  selectedSortLabel,
  setSort,
  sortOptions,
  total,
  totalPages,
} = useAuditLogPage()
</script>

<template>
  <div class="workspace-shell journal-shell journal-shell-admin journal-hero">
    <div class="workspace-grid">
      <main class="content-pane">
        <div class="audit-log-body">
          <AuditLogHeroPanel
            :current-count="list.length"
            :total="total"
            :total-pages="totalPages"
            :loading="loading"
            @sync="void loadLogs()"
          />

          <AuditLogDirectoryPanel
            :rows="filteredRows"
            :total="total"
            :page="page"
            :total-pages="totalPages"
            :loading="loading"
            :error="error"
            :keyword="keyword"
            :has-active-filters="hasActiveFilters"
            :selected-sort-label="selectedSortLabel"
            :sort-options="sortOptions"
            :action-filter="filters.action"
            :resource-type-filter="filters.resource_type"
            :actor-user-id-filter="filters.actor_user_id"
            :format-date="formatDate"
            :detail-preview="detailPreview"
            :actor-display-name="actorDisplayName"
            @update:keyword="keyword = $event"
            @update:action-filter="filters.action = $event"
            @update:resource-type-filter="filters.resource_type = $event"
            @update:actor-user-id-filter="filters.actor_user_id = $event"
            @select-sort="setSort"
            @reset-filters="void resetFilters()"
            @retry="loadLogs"
            @open-actor-detail="openActorDetail"
            @change-page="void changePage($event)"
          />
        </div>
      </main>
    </div>

    <AuditActorDetailModal
      :open="!!activeActorLog"
      :item="activeActorLog"
      :format-date="formatDate"
      :actor-display-name="actorDisplayName"
      :resource-display-name="resourceDisplayName"
      :detail-preview="detailPreview"
      @close="closeActorDetail"
      @update:open="!$event && closeActorDetail()"
    />
  </div>
</template>

<style scoped>
.audit-log-body {
  display: flex;
  flex-direction: column;
  gap: var(--workspace-directory-page-block-gap, var(--space-5));
}
</style>
