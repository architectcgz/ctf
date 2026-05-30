import { computed, onMounted, ref } from 'vue'

import { getTeacherOverview } from '@/api/teacher'
import type { TeacherOverviewData } from '@/api/contracts'
import { useAuthStore } from '@/stores/auth'
import { reportFrontendError } from '@/utils/reportFrontendError'
import { teacherClassManagementRoute } from './teacherDashboardRoutes'

export function useDashboardPage() {
  const authStore = useAuthStore()

  const overview = ref<TeacherOverviewData | null>(null)
  const error = ref<string | null>(null)
  const classManagementRoute = computed(() => teacherClassManagementRoute(authStore.user?.role))

  async function initialize(): Promise<void> {
    error.value = null

    try {
      overview.value = await getTeacherOverview()
    } catch (err) {
      reportFrontendError('加载教师概览失败:', err)
      error.value = '加载教师概览失败，请稍后重试'
      overview.value = null
    }
  }

  onMounted(() => {
    void initialize()
  })

  return {
    overview,
    error,
    classManagementRoute,
    initialize,
  }
}
