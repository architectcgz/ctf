interface ChallengeManageRouteTarget {
  name: 'ChallengeManage'
}

interface ChallengeImportManageRouteTarget {
  name: 'PlatformChallengeImportManage'
  hash?: string
}

interface ChallengeImportPreviewRouteTarget {
  name: 'PlatformChallengeImportPreview'
  params: {
    importId: string
  }
}

interface ChallengePackageFormatRouteTarget {
  name: 'PlatformChallengePackageFormat'
}

export function buildChallengeManageRoute(): ChallengeManageRouteTarget {
  return { name: 'ChallengeManage' }
}

export function buildChallengeImportManageRoute(hash?: string): ChallengeImportManageRouteTarget {
  return hash
    ? {
        name: 'PlatformChallengeImportManage',
        hash,
      }
    : {
        name: 'PlatformChallengeImportManage',
      }
}

export function buildChallengeImportPreviewRoute(
  importId: string
): ChallengeImportPreviewRouteTarget {
  return {
    name: 'PlatformChallengeImportPreview',
    params: { importId },
  }
}

export function buildChallengePackageFormatRoute(): ChallengePackageFormatRouteTarget {
  return { name: 'PlatformChallengePackageFormat' }
}
