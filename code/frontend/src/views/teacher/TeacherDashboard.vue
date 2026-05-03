<script setup lang="ts">
import TeacherDashboardPage from '@/components/teacher/dashboard/TeacherDashboardPage.vue'
import { useTeacherDashboardPage } from '@/features/teacher-dashboard'

const {
  classes,
  students,
  selectedClassName,
  selectedClass,
  review,
  summary,
  trend,
  error,
  initialize,
  openClassManagement,
} = useTeacherDashboardPage()
</script>

<template>
  <TeacherDashboardPage
    :classes="classes"
    :students="students"
    :selected-class-name="selectedClassName"
    :selected-class="selectedClass"
    :review="review"
    :summary="summary"
    :trend="trend"
    :error="error"
    @retry="initialize"
    @open-class-management="openClassManagement"
  />
</template>
