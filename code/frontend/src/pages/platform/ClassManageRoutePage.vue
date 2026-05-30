<template>
  <div class="workspace-shell journal-shell journal-shell-admin journal-hero admin-class-manage-shell">
    <main class="content-pane">
      <ClassManageHeroPanel
        :total="total"
        :total-students="totalStudents"
        @refresh="void loadClasses()"
      />

      <ClassManageWorkspacePanel
        :loading="loading"
        :has-classes="list.length > 0"
        :rows="rows"
        :page="page"
        :total-pages="totalPages"
        :total="total"
        :error="error"
        :build-class-route="buildClassRoute"
        @change-page="handlePageChange"
      />
    </main>
  </div>
</template>

<style scoped>
.admin-class-manage-shell {
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
}
</style>

<script setup lang="ts">
import {
  ClassManageHeroPanel,
  ClassManageWorkspacePanel,
  usePlatformClassManagementPage,
} from '@/features/platform/class-management'

const {
  list,
  total,
  page,
  totalPages,
  loading,
  error,
  totalStudents,
  rows,
  loadClasses,
  handlePageChange,
  buildClassRoute,
} = usePlatformClassManagementPage()
</script>
