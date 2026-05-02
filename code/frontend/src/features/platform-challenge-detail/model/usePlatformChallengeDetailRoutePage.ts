import { useRoute, useRouter } from 'vue-router'

import { useRouteQueryTabs } from '@/composables/useRouteQueryTabs'
import { usePlatformChallengeDetailPage } from './usePlatformChallengeDetailPage'

type ChallengePanelKey = 'detail' | 'writeup'

const panelTabs = [
  {
    key: 'detail' as const,
    label: '题目管理',
    tabId: 'admin-challenge-tab-detail',
    panelId: 'admin-challenge-panel-detail',
  },
  {
    key: 'writeup' as const,
    label: '题解管理',
    tabId: 'admin-challenge-tab-writeup',
    panelId: 'admin-challenge-panel-writeup',
  },
]

export function usePlatformChallengeDetailRoutePage() {
  const route = useRoute()
  const router = useRouter()
  const panelTabOrder = panelTabs.map((tab) => tab.key) as ChallengePanelKey[]
  const page = usePlatformChallengeDetailPage()
  const tabState = useRouteQueryTabs<ChallengePanelKey>({
    route,
    router,
    orderedTabs: panelTabOrder,
    defaultTab: 'detail',
    routeName: 'PlatformChallengeDetail',
    routeParams: route.params,
  })

  return {
    ...page,
    panelTabs,
    activePanel: tabState.activeTab,
    setTabButtonRef: tabState.setTabButtonRef,
    switchPanel: tabState.selectTab,
    handleTabKeydown: tabState.handleTabKeydown,
  }
}
