import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import awdChallengeConfigPanelSource from '@/features/platform/contests/ui/AWDChallengeConfigPanel.vue?raw'
import awdChallengeConfigDirectoryRowSource from '@/features/platform/contests/ui/AWDChallengeConfigDirectoryRow.vue?raw'
import awdChallengeConfigDirectorySectionSource from '@/features/platform/contests/ui/AWDChallengeConfigDirectorySection.vue?raw'
import awdChallengeConfigHeaderSource from '@/features/platform/contests/ui/AWDChallengeConfigHeader.vue?raw'
import awdReadinessOverrideDialogSource from '@/features/awd-readiness/ui/AWDReadinessOverrideDialog.vue?raw'
import awdReadinessChecklistSource from '@/features/awd-readiness/ui/AWDReadinessChecklist.vue?raw'

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

describe('contest ui primitive adoption phase 25', () => {
  it('awd challenge config panel should use the full shared metric panel class stack', () => {
    expect(awdChallengeConfigCombinedSource).toContain(
      'class="progress-strip metric-panel-grid metric-panel-default-surface"'
    )
    expect(awdChallengeConfigCombinedSource).toContain(
      'class="journal-note progress-card metric-panel-card"'
    )
    expect(awdChallengeConfigCombinedSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(awdChallengeConfigCombinedSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(awdChallengeConfigCombinedSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
  })

  it('awd readiness checklist should use the full shared metric panel class stack', () => {
    expect(awdReadinessChecklistSource).toContain(
      'class="progress-strip metric-panel-grid metric-panel-default-surface readiness-summary-grid"'
    )
    expect(awdReadinessChecklistSource).toContain(
      'class="journal-note progress-card metric-panel-card"'
    )
    expect(awdReadinessChecklistSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(awdReadinessChecklistSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(awdReadinessChecklistSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
  })

  it('awd readiness override dialog should use the full shared metric panel class stack', () => {
    expect(awdReadinessOverrideDialogSource).toContain(
      'class="progress-strip metric-panel-grid metric-panel-default-surface readiness-override-summary"'
    )
    expect(awdReadinessOverrideDialogSource).toContain(
      'class="journal-note progress-card metric-panel-card"'
    )
    expect(awdReadinessOverrideDialogSource).toContain(
      'class="journal-note-label progress-card-label metric-panel-label"'
    )
    expect(awdReadinessOverrideDialogSource).toContain(
      'class="journal-note-value progress-card-value metric-panel-value"'
    )
    expect(awdReadinessOverrideDialogSource).toContain(
      'class="journal-note-helper progress-card-hint metric-panel-helper"'
    )
  })
})
