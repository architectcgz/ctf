import { describe, expect, it } from 'vitest'

import userGovernanceSource from '@/components/platform/user/UserGovernancePage.vue?raw'
import cheatDetectionSource from '@/views/platform/CheatDetection.vue?raw'
import adminChallengeDetailSource from '@/views/platform/ChallengeDetail.vue?raw'

describe('route query tabs adoption', () => {
  it('admin 多 panel 页面应统一复用 useRouteQueryTabs，而不是继续在页面内手写状态机', () => {
    const tabSources = [userGovernanceSource, cheatDetectionSource, adminChallengeDetailSource]

    for (const source of tabSources) {
      expect(source).toContain(
        "import { useRouteQueryTabs } from '@/composables/useRouteQueryTabs'"
      )
      expect(source).toContain('setTabButtonRef')
      expect(source).toContain(
        ':ref="(element) => setTabButtonRef(tab.key, element as HTMLButtonElement | null)"'
      )
      expect(source).not.toContain('function handleTabKeydown(')
    }

    expect(userGovernanceSource).toContain('useRouteQueryTabs<UserPanelKey>({')
    expect(userGovernanceSource).not.toContain('const activePanel = computed<UserPanelKey>(() => {')
    expect(userGovernanceSource).not.toContain(
      'async function switchPanel(panelKey: UserPanelKey): Promise<void> {'
    )

    expect(cheatDetectionSource).toContain('useRouteQueryTabs<CheatPanelKey>({')
    expect(cheatDetectionSource).not.toContain(
      'const activePanel = computed<CheatPanelKey>(() => {'
    )
    expect(cheatDetectionSource).not.toContain(
      'async function switchPanel(panelKey: CheatPanelKey): Promise<void> {'
    )

    expect(adminChallengeDetailSource).toContain('useRouteQueryTabs<ChallengePanelKey>({')
    expect(adminChallengeDetailSource).not.toContain(
      'const activePanel = computed<ChallengePanelKey>(() => resolvePanel(route.query.panel))'
    )
    expect(adminChallengeDetailSource).not.toContain(
      'function handleTabKeydown(event: KeyboardEvent, index: number): void {'
    )
  })
})
