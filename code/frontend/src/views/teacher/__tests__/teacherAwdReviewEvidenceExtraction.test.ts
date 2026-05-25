import { describe, expect, it } from 'vitest'

import awdReviewDetailSource from '../TeacherAWDReviewDetail.vue?raw'
import teacherAwdReviewWorkspaceSource from '@/widgets/awd-review-workspace/AwdReviewWorkspace.vue?raw'
import awdReviewEvidenceGridSource from '@/components/teacher/awd-review/AwdReviewEvidenceGrid.vue?raw'

describe('Teacher AWD review evidence extraction', () => {
  it('应将服务/攻击/流量证据区下沉到独立组件', () => {
    expect(awdReviewDetailSource).toContain(
      "import { AwdReviewWorkspace } from '@/widgets/awd-review-workspace'"
    )
    expect(awdReviewDetailSource).toContain('<AwdReviewWorkspace')
    expect(awdReviewDetailSource).not.toContain('data-testid="awd-review-service-id"')
    expect(awdReviewDetailSource).not.toContain('data-testid="awd-review-attack-service-id"')
    expect(awdReviewDetailSource).not.toContain('data-testid="awd-review-traffic-service-id"')

    expect(teacherAwdReviewWorkspaceSource).toContain('<AwdReviewEvidenceGrid')
    expect(awdReviewEvidenceGridSource).toContain('data-testid="awd-review-service-id"')
    expect(awdReviewEvidenceGridSource).toContain('data-testid="awd-review-attack-service-id"')
    expect(awdReviewEvidenceGridSource).toContain('data-testid="awd-review-traffic-service-id"')
  })
})
