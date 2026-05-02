import { useRouter } from 'vue-router'

export function useAwdChallengeLibraryPage() {
  const router = useRouter()

  function openImportPage(): void {
    void router.push({ name: 'PlatformAwdChallengeImport' })
  }

  return {
    openImportPage,
  }
}
