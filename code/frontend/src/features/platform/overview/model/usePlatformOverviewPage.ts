import { buildPlatformAuditLogRoute, platformCheatDetectionRoute } from './platformOverviewRoutes'
import { usePlatformOverviewData } from './usePlatformOverviewData'

export function usePlatformOverviewPage() {
  const overviewData = usePlatformOverviewData()

  return {
    loading: overviewData.loading,
    error: overviewData.error,
    dashboard: overviewData.dashboard,
    loadDashboard: overviewData.loadDashboard,
    auditLogRoute: buildPlatformAuditLogRoute(),
    cheatDetectionRoute: platformCheatDetectionRoute,
  }
}
