import { describe, expect, it } from 'vitest'

import challengeDetailSource from '../ChallengeDetail.vue?raw'

describe('challenge detail shared shell alignment', () => {
  it('通过变量接入共享 workspace shell，而不是继续本地重写整套壳层样式', () => {
    expect(challengeDetailSource).toContain('--workspace-shell-border: var(--journal-line-soft);')
    expect(challengeDetailSource).toContain('--workspace-shell-bg: var(--bg-shell);')
    expect(challengeDetailSource).toContain('--workspace-shadow-shell: var(--journal-shadow);')
    expect(challengeDetailSource).toContain('min-height: max(100%, calc(100vh - 5rem));')
    expect(challengeDetailSource).not.toMatch(
      /\.workspace-shell\s*\{[\s\S]*border:\s*1px solid var\(--journal-line-soft\);/s
    )
    expect(challengeDetailSource).not.toMatch(
      /\.workspace-shell\s*\{[\s\S]*box-shadow:\s*var\(--journal-shadow\);/s
    )
    expect(challengeDetailSource).not.toMatch(
      /:global\(\[data-theme='dark'\]\) \.workspace-shell\s*\{[^}]*background:/s
    )
    expect(challengeDetailSource).toContain('--workspace-shell-radial-strength: 14%;')
    expect(challengeDetailSource).toContain('--workspace-shell-radial-size: 24rem;')
  })

  it('主标签轨道应通过 page-tabs 变量接入共享规则，而不是继续本地声明 top-tabs/top-tab', () => {
    expect(challengeDetailSource).toContain('--page-top-tabs-gap: var(--space-7);')
    expect(challengeDetailSource).toContain('--page-top-tabs-padding: 0 var(--space-7);')
    expect(challengeDetailSource).toContain('--page-top-tab-active-border: var(--brand);')
    expect(challengeDetailSource).not.toMatch(/^\.top-tabs\s*,/m)
    expect(challengeDetailSource).not.toMatch(/^\.top-tab\s*,/m)
  })
})
