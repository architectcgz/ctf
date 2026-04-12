<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { getClasses } from '@/api/teacher'
import type { TeacherClassItem } from '@/api/contracts'
import ClassManagementPage from '@/components/teacher/class-management/ClassManagementPage.vue'
import { DEFAULT_PAGE_SIZE } from '@/utils/constants'

const router = useRouter()

const classes = ref<TeacherClassItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(DEFAULT_PAGE_SIZE)
const loading = ref(false)
const error = ref<string | null>(null)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / Math.max(pageSize.value, 1))))

async function loadClasses(nextPage = page.value): Promise<void> {
  loading.value = true
  error.value = null

  try {
    const payload = await getClasses({ page: nextPage, page_size: pageSize.value })
    classes.value = payload.list
    total.value = payload.total
    page.value = payload.page
    pageSize.value = payload.page_size
  } catch (err) {
    console.error('加载班级列表失败:', err)
    error.value = '加载班级列表失败，请稍后重试'
    classes.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function handlePageChange(nextPage: number): void {
  const normalizedPage = Math.max(1, Math.floor(nextPage))
  if (normalizedPage === page.value || normalizedPage > totalPages.value) {
    return
  }
  void loadClasses(normalizedPage)
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
    :total="total"
    :page="page"
    :page-size="pageSize"
    :loading="loading"
    :error="error"
    @retry="loadClasses"
    @change-page="handlePageChange"
    @open-dashboard="router.push({ name: 'TeacherDashboard' })"
    @open-report-export="router.push({ name: 'TeacherAWDReviewIndex' })"
    @open-class="openClass"
  />
</template>
