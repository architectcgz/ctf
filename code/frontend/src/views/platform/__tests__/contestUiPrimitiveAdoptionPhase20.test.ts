import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import awdChallengeConfigPanelSource from '@/features/platform/contests/ui/AWDChallengeConfigPanel.vue?raw'
import awdChallengeConfigDirectoryRowSource from '@/features/platform/contests/ui/AWDChallengeConfigDirectoryRow.vue?raw'
import awdChallengeConfigDirectorySectionSource from '@/features/platform/contests/ui/AWDChallengeConfigDirectorySection.vue?raw'
import awdChallengeConfigHeaderSource from '@/features/platform/contests/ui/AWDChallengeConfigHeader.vue?raw'

const awdChallengeConfigCombinedSource = [
  awdChallengeConfigPanelSource,
  awdChallengeConfigHeaderSource,
  awdChallengeConfigDirectorySectionSource,
  awdChallengeConfigDirectoryRowSource,
  readFileSync(
    resolve(process.cwd(), 'src/features/platform/contests/ui/awdChallengeConfigPanel.css'),
    'utf8'
  ),
].join('\n')

describe('contest ui primitive adoption phase 20', () => {
  it('awd challenge config panel should no longer render the active challenge focus card block', () => {
    expect(awdChallengeConfigCombinedSource).not.toContain('config-focus-card')
    expect(awdChallengeConfigCombinedSource).not.toContain('当前焦点题目')
  })
})
