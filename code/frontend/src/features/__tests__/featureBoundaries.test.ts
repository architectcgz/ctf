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

function isTestFile(filePath: string): boolean {
  return filePath.endsWith('.test.ts') || filePath.endsWith('.spec.ts')
}

function extractImportSpecifiers(source: string): string[] {
  return Array.from(source.matchAll(/from\s+['"]([^'"]+)['"]/g)).map((match) => match[1])
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

  it('feature runtime sources should not import components layer', () => {
    const featuresRoot = join(sourceRoot, 'features')
    const violations = collectSourceFiles(featuresRoot)
      .filter((filePath) => !isTestFile(filePath))
      .filter((filePath) => {
        const source = readFileSync(filePath, 'utf-8')
        return /from\s+['"]@\/components\//.test(source)
      })
      .map((filePath) => relative(sourceRoot, filePath))

    expect(violations).toEqual([])
  })

  it('runtime sources should not deep import another feature model via alias path', () => {
    const violations = collectSourceFiles(sourceRoot)
      .filter((filePath) => !isTestFile(filePath))
      .flatMap((filePath) => {
        const source = readFileSync(filePath, 'utf-8')
        const imports = extractImportSpecifiers(source)
        const importerRelative = relative(sourceRoot, filePath)
        const importerFeature = importerRelative.startsWith('features/')
          ? importerRelative.split('/')[1]
          : null

        return imports
          .filter((importPath) => importPath.startsWith('@/features/'))
          .filter((importPath) => /@\/features\/[^/]+\/model\//.test(importPath))
          .filter((importPath) => {
            const importedFeature = importPath.replace('@/features/', '').split('/')[0]
            // Same feature internal imports are allowed; cross-feature deep imports are not.
            return !(importerFeature && importedFeature === importerFeature)
          })
          .map((importPath) => `${importerRelative} -> ${importPath}`)
      })

    expect(violations).toEqual([])
  })
})
