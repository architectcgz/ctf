import { describe, expect, it } from 'vitest'

import cheatDetectionSource from '../CheatDetection.vue?raw'

describe('cheat detection surface alignment', () => {
  it('softens risk card borders instead of using the full journal border contrast', () => {
    expect(cheatDetectionSource).toMatch(
      /--cheat-card-border:\s*color-mix\(in srgb,\s*var\(--journal-border\) 74%, transparent\);/,
    )
    expect(cheatDetectionSource).toMatch(
      /\.risk-row,\s*\.quick-action-row\s*\{[\s\S]*border:\s*1px solid var\(--cheat-card-border\);/s,
    )
    expect(cheatDetectionSource).not.toMatch(
      /\.risk-row,\s*\.quick-action-row\s*\{[\s\S]*border:\s*1px solid var\(--journal-border\);/s,
    )
  })

  it('uses a softer section divider so the content bands do not look boxed in', () => {
    expect(cheatDetectionSource).toMatch(
      /--cheat-divider:\s*color-mix\(in srgb,\s*var\(--journal-border\) 68%, transparent\);/,
    )
    expect(cheatDetectionSource).toMatch(
      /\.journal-divider\s*\{[\s\S]*border-top:\s*1px dashed var\(--cheat-divider\);/s,
    )
    expect(cheatDetectionSource).not.toMatch(
      /\.journal-divider\s*\{[^}]*88%, transparent\);/s,
    )
  })

  it('overrides the empty state top and bottom borders instead of inheriting AppEmpty bright border-y lines', () => {
    expect(cheatDetectionSource).toMatch(/<AppEmpty[\s\S]*class="cheat-empty-state"[\s\S]*title="当前没有超过阈值的高频提交账号"/s)
    expect(cheatDetectionSource).toMatch(/<AppEmpty[\s\S]*class="cheat-empty-state"[\s\S]*title="当前没有共享 IP 线索"/s)
    expect(cheatDetectionSource).toMatch(
      /\.cheat-empty-state\s*\{[\s\S]*border-top-color:\s*var\(--cheat-divider\);[\s\S]*border-bottom-color:\s*var\(--cheat-divider\);/s,
    )
  })
})
