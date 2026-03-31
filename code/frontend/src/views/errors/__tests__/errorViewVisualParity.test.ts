import { describe, expect, it } from 'vitest'

import shellSource from '@/components/errors/ErrorStatusShell.vue?raw'
import forbiddenSource from '../ForbiddenView.vue?raw'
import notFoundSource from '../NotFoundView.vue?raw'

describe('error view visual parity', () => {
  it('keeps status pages on a shared visual shell', () => {
    expect(shellSource).toMatch(
      /\.error-status-title\s*{[^}]*font-size:\s*clamp\(1\.7rem,\s*3\.2vw,\s*2\.35rem\);[^}]*line-height:\s*1\.18;[^}]*letter-spacing:\s*-0\.02em;/s
    )
    expect(shellSource).not.toMatch(/\.error-status-title\s*{[^}]*margin-top:/s)
    expect(forbiddenSource).toContain('<ErrorStatusShell')
    expect(notFoundSource).toContain('<ErrorStatusShell')
  })
})
