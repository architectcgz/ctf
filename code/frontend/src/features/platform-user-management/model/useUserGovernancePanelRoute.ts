import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

export type UserPanelKey = 'overview' | 'import'

export function useUserGovernancePanelRoute() {
  const route = useRoute()
  const router = useRouter()

  const activePanel = computed<UserPanelKey>(() => {
    const rawPanel = route.query.panel
    const panel = Array.isArray(rawPanel) ? rawPanel[0] : rawPanel
    if (panel === 'import') {
      return 'import'
    }
    return 'overview'
  })

  async function switchPanel(panel: UserPanelKey): Promise<void> {
    if (activePanel.value === panel) {
      return
    }

    const nextQuery = { ...route.query }
    if (panel === 'overview') {
      delete nextQuery.panel
    } else {
      nextQuery.panel = panel
    }

    await router.replace({ name: 'UserManage', query: nextQuery })
  }

  return {
    activePanel,
    switchPanel,
  }
}
