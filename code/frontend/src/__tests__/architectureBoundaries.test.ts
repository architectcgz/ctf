import { readdirSync, readFileSync, statSync } from 'node:fs'
import { dirname, join, normalize, relative, resolve, sep } from 'node:path'

import { describe, expect, it } from 'vitest'

const sourceRoot = join(process.cwd(), 'src')

type Layer = 'common' | 'entities' | 'features' | 'widgets' | 'views' | 'other'

interface SourceFile {
  absolutePath: string
  relativePath: string
  layer: Layer
}

const importPattern =
  /(?:import|export)\s+(?:type\s+)?(?:[^'"]*?\s+from\s+)?['"]([^'"]+)['"]|import\s*\(\s*['"]([^'"]+)['"]\s*\)/g

const viewLineLimit = 500

const oversizedViewAllowlist = new Set([
  'views/challenges/ChallengeDetail.vue',
  'views/contests/ContestDetail.vue',
  'views/instances/InstanceList.vue',
  'views/platform/ContestAwdConfig.vue',
  'views/profile/SkillProfile.vue',
  'views/profile/UserProfile.vue',
  'views/scoreboard/ScoreboardView.vue',
])

const componentFeatureImportAllowlist = new Set([
  'components/challenge/ChallengeSolutionsPanel.vue -> @/features/challenge-detail',
  'components/challenge/ChallengeSubmissionRecordsPanel.vue -> @/features/challenge-detail',
  'components/contests/ContestAWDWorkspacePanel.vue -> @/features/contest-awd-workspace',
  'components/contests/ContestAnnouncementRealtimeBridge.vue -> @/features/contest-announcements',
  'components/contests/awd/AWDDefenseOperationsPanel.vue -> @/features/contest-awd-workspace',
  'components/contests/awd/AWDDefenseServiceList.vue -> @/features/contest-awd-workspace',
  'components/dashboard/student/dashboardPanelRegistry.ts -> @/features/student-dashboard',
  'components/layout/AppLayout.vue -> @/features/notifications',
  'components/layout/NotificationDrawer.vue -> @/features/notifications',
  'components/layout/TopNav.vue -> @/features/auth',
  'components/notifications/AdminNotificationPublishDrawer.vue -> @/features/admin-notification-publisher',
  'components/platform/awd-service/AWDChallengeEditorDialog.vue -> @/features/platform-awd-challenges',
  'components/platform/awd-service/AWDChallengeLibraryPage.vue -> @/features/platform-awd-challenges',
  'components/platform/challenge/AdminChallengeProfilePanel.vue -> @/features/platform-challenge-detail',
  'components/platform/challenge/AdminChallengeWorkspaceTabs.vue -> @/features/platform-challenge-detail',
  'components/platform/challenge/ChallengeManageDirectoryPanel.vue -> @/features/platform-challenges',
  'components/platform/contest/AWDChallengeConfigDialog.vue -> @/features/awd-inspector',
  'components/platform/contest/AWDChallengeConfigDialog.vue -> @/features/contest-awd-config',
  'components/platform/contest/AWDChallengeConfigPanel.vue -> @/features/awd-inspector',
  'components/platform/contest/AWDOperationsPanel.vue -> @/features/contest-awd-admin',
  'components/platform/contest/AWDRoundInspector.vue -> @/features/awd-inspector',
  'components/platform/contest/AWDTrafficPanel.vue -> @/features/awd-inspector',
  'components/platform/contest/ContestAnnouncementManageDrawer.vue -> @/features/contest-announcements',
  'components/platform/contest/ContestChallengeFilterStrip.vue -> @/features/contest-workbench',
  'components/platform/contest/ContestChallengeOrchestrationPanel.vue -> @/features/contest-workbench',
  'components/platform/contest/ContestEditWorkspacePanel.vue -> @/features/contest-workbench',
  'components/platform/contest/ContestEditWorkspacePanel.vue -> @/features/platform-contests',
  'components/platform/contest/ContestOrchestrationPage.vue -> @/features/platform-contests',
  'components/platform/contest/ContestWorkbenchStageTabs.vue -> @/features/contest-workbench',
  'components/platform/contest/ContestWorkbenchSummaryStrip.vue -> @/features/contest-workbench',
  'components/platform/contest/PlatformContestFormDialog.vue -> @/features/platform-contests',
  'components/platform/contest/PlatformContestFormPanel.vue -> @/features/platform-contests',
  'components/platform/contest/awdInspector.types.ts -> @/features/awd-inspector',
  'components/platform/dashboard/PlatformOverviewPage.vue -> @/features/platform-overview',
  'components/platform/topology/ChallengeTopologyStudioPage.vue -> @/features/challenge-topology-studio',
  'components/platform/topology/ChallengeTopologyStudioPage.vue -> @/features/challenge-topology-studio/model',
  'components/platform/topology/TopologyCanvasBoard.vue -> @/features/challenge-topology-studio/model',
  'components/platform/topology/TopologyConnectivitySections.vue -> @/features/challenge-topology-studio/model',
  'components/platform/topology/TopologyNetworkSection.vue -> @/features/challenge-topology-studio/model',
  'components/platform/topology/TopologyNodeEditor.vue -> @/features/challenge-topology-studio/model',
  'components/platform/topology/TopologyNodeSection.vue -> @/features/challenge-topology-studio/model',
  'components/platform/user/PlatformUserFormDialog.vue -> @/features/platform-users',
  'components/platform/writeup/ChallengeWriteupEditorPage.vue -> @/features/challenge-writeup-editor',
  'components/platform/writeup/ChallengeWriteupManagePanel.vue -> @/features/challenge-writeup-editor',
  'components/platform/writeup/ChallengeWriteupViewPage.vue -> @/features/challenge-writeup-editor',
  'components/scoreboard/ScoreboardRealtimeBridge.vue -> @/features/scoreboard',
  'components/teacher/TeacherInterventionPanel.vue -> @/features/teacher-student-analysis',
  'components/teacher/dashboard/TeacherDashboardPage.vue -> @/features/teacher-dashboard',
  'components/teacher/reports/TeacherClassReportExportDialog.vue -> @/features/teacher-class-report-export',
])

const widgetLegacyComponentImportAllowlist = new Set([
  'widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.vue -> @/components/platform/challenge/AdminChallengeTopbarPanel.vue',
  'widgets/platform-challenge-detail/PlatformChallengeDetailWorkspace.vue -> @/components/platform/challenge/AdminChallengeWorkspaceTabs.vue',
  'widgets/teacher-awd-review/TeacherAWDReviewWorkspace.vue -> @/components/teacher/awd-review/TeacherAWDReviewAnalysisSection.vue',
  'widgets/teacher-awd-review/TeacherAWDReviewWorkspace.vue -> @/components/teacher/awd-review/TeacherAWDReviewEvidenceGrid.vue',
  'widgets/teacher-awd-review/TeacherAWDReviewWorkspace.vue -> @/components/teacher/awd-review/TeacherAWDReviewRoundSelector.vue',
  'widgets/teacher-awd-review/TeacherAWDReviewWorkspace.vue -> @/components/teacher/awd-review/TeacherAWDReviewTeamDrawer.vue',
  'widgets/teacher-review-archive/TeacherReviewArchiveWorkspace.vue -> @/components/teacher/review-archive/ReviewArchiveEvidencePanel.vue',
  'widgets/teacher-review-archive/TeacherReviewArchiveWorkspace.vue -> @/components/teacher/review-archive/ReviewArchiveHero.vue',
  'widgets/teacher-review-archive/TeacherReviewArchiveWorkspace.vue -> @/components/teacher/review-archive/ReviewArchiveObservationStrip.vue',
  'widgets/teacher-review-archive/TeacherReviewArchiveWorkspace.vue -> @/components/teacher/review-archive/ReviewArchiveReflectionPanel.vue',
])

const componentNonContractApiAllowlist = new Set([
  'components/teacher/StudentInsightPanel.vue -> @/api/teacher',
  'components/teacher/class-management/StudentAnalysisPage.vue -> @/api/teacher',
  'components/teacher/student-insight/StudentInsightAttackSessionsSection.vue -> @/api/teacher',
])

const widgetNonContractApiAllowlist = new Set([
  'widgets/teacher-student-review-workspace/TeacherStudentReviewWorkspace.vue -> @/api/teacher',
])

const commonForbiddenImportAllowlist = new Set([
  'components/common/InstancePanel.vue -> @/api/contracts',
  'entities/challenge/model/presentation.ts -> @/api/contracts',
  'entities/challenge/ui/ChallengeCategoryDifficultyPills.vue -> @/api/contracts',
  'entities/challenge/ui/ChallengeCategoryPill.vue -> @/api/contracts',
  'entities/challenge/ui/ChallengeCategoryText.vue -> @/api/contracts',
  'entities/challenge/ui/ChallengeDifficultyText.vue -> @/api/contracts',
  'entities/challenge/ui/ChallengeDirectoryRow.vue -> @/api/contracts',
  'entities/challenge/ui/ChallengeMetaStrip.vue -> @/api/contracts',
  'entities/challenge/ui/ChallengeProfileMetaGrid.vue -> @/api/contracts',
  'entities/challenge/ui/ChallengeProfileSummaryStrip.vue -> @/api/contracts',
])

const legacyComponentPageAllowlist = new Set([
  'components/dashboard/student/StudentCategoryProgressPage.vue',
  'components/dashboard/student/StudentDifficultyPage.vue',
  'components/dashboard/student/StudentOverviewPage.vue',
  'components/dashboard/student/StudentRecommendationPage.vue',
  'components/dashboard/student/StudentTimelinePage.vue',
  'components/platform/awd-service/AWDChallengeLibraryPage.vue',
  'components/platform/contest/ContestOrchestrationPage.vue',
  'components/platform/dashboard/PlatformOverviewPage.vue',
  'components/platform/topology/ChallengeTopologyStudioPage.vue',
  'components/platform/user/UserGovernancePage.vue',
  'components/platform/writeup/ChallengeWriteupEditorPage.vue',
  'components/platform/writeup/ChallengeWriteupViewPage.vue',
  'components/teacher/class-management/ClassManagementPage.vue',
  'components/teacher/class-management/ClassStudentsPage.vue',
  'components/teacher/class-management/StudentAnalysisPage.vue',
  'components/teacher/dashboard/TeacherDashboardPage.vue',
  'components/teacher/instance-management/TeacherInstanceManagementPage.vue',
  'components/teacher/student-management/StudentManagementPage.vue',
])

function collectSourceFiles(directory: string): SourceFile[] {
  return readdirSync(directory).flatMap((entry) => {
    const absolutePath = join(directory, entry)
    const stats = statSync(absolutePath)
    if (stats.isDirectory()) {
      if (entry === '__tests__') {
        return []
      }
      return collectSourceFiles(absolutePath)
    }
    if (!/\.(ts|vue)$/.test(entry) || /\.d\.ts$/.test(entry) || /\.test\.ts$/.test(entry)) {
      return []
    }

    const relativePath = normalize(relative(sourceRoot, absolutePath))
    return [{ absolutePath, relativePath, layer: classifyLayer(relativePath) }]
  })
}

function classifyLayer(relativePath: string): Layer {
  if (relativePath.startsWith(`components${sep}common${sep}`)) return 'common'
  if (relativePath.startsWith(`entities${sep}`)) return 'entities'
  if (relativePath.startsWith(`features${sep}`)) return 'features'
  if (relativePath.startsWith(`widgets${sep}`)) return 'widgets'
  if (relativePath.startsWith(`views${sep}`)) return 'views'
  return 'other'
}

function extractImports(source: string): string[] {
  return Array.from(source.matchAll(importPattern))
    .map((match) => match[1] ?? match[2])
    .filter(Boolean)
}

function resolveImportLayer(fromFile: SourceFile, importPath: string): Layer | null {
  if (importPath.startsWith('@/')) {
    return classifyLayer(normalize(importPath.slice(2)))
  }
  if (!importPath.startsWith('.')) {
    return null
  }

  const resolvedPath = normalize(
    relative(sourceRoot, resolve(dirname(fromFile.absolutePath), importPath))
  )
  if (resolvedPath.startsWith('..')) {
    return null
  }
  return classifyLayer(resolvedPath)
}

function collectLayerViolations(
  files: SourceFile[],
  forbiddenByLayer: Record<Layer, Layer[]>
): string[] {
  return files.flatMap((file) => {
    const forbiddenLayers = forbiddenByLayer[file.layer] ?? []
    if (forbiddenLayers.length === 0) {
      return []
    }

    const source = readFileSync(file.absolutePath, 'utf-8')
    return extractImports(source)
      .map((importPath) => ({ importPath, importedLayer: resolveImportLayer(file, importPath) }))
      .filter(
        ({ importedLayer }) => importedLayer !== null && forbiddenLayers.includes(importedLayer)
      )
      .map(
        ({ importPath, importedLayer }) =>
          `${file.relativePath} -> ${importPath} (${importedLayer})`
      )
  })
}

function collectImportKeys(files: SourceFile[], importPathPrefix: string): string[] {
  return files.flatMap((file) => {
    const source = readFileSync(file.absolutePath, 'utf-8')
    return extractImports(source)
      .filter((importPath) => importPath.startsWith(importPathPrefix))
      .map((importPath) => `${file.relativePath} -> ${importPath}`)
  })
}

function expectBaseline(
  actualEntries: string[],
  allowlist: Set<string>,
  violationLabel: string
): void {
  const violations = actualEntries.filter((key) => !allowlist.has(key))
  const staleAllowlistEntries = Array.from(allowlist).filter((key) => !actualEntries.includes(key))

  expect(violations, violationLabel).toEqual([])
  expect(staleAllowlistEntries, `${violationLabel} stale allowlist`).toEqual([])
}

describe('frontend architecture boundaries', () => {
  const sourceFiles = collectSourceFiles(sourceRoot)

  it('lower frontend layers should not import higher product layers', () => {
    const violations = collectLayerViolations(sourceFiles, {
      common: ['features', 'widgets', 'views'],
      entities: ['features', 'widgets', 'views'],
      features: ['widgets', 'views'],
      widgets: ['views'],
      views: [],
      other: [],
    })

    expect(violations).toEqual([])
  })

  it('legacy business components should not add new direct feature imports', () => {
    const componentFiles = sourceFiles.filter((file) =>
      file.relativePath.startsWith(`components${sep}`)
    )
    const featureImports = collectImportKeys(componentFiles, '@/features')
    expectBaseline(featureImports, componentFeatureImportAllowlist, 'component feature imports')
  })

  it('widgets should not add new dependencies on legacy business component directories', () => {
    const widgetFiles = sourceFiles.filter((file) => file.relativePath.startsWith(`widgets${sep}`))
    const legacyComponentImports = widgetFiles.flatMap((file) => {
      const source = readFileSync(file.absolutePath, 'utf-8')
      return extractImports(source)
        .filter((importPath) =>
          /^@\/components\/(platform|teacher|notifications|challenge|dashboard|contests|scoreboard)/.test(
            importPath
          )
        )
        .map((importPath) => `${file.relativePath} -> ${importPath}`)
    })
    expectBaseline(
      legacyComponentImports,
      widgetLegacyComponentImportAllowlist,
      'widget legacy component imports'
    )
  })

  it('new route views should stay below the page-size threshold', () => {
    const viewFiles = sourceFiles.filter((file) => file.relativePath.startsWith(`views${sep}`))
    const oversizedViews = viewFiles
      .map((file) => ({
        file: file.relativePath,
        lines: readFileSync(file.absolutePath, 'utf-8').split(/\r?\n/).length,
      }))
      .filter(({ lines }) => lines > viewLineLimit)

    const violations = oversizedViews
      .filter(({ file }) => !oversizedViewAllowlist.has(file))
      .map(({ file, lines }) => `${file} has ${lines} lines`)
    const staleAllowlistEntries = Array.from(oversizedViewAllowlist).filter(
      (file) => !oversizedViews.some((view) => view.file === file)
    )

    expect(violations).toEqual([])
    expect(staleAllowlistEntries).toEqual([])
  })

  it('components and widgets should not add new non-contract API imports', () => {
    const componentFiles = sourceFiles.filter((file) =>
      file.relativePath.startsWith(`components${sep}`)
    )
    const widgetFiles = sourceFiles.filter((file) => file.relativePath.startsWith(`widgets${sep}`))

    const componentApiImports = collectImportKeys(componentFiles, '@/api/').filter(
      (key) => !key.includes(' -> @/api/contracts')
    )
    const widgetApiImports = collectImportKeys(widgetFiles, '@/api/').filter(
      (key) => !key.includes(' -> @/api/contracts')
    )

    expectBaseline(componentApiImports, componentNonContractApiAllowlist, 'component API imports')
    expectBaseline(widgetApiImports, widgetNonContractApiAllowlist, 'widget API imports')
  })

  it('common and entity layers should stay free of app services, router, and stores', () => {
    const lowLevelFiles = sourceFiles.filter(
      (file) =>
        file.relativePath.startsWith(`components${sep}common${sep}`) ||
        file.relativePath.startsWith(`entities${sep}`)
    )
    const forbiddenImports = lowLevelFiles.flatMap((file) => {
      const source = readFileSync(file.absolutePath, 'utf-8')
      return extractImports(source)
        .filter(
          (importPath) =>
            /^@\/(api|features|widgets|views|stores|router)/.test(importPath) ||
            importPath === 'vue-router' ||
            importPath === 'pinia'
        )
        .map((importPath) => `${file.relativePath} -> ${importPath}`)
    })

    expectBaseline(forbiddenImports, commonForbiddenImportAllowlist, 'low-level forbidden imports')
  })

  it('feature UI files should not import non-contract API modules directly', () => {
    const featureUiFiles = sourceFiles.filter(
      (file) =>
        file.relativePath.startsWith(`features${sep}`) &&
        file.relativePath.includes(`${sep}ui${sep}`)
    )
    const apiImports = collectImportKeys(featureUiFiles, '@/api/').filter(
      (key) => !key.includes(' -> @/api/contracts')
    )

    expect(apiImports).toEqual([])
  })

  it('new page components should be route views or widgets instead of legacy component pages', () => {
    const componentPageFiles = sourceFiles
      .map((file) => file.relativePath)
      .filter((relativePath) => relativePath.startsWith(`components${sep}`))
      .filter((relativePath) => /Page\.vue$/.test(relativePath))

    expectBaseline(componentPageFiles, legacyComponentPageAllowlist, 'legacy component page files')
  })
})
