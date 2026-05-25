import { computed, ref } from 'vue'

import {
  getClassStudents,
  getStudentProgress,
  getStudentRecommendations,
  getStudentSkillProfile,
  getStudentTimeline,
} from '@/api/teaching'
import type {
  MyProgressData,
  RecommendationItem,
  RecommendationWeakDimension,
  SkillProfileData,
  TeacherStudentItem,
  TimelineEvent,
} from '@/api/contracts'
import { getWeakDimensionLabels } from '@/utils/skillProfile'

interface UseStudentAnalysisDataStateOptions {
  classNameFromRoute: () => string
  studentIdFromRoute: () => string
}

export function useStudentAnalysisDataState(options: UseStudentAnalysisDataStateOptions) {
  const { classNameFromRoute, studentIdFromRoute } = options

  const students = ref<TeacherStudentItem[]>([])
  const selectedClassName = ref('')
  const selectedStudentId = ref('')
  const loadingDetails = ref(false)
  const progress = ref<MyProgressData | null>(null)
  const skillProfile = ref<SkillProfileData | null>(null)
  const recommendations = ref<RecommendationItem[]>([])
  const weakDimensionAdvice = ref<RecommendationWeakDimension[]>([])
  const timeline = ref<TimelineEvent[]>([])

  const selectedStudent = computed(
    () => students.value.find((item) => item.id === selectedStudentId.value) ?? null
  )
  const solvedRate = computed(() => {
    if (!progress.value?.total_challenges) return 0
    return Math.round(
      ((progress.value.solved_challenges ?? 0) / progress.value.total_challenges) * 100
    )
  })
  const weakDimensions = computed(() => getWeakDimensionLabels(weakDimensionAdvice.value))

  async function loadStudents(className = classNameFromRoute()): Promise<void> {
    if (!className) {
      selectedClassName.value = ''
      students.value = []
      return
    }

    selectedClassName.value = className
    students.value = await getClassStudents(className)
  }

  async function loadStudentDetails(studentId = studentIdFromRoute()): Promise<void> {
    if (!studentId) {
      progress.value = null
      skillProfile.value = null
      recommendations.value = []
      weakDimensionAdvice.value = []
      timeline.value = []
      selectedStudentId.value = ''
      return
    }

    loadingDetails.value = true
    selectedStudentId.value = studentId

    try {
      const [nextProgress, nextProfile, nextRecommendations, nextTimeline] = await Promise.all([
        getStudentProgress(studentId),
        getStudentSkillProfile(studentId),
        getStudentRecommendations(studentId),
        getStudentTimeline(studentId),
      ])

      progress.value = nextProgress
      skillProfile.value = nextProfile
      recommendations.value = nextRecommendations.challenges
      weakDimensionAdvice.value = nextRecommendations.weak_dimensions
      timeline.value = nextTimeline
    } finally {
      loadingDetails.value = false
    }
  }

  return {
    selectedClassName,
    selectedStudentId,
    selectedStudent,
    loadingDetails,
    progress,
    skillProfile,
    recommendations,
    timeline,
    solvedRate,
    weakDimensions,
    loadStudents,
    loadStudentDetails,
  }
}
