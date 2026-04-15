<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import {
  getClasses,
  getClassReview,
  getClassStudents,
  getClassSummary,
  getClassTrend,
} from '@/api/teacher'
import type {
  TeacherClassItem,
  TeacherClassReviewData,
  TeacherClassSummaryData,
  TeacherClassTrendData,
  TeacherStudentItem,
} from '@/api/contracts'
import TeacherDashboardPage from '@/components/teacher/dashboard/TeacherDashboardPage.vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const classes = ref<TeacherClassItem[]>([])
const students = ref<TeacherStudentItem[]>([])
const review = ref<TeacherClassReviewData | null>(null)
const summary = ref<TeacherClassSummaryData | null>(null)
const trend = ref<TeacherClassTrendData | null>(null)
const selectedClassName = ref('')
const error = ref<string | null>(null)

const selectedClass = computed(
  () => classes.value.find((item) => item.name === selectedClassName.value) ?? null
)

async function initialize(): Promise<void> {
  error.value = null

  try {
    classes.value = await getClasses()
    const preferredClass = authStore.user?.class_name || classes.value[0]?.name || ''
    selectedClassName.value = preferredClass

    if (!preferredClass) {
      students.value = []
      review.value = null
      summary.value = null
      trend.value = null
      return
    }

    const [nextStudents, nextReview, nextSummary, nextTrend] = await Promise.all([
      getClassStudents(preferredClass),
      getClassReview(preferredClass),
      getClassSummary(preferredClass),
      getClassTrend(preferredClass),
    ])
    students.value = nextStudents
    review.value = nextReview
    summary.value = nextSummary
    trend.value = nextTrend
  } catch (err) {
    console.error('加载教师概览失败:', err)
    error.value = '加载教师概览失败，请稍后重试'
    classes.value = []
    students.value = []
    review.value = null
    summary.value = null
    trend.value = null
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
    :review="review"
    :summary="summary"
    :trend="trend"
    :error="error"
    @retry="initialize"
    @open-class-management="router.push({ name: 'ClassManagement' })"
  />
</template>
