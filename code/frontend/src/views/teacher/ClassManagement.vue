<script setup lang="ts">
import { ClassReportExportDialog } from '@/features/teacher-class-report-export'
import {
  ClassManagementPage,
  useClassManagementPage,
} from '@/features/teacher-class-management'

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
  openClass,
  openDashboard,
  openClassReportDialog,
} = useClassManagementPage()
</script>

<template>
  <section class="teacher-route-root">
    <ClassManagementPage
      :classes="classes"
      :total="total"
      :page="page"
      :page-size="pageSize"
      :loading="loading"
      :error="error"
      @retry="loadClasses"
      @change-page="handlePageChange"
      @open-dashboard="openDashboard"
      @open-report-export="openClassReportDialog"
      @open-class="openClass"
    />
    <ClassReportExportDialog
      v-model="reportDialogVisible"
      :default-class-name="defaultReportClassName"
    />
  </section>
</template>
