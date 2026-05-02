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
})
