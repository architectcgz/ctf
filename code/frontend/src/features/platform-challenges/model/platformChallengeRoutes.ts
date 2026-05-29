export const platformChallengeImportManageRoute = {
  name: 'PlatformChallengeImportManage',
} as const

export function platformChallengeImportPreviewRoute(importId: string) {
  return {
    name: 'PlatformChallengeImportPreview',
    params: {
      importId,
    },
  } as const
}

export function platformChallengeDetailRoute(challengeId: string) {
  return {
    name: 'PlatformChallengeDetail',
    params: {
      id: challengeId,
    },
  } as const
}

export function platformChallengeTopologyRoute(challengeId: string) {
  return {
    name: 'PlatformChallengeTopologyStudio',
    params: {
      id: challengeId,
    },
  } as const
}

export function platformChallengeWriteupPanelRoute(challengeId: string) {
  return {
    name: 'PlatformChallengeDetail',
    params: {
      id: challengeId,
    },
    query: {
      panel: 'writeup',
    },
  } as const
}

export function platformChallengeWriteupEditorRoute(challengeId: string) {
  return {
    name: 'PlatformChallengeWriteup',
    params: {
      id: challengeId,
    },
  } as const
}
