export { buildContestOperationsRoute } from './contestManageRoutes'

export function buildContestManageListRoute() {
  return {
    name: 'ContestManage',
    query: { panel: 'list' },
  } as const
}
