import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

import cheatDetectionSource from '../CheatDetection.vue?raw'

const journalNotesSource = readFileSync(
  `${process.cwd()}/src/assets/styles/journal-notes.css`,
  'utf-8'
)
const pageTabsSource = readFileSync(`${process.cwd()}/src/assets/styles/page-tabs.css`, 'utf-8')

describe('cheat detection surface alignment', () => {
  it('softens risk card borders instead of using the full journal border contrast', () => {
    expect(cheatDetectionSource).toMatch(
      /--cheat-card-border:\s*color-mix\(in srgb,\s*var\(--journal-border\) 74%, transparent\);/
    )
    expect(cheatDetectionSource).toMatch(
      /\.risk-row,\s*\.quick-action-row\s*\{[\s\S]*border:\s*1px solid var\(--cheat-card-border\);/s
    )
    expect(cheatDetectionSource).not.toMatch(
      /\.risk-row,\s*\.quick-action-row\s*\{[\s\S]*border:\s*1px solid var\(--journal-border\);/s
    )
  })

  it('uses a softer section divider so the content bands do not look boxed in', () => {
    expect(cheatDetectionSource).toMatch(
      /--cheat-divider:\s*color-mix\(in srgb,\s*var\(--journal-border\) 68%, transparent\);/
    )
    expect(cheatDetectionSource).toContain(
      '--journal-divider-border: 1px dashed var(--cheat-divider);'
    )
    expect(journalNotesSource).toMatch(
      /\.journal-shell-admin \.journal-divider\s*\{[\s\S]*border-top:\s*var\(/s
    )
    expect(cheatDetectionSource).not.toMatch(/\.journal-divider\s*\{/s)
  })

  it('overrides the empty state top and bottom borders instead of inheriting AppEmpty bright border-y lines', () => {
    expect(cheatDetectionSource).toMatch(
      /<AppEmpty[\s\S]*class="cheat-empty-state"[\s\S]*title="当前没有超过阈值的高频提交账号"/s
    )
    expect(cheatDetectionSource).toMatch(
      /<AppEmpty[\s\S]*class="cheat-empty-state"[\s\S]*title="当前没有共享 IP 线索"/s
    )
    expect(cheatDetectionSource).toMatch(
      /\.cheat-empty-state\s*\{[\s\S]*border-top-color:\s*var\(--cheat-divider\);[\s\S]*border-bottom-color:\s*var\(--cheat-divider\);/s
    )
  })

  it('aligns the tab rail with the teacher dashboard underline style instead of pill tabs', () => {
    expect(pageTabsSource).toMatch(
      /\.top-tabs\s*\{[\s\S]*overflow-x:\s*auto;[\s\S]*scrollbar-width:\s*none;/s
    )
    expect(pageTabsSource).toMatch(
      /\.top-tab\s*\{[\s\S]*border-bottom:\s*2px solid transparent;[\s\S]*white-space:\s*nowrap;/s
    )
    expect(pageTabsSource).toContain('--page-top-tab-active-border,')
    expect(pageTabsSource).toMatch(
      /\.top-tab:hover,\s*\.top-tab.active,\s*\.top-tab:focus-visible\s*\{[\s\S]*border-bottom-color:\s*var\(/s
    )
    expect(cheatDetectionSource).toContain('--page-top-tabs-gap: 28px;')
    expect(cheatDetectionSource).toContain('--page-top-tab-font-size: 15px;')
    expect(cheatDetectionSource).toContain('--page-top-tab-active-border:')
    expect(cheatDetectionSource).not.toMatch(/\.top-tabs\s*\{[^}]*display:\s*flex;/s)
    expect(cheatDetectionSource).not.toMatch(
      /\.top-tab\s*\{[^}]*border-bottom:\s*2px solid transparent;/s
    )
  })
})
