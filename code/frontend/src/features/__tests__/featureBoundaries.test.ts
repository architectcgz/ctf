import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, relative } from 'node:path'

import { describe, expect, it } from 'vitest'

const sourceRoot = join(process.cwd(), 'src')
const checkedRoots = ['views', 'components'].map((item) => join(sourceRoot, item))
const migratedComposableImports = [
  '@/composables/useChallengeDetailInteractions',
  '@/composables/useChallengeDetailPresentation',
  '@/composables/useChallengeInstance',
  '@/composables/useChallengeTopologyStudioPage',
  '@/composables/useContestAWDWorkspace',
  '@/composables/useChallengeManagePage',
  '@/composables/useChallengePackageImport',
  '@/composables/useChallengeWriteupEditorPage',
  '@/composables/useContestAnnouncementManagement',
  '@/composables/useContestDetailPage',
  '@/composables/useAuditLogPage',
  '@/composables/useAdminNotificationPublisher',
  '@/composables/useAuth',
  '@/composables/useAwdCheckResultPresentation',
  '@/composables/useAwdInspectorCoreState',
  '@/composables/useAwdInspectorDerivedData',
  '@/composables/useAwdInspectorExports',
  '@/composables/useAwdInspectorFilters',
  '@/composables/useAwdInspectorFormatting',
  '@/composables/useAwdInspectorSummaryMetrics',
  '@/composables/useAwdTrafficPanel',
  '@/composables/useContestAnnouncementRealtime',
  '@/composables/useContestAwdChallengePicker',
  '@/composables/useContestAwdPreviewRealtime',
  '@/composables/useContestChallengePool',
  '@/composables/useContestEditAwdWorkspace',
  '@/composables/useContestExportFlow',
  '@/composables/useContestProjectorData',
  '@/composables/useContestProjectorDerived',
  '@/composables/useContestScoreboardRealtime',
  '@/composables/useContestWorkbench',
  '@/composables/useImageManagePage',
  '@/composables/useInstanceListPage',
  '@/composables/useNotificationDropdown',
  '@/composables/useNotificationRealtime',
  '@/composables/usePlatformContestAwd',
  '@/composables/usePlatformAwdChallenges',
  '@/composables/usePlatformChallenges',
  '@/composables/usePlatformContests',
  '@/composables/usePlatformOverviewWorkspace',
  '@/composables/usePlatformUsers',
  '@/composables/useScoreboardView',
  '@/composables/useSkillProfilePage',
  '@/composables/useStudentDirectoryQuery',
  '@/composables/useStudentFilters',
  '@/composables/useStudentListQuery',
  '@/composables/useTeacherAwdReviewDetail',
  '@/composables/useTeacherAwdReviewIndex',
  '@/composables/useTeacherClassReportExport',
  '@/composables/useTeacherDashboardMetrics',
  '@/composables/useTeacherInstances',
  '@/composables/useTeacherStudentAnalysisPage',
  '@/composables/useTeacherStudentReviewArchive',
  '@/composables/useTeacherWorkspace',
]

function collectSourceFiles(directory: string): string[] {
  return readdirSync(directory).flatMap((entry) => {
    const path = join(directory, entry)
    const stats = statSync(path)
    if (stats.isDirectory()) {
      return collectSourceFiles(path)
    }
    if (/\.(ts|vue)$/.test(entry)) {
      return [path]
    }
    return []
  })
}

describe('feature boundaries', () => {
  it('route views and components import migrated feature models through feature public APIs', () => {
    const violations = checkedRoots
      .flatMap(collectSourceFiles)
      .flatMap((filePath) => {
        const source = readFileSync(filePath, 'utf-8')
        return migratedComposableImports
          .filter((importPath) => source.includes(importPath))
          .map((importPath) => `${relative(sourceRoot, filePath)} -> ${importPath}`)
      })

    expect(violations).toEqual([])
  })
})
