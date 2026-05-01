import { describe, expect, it } from 'vitest'

import teacherInterventionPanelSource from '@/components/teacher/TeacherInterventionPanel.vue?raw'

describe('teacher intervention panel layout', () => {
  it('应采用紧凑可穿透的介入列表布局', () => {
    expect(teacherInterventionPanelSource).toContain('intervention-item__header')
    expect(teacherInterventionPanelSource).toContain('intervention-item__name-button')
    expect(teacherInterventionPanelSource).toContain("@click=\"openStudent(item.student.id)\"")
    expect(teacherInterventionPanelSource).toContain('name: \'TeacherStudentAnalysis\'')
    expect(teacherInterventionPanelSource).toContain('intervention-item__signal-inline')
    expect(teacherInterventionPanelSource).toContain('intervention-item__meta-inline--username')
    expect(teacherInterventionPanelSource).not.toContain('intervention-item__diagnosis')
    expect(teacherInterventionPanelSource).not.toContain('intervention-item__meta-chip--username')
    expect(teacherInterventionPanelSource).toMatch(
      /\.stat-row\s*\{[\s\S]*align-items:\s*baseline;[\s\S]*gap:\s*var\(--space-1-5\);/s
    )
  })
})
