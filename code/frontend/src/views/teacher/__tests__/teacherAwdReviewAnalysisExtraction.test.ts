import { describe, expect, it } from 'vitest'

import awdReviewDetailSource from '@/pages/awd-review/TeacherAwdReviewDetailRoutePage.vue?raw'
import teacherAwdReviewWorkspaceSource from '@/widgets/awd-review-workspace/AwdReviewWorkspace.vue?raw'
import awdReviewAnalysisSectionSource from '@/components/teacher/awd-review/AwdReviewAnalysisSection.vue?raw'

describe('Teacher AWD review analysis extraction', () => {
  it('应将轮次分析与队伍目录区下沉到独立组件', () => {
    expect(awdReviewDetailSource).toContain(
      "import { AwdReviewWorkspace } from '@/widgets/awd-review-workspace'"
    )
    expect(awdReviewDetailSource).toContain('<AwdReviewWorkspace')
    expect(awdReviewDetailSource).not.toContain('class="awd-review-round-grid"')
    expect(awdReviewDetailSource).not.toContain('class="teacher-directory"')

    expect(teacherAwdReviewWorkspaceSource).toContain('<AwdReviewAnalysisSection')
    expect(awdReviewAnalysisSectionSource).toContain('class="awd-review-round-grid"')
    expect(awdReviewAnalysisSectionSource).toContain('class="teacher-directory"')
    expect(awdReviewAnalysisSectionSource).toContain('Performance Analysis')
  })
})
