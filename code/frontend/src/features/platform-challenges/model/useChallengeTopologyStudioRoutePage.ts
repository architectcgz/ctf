import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

export function useChallengeTopologyStudioRoutePage() {
  const route = useRoute()
  const router = useRouter()
  const challengeId = computed(() => String(route.params.id ?? ''))

  function backToChallengeDetail(): void {
    void router.push({ name: 'PlatformChallengeDetail', params: { id: challengeId.value } })
  }

  return {
    challengeId,
    backToChallengeDetail,
  }
}
