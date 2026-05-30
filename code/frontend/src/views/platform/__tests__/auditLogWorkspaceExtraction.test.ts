import { describe, expect, it } from 'vitest'

import auditLogSource from '@/pages/platform/AuditLogRoutePage.vue?raw'
import auditActorDetailModalSource from '@/features/audit-log/ui/AuditActorDetailModal.vue?raw'
import auditLogHeroPanelSource from '@/features/audit-log/ui/AuditLogHeroPanel.vue?raw'

describe('AuditLog workspace extraction', () => {
  it('应将操作流水工作区抽到独立平台组件', () => {
    expect(auditLogSource).toContain("from '@/features/audit-log'")
    expect(auditLogSource).toContain('AuditLogDirectoryPanel')
    expect(auditLogSource).toContain('<AuditLogDirectoryPanel')
  })

  it('应将执行人详情弹窗抽到独立平台组件', () => {
    expect(auditLogSource).toContain("from '@/features/audit-log'")
    expect(auditLogSource).toContain('AuditActorDetailModal')
    expect(auditLogSource).toContain('<AuditActorDetailModal')
    expect(auditLogSource).not.toContain('<AdminSurfaceModal')
    expect(auditActorDetailModalSource).toContain('<AdminSurfaceModal')
    expect(auditActorDetailModalSource).toContain('执行人详情')
  })

  it('应将 hero 与审计摘要抽到独立平台组件', () => {
    expect(auditLogSource).toContain("from '@/features/audit-log'")
    expect(auditLogSource).toContain('AuditLogHeroPanel')
    expect(auditLogSource).toContain('<AuditLogHeroPanel')
    expect(auditLogHeroPanelSource).toContain('<div class="workspace-overline">Audit Log</div>')
    expect(auditLogHeroPanelSource).toContain('同步日志')
    expect(auditLogHeroPanelSource).toContain('本页已加载的日志条数')
  })
})
