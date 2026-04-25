import { describe, expect, it } from 'vitest'

import adminChallengeDetailSource from '@/views/platform/ChallengeDetail.vue?raw'
import adminChallengeWorkspaceTabsSource from '@/components/platform/challenge/AdminChallengeWorkspaceTabs.vue?raw'

describe('route query tabs adoption', () => {
  it('admin 多 panel 页面应统一复用 useRouteQueryTabs，而不是继续在页面内手写状态机', () => {
    const tabSources = [adminChallengeDetailSource]

    for (const source of tabSources) {
      expect(source).toContain(
        "import { useRouteQueryTabs } from '@/composables/useRouteQueryTabs'"
      )
      expect(source).toContain('setTabButtonRef')
      expect(source).not.toContain('function handleTabKeydown(')
    }
    expect(adminChallengeWorkspaceTabsSource).toContain(
      ':ref="(element) => setTabButtonRef(tab.key, element as HTMLButtonElement | null)"'
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
