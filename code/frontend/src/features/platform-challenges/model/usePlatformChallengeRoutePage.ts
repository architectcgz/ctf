import { computed, type ComputedRef } from 'vue'
import { useRoute, useRouter } from 'vue-router'

type PlatformChallengeRoutePageMode = 'topology-studio' | 'writeup-editor' | 'writeup-view'

type PlatformChallengeRoutePageBase = {
  challengeId: ComputedRef<string>
  backToChallengeDetail: () => void
}

type PlatformChallengeWriteupViewRoutePage = PlatformChallengeRoutePageBase & {
  goToWriteupEditor: () => void
}

export function usePlatformChallengeRoutePage(
  mode: 'writeup-view'
): PlatformChallengeWriteupViewRoutePage
export function usePlatformChallengeRoutePage(
  mode: 'topology-studio' | 'writeup-editor'
): PlatformChallengeRoutePageBase
export function usePlatformChallengeRoutePage(mode: PlatformChallengeRoutePageMode) {
  const route = useRoute()
  const router = useRouter()
  const challengeId = computed(() => String(route.params.id ?? ''))

  function backToChallengeDetail(): void {
    void router.push({
      name: 'PlatformChallengeDetail',
      params: { id: challengeId.value },
      query: mode === 'topology-studio' ? undefined : { panel: 'writeup' },
    })
  }

  if (mode === 'writeup-view') {
    return {
      challengeId,
      backToChallengeDetail,
      goToWriteupEditor: () => {
        void router.push({ name: 'PlatformChallengeWriteup', params: { id: challengeId.value } })
      },
    }
  }

  return {
    challengeId,
    backToChallengeDetail,
  }
}
