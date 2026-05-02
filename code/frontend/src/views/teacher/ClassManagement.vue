<script setup lang="ts">
import ClassManagementPage from '@/components/teacher/class-management/ClassManagementPage.vue'
import TeacherClassReportExportDialog from '@/components/teacher/reports/TeacherClassReportExportDialog.vue'
import { useTeacherClassManagementPage } from '@/features/teacher-class-management'

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
} = useTeacherClassManagementPage()
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
    <TeacherClassReportExportDialog
      v-model="reportDialogVisible"
      :default-class-name="defaultReportClassName"
    />
  </section>
</template>
