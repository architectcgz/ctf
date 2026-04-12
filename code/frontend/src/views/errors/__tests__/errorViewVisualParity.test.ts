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
    expect(forbiddenSource).toContain('<ErrorStatusShell')
    expect(notFoundSource).toContain('<ErrorStatusShell')
  })
})
