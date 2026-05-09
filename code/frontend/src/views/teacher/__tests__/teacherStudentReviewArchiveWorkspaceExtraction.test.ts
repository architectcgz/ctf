import { describe, expect, it } from 'vitest'

import reviewArchiveSource from '../TeacherStudentReviewArchive.vue?raw'
import reviewArchiveWorkspaceSource from '@/widgets/teacher-review-archive/TeacherReviewArchiveWorkspace.vue?raw'
import reviewArchiveStateSource from '@/widgets/teacher-review-archive/TeacherReviewArchiveState.vue?raw'
import reviewArchiveSummarySectionSource from '@/widgets/teacher-review-archive/TeacherReviewArchiveSummarySection.vue?raw'
import reviewArchiveHeroSource from '@/components/teacher/review-archive/ReviewArchiveHero.vue?raw'

describe('Teacher student review archive workspace extraction', () => {
  it('路由页应收敛为 feature model 与 widget 组合层', () => {
    expect(reviewArchiveSource).toContain(
      "import { useTeacherStudentReviewArchivePage } from '@/features/teacher-student-review-archive'"
    )
    expect(reviewArchiveSource).toContain(
      "import { TeacherReviewArchiveWorkspace } from '@/widgets/teacher-review-archive'"
    )
    expect(reviewArchiveSource).toContain('<TeacherReviewArchiveWorkspace')
    expect(reviewArchiveSource).not.toContain('exportStudentReviewArchive')
    expect(reviewArchiveSource).not.toContain('<ReviewArchiveHero')
    expect(reviewArchiveSource).not.toContain('class="review-archive-shell')

    expect(reviewArchiveWorkspaceSource).toContain('<ReviewArchiveHero')
    expect(reviewArchiveWorkspaceSource).toContain('<ReviewArchiveObservationStrip')
    expect(reviewArchiveWorkspaceSource).toContain('<TeacherReviewArchiveState')
    expect(reviewArchiveWorkspaceSource).toContain('<TeacherReviewArchiveSummarySection')
    expect(reviewArchiveWorkspaceSource).toContain('<ReviewArchiveEvidencePanel')
    expect(reviewArchiveWorkspaceSource).toContain('<ReviewArchiveReflectionPanel')
    expect(reviewArchiveWorkspaceSource).toContain('class="review-archive-shell')
    expect(reviewArchiveStateSource).toContain('class="ui-btn ui-btn--primary"')
    expect(reviewArchiveStateSource).toContain('class="review-archive-loading__hero"')
    expect(reviewArchiveWorkspaceSource).not.toContain('<ElButton')
    expect(reviewArchiveSummarySectionSource).toContain('class="review-archive-summary-grid"')
    expect(reviewArchiveSummarySectionSource).toContain('class="skill-bars"')
    expect(reviewArchiveSummarySectionSource).toContain(
      'class="summary-card progress-card metric-panel-card"'
    )
    expect(reviewArchiveSummarySectionSource).toContain(
      'class="summary-card__label progress-card-label metric-panel-label"'
    )
    expect(reviewArchiveSummarySectionSource).toContain('<component :is="card.icon" class="h-4 w-4" />')

    expect(reviewArchiveHeroSource).toContain('class="header-actions archive-hero__actions"')
    expect(reviewArchiveHeroSource).toContain('class="header-btn header-btn--ghost"')
    expect(reviewArchiveHeroSource).toContain('class="header-btn header-btn--primary"')
  })
})
