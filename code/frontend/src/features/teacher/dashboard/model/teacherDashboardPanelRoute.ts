export type TeacherDashboardPanelKey =
  | 'overview'
  | 'portrait'
  | 'insight'
  | 'trend'
  | 'review'
  | 'intervention'

export function resolveTeacherDashboardPanel(rawPanel: unknown): TeacherDashboardPanelKey {
  const panel = Array.isArray(rawPanel) ? rawPanel[0] : rawPanel
  switch (panel) {
    case 'portrait':
    case 'insight':
    case 'trend':
    case 'review':
    case 'intervention':
      return panel
    default:
      return 'overview'
  }
}

export function buildTeacherDashboardPanelQuery(
  query: Record<string, unknown>,
  panel: TeacherDashboardPanelKey
): Record<string, unknown> {
  const nextQuery = { ...query }
  if (panel === 'overview') {
    delete nextQuery.panel
  } else {
    nextQuery.panel = panel
  }
  return nextQuery
}
