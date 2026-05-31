import { onMounted, ref } from 'vue'

import { getTeacherOverview } from '@/api/teacher'
import type { TeacherOverviewData } from '@/api/contracts'
import { reportFrontendError } from '@/utils/reportFrontendError'

export function useTeacherOverviewData() {
  const overview = ref<TeacherOverviewData | null>(null)
  const error = ref<string | null>(null)

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
    initialize,
  }
}
