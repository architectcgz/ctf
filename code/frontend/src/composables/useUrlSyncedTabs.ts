import { ref, type Ref } from 'vue'

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

  const activeTab = ref<T>(resolveTabFromLocation())
  const tabButtonRefs: Partial<Record<T, HTMLButtonElement | null>> = {}

  function setTabButtonRef(tab: T, element: HTMLButtonElement | null): void {
    tabButtonRefs[tab] = element
  }

  function focusTab(tab: T): void {
    tabButtonRefs[tab]?.focus()
  }

  function selectTab(tab: T): void {
    if (activeTab.value === tab) return
    activeTab.value = tab
    syncPanelToLocation(tab)
  }

  function handleTabKeydown(event: KeyboardEvent, index: number): void {
    if (
      event.key !== 'ArrowRight' &&
      event.key !== 'ArrowLeft' &&
      event.key !== 'Home' &&
      event.key !== 'End'
    ) {
      return
    }

    event.preventDefault()

    if (event.key === 'Home') {
      const firstTab = orderedTabs[0]
      if (!firstTab) return
      selectTab(firstTab)
      focusTab(firstTab)
      return
    }

    if (event.key === 'End') {
      const lastTab = orderedTabs[orderedTabs.length - 1]
      if (!lastTab) return
      selectTab(lastTab)
      focusTab(lastTab)
      return
    }

    const direction = event.key === 'ArrowRight' ? 1 : -1
    const nextIndex = (index + direction + orderedTabs.length) % orderedTabs.length
    const nextTab = orderedTabs[nextIndex]
    if (!nextTab) return
    selectTab(nextTab)
    focusTab(nextTab)
  }

  return {
    activeTab,
    setTabButtonRef,
    selectTab,
    handleTabKeydown,
  }
}
