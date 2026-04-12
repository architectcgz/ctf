import { describe, expect, it } from 'vitest'

import studentAnalysisPageSource from '@/components/teacher/class-management/StudentAnalysisPage.vue?raw'

describe('student analysis panel extraction', () => {
  it('StudentAnalysisPage 应复用单个 StudentInsightPanel 挂载，而不是在每个 tab 内复制同一组 props', () => {
    const insightPanelMounts = studentAnalysisPageSource.match(/<StudentInsightPanel/g) ?? []

    expect(insightPanelMounts).toHaveLength(1)
    expect(studentAnalysisPageSource).toContain(':active-section="tab.key"')
    expect(studentAnalysisPageSource).not.toContain('active-section="overview"')
    expect(studentAnalysisPageSource).not.toContain('active-section="recommendations"')
    expect(studentAnalysisPageSource).not.toContain('active-section="writeups"')
    expect(studentAnalysisPageSource).not.toContain('active-section="evidence"')
    expect(studentAnalysisPageSource).not.toContain('active-section="timeline"')
  })
})
