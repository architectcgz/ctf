import { describe, expect, it } from 'vitest'

import contestEditSource from '../ContestEdit.vue?raw'
import contestEditWorkspacePanelSource from '@/components/platform/contest/ContestEditWorkspacePanel.vue?raw'
import awdChallengeConfigDialogSource from '@/components/platform/contest/AWDChallengeConfigDialog.vue?raw'
import awdReadinessOverrideDialogSource from '@/components/platform/contest/AWDReadinessOverrideDialog.vue?raw'
import awdReadinessSummarySource from '@/components/platform/contest/AWDReadinessSummary.vue?raw'
import contestChallengeOrchestrationPanelSource from '@/components/platform/contest/ContestChallengeOrchestrationPanel.vue?raw'

describe('contest ui primitive adoption phase 2', () => {
  const contestEditCombinedSource = [contestEditSource, contestEditWorkspacePanelSource].join('\n')

  it('contest edit workspace should consume shared ui buttons for back and retry actions', () => {
    expect(contestEditCombinedSource).toContain('class="ui-btn ui-btn--ghost"')
    expect(contestEditCombinedSource).not.toContain('class="admin-btn admin-btn-ghost"')
  })

  it('contest edit workspace should keep the inner scroll container shrinkable so basics panel can scroll', () => {
    expect(contestEditSource).toMatch(
      /\.studio-content\s*\{[^}]*min-height:\s*0;[^}]*\}/
    )
    expect(contestEditWorkspacePanelSource).toMatch(
      /\.studio-scroll-area\s*\{[^}]*overflow-y:\s*auto;[^}]*\}/
    )
  })

  it('contest edit source should not contain truncated placeholder fragments', () => {
    expect(contestEditSource).not.toContain('\n...\n')
  })

  it('contest challenge orchestration panel should consume shared ui buttons and row actions', () => {
    expect(contestChallengeOrchestrationPanelSource).toContain(
      "from '@/components/common/menus/CActionMenu.vue'"
    )
    expect(contestChallengeOrchestrationPanelSource).toContain('class="ui-btn ui-btn--ghost"')
    expect(contestChallengeOrchestrationPanelSource).toContain('class="ui-btn ui-btn--primary"')
    expect(contestChallengeOrchestrationPanelSource).toContain(
      'class="ui-row-actions contest-challenge-row__actions"'
    )
    expect(contestChallengeOrchestrationPanelSource).toContain(
      'class="c-action-menu__trigger c-action-menu__trigger--icon"'
    )
    expect(contestChallengeOrchestrationPanelSource).toContain(
      'class="c-action-menu__item c-action-menu__item--danger"'
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
