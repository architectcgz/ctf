export const challengePackageImportManageRoute = {
  name: 'PlatformChallengeImportManage',
} as const

export function useChallengePackageFormatPage() {
  return {
    backToImportManage: challengePackageImportManageRoute,
  }
}
