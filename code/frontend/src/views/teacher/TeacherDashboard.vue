<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'

import TeacherDashboardPage from '@/components/teacher/dashboard/TeacherDashboardPage.vue'
import { useTeacherWorkspace } from '@/composables/useTeacherWorkspace'

const router = useRouter()
const {
  classes,
  students,
  selectedClassName,
  selectedStudentId,
  selectedClass,
  selectedStudent,
  loadingClasses,
  loadingStudents,
  loadingDetails,
  error,
  progress,
  skillProfile,
  recommendations,
  solvedRate,
  weakDimensions,
  initialize,
  loadStudents,
  loadStudentDetails,
} = useTeacherWorkspace()

function openChallenge(challengeId: string): void {
  router.push(`/challenges/${challengeId}`)
}

onMounted(() => {
  initialize()
})
</script>

<template>
  <TeacherDashboardPage
    :classes="classes"
    :students="students"
    :selected-class-name="selectedClassName"
    :selected-class="selectedClass"
    :selected-student-id="selectedStudentId"
    :selected-student="selectedStudent"
    :loading-classes="loadingClasses"
    :loading-students="loadingStudents"
    :loading-details="loadingDetails"
    :error="error"
    :progress="progress"
    :skill-profile="skillProfile"
    :recommendations="recommendations"
    :solved-rate="solvedRate"
    :weak-dimensions="weakDimensions"
    @retry="initialize"
    @open-class-management="router.push({ name: 'ClassManagement' })"
    @open-report-export="router.push({ name: 'ReportExport' })"
    @open-challenge="openChallenge"
    @select-class="loadStudents"
    @select-student="loadStudentDetails"
  />
</template>
