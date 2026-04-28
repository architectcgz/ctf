import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

const themeSource = readFileSync(`${process.cwd()}/src/assets/styles/theme.css`, 'utf-8')
const workspaceShellSource = readFileSync(
  `${process.cwd()}/src/assets/styles/workspace-shell.css`,
  'utf-8'
)
const adminShellSource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-admin-shell.css`,
  'utf-8'
)
const userShellSource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-user-shell.css`,
  'utf-8'
)

describe('misc design style system', () => {
  it('should define shared tokens for controls, dialogs, badges, and action groups', () => {
    expect(themeSource).toContain('--ui-control-height-md: 2.75rem;')
    expect(themeSource).toContain('--ui-dialog-wide-width: 72rem;')
    expect(themeSource).toContain('--ui-badge-radius-pill: 999px;')
    expect(themeSource).toContain('--ui-action-gap: 0.5rem;')
  })

  it('should expose shared class primitives for buttons, fields, controls, badges, dialogs, and action groups', () => {
    for (const selector of [
      '.ui-btn',
      '.ui-btn--primary',
      '.ui-field',
      '.ui-control-wrap',
      '.ui-badge',
      '.ui-confirm-panel',
      '.ui-workbench-modal',
      '.ui-toolbar-actions',
      '.ui-row-actions',
      '.ui-row-actions--fixed',
      '.ui-row-action--main',
      '.ui-row-action--default',
      '.ui-row-action--menu',
      '.ui-card-actions',
    ]) {
      expect(workspaceShellSource).toContain(selector)
    }

    expect(workspaceShellSource).toContain('--ui-row-action-fixed-width')
    expect(workspaceShellSource).toContain(
      'grid-template-columns: var(--ui-row-action-main-width) var(--ui-row-action-button-width) var(--ui-row-action-menu-width);'
    )
    expect(workspaceShellSource).toContain('.ui-workbench-modal__nav-button.is-active')
    expect(workspaceShellSource).toContain('.ui-card-action.is-danger:hover')
  })

  it('should split admin and student visual variants into dedicated shell style files', () => {
    expect(adminShellSource).toContain('--ui-btn-primary-background: var(--color-primary);')
    expect(adminShellSource).toContain('--ui-btn-primary-hover-background: var(--color-primary-hover);')
    expect(adminShellSource).toContain('--ui-control-focus-border: #3b82f6;')
    expect(adminShellSource).toContain('--ui-dialog-radius-wide: 1.75rem;')

    expect(userShellSource).toContain('--ui-btn-primary-background: var(--color-primary);')
    expect(userShellSource).toContain('--ui-btn-primary-hover-background: var(--color-primary-hover);')
    expect(userShellSource).toContain('--ui-control-focus-border: #2a7a58;')
    expect(userShellSource).toContain('--ui-badge-radius: var(--ui-badge-radius-pill);')
  })
})
