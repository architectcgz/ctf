import { describe, expect, it } from 'vitest'

import auditLogSource from '@/views/platform/AuditLog.vue?raw'

describe('AuditLog workspace extraction', () => {
  it('应将操作流水工作区抽到独立平台组件', () => {
    expect(auditLogSource).toContain(
      "import AuditLogDirectoryPanel from '@/components/platform/audit/AuditLogDirectoryPanel.vue'"
    )
    expect(auditLogSource).toContain('<AuditLogDirectoryPanel')
  })
})
