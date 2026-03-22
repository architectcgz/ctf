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
  if (/^\/dashboard(?:\/\d+)?$/.test(route.path)) {
    const panel = route.query?.panel
    if (typeof panel === 'string' && panel in dashboardPanelTitleMap) {
      return dashboardPanelTitleMap[panel as keyof typeof dashboardPanelTitleMap]
    }

    const variantMatch = route.path.match(/^\/dashboard\/(\d+)$/)
    if (variantMatch) {
      return `仪表盘 · 风格 ${variantMatch[1]}`
    }

    return '仪表盘'
  }

  return typeof route.meta?.title === 'string' ? route.meta.title : ''
}
