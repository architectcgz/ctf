import { describe, expect, it } from 'vitest'

import teacherClassWorkspaceSectionSource from '../TeacherClassWorkspaceSection.vue?raw'
import classWorkspaceSectionSource from '@/features/class-workspace-redirect/model/useClassWorkspaceSection.ts?raw'

describe('TeacherClassWorkspaceSection', () => {
  it('maps legacy detail entry routes back to the canonical class workspace with panel query state', () => {
    expect(teacherClassWorkspaceSectionSource).toContain('useClassWorkspaceSection({')
    expect(teacherClassWorkspaceSectionSource).toContain(
      "import { useClassWorkspaceSection } from '@/features/class-workspace-redirect'"
    )
    expect(teacherClassWorkspaceSectionSource).toContain("workspaceRouteName: 'TeacherClassStudents'")
    expect(teacherClassWorkspaceSectionSource).not.toContain('router.replace')
    expect(classWorkspaceSectionSource).toContain("TeacherClassTrend: 'trend'")
    expect(classWorkspaceSectionSource).toContain("PlatformClassTrend: 'trend'")
    expect(classWorkspaceSectionSource).toContain('workspaceRouteName: ClassWorkspaceCanonicalRouteName')
    expect(classWorkspaceSectionSource).toContain('name: workspaceRouteName')
    expect(classWorkspaceSectionSource).toContain('panel: panel.value')
    expect(classWorkspaceSectionSource).toContain('router.replace')
  })
})
