<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { getClasses, getClassStudents } from '@/api/teacher'
import type { TeacherClassItem, TeacherStudentItem } from '@/api/contracts'
import TeacherDashboardPage from '@/components/teacher/dashboard/TeacherDashboardPage.vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const classes = ref<TeacherClassItem[]>([])
const students = ref<TeacherStudentItem[]>([])
const selectedClassName = ref('')
const error = ref<string | null>(null)

const selectedClass = computed(() => classes.value.find((item) => item.name === selectedClassName.value) ?? null)

async function initialize(): Promise<void> {
  error.value = null

  try {
    classes.value = await getClasses()
    const preferredClass = authStore.user?.class_name || classes.value[0]?.name || ''
    selectedClassName.value = preferredClass

    if (!preferredClass) {
      students.value = []
      return
    }

    students.value = await getClassStudents(preferredClass)
  } catch (err) {
    console.error('加载教师概览失败:', err)
    error.value = '加载教师概览失败，请稍后重试'
    classes.value = []
    students.value = []
    selectedClassName.value = ''
  }
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
    :error="error"
    @retry="initialize"
    @open-class-management="router.push({ name: 'ClassManagement' })"
    @open-report-export="router.push({ name: 'ReportExport' })"
  />
</template>
