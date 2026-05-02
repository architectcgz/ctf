<script setup lang="ts">
import TeacherInstanceManagementPage from '@/components/teacher/instance-management/TeacherInstanceManagementPage.vue'
import { useTeacherInstanceManagementPage } from '@/features/teacher-instances'

const {
  classes,
  filters,
  page,
  totalPages,
  loadingClasses,
  loadingInstances,
  destroyingId,
  error,
  isAdmin,
  totalCount,
  runningCount,
  expiringSoonCount,
  paginatedInstances,
  initialize,
  updateFilter,
  handlePageChange,
  handleDestroy,
  openDashboard,
} = useTeacherInstanceManagementPage()
</script>

<template>
  <TeacherInstanceManagementPage
    :classes="classes"
    :class-name="filters.className"
    :keyword="filters.keyword"
    :student-no="filters.studentNo"
    :page="page"
    :total-pages="totalPages"
    :loading-classes="loadingClasses"
    :loading-instances="loadingInstances"
    :destroying-id="destroyingId"
    :error="error"
    :is-admin="isAdmin"
    :total-count="totalCount"
    :running-count="runningCount"
    :expiring-soon-count="expiringSoonCount"
    :instances="paginatedInstances"
    @retry="initialize"
    @update-class-name="updateFilter('className', $event)"
    @update-keyword="updateFilter('keyword', $event)"
    @update-student-no="updateFilter('studentNo', $event)"
    @change-page="handlePageChange"
    @destroy="handleDestroy"
    @open-dashboard="openDashboard"
  />
</template>
