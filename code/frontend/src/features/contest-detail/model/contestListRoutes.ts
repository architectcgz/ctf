import type { AppRouteTarget } from '@/shared/lib/navigation/routeTarget'

export function buildContestDetailRoute(id: string): AppRouteTarget {
  return {
    name: 'ContestDetail',
    params: { id },
  }
}
