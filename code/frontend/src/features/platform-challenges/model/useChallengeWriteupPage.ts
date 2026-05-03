import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

export function useChallengeWriteupPage() {
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

  return {
    challengeId,
    backToChallengeDetail,
  }
}
