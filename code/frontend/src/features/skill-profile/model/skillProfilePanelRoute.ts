export type SkillProfilePanelKey = 'analysis' | 'weakness' | 'recommendations'

export function resolveSkillProfilePanel(rawPanel: unknown): SkillProfilePanelKey {
  const panel = Array.isArray(rawPanel) ? rawPanel[0] : rawPanel
  switch (panel) {
    case 'weakness':
    case 'recommendations':
      return panel
    default:
      return 'analysis'
  }
}

export function buildSkillProfilePanelQuery(
  query: Record<string, unknown>,
  panel: SkillProfilePanelKey
): Record<string, unknown> {
  const nextQuery = { ...query }
  if (panel === 'analysis') {
    delete nextQuery.panel
  } else {
    nextQuery.panel = panel
  }
  return nextQuery
}
