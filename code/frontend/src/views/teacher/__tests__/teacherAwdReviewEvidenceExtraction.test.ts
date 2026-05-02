import { describe, expect, it } from 'vitest'

import awdReviewDetailSource from '../TeacherAWDReviewDetail.vue?raw'
import teacherAwdReviewWorkspaceSource from '@/widgets/teacher-awd-review/TeacherAWDReviewWorkspace.vue?raw'
import teacherAwdReviewEvidenceGridSource from '@/components/teacher/awd-review/TeacherAWDReviewEvidenceGrid.vue?raw'

describe('Teacher AWD review evidence extraction', () => {
  it('应将服务/攻击/流量证据区下沉到独立组件', () => {
    expect(awdReviewDetailSource).toContain(
      "import { TeacherAWDReviewWorkspace } from '@/widgets/teacher-awd-review'"
    )
    expect(awdReviewDetailSource).toContain('<TeacherAWDReviewWorkspace')
    expect(awdReviewDetailSource).not.toContain('data-testid="awd-review-service-id"')
    expect(awdReviewDetailSource).not.toContain('data-testid="awd-review-attack-service-id"')
    expect(awdReviewDetailSource).not.toContain('data-testid="awd-review-traffic-service-id"')

    expect(teacherAwdReviewWorkspaceSource).toContain('<TeacherAWDReviewEvidenceGrid')
    expect(teacherAwdReviewEvidenceGridSource).toContain('data-testid="awd-review-service-id"')
    expect(teacherAwdReviewEvidenceGridSource).toContain('data-testid="awd-review-attack-service-id"')
    expect(teacherAwdReviewEvidenceGridSource).toContain('data-testid="awd-review-traffic-service-id"')
  })
})
