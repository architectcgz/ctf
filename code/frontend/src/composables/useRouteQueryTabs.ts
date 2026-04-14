import { computed, type ComputedRef } from 'vue'
import type { RouteLocationNormalizedLoaded, Router } from 'vue-router'
import { useTabKeyboardNavigation } from '@/composables/useTabKeyboardNavigation'

interface UseRouteQueryTabsOptions<T extends string> {
  route: RouteLocationNormalizedLoaded
  router: Router
  orderedTabs: readonly T[]
  defaultTab: T
  routeName?: string
  routeParams?: RouteLocationNormalizedLoaded['params']
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
  routeParams,
  queryKey = 'panel',
}: UseRouteQueryTabsOptions<T>): UseRouteQueryTabsResult<T> {
  const tabSet = new Set<T>(orderedTabs)

  const activeTab = computed<T>(() => {
    const rawPanel = route.query[queryKey]
    const panel = Array.isArray(rawPanel) ? rawPanel[0] : rawPanel
    if (typeof panel === 'string' && tabSet.has(panel as T)) {
      return panel as T
    }
    return defaultTab
  })

  async function selectTab(tab: T): Promise<void> {
    if (activeTab.value === tab) return

    const nextQuery = { ...route.query }
    if (tab === defaultTab) {
      delete nextQuery[queryKey]
    } else {
      nextQuery[queryKey] = tab
    }

    if (routeName) {
      if (routeParams) {
        await router.replace({ name: routeName, params: routeParams, query: nextQuery })
        return
      }

      await router.replace({ name: routeName, query: nextQuery })
      return
    }

    await router.replace({ query: nextQuery })
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
