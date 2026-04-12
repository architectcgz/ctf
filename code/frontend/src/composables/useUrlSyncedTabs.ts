import { ref, type Ref } from 'vue'
import { useTabKeyboardNavigation } from '@/composables/useTabKeyboardNavigation'

interface UseUrlSyncedTabsOptions<T extends string> {
  orderedTabs: readonly T[]
  defaultTab: T
  queryKey?: string
}

interface UseUrlSyncedTabsResult<T extends string> {
  activeTab: Ref<T>
  setTabButtonRef: (tab: T, element: HTMLButtonElement | null) => void
  selectTab: (tab: T) => void
  handleTabKeydown: (event: KeyboardEvent, index: number) => void
}

export function useUrlSyncedTabs<T extends string>({
  orderedTabs,
  defaultTab,
  queryKey = 'panel',
}: UseUrlSyncedTabsOptions<T>): UseUrlSyncedTabsResult<T> {
  const tabSet = new Set<T>(orderedTabs)

  function resolveTabFromLocation(): T {
    if (typeof window === 'undefined') return defaultTab
    if (!window.location.pathname || window.location.pathname === '/') return defaultTab
    const panel = new URLSearchParams(window.location.search).get(queryKey)
    if (panel && tabSet.has(panel as T)) {
      return panel as T
    }
    return defaultTab
  }

  function syncPanelToLocation(tab: T): void {
    if (typeof window === 'undefined') return
    const url = new URL(window.location.href)
    url.searchParams.set(queryKey, tab)
    window.history.replaceState(window.history.state, '', `${url.pathname}${url.search}${url.hash}`)
  }

  const activeTab = ref(resolveTabFromLocation()) as Ref<T>

  function selectTab(tab: T): void {
    if (activeTab.value === tab) return
    activeTab.value = tab
    syncPanelToLocation(tab)
  }
  const { setTabButtonRef, handleTabKeydown } = useTabKeyboardNavigation<T>({
    orderedTabs,
    selectTab,
  })

  return {
    activeTab,
    setTabButtonRef,
    selectTab,
    handleTabKeydown,
  }
}
