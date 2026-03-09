<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { getClasses } from '@/api/teacher'
import type { TeacherClassItem } from '@/api/contracts'
import ClassManagementPage from '@/components/teacher/class-management/ClassManagementPage.vue'

const router = useRouter()

const classes = ref<TeacherClassItem[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

async function loadClasses(): Promise<void> {
  loading.value = true
  error.value = null

  try {
    classes.value = await getClasses()
  } catch (err) {
    console.error('加载班级列表失败:', err)
    error.value = '加载班级列表失败，请稍后重试'
    classes.value = []
  } finally {
    loading.value = false
  }
}

function openClass(className: string): void {
  router.push({ name: 'TeacherClassStudents', params: { className } })
}

onMounted(() => {
  loadClasses()
})
</script>

<template>
  <ClassManagementPage
    :classes="classes"
    :loading="loading"
    :error="error"
    @retry="loadClasses"
    @open-dashboard="router.push({ name: 'TeacherDashboard' })"
    @open-report-export="router.push({ name: 'ReportExport' })"
    @open-class="openClass"
  />
</template>
