interface ContestDetailRouteTarget {
  name: 'ContestDetail'
  params: {
    id: string
  }
}

export function buildContestDetailRoute(id: string): ContestDetailRouteTarget {
  return {
    name: 'ContestDetail',
    params: { id },
  }
}
