<script setup lang="ts">
import StudentManagementPage from '@/components/teacher/student-management/StudentManagementPage.vue'
import TeacherClassReportExportDialog from '@/components/teacher/reports/TeacherClassReportExportDialog.vue'
import { useTeacherStudentManagementPage } from '@/features/teacher-student-management'

const {
  classes,
  selectedClassName,
  searchQuery,
  studentNoQuery,
  students,
  filteredTotal,
  totalStudents,
  page,
  totalPages,
  loadingClasses,
  loadingStudents,
  error,
  reportDialogVisible,
  initialize,
  openClassManagement,
  openClassReportDialog,
  updateSearchQuery,
  updateStudentNoQuery,
  selectClass,
  handlePageChange,
  openStudent,
} = useTeacherStudentManagementPage()
</script>

<template>
  <section class="teacher-route-root">
    <StudentManagementPage
      :classes="classes"
      :selected-class-name="selectedClassName"
      :search-query="searchQuery"
      :student-no-query="studentNoQuery"
      :filtered-students="students"
      :filtered-total="filteredTotal"
      :total-students="totalStudents"
      :page="page"
      :total-pages="totalPages"
      :loading-classes="loadingClasses"
      :loading-students="loadingStudents"
      :error="error"
      @retry="initialize"
      @open-class-management="openClassManagement"
      @open-report-export="openClassReportDialog"
      @update-search-query="updateSearchQuery"
      @update-student-no-query="updateStudentNoQuery"
      @select-class="selectClass"
      @change-page="handlePageChange"
      @open-student="openStudent"
    />
    <TeacherClassReportExportDialog
      v-model="reportDialogVisible"
      :default-class-name="selectedClassName"
    />
  </section>
</template>
