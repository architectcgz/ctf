import { describe, expect, it } from 'vitest'

import teacherClassWorkspaceSectionSource from '../TeacherClassWorkspaceSection.vue?raw'

describe('TeacherClassWorkspaceSection', () => {
  it('maps legacy detail entry routes back to the canonical class workspace with panel query state', () => {
    expect(teacherClassWorkspaceSectionSource).toContain("TeacherClassTrend: 'trend'")
    expect(teacherClassWorkspaceSectionSource).toContain("TeacherClassReview: 'review'")
    expect(teacherClassWorkspaceSectionSource).toContain("TeacherClassInsights: 'insight'")
    expect(teacherClassWorkspaceSectionSource).toContain("TeacherClassIntervention: 'action'")
    expect(teacherClassWorkspaceSectionSource).toContain("name: 'TeacherClassStudents'")
    expect(teacherClassWorkspaceSectionSource).toContain('panel: targetPanel')
    expect(teacherClassWorkspaceSectionSource).toContain('router.replace')
  })
})
