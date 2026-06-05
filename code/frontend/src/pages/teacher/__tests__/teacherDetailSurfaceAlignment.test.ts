import { describe, expect, it } from 'vitest'

import classTrendPanelSource from '@/entities/class-insight/ui/ClassTrendPanel.vue?raw'
import classInsightsPanelSource from '@/entities/class-insight/ui/ClassInsightsPanel.vue?raw'
import classReviewPanelSource from '@/entities/class-insight/ui/ClassReviewPanel.vue?raw'
import interventionPanelSource from '@/features/teaching/student-analysis-review/ui/InterventionPanel.vue?raw'
import studentInsightRecommendationsSectionSource from '@/features/teaching/student-analysis-workspace/ui/StudentInsightRecommendationsSection.vue?raw'
import studentInsightAttackSessionsSectionSource from '@/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.vue?raw'
import studentInsightWriteupsSource from '@/features/teaching/student-analysis-review/ui/StudentInsightWriteupsSection.vue?raw'
import studentInsightManualReviewSource from '@/features/teaching/student-analysis-review/ui/StudentInsightManualReviewSection.vue?raw'
import reviewArchiveSource from '@/pages/review-archive/StudentReviewArchiveRoutePage.vue?raw'
import reviewArchiveWorkspaceSource from '@/widgets/review-archive-workspace/ReviewArchiveWorkspace.vue?raw'
import reviewArchiveSummarySectionSource from '@/widgets/review-archive-workspace/ReviewArchiveSummarySection.vue?raw'
import sectionCardSource from '@/shared/ui/common/SectionCard.vue?raw'

const reviewArchiveCombinedSource = [
  reviewArchiveSource,
  reviewArchiveWorkspaceSource,
  reviewArchiveSummarySectionSource,
].join('\n')

describe('teacher detail surface alignment', () => {
  it('student analysis and review archive should keep section surface owner on shared section-card variants', () => {
    expect(sectionCardSource).toContain(
      "type SectionCardVariant = 'default' | 'teacher-flat' | 'teacher-surface'"
    )
    expect(studentInsightRecommendationsSectionSource).toContain('variant="teacher-flat"')
    expect(studentInsightRecommendationsSectionSource).toContain(
      '<style src="@/features/teaching/student-analysis-shared/ui/studentInsightSections.css"></style>'
    )
    expect(reviewArchiveCombinedSource).toContain('variant="teacher-surface"')
  })

  it('teacher detail panels should keep shared shell and section CSS as the only style owner', () => {
    expect(classTrendPanelSource).toContain("@import '@/assets/styles/teacher-panel-shell.css';")
    expect(classInsightsPanelSource).toContain("@import '@/assets/styles/teacher-panel-shell.css';")
    expect(classReviewPanelSource).toContain("@import '@/assets/styles/teacher-panel-shell.css';")
    expect(interventionPanelSource).toContain("@import '@/assets/styles/teacher-panel-shell.css';")
    expect(studentInsightWriteupsSource).toContain(
      '<style src="@/features/teaching/student-analysis-shared/ui/studentInsightSections.css"></style>'
    )
    expect(studentInsightManualReviewSource).toContain(
      '<style src="@/features/teaching/student-analysis-shared/ui/studentInsightSections.css"></style>'
    )
    expect(studentInsightAttackSessionsSectionSource).toContain(
      '<style src="@/features/teaching/student-analysis-shared/ui/studentInsightSections.css"></style>'
    )
  })
})
