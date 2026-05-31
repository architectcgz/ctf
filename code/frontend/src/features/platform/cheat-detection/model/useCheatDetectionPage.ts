import { buildPlatformAuditLogRoute } from './cheatDetectionRoutes'
import { useCheatDetectionData } from './useCheatDetectionData'

type CheatQuickAction = {
  title: string
  description: string
  actionLabel: string
  route: ReturnType<typeof buildPlatformAuditLogRoute>
}

export function useCheatDetectionPage() {
  const cheatDetectionData = useCheatDetectionData()

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

  function formatDateTime(value: string): string {
    return new Date(value).toLocaleString('zh-CN')
  }

  return {
    riskData: cheatDetectionData.riskData,
    loading: cheatDetectionData.loading,
    error: cheatDetectionData.error,
    auditLogRoute: buildPlatformAuditLogRoute(),
    buildAuditRoute: buildPlatformAuditLogRoute,
    quickActions,
    loadRiskData: cheatDetectionData.loadRiskData,
    formatDateTime,
  }
}
