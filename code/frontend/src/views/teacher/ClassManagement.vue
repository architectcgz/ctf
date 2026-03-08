<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import ClassManagementPage from '@/components/teacher/class-management/ClassManagementPage.vue'
import { useTeacherWorkspace } from '@/composables/useTeacherWorkspace'

const router = useRouter()
const searchQuery = ref('')

const {
  classes,
  students,
  selectedClassName,
  selectedStudentId,
  selectedStudent,
  loadingClasses,
  loadingStudents,
  loadingDetails,
  error,
  progress,
  skillProfile,
  recommendations,
  initialize,
  loadStudents,
  loadStudentDetails,
} = useTeacherWorkspace()

const filteredStudents = computed(() => {
  const keyword = searchQuery.value.trim().toLowerCase()
  if (!keyword) return students.value
  return students.value.filter((student) => {
    const label = `${student.name || ''} ${student.username}`.toLowerCase()
    return label.includes(keyword)
  })
})

function openChallenge(challengeId: string): void {
  router.push(`/challenges/${challengeId}`)
}

function updateSearchQuery(value: string): void {
  searchQuery.value = value
}

onMounted(() => {
  initialize()
})
</script>

<template>
  <ClassManagementPage
    :classes="classes"
    :selected-class-name="selectedClassName"
    :search-query="searchQuery"
    :filtered-students="filteredStudents"
    :selected-student-id="selectedStudentId"
    :selected-student="selectedStudent"
    :loading-classes="loadingClasses"
    :loading-students="loadingStudents"
    :loading-details="loadingDetails"
    :error="error"
    :progress="progress"
    :skill-profile="skillProfile"
    :recommendations="recommendations"
    @retry="initialize"
    @update-search-query="updateSearchQuery"
    @select-class="loadStudents"
    @select-student="loadStudentDetails"
    @open-challenge="openChallenge"
    @open-dashboard="router.push({ name: 'TeacherDashboard' })"
    @open-report-export="router.push({ name: 'ReportExport' })"
  />
</template>
