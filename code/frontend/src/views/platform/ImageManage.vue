<script setup lang="ts">
import ImageCreateModal from '@/components/platform/images/ImageCreateModal.vue'
import ImageDetailModal from '@/components/platform/images/ImageDetailModal.vue'
import ImageDirectoryPanel from '@/components/platform/images/ImageDirectoryPanel.vue'
import ImageManageHeroPanel from '@/components/platform/images/ImageManageHeroPanel.vue'
import { useImageManagePage } from '@/composables/useImageManagePage'

const {
  activeImage,
  changePage,
  closeDetail,
  creating,
  dialogVisible,
  filteredRows,
  filteredTotal,
  form,
  formatDateTime,
  formatSize,
  getStatusLabel,
  getStatusStyle,
  handleCreate,
  handleDelete,
  hasActiveFilters,
  keyword,
  list,
  loading,
  openDetail,
  page,
  refresh,
  refreshHint,
  resetFilters,
  selectedSortLabel,
  setSort,
  sortOptions,
  statusFilter,
  statusSummary,
  total,
  totalPages,
} = useImageManagePage()
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-admin journal-notes-rail journal-hero flex min-h-full flex-1 flex-col"
  >
    <main class="content-pane">
      <ImageManageHeroPanel
        :loading="loading"
        :refresh-hint="refreshHint"
        :status-summary="statusSummary"
        @refresh="void refresh()"
        @create="dialogVisible = true"
      />

      <!--
        class="image-board workspace-directory-section"
        class="image-list workspace-directory-list"
        class="admin-pagination workspace-directory-pagination"
        PlatformPaginationControls
      -->
      <ImageDirectoryPanel
        :list="list"
        :rows="filteredRows"
        :total="total"
        :filtered-total="filteredTotal"
        :page="page"
        :total-pages="totalPages"
        :loading="loading"
        :keyword="keyword"
        :status-filter="statusFilter"
        :has-active-filters="hasActiveFilters"
        :sort-options="sortOptions"
        :selected-sort-label="selectedSortLabel"
        :get-status-label="getStatusLabel"
        :get-status-style="getStatusStyle"
        :format-date-time="formatDateTime"
        @update:keyword="keyword = $event"
        @update:status-filter="statusFilter = $event"
        @select-sort="setSort"
        @reset-filters="resetFilters"
        @open-detail="openDetail"
        @delete-image="handleDelete"
        @change-page="void changePage($event)"
      />
    </main>

    <ImageDetailModal
      :open="!!activeImage"
      :image="activeImage"
      :format-size="formatSize"
      :format-date-time="formatDateTime"
      :get-status-label="getStatusLabel"
      :get-status-style="getStatusStyle"
      @close="closeDetail"
      @update:open="!$event && closeDetail()"
    />

    <ImageCreateModal
      :open="dialogVisible"
      :creating="creating"
      :form="form"
      @close="dialogVisible = false"
      @update:open="dialogVisible = $event"
      @update:name="form.name = $event"
      @update:tag="form.tag = $event"
      @update:description="form.description = $event"
      @submit="handleCreate"
    />
  </section>
</template>

<style scoped>
.journal-shell {
  --admin-summary-grid-columns: repeat(2, minmax(0, 1fr));
  --admin-control-border: color-mix(in srgb, var(--journal-border) 78%, transparent);
  --journal-divider-border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  --journal-shell-hero-radial-strength: 7%;
  --journal-shell-hero-radial-size: 22rem;
  --journal-shell-hero-end: var(--journal-surface);
  --journal-shell-hero-shadow: 0 22px 50px var(--color-shadow-soft);
  --journal-shell-dark-ink: var(--color-text-primary);
  --journal-shell-dark-accent: var(--color-primary-hover);
  --journal-shell-dark-surface: color-mix(
    in srgb,
    var(--color-bg-surface) 92%,
    var(--color-bg-base)
  );
  --journal-shell-dark-surface-subtle: color-mix(
    in srgb,
    var(--color-bg-surface) 78%,
    var(--color-bg-base)
  );
  --journal-shell-dark-hero-radial-strength: 10%;
  --journal-shell-dark-hero-top: color-mix(
    in srgb,
    var(--journal-surface) 97%,
    var(--color-bg-base)
  );
  --journal-shell-dark-hero-end: color-mix(
    in srgb,
    var(--journal-surface-subtle) 95%,
    var(--color-bg-base)
  );
}
</style>
