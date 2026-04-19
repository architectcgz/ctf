import { describe, expect, it } from 'vitest'

import awdReadinessSummarySource from '@/components/platform/contest/AWDReadinessSummary.vue?raw'

describe('contest ui primitive adoption phase 19', () => {
  it('awd readiness summary should use a header element for readiness decision intro block', () => {
    expect(awdReadinessSummarySource).toContain(
      '<header class="list-heading readiness-decision__head">'
    )
    expect(awdReadinessSummarySource).not.toContain(`>
        <div>
          <div class="journal-note-label">Start Decision</div>`)
  })
})
