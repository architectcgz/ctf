import { computed, ref } from 'vue'

import {
  getStudentRecommendations,
  getStudentProgress,
  getStudentSkillProfile,
  getClasses,
  getClassStudents,
} from '@/api/teacher'
import type {
  MyProgressData,
  RecommendationItem,
  SkillProfileData,
  TeacherClassItem,
  TeacherStudentItem,
} from '@/api/contracts'
import { useAuthStore } from '@/stores/auth'
import { getWeakDimensions } from '@/utils/skillProfile'

export function useTeacherWorkspace() {
  const authStore = useAuthStore()

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

  const selectedClass = computed(
    () => classes.value.find((item) => item.name === selectedClassName.value) ?? null
  )
  const selectedStudent = computed(
    () => students.value.find((item) => item.id === selectedStudentId.value) ?? null
  )
  const solvedRate = computed(() => {
    if (!progress.value?.total_challenges) return 0
    return Math.round(
      ((progress.value.solved_challenges ?? 0) / progress.value.total_challenges) * 100
    )
  })
  const weakDimensions = computed(() => getWeakDimensions(skillProfile.value))

  async function initialize(): Promise<void> {
    loadingClasses.value = true
    error.value = null

    try {
      classes.value = await getClasses()
      const preferredClass = authStore.user?.class_name || classes.value[0]?.name || ''
      selectedClassName.value = preferredClass

      if (preferredClass) {
        await loadStudents(preferredClass)
      } else {
        students.value = []
        clearStudentDetails()
      }
    } catch (err) {
      console.error('加载教师工作区失败:', err)
      error.value = '加载教师数据失败，请稍后重试'
      classes.value = []
      students.value = []
      clearStudentDetails()
    } finally {
      loadingClasses.value = false
    }
  }

  async function loadStudents(className: string): Promise<void> {
    selectedClassName.value = className
    loadingStudents.value = true
    error.value = null

    try {
      students.value = await getClassStudents(className)
      const fallbackStudentId = students.value[0]?.id || ''
      selectedStudentId.value = students.value.some((item) => item.id === selectedStudentId.value)
        ? selectedStudentId.value
        : fallbackStudentId

      if (selectedStudentId.value) {
        await loadStudentDetails(selectedStudentId.value)
      } else {
        clearStudentDetails()
      }
    } catch (err) {
      console.error('加载班级学员失败:', err)
      error.value = '加载班级学员失败，请稍后重试'
      students.value = []
      clearStudentDetails()
    } finally {
      loadingStudents.value = false
    }
  }

  async function loadStudentDetails(studentId: string): Promise<void> {
    if (!studentId) {
      clearStudentDetails()
      return
    }

    selectedStudentId.value = studentId
    loadingDetails.value = true
    error.value = null

    try {
      const [nextProgress, nextProfile, nextRecommendations] = await Promise.all([
        getStudentProgress(studentId),
        getStudentSkillProfile(studentId),
        getStudentRecommendations(studentId),
      ])

      progress.value = nextProgress
      skillProfile.value = nextProfile
      recommendations.value = nextRecommendations
    } catch (err) {
      console.error('加载学员详情失败:', err)
      error.value = '加载学员详情失败，请稍后重试'
      clearStudentDetails()
    } finally {
      loadingDetails.value = false
    }
  }

  function clearStudentDetails(): void {
    progress.value = null
    skillProfile.value = null
    recommendations.value = []
  }

  return {
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
  }
}
