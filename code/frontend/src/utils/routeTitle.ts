type RouteLike = {
  path: string
  meta?: {
    title?: unknown
  }
  query?: Record<string, unknown>
}

const dashboardPanelTitleMap = {
  recommendation: '训练建议',
  category: '分类进度',
  timeline: '近期动态',
  difficulty: '难度分布',
} as const

export function resolveRouteTitle(route: RouteLike): string {
  if (route.path === '/dashboard') {
    const panel = route.query?.panel
    if (typeof panel === 'string' && panel in dashboardPanelTitleMap) {
      return dashboardPanelTitleMap[panel as keyof typeof dashboardPanelTitleMap]
    }
  }

  return typeof route.meta?.title === 'string' ? route.meta.title : ''
}
