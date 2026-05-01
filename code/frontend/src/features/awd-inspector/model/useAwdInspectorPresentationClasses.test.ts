import { describe, expect, it } from 'vitest'

import awdInspectorDerivedDataSource from '@/features/awd-inspector/model/useAwdInspectorDerivedData.ts?raw'
import awdInspectorFormattingSource from '@/features/awd-inspector/model/useAwdInspectorFormatting.ts?raw'

const forbiddenClassFragments = [
  'bg-[var(',
  'text-[var(',
  'border-[var(',
]

function expectNoArbitraryThemeClass(source: string, label: string): void {
  forbiddenClassFragments.forEach((fragment) => {
    expect(source, `${label} 不应返回 Tailwind 任意主题类 ${fragment}`).not.toContain(fragment)
  })
}

describe('AWD inspector presentation class helpers', () => {
  it('格式化与派生 helper 应只返回语义类，不应直接拼 Tailwind 任意主题类', () => {
    expectNoArbitraryThemeClass(awdInspectorFormattingSource, 'useAwdInspectorFormatting')
    expectNoArbitraryThemeClass(awdInspectorDerivedDataSource, 'useAwdInspectorDerivedData')
    expect(awdInspectorFormattingSource).toContain('awd-status-pill--success')
    expect(awdInspectorDerivedDataSource).toContain('awd-service-alert--warning')
  })
})
