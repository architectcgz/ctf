import { existsSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

describe('PlatformClassWorkspaceSection route owner', () => {
  it('应使用平台 route view，并通过中性 feature 承接班级工作台别名页重定向 owner', () => {
    const platformViewPath = resolve(
      process.cwd(),
      'src/views/platform/PlatformClassWorkspaceSection.vue'
    )
    const platformRoutesPath = resolve(process.cwd(), 'src/router/routes/platformRoutes.ts')

    expect(existsSync(platformViewPath)).toBe(true)

    const platformRoutesSource = readFileSync(platformRoutesPath, 'utf-8')
    expect(platformRoutesSource).toContain(
      "component: () => import('@/views/platform/PlatformClassWorkspaceSection.vue')"
    )

    if (!existsSync(platformViewPath)) {
      return
    }

    const platformViewSource = readFileSync(platformViewPath, 'utf-8')
    expect(platformViewSource).toContain(
      "import PlatformClassStudents from '@/views/platform/PlatformClassStudents.vue'"
    )
    expect(platformViewSource).not.toContain('useClassWorkspaceSection(')
    expect(platformViewSource).not.toContain('useRoute(')
    expect(platformViewSource).not.toContain('useRouter(')
    expect(platformViewSource).not.toContain('router.replace')
    expect(platformViewSource).not.toContain("from '@/views/teacher/TeacherClassWorkspaceSection.vue'")
    expect(platformViewSource).not.toContain("from '@/api/teacher'")
  })
})
