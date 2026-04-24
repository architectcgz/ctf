import { describe, expect, it } from 'vitest'

import auditLogSource from '@/views/platform/AuditLog.vue?raw'
import auditActorDetailModalSource from '@/components/platform/audit/AuditActorDetailModal.vue?raw'
import auditLogHeroPanelSource from '@/components/platform/audit/AuditLogHeroPanel.vue?raw'

describe('AuditLog workspace extraction', () => {
  it('应将操作流水工作区抽到独立平台组件', () => {
    expect(auditLogSource).toContain(
      "import AuditLogDirectoryPanel from '@/components/platform/audit/AuditLogDirectoryPanel.vue'"
    )
    expect(auditLogSource).toContain('<AuditLogDirectoryPanel')
  })

  it('应将执行人详情弹窗抽到独立平台组件', () => {
    expect(auditLogSource).toContain(
      "import AuditActorDetailModal from '@/components/platform/audit/AuditActorDetailModal.vue'"
    )
    expect(auditLogSource).toContain('<AuditActorDetailModal')
    expect(auditLogSource).not.toContain('<AdminSurfaceModal')
    expect(auditActorDetailModalSource).toContain('<AdminSurfaceModal')
    expect(auditActorDetailModalSource).toContain('执行人详情')
  })

  it('应将 hero 与审计摘要抽到独立平台组件', () => {
    expect(auditLogSource).toContain(
      "import AuditLogHeroPanel from '@/components/platform/audit/AuditLogHeroPanel.vue'"
    )
    expect(auditLogSource).toContain('<AuditLogHeroPanel')
    expect(auditLogHeroPanelSource).toContain('<div class="workspace-overline">Audit Log</div>')
    expect(auditLogHeroPanelSource).toContain('同步日志')
    expect(auditLogHeroPanelSource).toContain('本页已加载的日志条数')
  })
})
