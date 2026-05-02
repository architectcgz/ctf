import { describe, expect, it } from 'vitest'

import source from './useAdminNotificationPublisher.ts?raw'

describe('useAdminNotificationPublisher boundary', () => {
  it('应组合发布 support 模块，避免主组合器内联 payload 组装与校验细节', () => {
    expect(source).toContain("from './adminNotificationPublishSupport'")
    expect(source).toContain('buildAdminNotificationPublishPayload(')
    expect(source).toContain('validateNotificationPublishForm(')
    expect(source).toContain('createDefaultNotificationPublishForm(')
    expect(source).not.toContain('function uniqueValues(')
    expect(source).not.toContain('function createDefaultForm(')
  })
})
