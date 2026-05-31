import { onMounted, ref } from 'vue'

import { getCheatDetection } from '@/api/admin/platform'
import type { AdminCheatDetectionData } from '@/api/contracts'
import { reportFrontendError } from '@/utils/reportFrontendError'

export function useCheatDetectionData() {
  const loading = ref(false)
  const error = ref('')
  const riskData = ref<AdminCheatDetectionData | null>(null)

  async function loadRiskData() {
    loading.value = true
    error.value = ''
    try {
      riskData.value = await getCheatDetection()
    } catch (err) {
      reportFrontendError('加载作弊检测结果失败:', err)
      error.value = '加载作弊检测结果失败，请稍后重试。'
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    void loadRiskData()
  })

  return {
    riskData,
    loading,
    error,
    loadRiskData,
  }
}
