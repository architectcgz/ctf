type RouteLike = {
  path: string
  meta?: {
    title?: unknown
  }
  query?: Record<string, unknown>
}

const dashboardPanelTitleMap = {
  overview: '训练总览',
  recommendation: '训练队列',
  category: '分类补强',
  timeline: '训练记录',
  difficulty: '强度推进',
} as const

export function resolveRouteTitle(route: RouteLike): string {
  if (/^(?:\/student)?\/dashboard(?:\/\d+)?$/.test(route.path)) {
    const panel = route.query?.panel
    if (typeof panel === 'string' && panel in dashboardPanelTitleMap) {
      return dashboardPanelTitleMap[panel as keyof typeof dashboardPanelTitleMap]
    }

    const variantMatch = route.path.match(/^(?:\/student)?\/dashboard\/(\d+)$/)
    if (variantMatch) {
      return `仪表盘 · 风格 ${variantMatch[1]}`
    }

    return '仪表盘'
  }

  return typeof route.meta?.title === 'string' ? route.meta.title : ''
}
