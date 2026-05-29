import { onMounted, ref } from 'vue'

import type { AdminCheatDetectionData } from '@/api/contracts'
import { getCheatDetection } from '@/api/admin/platform'
import { buildPlatformAuditLogRoute } from './platformOverviewRoutes'

type CheatQuickAction = {
  title: string
  description: string
  actionLabel: string
  route: ReturnType<typeof buildPlatformAuditLogRoute>
}

export function useCheatDetectionPage() {
  const loading = ref(false)
  const error = ref('')
  const riskData = ref<AdminCheatDetectionData | null>(null)

  const quickActions: ReadonlyArray<CheatQuickAction> = [
    {
      title: '查看提交记录',
      description: '直接打开审计日志中的 submit 动作，复核高频提交账号。',
      actionLabel: '提交审计',
      route: buildPlatformAuditLogRoute({ action: 'submit' }),
    },
    {
      title: '查看登录记录',
      description: '回看 login 日志，继续确认共享 IP 或短时集中登录。',
      actionLabel: '登录审计',
      route: buildPlatformAuditLogRoute({ action: 'login' }),
    },
  ] as const

  async function loadRiskData() {
    loading.value = true
    error.value = ''
    try {
      riskData.value = await getCheatDetection()
    } catch (err) {
      console.error(err)
      error.value = '加载作弊检测结果失败，请稍后重试。'
    } finally {
      loading.value = false
    }
  }

  function formatDateTime(value: string): string {
    return new Date(value).toLocaleString('zh-CN')
  }

  onMounted(() => {
    void loadRiskData()
  })

  return {
    riskData,
    loading,
    error,
    auditLogRoute: buildPlatformAuditLogRoute(),
    buildAuditRoute: buildPlatformAuditLogRoute,
    quickActions,
    loadRiskData,
    formatDateTime,
  }
}
