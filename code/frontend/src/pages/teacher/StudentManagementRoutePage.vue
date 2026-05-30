<template>
  <section class="teacher-route-root">
    <StudentManagementPage
      :class-management-route="classManagementRoute"
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
      :build-student-route="buildStudentRoute"
      @retry="initialize"
      @open-report-export="openClassReportDialog"
      @update-search-query="updateSearchQuery"
      @update-student-no-query="updateStudentNoQuery"
      @select-class="selectClass"
      @change-page="handlePageChange"
    />
    <ClassReportExportDialog
      v-model="reportDialogVisible"
      :default-class-name="selectedClassName"
    />
  </section>
</template>

<script setup lang="ts">
import { ClassReportExportDialog } from '@/features/teaching/class-report-export'
import {
  StudentManagementPage,
  useStudentManagementPage,
} from '@/features/teacher/student-management'

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
  classManagementRoute,
  reportDialogVisible,
  initialize,
  openClassReportDialog,
  updateSearchQuery,
  updateStudentNoQuery,
  selectClass,
  handlePageChange,
  buildStudentRoute,
} = useStudentManagementPage()
</script>
