import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { getDashboard } from '@/api/admin/platform'
import type { AdminDashboardData } from '@/api/contracts'

export function usePlatformOverviewPage() {
  const router = useRouter()
  const loading = ref(false)
  const error = ref<string | null>(null)
  const dashboard = ref<AdminDashboardData | null>(null)

  async function loadDashboard(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      dashboard.value = await getDashboard()
    } catch (err) {
      console.error('加载系统概览失败:', err)
      error.value = '加载系统概览失败，请稍后重试'
    } finally {
      loading.value = false
    }
  }

  function openAuditLog(): void {
    router.push({ name: 'AuditLog' })
  }

  function openCheatDetection(): void {
    router.push({ name: 'CheatDetection' })
  }

  onMounted(() => {
    void loadDashboard()
  })

  return {
    loading,
    error,
    dashboard,
    loadDashboard,
    openAuditLog,
    openCheatDetection,
  }
}
