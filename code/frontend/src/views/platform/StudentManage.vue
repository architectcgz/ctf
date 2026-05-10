<script setup lang="ts">
import StudentManageHeroPanel from '@/components/platform/student/StudentManageHeroPanel.vue'
import StudentManageWorkspacePanel from '@/components/platform/student/StudentManageWorkspacePanel.vue'
import { usePlatformStudentManagementPage } from '@/features/platform-users'

const {
  classes,
  loadingClasses,
  keyword,
  classFilter,
  total,
  loading,
  error,
  hasActiveFilters,
  rows,
  page,
  totalPages,
  activeStudents,
  assignedClassCount,
  initialize,
  handleKeywordChange,
  handleClassFilterChange,
  resetFilters,
  handlePageChange,
  openStudent,
} = usePlatformStudentManagementPage()
</script>

<template>
  <div class="workspace-shell journal-shell journal-shell-admin journal-hero admin-student-manage-shell">
    <main class="content-pane">
        <StudentManageHeroPanel
          :total="total"
          :active-students="activeStudents"
          :assigned-class-count="assignedClassCount"
          @refresh="void initialize()"
        />

        <StudentManageWorkspacePanel
          :classes="classes"
          :loading="loading"
          :loading-classes="loadingClasses"
          :error="error"
          :keyword="keyword"
          :class-filter="classFilter"
          :total="total"
          :has-active-filters="hasActiveFilters"
          :rows="rows"
          :page="page"
          :total-pages="totalPages"
          @update:keyword="handleKeywordChange"
          @change:class-filter="handleClassFilterChange"
          @reset-filters="resetFilters"
          @change-page="handlePageChange"
          @open-student="openStudent"
        />
    </main>
  </div>
</template>

<style scoped>
.admin-student-manage-shell {
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
}
</style>
