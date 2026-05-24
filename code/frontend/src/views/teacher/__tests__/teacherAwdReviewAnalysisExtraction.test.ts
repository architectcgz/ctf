import { describe, expect, it } from 'vitest'

import awdReviewDetailSource from '../TeacherAWDReviewDetail.vue?raw'
import teacherAwdReviewWorkspaceSource from '@/widgets/awd-review-workspace/AwdReviewWorkspace.vue?raw'
import teacherAwdReviewAnalysisSectionSource from '@/components/teacher/awd-review/TeacherAWDReviewAnalysisSection.vue?raw'

describe('Teacher AWD review analysis extraction', () => {
  it('应将轮次分析与队伍目录区下沉到独立组件', () => {
    expect(awdReviewDetailSource).toContain(
      "import { AwdReviewWorkspace } from '@/widgets/awd-review-workspace'"
    )
    expect(awdReviewDetailSource).toContain('<AwdReviewWorkspace')
    expect(awdReviewDetailSource).not.toContain('class="awd-review-round-grid"')
    expect(awdReviewDetailSource).not.toContain('class="teacher-directory"')

    expect(teacherAwdReviewWorkspaceSource).toContain('<TeacherAWDReviewAnalysisSection')
    expect(teacherAwdReviewAnalysisSectionSource).toContain('class="awd-review-round-grid"')
    expect(teacherAwdReviewAnalysisSectionSource).toContain('class="teacher-directory"')
    expect(teacherAwdReviewAnalysisSectionSource).toContain('Performance Analysis')
  })
})
