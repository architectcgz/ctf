import { describe, expect, it } from 'vitest'

import source from './useAwdInspectorDerivedData.ts?raw'

describe('useAwdInspectorDerivedData boundary', () => {
  it('应组合服务侧 derived builders，避免主组合器内联服务告警聚合', () => {
    expect(source).toContain("from './awdInspectorServiceDerived'")
    expect(source).toContain('buildServiceAlerts(')
    expect(source).toContain('filterServices(')
    expect(source).not.toContain('function getServiceAlertReason(')
    expect(source).not.toContain('function getServiceCheckSourceValue(')
    expect(source).not.toContain('const grouped = new Map<string, AWDServiceAlertView>()')
  })

  it('应组合攻击与流量派生 builders，避免主组合器内联攻击筛选与队伍选项拼装', () => {
    expect(source).toContain("from './awdInspectorAttackDerived'")
    expect(source).toContain('buildAttackTeamOptions(')
    expect(source).toContain('buildAttackSourceOptions(')
    expect(source).toContain('buildTrafficTeamOptions(')
    expect(source).toContain('filterAttacks(')
    expect(source).not.toContain('const entries = new Map<string, string>()')
  })
})
