import { useRouter } from 'vue-router'

export function useChallengePackageFormatPage() {
  const router = useRouter()

  function backToImportManage(): void {
    void router.push({ name: 'PlatformChallengeImportManage' })
  }

  return {
    backToImportManage,
  }
}
