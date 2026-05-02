import { useRoute, useRouter } from 'vue-router'

import { useRouteQueryTabs } from '@/composables/useRouteQueryTabs'

type ScoreboardPanelKey = 'contest' | 'points'

const panelTabs: Array<{ key: ScoreboardPanelKey; label: string; panelId: string; tabId: string }> =
  [
    {
      key: 'contest',
      label: '竞赛排行榜',
      panelId: 'scoreboard-panel-contest',
      tabId: 'scoreboard-tab-contest',
    },
    {
      key: 'points',
      label: '积分排行榜',
      panelId: 'scoreboard-panel-points',
      tabId: 'scoreboard-tab-points',
    },
  ]

export function useScoreboardRoutePage() {
  const route = useRoute()
  const router = useRouter()
  const tabState = useRouteQueryTabs<ScoreboardPanelKey>({
    route,
    router,
    orderedTabs: panelTabs.map((tab) => tab.key) as ScoreboardPanelKey[],
    defaultTab: 'contest',
  })

  return {
    panelTabs,
    activeTab: tabState.activeTab,
    setTabButtonRef: tabState.setTabButtonRef,
    selectTab: tabState.selectTab,
    handleTabKeydown: tabState.handleTabKeydown,
  }
}
