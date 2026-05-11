import { describe, expect, it } from 'vitest'

import shellSource from '@/components/errors/ErrorStatusShell.vue?raw'
import forbiddenSource from '../ForbiddenView.vue?raw'
import notFoundSource from '../NotFoundView.vue?raw'

describe('error view visual parity', () => {
  it('keeps status pages on a shared visual shell', () => {
    expect(shellSource).toContain('<h1 class="error-status-title workspace-page-title">')
    expect(shellSource).toContain('<p class="error-status-text workspace-page-copy">')
    expect(shellSource).not.toMatch(/\.error-status-title\s*{[^}]*font-size:/s)
    expect(shellSource).not.toMatch(/\.error-status-title\s*{[^}]*line-height:/s)
    expect(shellSource).not.toMatch(/\.error-status-title\s*{[^}]*letter-spacing:/s)
    expect(shellSource).toContain('class="ui-btn ui-btn--primary error-status-action"')
    expect(shellSource).toContain('class="ui-btn ui-btn--secondary error-status-action"')
    expect(shellSource).not.toContain('error-status-action-primary')
    expect(shellSource).not.toContain('error-status-action-secondary')
    expect(shellSource).toMatch(
      /:global\(\[data-theme='dark'\]\) \.error-status-view\s*\{[\s\S]*--ui-btn-primary-hover-border:\s*color-mix\(/s
    )
    expect(forbiddenSource).toContain('<ErrorStatusShell')
    expect(notFoundSource).toContain('<ErrorStatusShell')
  })
})
