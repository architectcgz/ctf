<template>
  <TeacherInstanceManagementPage
    :classes="classes"
    :dashboard-route="dashboardRoute"
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
    :instances="instances"
    @retry="initialize"
    @update-class-name="updateFilter('className', $event)"
    @update-keyword="updateFilter('keyword', $event)"
    @update-student-no="updateFilter('studentNo', $event)"
    @change-page="handlePageChange"
    @destroy="handleDestroy"
  />
</template>

<script setup lang="ts">
import {
  TeacherInstanceManagementPage,
  useInstanceManagementPage,
} from '@/features/teacher/instances'

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
  dashboardRoute,
  totalCount,
  runningCount,
  expiringSoonCount,
  instances,
  initialize,
  updateFilter,
  handlePageChange,
  handleDestroy,
} = useInstanceManagementPage()
</script>
