import { useRouter, type RouteLocationRaw } from 'vue-router'

import type { AppRouteTarget } from '@/shared/lib/navigation/routeTarget'

export function useRouteNavigationTransport() {
  const router = useRouter()

  async function push(target: AppRouteTarget): Promise<void> {
    await router.push(target as RouteLocationRaw)
  }

  async function replace(target: AppRouteTarget): Promise<void> {
    await router.replace(target as RouteLocationRaw)
  }

  return {
    push,
    replace,
  }
}
