export type UserPanelKey = 'overview' | 'import'

export function resolveUserGovernancePanel(rawPanel: unknown): UserPanelKey {
  const panel = Array.isArray(rawPanel) ? rawPanel[0] : rawPanel
  if (panel === 'import') {
    return 'import'
  }
  return 'overview'
}

export function buildUserGovernancePanelQuery(
  query: Record<string, unknown>,
  panel: UserPanelKey
): Record<string, unknown> {
  const nextQuery = { ...query }
  if (panel === 'overview') {
    delete nextQuery.panel
  } else {
    nextQuery.panel = panel
  }
  return nextQuery
}
