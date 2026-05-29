export interface ContestEditRouteTarget {
  name: 'ContestEdit'
  params: {
    id: string
  }
}

export interface ContestOperationsRouteTarget {
  name: 'ContestOperations'
  params: {
    id: string
  }
}

export interface ContestAnnouncementsRouteTarget {
  name: 'ContestAnnouncements'
  params: {
    id: string
  }
}

export function buildContestEditRoute(id: string): ContestEditRouteTarget {
  return {
    name: 'ContestEdit',
    params: { id },
  }
}

export function buildContestOperationsRoute(id: string): ContestOperationsRouteTarget {
  return {
    name: 'ContestOperations',
    params: { id },
  }
}

export function buildContestAnnouncementsRoute(id: string): ContestAnnouncementsRouteTarget {
  return {
    name: 'ContestAnnouncements',
    params: { id },
  }
}
