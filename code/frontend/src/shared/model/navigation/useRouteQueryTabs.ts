import { computed, type ComputedRef } from 'vue'
import { useRoute, useRouter, type RouteLocationNormalizedLoaded, type Router } from 'vue-router'

import { useTabKeyboardNavigation } from '@/shared/lib/keyboard/useTabKeyboardNavigation'

interface UseRouteQueryTabsOptions<T extends string> {
  route?: RouteLocationNormalizedLoaded
  router?: Router
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
  const currentRoute = route ?? useRoute()
  const currentRouter = router ?? useRouter()
  const tabSet = new Set<T>(orderedTabs)

  const activeTab = computed<T>(() => {
    const rawPanel = currentRoute.query[queryKey]
    const panel = Array.isArray(rawPanel) ? rawPanel[0] : rawPanel
    if (typeof panel === 'string' && tabSet.has(panel as T)) {
      return panel as T
    }
    return defaultTab
  })

  async function selectTab(tab: T): Promise<void> {
    if (activeTab.value === tab) return

    const nextQuery = { ...currentRoute.query }
    if (tab === defaultTab) {
      delete nextQuery[queryKey]
    } else {
      nextQuery[queryKey] = tab
    }

    if (routeName) {
      const currentRouteParams = routeParams ?? currentRoute.params

      if (Object.keys(currentRouteParams).length > 0) {
        await currentRouter.replace({ name: routeName, params: currentRouteParams, query: nextQuery })
        return
      }

      await currentRouter.replace({ name: routeName, query: nextQuery })
      return
    }

    await currentRouter.replace({ query: nextQuery })
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
