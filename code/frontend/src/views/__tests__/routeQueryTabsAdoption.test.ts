import { describe, expect, it } from 'vitest'

import adminChallengeDetailSource from '@/pages/platform/challenges/ChallengeDetailRoutePage.vue?raw'
import adminChallengeRoutePageSource from '@/features/platform/challenge-detail/model/usePlatformChallengeDetailRoutePage.ts?raw'
import adminChallengeWorkspaceTabsSource from '@/features/platform/challenge-detail/ui/AdminChallengeWorkspaceTabs.vue?raw'

describe('route query tabs adoption', () => {
  it('admin 多 panel 页面应统一通过 feature route model 复用 useRouteQueryTabs，而不是继续在 route view 内手写状态机', () => {
    expect(adminChallengeRoutePageSource).toContain(
      "import { useRouteQueryTabs } from '@/composables/useRouteQueryTabs'"
    )
    expect(adminChallengeRoutePageSource).toContain('setTabButtonRef: tabState.setTabButtonRef')
    expect(adminChallengeRoutePageSource).toContain('handleTabKeydown: tabState.handleTabKeydown')
    expect(adminChallengeRoutePageSource).toContain('useRouteQueryTabs<ChallengePanelKey>({')
    expect(adminChallengeRoutePageSource).not.toContain("from 'vue-router'")
    expect(adminChallengeRoutePageSource).not.toContain('useRoute(')
    expect(adminChallengeRoutePageSource).not.toContain('useRouter(')
    expect(adminChallengeRoutePageSource).not.toContain(
      'const activePanel = computed<ChallengePanelKey>(() => resolvePanel(route.query.panel))'
    )
    expect(adminChallengeRoutePageSource).not.toContain(
      'function handleTabKeydown(event: KeyboardEvent, index: number): void {'
    )

    expect(adminChallengeDetailSource).toContain(
      "import { usePlatformChallengeDetailRoutePage } from '@/features/platform/challenge-detail'"
    )
    expect(adminChallengeDetailSource).toContain('setTabButtonRef')
    expect(adminChallengeDetailSource).not.toContain(
      "import { useRouteQueryTabs } from '@/composables/useRouteQueryTabs'"
    )
    expect(adminChallengeDetailSource).not.toContain('useRouteQueryTabs<ChallengePanelKey>({')
    expect(adminChallengeDetailSource).not.toContain('function handleTabKeydown(')

    expect(adminChallengeWorkspaceTabsSource).toContain(
      ':ref="(element) => setTabButtonRef(tab.key, element as HTMLButtonElement | null)"'
    )
  })
})
