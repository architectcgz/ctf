import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import studentDifficultySource from '@/features/student-dashboard/ui/StudentDifficultyContent.vue?raw'
import trainingTimelineSource from '@/entities/training-timeline/ui/TrainingTimelinePanel.vue?raw'
import studentOverviewSource from '@/features/student-dashboard/ui/StudentOverviewContent.vue?raw'
import studentRecommendationSource from '@/features/student-dashboard/ui/StudentRecommendationContent.vue?raw'
import studentCategoryProgressSource from '@/features/student-dashboard/ui/StudentCategoryProgressContent.vue?raw'
import instanceListWorkspaceShellSource from '@/features/instance-list/ui/InstanceListWorkspaceShell.vue?raw'
import challengePresentationSource from '@/entities/challenge/model/presentation.ts?raw'
import instanceListSource from '@/pages/instances/InstanceListRoutePage.vue?raw'
import notificationListSource from '@/pages/notifications/NotificationListRoutePage.vue?raw'
import notificationListWorkspaceSourceBase from '@/widgets/notification-list-workspace/NotificationListWorkspace.vue?raw'

const themeSource = readFileSync(`${process.cwd()}/src/assets/styles/theme.css`, 'utf-8')
const styleSource = readFileSync(`${process.cwd()}/src/style.css`, 'utf-8')
const journalSoftSurfacesSource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-soft-surfaces.css`,
  'utf-8'
)
const journalUserShellSource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-user-shell.css`,
  'utf-8'
)

const instanceListWorkspaceSource = `${instanceListSource}\n${instanceListWorkspaceShellSource}`
const notificationListWorkspaceSource = `${notificationListSource}\n${notificationListWorkspaceSourceBase}`

describe('student and user surface alignment', () => {
  it('student dashboard panel headers 应复用共享 header 与 divider 原语', () => {
    expect(themeSource).toContain('--space-workspace-panel-title-gap:')
    expect(themeSource).toContain('--space-workspace-panel-copy-gap:')
    expect(styleSource).toContain('.workspace-panel-header__intro')
    expect(styleSource).toContain('.workspace-panel-divider')
    expect(studentOverviewSource).toContain('class="workspace-panel-divider"')
    expect(studentRecommendationSource).toContain('workspace-panel-header')
    expect(studentCategoryProgressSource).toContain('workspace-panel-header')
    expect(studentDifficultySource).toContain('workspace-panel-header')
  })

  it('student dashboard panels should continue to use shared journal soft surface owner', () => {
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-shell')
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-soft-panel-shell')
    expect(journalSoftSurfacesSource).toContain('.journal-soft-surface .journal-soft-panel-item')
    expect(studentRecommendationSource).toContain('journal-soft-surface')
    expect(studentCategoryProgressSource).toContain('journal-soft-surface')
    expect(studentDifficultySource).toContain('journal-soft-surface')
    expect(trainingTimelineSource).toContain('journal-soft-surface')
  })

  it('student recommendation 与难度/分类面板应继续通过 challenge entity 承接胶囊与映射 owner', () => {
    expect(challengePresentationSource).toContain('var(--challenge-category-pill-web)')
    expect(challengePresentationSource).toContain('var(--challenge-difficulty-pill-easy)')
    expect(studentRecommendationSource).toContain("from '@/entities/challenge'")
    expect(studentRecommendationSource).toContain('<ChallengeCategoryDifficultyPills')
    expect(studentRecommendationSource).toContain('toChallengeCategory')
    expect(studentRecommendationSource).not.toContain('categoryPillStyle(item.category)')
    expect(studentDifficultySource).toContain('getChallengeDifficultyColor')
    expect(studentCategoryProgressSource).toContain('category-action')
  })

  it('instance 与 notification 页面应继续通过 journal-shell-user 承接用户侧 surface', () => {
    expect(journalUserShellSource).toContain('.journal-shell.journal-shell-user')
    expect(instanceListWorkspaceSource).toContain('journal-shell-user')
    expect(notificationListWorkspaceSource).toContain('journal-shell-user')
    expect(instanceListWorkspaceSource).not.toContain('border-[var(--journal-border)]')
    expect(notificationListWorkspaceSource).not.toContain('rgba(148, 163, 184, 0.58)')
  })
})
