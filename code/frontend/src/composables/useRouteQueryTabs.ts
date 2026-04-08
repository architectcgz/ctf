import { computed, type ComputedRef } from 'vue'
import type { RouteLocationNormalizedLoaded, Router } from 'vue-router'

interface UseRouteQueryTabsOptions<T extends string> {
  route: RouteLocationNormalizedLoaded
  router: Router
  orderedTabs: readonly T[]
  defaultTab: T
  routeName?: string
  queryKey?: string
}

interface UseRouteQueryTabsResult<T extends string> {
  activeTab: ComputedRef<T>
  setTabButtonRef: (tab: T, element: HTMLButtonElement | null) => void
  selectTab: (tab: T) => Promise<void>
  handleTabKeydown: (event: KeyboardEvent, index: number) => void
}

export function useRouteQueryTabs<T extends string>({
  route,
  router,
  orderedTabs,
  defaultTab,
  routeName,
  queryKey = 'panel',
}: UseRouteQueryTabsOptions<T>): UseRouteQueryTabsResult<T> {
  const tabSet = new Set<T>(orderedTabs)
  const tabButtonRefs: Partial<Record<T, HTMLButtonElement | null>> = {}

  const activeTab = computed<T>(() => {
    const panel = route.query[queryKey]
    if (typeof panel === 'string' && tabSet.has(panel as T)) {
      return panel as T
    }
    return defaultTab
  })

  function setTabButtonRef(tab: T, element: HTMLButtonElement | null): void {
    tabButtonRefs[tab] = element
  }

  function focusTab(tab: T): void {
    tabButtonRefs[tab]?.focus()
  }

  async function selectTab(tab: T): Promise<void> {
    if (activeTab.value === tab) return

    const nextQuery = { ...route.query }
    if (tab === defaultTab) {
      delete nextQuery[queryKey]
    } else {
      nextQuery[queryKey] = tab
    }

    if (routeName) {
      await router.replace({ name: routeName, query: nextQuery })
      return
    }

    await router.replace({ query: nextQuery })
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
      void selectTab(firstTab)
      focusTab(firstTab)
      return
    }

    if (event.key === 'End') {
      const lastTab = orderedTabs[orderedTabs.length - 1]
      if (!lastTab) return
      void selectTab(lastTab)
      focusTab(lastTab)
      return
    }

    const direction = event.key === 'ArrowRight' ? 1 : -1
    const nextIndex = (index + direction + orderedTabs.length) % orderedTabs.length
    const nextTab = orderedTabs[nextIndex]
    if (!nextTab) return
    void selectTab(nextTab)
    focusTab(nextTab)
  }

  return {
    activeTab,
    setTabButtonRef,
    selectTab,
    handleTabKeydown,
  }
}
