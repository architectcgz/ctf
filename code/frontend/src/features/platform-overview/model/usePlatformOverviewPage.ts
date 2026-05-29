import { onMounted, ref } from 'vue'

import { getDashboard } from '@/api/admin/platform'
import type { AdminDashboardData } from '@/api/contracts'
import { reportFrontendError } from '@/utils/reportFrontendError'
import { buildPlatformAuditLogRoute, platformCheatDetectionRoute } from './platformOverviewRoutes'

export function usePlatformOverviewPage() {
  const loading = ref(false)
  const error = ref<string | null>(null)
  const dashboard = ref<AdminDashboardData | null>(null)

  async function loadDashboard(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      dashboard.value = await getDashboard()
    } catch (err) {
      reportFrontendError('加载系统概览失败:', err)
      error.value = '加载系统概览失败，请稍后重试'
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    void loadDashboard()
  })

  return {
    loading,
    error,
    dashboard,
    loadDashboard,
    auditLogRoute: buildPlatformAuditLogRoute(),
    cheatDetectionRoute: platformCheatDetectionRoute,
  }
}
