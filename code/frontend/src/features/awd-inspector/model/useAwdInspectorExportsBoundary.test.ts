import { describe, expect, it } from 'vitest'

import source from './useAwdInspectorExports.ts?raw'

describe('useAwdInspectorExports boundary', () => {
  it('应组合 export payload builders，避免主模块内联大段导出映射逻辑', () => {
    expect(source).toContain("from './awdInspectorExportPayloads'")
    expect(source).toContain('buildServiceExportRows(')
    expect(source).toContain('buildAttackExportRows(')
    expect(source).toContain('buildReviewPackagePayload(')
  })
})
