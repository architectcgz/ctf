<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import {
  getClasses,
  getClassStudents,
  getStudentProgress,
  getStudentRecommendations,
  getStudentSkillProfile,
} from '@/api/teacher'
import type {
  MyProgressData,
  RecommendationItem,
  SkillProfileData,
  TeacherClassItem,
  TeacherStudentItem,
} from '@/api/contracts'
import StudentAnalysisPage from '@/components/teacher/class-management/StudentAnalysisPage.vue'
import { getWeakDimensions } from '@/utils/skillProfile'

const route = useRoute()
const router = useRouter()

const classes = ref<TeacherClassItem[]>([])
const students = ref<TeacherStudentItem[]>([])
const selectedClassName = ref('')
const selectedStudentId = ref('')

const loadingClasses = ref(false)
const loadingStudents = ref(false)
const loadingDetails = ref(false)
const error = ref<string | null>(null)

const progress = ref<MyProgressData | null>(null)
const skillProfile = ref<SkillProfileData | null>(null)
const recommendations = ref<RecommendationItem[]>([])

const selectedStudent = computed(() => students.value.find((item) => item.id === selectedStudentId.value) ?? null)
const solvedRate = computed(() => {
  if (!progress.value?.total_challenges) return 0
  return Math.round(((progress.value.solved_challenges ?? 0) / progress.value.total_challenges) * 100)
})
const weakDimensions = computed(() => getWeakDimensions(skillProfile.value))

function classNameFromRoute(): string {
  return decodeURIComponent(String(route.params.className || ''))
}

function studentIdFromRoute(): string {
  return String(route.params.studentId || '')
}

async function loadClasses(): Promise<void> {
  loadingClasses.value = true
  try {
    classes.value = await getClasses()
  } finally {
    loadingClasses.value = false
  }
}

async function loadStudents(className = classNameFromRoute()): Promise<void> {
  if (!className) {
    selectedClassName.value = ''
    students.value = []
    return
  }

  loadingStudents.value = true
  selectedClassName.value = className

  try {
    students.value = await getClassStudents(className)
  } finally {
    loadingStudents.value = false
  }
}

async function loadStudentDetails(studentId = studentIdFromRoute()): Promise<void> {
  if (!studentId) {
    progress.value = null
    skillProfile.value = null
    recommendations.value = []
    selectedStudentId.value = ''
    return
  }

  loadingDetails.value = true
  selectedStudentId.value = studentId

  try {
    const [nextProgress, nextProfile, nextRecommendations] = await Promise.all([
      getStudentProgress(studentId),
      getStudentSkillProfile(studentId),
      getStudentRecommendations(studentId),
    ])

    progress.value = nextProgress
    skillProfile.value = nextProfile
    recommendations.value = nextRecommendations
  } finally {
    loadingDetails.value = false
  }
}

async function initialize(): Promise<void> {
  error.value = null

  try {
    await loadClasses()
    await loadStudents()
    await loadStudentDetails()
  } catch (err) {
    console.error('加载学员分析失败:', err)
    error.value = '加载学员分析失败，请稍后重试'
  }
}

function selectClass(className: string): void {
  router.push({
    name: 'TeacherClassStudents',
    params: { className },
  })
}

function selectStudent(studentId: string): void {
  router.push({
    name: 'TeacherStudentAnalysis',
    params: {
      className: selectedClassName.value,
      studentId,
    },
  })
}

function openChallenge(challengeId: string): void {
  router.push(`/challenges/${challengeId}`)
}

watch(
  () => [route.params.className, route.params.studentId],
  () => {
    initialize()
  },
)

onMounted(() => {
  initialize()
})
</script>

<template>
  <StudentAnalysisPage
    :classes="classes"
    :students="students"
    :selected-class-name="selectedClassName"
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
    @open-class-students="router.push({ name: 'TeacherClassStudents', params: { className: selectedClassName } })"
    @open-report-export="router.push({ name: 'ReportExport' })"
    @select-class="selectClass"
    @select-student="selectStudent"
    @open-challenge="openChallenge"
  />
</template>
