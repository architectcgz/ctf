import { describe, expect, it } from 'vitest'

import teacherClassWorkspaceSectionSource from '../TeacherClassWorkspaceSection.vue?raw'
import classWorkspaceSectionSource from '@/features/class-workspace-redirect/model/useClassWorkspaceSection.ts?raw'
import teacherClassWorkspaceSectionModelSource from '@/features/teacher-class-workspace/model/useTeacherClassWorkspaceSection.ts?raw'

describe('TeacherClassWorkspaceSection', () => {
  it('maps legacy detail entry routes back to the canonical class workspace with panel query state', () => {
    expect(teacherClassWorkspaceSectionSource).toContain('useTeacherClassWorkspaceSection()')
    expect(teacherClassWorkspaceSectionSource).not.toContain('router.replace')
    expect(teacherClassWorkspaceSectionModelSource).toContain(
      "export { useClassWorkspaceSection as useTeacherClassWorkspaceSection } from '@/features/class-workspace-redirect'"
    )
    expect(classWorkspaceSectionSource).toContain("TeacherClassTrend: {")
    expect(classWorkspaceSectionSource).toContain("PlatformClassTrend: {")
    expect(classWorkspaceSectionSource).toContain("routeName: 'TeacherClassStudents'")
    expect(classWorkspaceSectionSource).toContain("routeName: 'PlatformClassStudents'")
    expect(classWorkspaceSectionSource).toContain('panel: target.value.panel')
    expect(classWorkspaceSectionSource).toContain('router.replace')
  })
})
