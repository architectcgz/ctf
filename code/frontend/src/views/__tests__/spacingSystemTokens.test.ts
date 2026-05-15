import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

const themeSource = readFileSync(`${process.cwd()}/src/assets/styles/theme.css`, 'utf-8')
const workspaceShellSource = readFileSync(
  `${process.cwd()}/src/assets/styles/workspace-shell.css`,
  'utf-8'
)
const pageTabsSource = readFileSync(`${process.cwd()}/src/assets/styles/page-tabs.css`, 'utf-8')
const teacherSurfaceSource = readFileSync(
  `${process.cwd()}/src/assets/styles/teacher-surface.css`,
  'utf-8'
)
const journalNotesSource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-notes.css`,
  'utf-8'
)
const teacherAwdReviewSurfaceShellSource = readFileSync(
  `${process.cwd()}/src/widgets/teacher-awd-review/TeacherAWDReviewSurfaceShell.vue`,
  'utf-8'
)

describe('spacing system tokens', () => {
  it('should define global spacing scale and semantic spacing tokens', () => {
    expect(themeSource).toContain('--space-0: 0;')
    expect(themeSource).toContain('--space-2: 0.5rem;')
    expect(themeSource).toContain('--space-5-5: 1.375rem;')
    expect(themeSource).toContain('--space-workspace-side-padding: var(--space-7);')
    expect(themeSource).toContain('--space-workspace-content-padding: var(--space-7);')
    expect(themeSource).toContain('--space-workspace-content-start-padding-top: var(--space-4);')
    expect(themeSource).toContain('--space-workspace-tabs-panel-gap: var(--space-3-5);')
    expect(themeSource).toContain(
      '--space-workspace-panel-title-gap: var(--workspace-page-title-margin-top);'
    )
    expect(themeSource).toContain(
      '--space-workspace-panel-copy-gap: var(--workspace-page-copy-margin-top);'
    )
    expect(themeSource).toContain('--space-workspace-panel-block-gap: var(--space-5);')
    expect(themeSource).toContain('--space-workspace-panel-divider-gap: var(--space-1);')
    expect(themeSource).toContain('--space-divider-gap: var(--space-4);')
  })

  it('should make workspace shell use semantic spacing tokens instead of fixed pixel values', () => {
    expect(workspaceShellSource).toContain('--workspace-topbar-gap,')
    expect(workspaceShellSource).toContain('var(--space-workspace-topbar-padding-top)')
    expect(workspaceShellSource).toContain('var(--space-workspace-side-padding)')
    expect(workspaceShellSource).toContain('var(--space-workspace-tabs-gap)')
    expect(workspaceShellSource).toContain('var(--space-workspace-content-padding)')
    expect(workspaceShellSource).toContain('var(--space-workspace-content-start-padding-top)')
    expect(workspaceShellSource).toContain(
      'var(--workspace-tabs-panel-gap, var(--space-workspace-tabs-panel-gap))'
    )
    expect(workspaceShellSource).toContain('.workspace-shell > .content-pane:first-child')
    expect(workspaceShellSource).not.toContain('.workspace-shell > .workspace-grid')
    expect(workspaceShellSource).not.toContain('padding: 22px 28px 0;')
    expect(workspaceShellSource).not.toContain('gap: 28px;')
    expect(workspaceShellSource).not.toContain('padding: 28px;')
  })

  it('should make tab rail and teacher shared surfaces inherit spacing tokens', () => {
    expect(pageTabsSource).toContain('var(--space-workspace-tabs-gap)')
    expect(pageTabsSource).toContain('var(--space-workspace-tab-padding-top)')
    expect(pageTabsSource).toContain('var(--space-workspace-tab-padding-bottom)')

    expect(teacherSurfaceSource).toContain('var(--space-section-gap')
    expect(teacherSurfaceSource).toContain('var(--space-divider-gap')
    expect(teacherSurfaceSource).toContain('var(--space-5)')

    expect(journalNotesSource).toContain('var(--space-divider-gap')
    expect(teacherAwdReviewSurfaceShellSource).toContain(
      'var(--workspace-content-start-padding-top, var(--space-workspace-content-start-padding-top))'
    )
    expect(teacherAwdReviewSurfaceShellSource).toContain(
      'var(--workspace-side-padding, var(--space-workspace-side-padding))'
    )
    expect(teacherAwdReviewSurfaceShellSource).toContain(
      'var(--workspace-content-padding, var(--space-workspace-content-padding))'
    )
  })
})
