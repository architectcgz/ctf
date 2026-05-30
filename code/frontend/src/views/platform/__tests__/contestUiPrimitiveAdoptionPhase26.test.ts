import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import awdChallengeConfigPanelSource from '@/features/platform/contests/ui/AWDChallengeConfigPanel.vue?raw'
import awdChallengeConfigDirectoryRowSource from '@/features/platform/contests/ui/AWDChallengeConfigDirectoryRow.vue?raw'
import awdChallengeConfigDirectorySectionSource from '@/features/platform/contests/ui/AWDChallengeConfigDirectorySection.vue?raw'
import awdChallengeConfigHeaderSource from '@/features/platform/contests/ui/AWDChallengeConfigHeader.vue?raw'
import awdReadinessSummarySource from '@/features/awd-readiness/ui/AWDReadinessSummary.vue?raw'

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

describe('contest ui primitive adoption phase 26', () => {
  it('awd challenge config panel should use workspace overline in the active tab panel header', () => {
    expect(awdChallengeConfigCombinedSource).toMatch(
      /<div class="workspace-overline">\s*AWD Service Config\s*<\/div>/
    )
    expect(awdChallengeConfigCombinedSource).not.toContain(
      '<div class="journal-eyebrow">AWD Service Config</div>'
    )
  })

  it('awd readiness summary should use workspace overline in the active tab panel header', () => {
    expect(awdReadinessSummarySource).toMatch(
      /<div class="workspace-overline">\s*AWD Readiness\s*<\/div>/
    )
    expect(awdReadinessSummarySource).not.toContain(
      '<div class="journal-eyebrow">AWD Readiness</div>'
    )
  })
})
