<template>
  <section class="teacher-route-root">
    <ClassManagementPage
      :classes="classes"
      :total="total"
      :page="page"
      :page-size="pageSize"
      :loading="loading"
      :error="error"
      :dashboard-route="dashboardRoute"
      :build-class-route="buildClassRoute"
      @retry="loadClasses"
      @change-page="handlePageChange"
      @open-report-export="openClassReportDialog"
    />
    <ClassReportExportDialog
      v-model="reportDialogVisible"
      :default-class-name="defaultReportClassName"
    />
  </section>
</template>

<script setup lang="ts">
import { ClassReportExportDialog } from '@/features/teaching/class-report-export'
import {
  ClassManagementPage,
  useClassManagementPage,
} from '@/features/teacher/class-management'

const {
  classes,
  total,
  page,
  pageSize,
  loading,
  error,
  reportDialogVisible,
  defaultReportClassName,
  loadClasses,
  handlePageChange,
  buildClassRoute,
  dashboardRoute,
  openClassReportDialog,
} = useClassManagementPage()
</script>
