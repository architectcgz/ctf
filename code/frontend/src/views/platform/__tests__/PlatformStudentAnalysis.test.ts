import { existsSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

describe('PlatformStudentAnalysis route owner', () => {
  it('应使用平台 route view，并通过中性 feature 承接共享 page workflow', () => {
    const platformViewPath = resolve(process.cwd(), 'src/views/platform/PlatformStudentAnalysis.vue')
    const platformRoutesPath = resolve(process.cwd(), 'src/router/routes/platformRoutes.ts')

    expect(existsSync(platformViewPath)).toBe(true)

    const platformRoutesSource = readFileSync(platformRoutesPath, 'utf-8')
    expect(platformRoutesSource).toContain(
      "component: () => import('@/views/platform/PlatformStudentAnalysis.vue')"
    )

    if (!existsSync(platformViewPath)) {
      return
    }

    const platformViewSource = readFileSync(platformViewPath, 'utf-8')
    expect(platformViewSource).toContain(
      "import { StudentAnalysisPage } from '@/components/class-management'"
    )
    expect(platformViewSource).toContain(
      "import { useStudentAnalysisPage } from '@/features/student-analysis-workspace'"
    )
    expect(platformViewSource).toContain(
      "import { ClassReportExportDialog } from '@/components/teacher/reports'"
    )
    expect(platformViewSource).not.toContain("from '@/views/teacher/TeacherStudentAnalysis.vue'")
    expect(platformViewSource).not.toContain("from '@/api/teacher'")
    expect(platformViewSource).not.toContain(
      '@/components/teacher/class-management/StudentAnalysisPage.vue'
    )
    expect(platformViewSource).not.toContain('ClassReportExportDialog.vue')
    expect(platformViewSource).not.toContain(':classes="classes"')
    expect(platformViewSource).not.toContain(':students="students"')
    expect(platformViewSource).not.toContain(':selected-class-name="selectedClassName"')
    expect(platformViewSource).not.toContain(':selected-student-id="selectedStudentId"')
    expect(platformViewSource).not.toContain(':loading-classes="loadingClasses"')
    expect(platformViewSource).not.toContain(':loading-students="loadingStudents"')
    expect(platformViewSource).not.toContain('@open-class-management="openClassManagement"')
    expect(platformViewSource).not.toContain('@select-class="selectClass"')
    expect(platformViewSource).not.toContain('@select-student="selectStudent"')
  })
})
