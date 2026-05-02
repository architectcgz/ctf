import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

export function useChallengeWriteupViewPage() {
  const route = useRoute()
  const router = useRouter()
  const challengeId = computed(() => String(route.params.id ?? ''))

  function backToChallengeDetail(): void {
    void router.push({
      name: 'PlatformChallengeDetail',
      params: { id: challengeId.value },
      query: { panel: 'writeup' },
    })
  }

  function goToWriteupEditor(): void {
    void router.push({ name: 'PlatformChallengeWriteup', params: { id: challengeId.value } })
  }

  return {
    challengeId,
    backToChallengeDetail,
    goToWriteupEditor,
  }
}
