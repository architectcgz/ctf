export interface ContestEditRouteTarget {
  name: 'ContestEdit'
  params: {
    id: string
  }
}

export interface ContestManageListRouteTarget {
  name: 'ContestManage'
  query: {
    panel: 'list'
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

export interface ContestAwdConfigRouteTarget {
  name: 'ContestAWDConfig'
  params: {
    id: string
  }
  query?: {
    service: string
  }
}

export function buildContestEditRoute(id: string): ContestEditRouteTarget {
  return {
    name: 'ContestEdit',
    params: { id },
  }
}

export function buildContestManageListRoute(): ContestManageListRouteTarget {
  return {
    name: 'ContestManage',
    query: { panel: 'list' },
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

export function buildContestAwdConfigRoute(
  id: string,
  serviceId?: string
): ContestAwdConfigRouteTarget {
  return {
    name: 'ContestAWDConfig',
    params: { id },
    query: serviceId ? { service: serviceId } : undefined,
  }
}
