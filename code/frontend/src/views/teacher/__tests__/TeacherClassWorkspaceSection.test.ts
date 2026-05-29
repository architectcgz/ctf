import { describe, expect, it } from 'vitest'

import teacherClassWorkspaceSectionSource from '../TeacherClassWorkspaceSection.vue?raw'
import classWorkspaceSectionSource from '@/features/class-students-workspace/model/useClassWorkspaceSection.ts?raw'
import classStudentsPageModelSource from '@/features/class-students-workspace/model/useClassStudentsPage.ts?raw'

describe('TeacherClassWorkspaceSection', () => {
  it('maps legacy detail entry routes back to the canonical class workspace with panel query state', () => {
    expect(teacherClassWorkspaceSectionSource).toContain("import TeacherClassStudents from '@/views/teacher/TeacherClassStudents.vue'")
    expect(teacherClassWorkspaceSectionSource).not.toContain('useClassWorkspaceSection(')
    expect(teacherClassWorkspaceSectionSource).not.toContain('useRoute(')
    expect(teacherClassWorkspaceSectionSource).not.toContain('useRouter(')
    expect(teacherClassWorkspaceSectionSource).not.toContain('router.replace')
    expect(classWorkspaceSectionSource).toContain("TeacherClassTrend: 'trend'")
    expect(classWorkspaceSectionSource).toContain("PlatformClassTrend: 'trend'")
    expect(classWorkspaceSectionSource).toContain('route: ClassWorkspaceRouteLike')
    expect(classWorkspaceSectionSource).toContain("TeacherClassTrend: 'TeacherClassStudents'")
    expect(classWorkspaceSectionSource).toContain("PlatformClassTrend: 'PlatformClassStudents'")
    expect(classWorkspaceSectionSource).toContain('panel: panel.value')
    expect(classWorkspaceSectionSource).toContain('canonicalWorkspaceTarget')
    expect(classWorkspaceSectionSource).not.toContain("from 'vue-router'")
    expect(classWorkspaceSectionSource).not.toContain('router.replace')
    expect(classStudentsPageModelSource).toContain('useClassWorkspaceSection')
    expect(classStudentsPageModelSource).toContain('canonicalWorkspaceTarget')
    expect(classStudentsPageModelSource).toContain("from './useClassWorkspaceSection'")
    expect(classStudentsPageModelSource).toContain('router.replace(canonicalWorkspaceTarget.value)')
  })
})
