import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import { getRecommendations, getSkillProfile } from '@/api/assessment'
import { getClassStudents, getStudentRecommendations, getStudentSkillProfile } from '@/api/teacher'
import type { RecommendationItem, SkillProfileData, TeacherStudentItem } from '@/api/contracts'
import { useAuthStore } from '@/stores/auth'
import { getWeakDimensions } from '@/utils/skillProfile'

let loadToken = 0

export function useSkillProfilePage() {
  const authStore = useAuthStore()
  const router = useRouter()

  const isTeacher = computed(() => authStore.isTeacher)
  const selectedStudentId = ref('')
  const students = ref<TeacherStudentItem[]>([])

  const loading = ref(false)
  const error = ref<string | null>(null)
  const skillProfile = ref<SkillProfileData | null>(null)

  const loadingRecommendations = ref(false)
  const recommendations = ref<RecommendationItem[]>([])

  const weakDimensions = computed(() => getWeakDimensions(skillProfile.value))
  const radarIndicators = computed(
    () =>
      skillProfile.value?.dimensions.map((dimension) => ({
        name: dimension.name,
        max: 100,
      })) ?? []
  )
  const radarValues = computed(
    () => skillProfile.value?.dimensions.map((dimension) => dimension.value) ?? []
  )

  async function loadStudents() {
    if (!isTeacher.value) {
      students.value = []
      return
    }

    try {
      const className = authStore.user?.class_name
      if (className) {
        students.value = await getClassStudents(className)
      } else {
        students.value = []
      }
    } catch (error) {
      console.error('加载学员列表失败:', error)
      students.value = []
    }
  }

  async function loadSkillProfileData(token: number) {
    loading.value = true
    error.value = null

    try {
      const nextProfile = selectedStudentId.value
        ? await getStudentSkillProfile(selectedStudentId.value)
        : await getSkillProfile()

      if (token !== loadToken) {
        return
      }

      skillProfile.value = nextProfile
    } catch (err) {
      console.error('加载能力画像失败:', err)
      if (token !== loadToken) {
        return
      }
      error.value = '加载能力画像失败，请稍后重试'
      skillProfile.value = null
    } finally {
      if (token === loadToken) {
        loading.value = false
      }
    }
  }

  async function loadRecommendationsData(token: number) {
    loadingRecommendations.value = true

    try {
      const nextRecommendations = selectedStudentId.value
        ? await getStudentRecommendations(selectedStudentId.value)
        : await getRecommendations()

      if (token !== loadToken) {
        return
      }

      recommendations.value = nextRecommendations
    } catch (err) {
      console.error('加载推荐靶场失败:', err)
      if (token !== loadToken) {
        return
      }
      recommendations.value = []
    } finally {
      if (token === loadToken) {
        loadingRecommendations.value = false
      }
    }
  }

  async function loadCurrentData() {
    const token = ++loadToken
    await Promise.all([loadSkillProfileData(token), loadRecommendationsData(token)])
  }

  function goToChallenge(id: string) {
    void router.push(`/challenges/${id}`)
  }

  function goToChallenges() {
    void router.push('/challenges')
  }

  watch(selectedStudentId, () => {
    void loadCurrentData()
  })

  onMounted(() => {
    void loadStudents()
    void loadCurrentData()
  })

  return {
    isTeacher,
    selectedStudentId,
    students,
    loading,
    error,
    skillProfile,
    loadingRecommendations,
    recommendations,
    weakDimensions,
    radarIndicators,
    radarValues,
    loadCurrentData,
    goToChallenge,
    goToChallenges,
  }
}
