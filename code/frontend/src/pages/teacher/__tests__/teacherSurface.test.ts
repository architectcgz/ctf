import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import classManagementSource from '@/features/teacher/class-management/ui/ClassManagementPage.vue?raw'
import classStudentsSourceBase from '@/features/teaching/class-students-workspace/ui/ClassStudentsPage.vue?raw'
import classStudentsOverviewPanelSource from '@/features/teaching/class-students-workspace/ui/ClassStudentsOverviewPanel.vue?raw'
import classStudentsInsightWindowPanelSource from '@/features/teaching/class-students-workspace/ui/ClassStudentsInsightWindowPanel.vue?raw'
import classStudentsDirectoryPanelSource from '@/features/teaching/class-students-workspace/ui/ClassStudentsDirectoryPanel.vue?raw'
import studentAnalysisSource from '@/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue?raw'
import studentManagementSource from '@/features/teacher/student-management/ui/StudentManagementPage.vue?raw'
import instanceManagementSourceBase from '@/features/teacher/instances/ui/TeacherInstanceManagementPage.vue?raw'
import teacherInstanceDirectorySectionSource from '@/features/teacher/instances/ui/TeacherInstanceDirectorySection.vue?raw'
import teacherInstanceHeroPanelSource from '@/features/teacher/instances/ui/TeacherInstanceHeroPanel.vue?raw'
import awdReviewIndexWorkspaceSource from '@/widgets/awd-review-workspace/AwdReviewIndexWorkspace.vue?raw'
import awdReviewWorkspaceSource from '@/widgets/awd-review-workspace/AwdReviewWorkspace.vue?raw'
import awdReviewSurfaceShellSource from '@/widgets/awd-review-workspace/AwdReviewSurfaceShell.vue?raw'
import reviewArchiveSource from '@/pages/review-archive/StudentReviewArchiveRoutePage.vue?raw'
import reviewArchiveWorkspaceSource from '@/widgets/teacher-review-archive/ReviewArchiveWorkspace.vue?raw'

const classStudentsSource = [
  classStudentsSourceBase,
  classStudentsOverviewPanelSource,
  classStudentsInsightWindowPanelSource,
  classStudentsDirectoryPanelSource,
].join('\n')

const instanceManagementSource = [
  instanceManagementSourceBase,
  teacherInstanceHeroPanelSource,
  teacherInstanceDirectorySectionSource,
].join('\n')

const reviewArchiveCombinedSource = [reviewArchiveSource, reviewArchiveWorkspaceSource].join('\n')

const teacherSurfaceSource = readFileSync(
  `${process.cwd()}/src/assets/styles/teacher-surface.css`,
  'utf-8'
)

const teacherSurfacePattern =
  /--journal-ink:\s*var\(--color-text-primary\);[\s\S]*--journal-surface:\s*color-mix\(in srgb, var\(--color-bg-surface\) 88%, var\(--color-bg-base\)\);/s

const forbiddenTeacherSurfaceLiterals = ['rgba(255, 255, 255, 0.98)', '#ffffff', '#f8fafc']

const teacherSurfaceSources = [
  ['teacher-surface.css', teacherSurfaceSource],
  ['ClassStudentsPage.vue', classStudentsSource],
  ['StudentAnalysisPage.vue', studentAnalysisSource],
  ['TeacherStudentReviewArchive.vue', reviewArchiveCombinedSource],
] as const

const teacherManagementSources = [
  ['ClassManagementPage.vue', classManagementSource],
  ['StudentManagementPage.vue', studentManagementSource],
  ['TeacherInstanceManagementPage.vue', instanceManagementSource],
  ['AwdReviewSurfaceShell.vue', awdReviewSurfaceShellSource],
] as const

const teacherSurfaceForbiddenLiteralCases = teacherSurfaceSources.flatMap(([sourceName, source]) =>
  forbiddenTeacherSurfaceLiterals.map(
    (forbiddenLiteral) => [sourceName, forbiddenLiteral, source] as const
  )
)

describe('teacher surface source regression', () => {
  it.each(teacherSurfaceSources)('%s 应命中教师端 surface 主题模式', (_name, source) => {
    expect(teacherSurfacePattern.test(source)).toBe(true)
  })

  it('教师端共享 hero 外壳应统一改为直角', () => {
    expect(teacherSurfaceSource).toMatch(
      /\.teacher-surface-hero\s*\{[\s\S]*border-radius:\s*0\s*!important;/s
    )
  })

  it.each(teacherManagementSources)(
    '%s 应通过共享 teacher-management-shell 承接教师端 surface token',
    (_name, source) => {
      expect(source).toContain('teacher-management-shell')
      expect(source).toContain('workspace-shell')
      expect(source).not.toContain('--journal-ink: var(--color-text-primary);')
    }
  )

  it('AWD 复盘 widgets 应通过共享 surface shell 承接教师端外层壳', () => {
    expect(awdReviewIndexWorkspaceSource).toContain('<AwdReviewSurfaceShell')
    expect(awdReviewWorkspaceSource).toContain('<AwdReviewSurfaceShell')
  })

  it.each(teacherSurfaceForbiddenLiteralCases)(
    '%s 不应包含教师端高对比亮色 surface 硬编码: %s',
    (_name, forbiddenLiteral, source) => {
      expect(source.includes(forbiddenLiteral)).toBe(false)
    }
  )
})
