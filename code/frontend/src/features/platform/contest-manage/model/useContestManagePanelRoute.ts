export type ContestManagePanelKey = 'overview' | 'create'

export function resolveContestManagePanel(rawPanel: unknown): ContestManagePanelKey {
  const panel = Array.isArray(rawPanel) ? rawPanel[0] : rawPanel
  if (panel === 'create') {
    return 'create'
  }
  return 'overview'
}

export function buildContestManagePanelQuery(
  query: Record<string, unknown>,
  panel: ContestManagePanelKey
): Record<string, unknown> {
  const nextQuery = { ...query }
  if (panel === 'overview') {
    delete nextQuery.panel
  } else {
    nextQuery.panel = panel
  }
  return nextQuery
}
