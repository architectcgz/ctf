import { describe, expect, it } from 'vitest'

import awdReviewDetailSource from '../TeacherAWDReviewDetail.vue?raw'
import teacherAwdReviewAnalysisSectionSource from '@/components/teacher/awd-review/TeacherAWDReviewAnalysisSection.vue?raw'

describe('Teacher AWD review analysis extraction', () => {
  it('应将轮次分析与队伍目录区下沉到独立组件', () => {
    expect(awdReviewDetailSource).toContain(
      "import TeacherAWDReviewAnalysisSection from '@/components/teacher/awd-review/TeacherAWDReviewAnalysisSection.vue'"
    )
    expect(awdReviewDetailSource).toContain('<TeacherAWDReviewAnalysisSection')
    expect(awdReviewDetailSource).not.toContain('class="awd-review-round-grid"')
    expect(awdReviewDetailSource).not.toContain('class="teacher-directory"')

    expect(teacherAwdReviewAnalysisSectionSource).toContain('class="awd-review-round-grid"')
    expect(teacherAwdReviewAnalysisSectionSource).toContain('class="teacher-directory"')
    expect(teacherAwdReviewAnalysisSectionSource).toContain('Performance Analysis')
  })
})
