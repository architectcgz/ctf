import { describe, expect, it } from 'vitest'

import auditLogSource from '../AuditLog.vue?raw'
import contestOrchestrationSource from '@/components/admin/contest/ContestOrchestrationPage.vue?raw'
import userGovernanceSource from '@/components/admin/user/UserGovernancePage.vue?raw'

describe('admin management surface alignment', () => {
  it('audit log should soften table and empty-state borders on dark surfaces', () => {
    expect(auditLogSource).toMatch(/--audit-table-border:\s*color-mix\(in srgb,\s*var\(--journal-border\) 74%, transparent\);/)
    expect(auditLogSource).toMatch(/--audit-row-divider:\s*color-mix\(in srgb,\s*var\(--journal-border\) 62%, transparent\);/)
    expect(auditLogSource).toContain('class="audit-empty-state"')
    expect(auditLogSource).toContain('border-[var(--audit-table-border)]')
    expect(auditLogSource).toContain('divide-y divide-[var(--audit-row-divider)]')
  })

  it('user governance should soften control and table shell borders', () => {
    expect(userGovernanceSource).toMatch(/--admin-control-border:\s*color-mix\(in srgb,\s*var\(--journal-border\) 76%, transparent\);/)
    expect(userGovernanceSource).toMatch(/--user-table-border:\s*color-mix\(in srgb,\s*var\(--journal-border\) 72%, transparent\);/)
    expect(userGovernanceSource).toMatch(/--user-row-divider:\s*color-mix\(in srgb,\s*var\(--journal-border\) 58%, transparent\);/)
    expect(userGovernanceSource).toMatch(/\.admin-btn-ghost\s*\{[\s\S]*border:\s*1px solid var\(--admin-control-border\);/s)
    expect(userGovernanceSource).toMatch(/\.admin-input\s*\{[\s\S]*border:\s*1px solid var\(--admin-control-border\);/s)
    expect(userGovernanceSource).toMatch(/\.user-table-shell\s*\{[\s\S]*border:\s*1px solid var\(--user-table-border\);/s)
    expect(userGovernanceSource).toMatch(/\.user-table-row\s*\{[\s\S]*border-top:\s*1px solid var\(--user-row-divider\);/s)
  })

  it('contest orchestration should soften control and empty-state borders', () => {
    expect(contestOrchestrationSource).toMatch(/--admin-control-border:\s*color-mix\(in srgb,\s*var\(--journal-border\) 76%, transparent\);/)
    expect(contestOrchestrationSource).toContain('class="contest-empty-state"')
    expect(contestOrchestrationSource).toMatch(/\.admin-btn-ghost\s*\{[\s\S]*border:\s*1px solid var\(--admin-control-border\);/s)
    expect(contestOrchestrationSource).toMatch(/\.admin-input\s*\{[\s\S]*border:\s*1px solid var\(--admin-control-border\);/s)
    expect(contestOrchestrationSource).toMatch(
      /\.contest-empty-state\s*\{[\s\S]*border-top-color:\s*color-mix\(in srgb,\s*var\(--journal-border\) 68%, transparent\);[\s\S]*border-bottom-color:\s*color-mix\(in srgb,\s*var\(--journal-border\) 68%, transparent\);/s,
    )
  })
})
