import { describe, expect, it } from 'vitest'

import awdChallengeConfigPanelSource from '@/components/platform/contest/AWDChallengeConfigPanel.vue?raw'

describe('contest ui primitive adoption phase 20', () => {
  it('awd challenge config panel should no longer render the active challenge focus card block', () => {
    expect(awdChallengeConfigPanelSource).not.toContain('config-focus-card')
    expect(awdChallengeConfigPanelSource).not.toContain('当前焦点题目')
  })
})
