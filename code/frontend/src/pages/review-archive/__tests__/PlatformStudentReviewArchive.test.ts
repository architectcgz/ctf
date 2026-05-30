import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, it } from 'vitest'

describe('PlatformStudentReviewArchive route owner', () => {
  it('应使用平台 route view，并通过中性 feature 承接复盘归档页 workflow', () => {
    const platformRoutesPath = resolve(process.cwd(), 'src/router/routes/platformRoutes.ts')

    const platformRoutesSource = readFileSync(platformRoutesPath, 'utf-8')
    expect(platformRoutesSource).toContain(
      "component: () => import('@/pages/review-archive/StudentReviewArchiveRoutePage.vue')"
    )

    const platformViewSource = readFileSync(
      resolve(
        process.cwd(),
        'src/pages/review-archive/StudentReviewArchiveRoutePage.vue'
      ),
      'utf-8'
    )
    expect(platformViewSource).toContain(
      "import { useStudentReviewArchivePage } from '@/features/teaching/student-review-archive-workspace'"
    )
    expect(platformViewSource).toContain(
      "import { ReviewArchiveWorkspace } from '@/widgets/review-archive-workspace'"
    )
    expect(platformViewSource).not.toContain("from '@/views/teacher/TeacherStudentReviewArchive.vue'")
    expect(platformViewSource).not.toContain("from '@/api/teacher'")
  })
})
