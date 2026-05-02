import { describe, expect, it } from 'vitest'

import teacherClassWorkspaceSectionSource from '../TeacherClassWorkspaceSection.vue?raw'
import teacherClassWorkspaceSectionModelSource from '@/features/teacher-class-workspace/model/useTeacherClassWorkspaceSection.ts?raw'

describe('TeacherClassWorkspaceSection', () => {
  it('maps legacy detail entry routes back to the canonical class workspace with panel query state', () => {
    expect(teacherClassWorkspaceSectionSource).toContain('useTeacherClassWorkspaceSection()')
    expect(teacherClassWorkspaceSectionSource).not.toContain('router.replace')
    expect(teacherClassWorkspaceSectionModelSource).toContain("TeacherClassTrend: 'trend'")
    expect(teacherClassWorkspaceSectionModelSource).toContain("TeacherClassReview: 'review'")
    expect(teacherClassWorkspaceSectionModelSource).toContain("TeacherClassInsights: 'insight'")
    expect(teacherClassWorkspaceSectionModelSource).toContain("TeacherClassIntervention: 'action'")
    expect(teacherClassWorkspaceSectionModelSource).toContain("name: 'TeacherClassStudents'")
    expect(teacherClassWorkspaceSectionModelSource).toContain('panel: targetPanel.value')
    expect(teacherClassWorkspaceSectionModelSource).toContain('router.replace')
  })
})
