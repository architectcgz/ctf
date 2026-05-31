import { describe, expect, it } from 'vitest'

import interventionPanelSource from '@/features/teaching/student-analysis-review/ui/InterventionPanel.vue?raw'
import userPresentationSource from '@/entities/user/model/presentation.ts?raw'

describe('teacher intervention panel layout', () => {
  it('应采用紧凑可穿透的介入列表布局', () => {
    expect(interventionPanelSource).toContain(
      "from '@/features/teaching/student-analysis-review'"
    )
    expect(interventionPanelSource).toContain('useInterventionRecommendations')
    expect(interventionPanelSource).not.toContain('useTeacherInterventionRecommendations')
    expect(interventionPanelSource).not.toContain("from '@/api/teacher'")
    expect(interventionPanelSource).toContain('intervention-item__header')
    expect(interventionPanelSource).toContain('intervention-item__name-button')
    expect(interventionPanelSource).toContain("@click=\"openStudent(item.student.id)\"")
    expect(interventionPanelSource).toContain('const emit = defineEmits<{')
    expect(interventionPanelSource).toContain('openStudent: [studentId: string]')
    expect(interventionPanelSource).toContain("from '@/entities/user'")
    expect(interventionPanelSource).toContain('getUserDisplayName')
    expect(interventionPanelSource).toContain('getUserUsername')
    expect(interventionPanelSource).not.toContain('item.student.name || item.student.username')
    expect(interventionPanelSource).toContain('intervention-item__signal-inline')
    expect(interventionPanelSource).toContain('intervention-item__meta-inline--username')
    expect(interventionPanelSource).not.toContain('intervention-item__diagnosis')
    expect(interventionPanelSource).not.toContain('intervention-item__meta-chip--username')
    expect(userPresentationSource).toContain('getUserDisplayName')
    expect(userPresentationSource).toContain('getUserUsername')
    expect(interventionPanelSource).toMatch(
      /\.stat-row\s*\{[\s\S]*align-items:\s*baseline;[\s\S]*gap:\s*var\(--space-1-5\);/s
    )
  })
})
