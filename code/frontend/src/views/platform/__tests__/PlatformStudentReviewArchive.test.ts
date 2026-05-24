import { existsSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

describe('PlatformStudentReviewArchive route owner', () => {
  it('应使用平台 route view，并通过中性 feature 承接复盘归档页 workflow', () => {
    const platformViewPath = resolve(
      process.cwd(),
      'src/views/platform/PlatformStudentReviewArchive.vue'
    )
    const platformRoutesPath = resolve(process.cwd(), 'src/router/routes/platformRoutes.ts')

    expect(existsSync(platformViewPath)).toBe(true)

    const platformRoutesSource = readFileSync(platformRoutesPath, 'utf-8')
    expect(platformRoutesSource).toContain(
      "component: () => import('@/views/platform/PlatformStudentReviewArchive.vue')"
    )

    if (!existsSync(platformViewPath)) {
      return
    }

    const platformViewSource = readFileSync(platformViewPath, 'utf-8')
    expect(platformViewSource).toContain(
      "import { useStudentReviewArchivePage } from '@/features/student-review-archive-workspace'"
    )
    expect(platformViewSource).toContain(
      "import { ReviewArchiveWorkspace } from '@/widgets/teacher-review-archive'"
    )
    expect(platformViewSource).not.toContain("from '@/views/teacher/TeacherStudentReviewArchive.vue'")
    expect(platformViewSource).not.toContain("from '@/api/teacher'")
  })
})
