import { describe, expect, it } from 'vitest'

import auditLogSource from '@/views/platform/AuditLog.vue?raw'

describe('AuditLog page state extraction', () => {
  it('应将审计日志页面状态与路由同步逻辑抽到独立 composable', () => {
    expect(auditLogSource).toContain(
      "import { useAuditLogPage } from '@/features/audit-log'"
    )
    expect(auditLogSource).toContain('} = useAuditLogPage()')
  })
})
