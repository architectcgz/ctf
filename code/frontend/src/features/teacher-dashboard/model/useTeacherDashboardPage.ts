import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { getTeacherOverview } from '@/api/teacher'
import type { TeacherOverviewData } from '@/api/contracts'
import { useAuthStore } from '@/stores/auth'
import { resolveClassManagementRouteName } from '@/utils/classManagementRouting'

export function useTeacherDashboardPage() {
  const router = useRouter()
  const authStore = useAuthStore()

  const overview = ref<TeacherOverviewData | null>(null)
  const error = ref<string | null>(null)

  async function initialize(): Promise<void> {
    error.value = null

    try {
      overview.value = await getTeacherOverview()
    } catch (err) {
      console.error('加载教师概览失败:', err)
      error.value = '加载教师概览失败，请稍后重试'
      overview.value = null
    }
  }

  function openClassManagement(): void {
    router.push({ name: resolveClassManagementRouteName(authStore.user?.role) })
  }

  onMounted(() => {
    void initialize()
  })

  return {
    overview,
    error,
    initialize,
    openClassManagement,
  }
}
