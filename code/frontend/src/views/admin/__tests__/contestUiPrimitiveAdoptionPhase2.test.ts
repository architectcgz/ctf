import { describe, expect, it } from 'vitest'

import contestEditSource from '../ContestEdit.vue?raw'
import awdChallengeConfigDialogSource from '@/components/admin/contest/AWDChallengeConfigDialog.vue?raw'
import awdReadinessOverrideDialogSource from '@/components/admin/contest/AWDReadinessOverrideDialog.vue?raw'
import awdReadinessSummarySource from '@/components/admin/contest/AWDReadinessSummary.vue?raw'
import contestChallengeOrchestrationPanelSource from '@/components/admin/contest/ContestChallengeOrchestrationPanel.vue?raw'

describe('contest ui primitive adoption phase 2', () => {
  it('contest edit workspace should consume shared ui buttons for back and retry actions', () => {
    expect(contestEditSource).toContain('class="ui-btn ui-btn--ghost"')
    expect(contestEditSource).not.toContain('class="admin-btn admin-btn-ghost"')
  })

  it('contest challenge orchestration panel should consume shared ui buttons and row actions', () => {
    expect(contestChallengeOrchestrationPanelSource).toContain('class="ui-btn ui-btn--ghost"')
    expect(contestChallengeOrchestrationPanelSource).toContain('class="ui-btn ui-btn--primary"')
    expect(contestChallengeOrchestrationPanelSource).toContain(
      'class="ui-row-actions contest-challenge-row__actions"'
    )
    expect(contestChallengeOrchestrationPanelSource).toContain(
      'class="ui-btn ui-btn--sm ui-btn--secondary contest-challenge-row__button'
    )
    expect(contestChallengeOrchestrationPanelSource).toContain(
      'class="ui-btn ui-btn--sm ui-btn--danger contest-challenge-row__button'
    )
    expect(contestChallengeOrchestrationPanelSource).not.toContain('class="admin-btn')
  })

  it('awd readiness summary should consume shared badges and action primitives', () => {
    expect(awdReadinessSummarySource).toContain('class="ui-badge readiness-status-chip')
    expect(awdReadinessSummarySource).toContain('class="ui-row-actions readiness-row__actions"')
    expect(awdReadinessSummarySource).toContain('class="ui-btn ui-btn--sm ui-btn--secondary"')
  })

  it('awd readiness override dialog should consume shared field and button primitives', () => {
    expect(awdReadinessOverrideDialogSource).toContain(
      'class="ui-field readiness-override-form"'
    )
    expect(awdReadinessOverrideDialogSource).toContain('class="ui-control-wrap"')
    expect(awdReadinessOverrideDialogSource).toContain('class="ui-control readiness-override-textarea"')
    expect(awdReadinessOverrideDialogSource).toContain(
      'class="ui-btn ui-btn--secondary readiness-override-footer__button"'
    )
    expect(awdReadinessOverrideDialogSource).toContain(
      'class="ui-btn ui-btn--primary readiness-override-footer__button"'
    )
  })

  it('awd challenge config dialog should adopt admin modal shell and shared form primitives', () => {
    expect(awdChallengeConfigDialogSource).toContain(
      "from '@/components/common/modal-templates/AdminSurfaceModal.vue'"
    )
    expect(awdChallengeConfigDialogSource).toContain('<AdminSurfaceModal')
    expect(awdChallengeConfigDialogSource).not.toContain('<ElDialog')
    expect(awdChallengeConfigDialogSource).toContain('class="ui-field awd-config-field')
    expect(awdChallengeConfigDialogSource).toContain('class="ui-control-wrap"')
    expect(awdChallengeConfigDialogSource).toContain('class="ui-control"')
    expect(awdChallengeConfigDialogSource).toContain('class="ui-btn ui-btn--secondary"')
    expect(awdChallengeConfigDialogSource).toContain('class="ui-btn ui-btn--primary"')
  })
})
