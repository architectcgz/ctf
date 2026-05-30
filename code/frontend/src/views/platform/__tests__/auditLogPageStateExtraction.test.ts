import { describe, expect, it } from 'vitest'

import auditLogSource from '@/pages/platform/AuditLogRoutePage.vue?raw'
import auditLogPageSource from '@/features/audit-log/model/useAuditLogPage.ts?raw'

describe('AuditLog page state extraction', () => {
  it('应将审计日志页面状态与路由同步逻辑抽到独立 composable', () => {
    expect(auditLogSource).toContain("from '@/features/audit-log'")
    expect(auditLogSource).toContain('useAuditLogPage')
    expect(auditLogSource).toContain('} = useAuditLogPage()')
    expect(auditLogPageSource).not.toContain("from 'vue-router'")
    expect(auditLogPageSource).toContain('useRouteQueryTransport')
  })
})
