export function contestAwdConfigBackToStudioRoute(contestId: string) {
  return {
    name: 'ContestEdit',
    params: {
      id: contestId,
    },
    query: {
      panel: 'awd-config',
    },
  } as const
}
