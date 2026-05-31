import { computed, type ComputedRef } from 'vue'

import { useRouteNavigationTransport } from '@/shared/model/navigation/useRouteNavigationTransport'
import { useRouteQueryTransport } from '@/shared/model/navigation/useRouteQueryTransport'
import {
  platformChallengeDetailRoute,
  platformChallengeWriteupEditorRoute,
  platformChallengeWriteupPanelRoute,
} from './platformChallengeRoutes'

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
  const { params } = useRouteQueryTransport()
  const { push } = useRouteNavigationTransport()
  const challengeId = computed(() => String(params.value.id ?? ''))

  function backToChallengeDetail(): void {
    void push(
      mode === 'topology-studio'
        ? platformChallengeDetailRoute(challengeId.value)
        : platformChallengeWriteupPanelRoute(challengeId.value)
    )
  }

  if (mode === 'writeup-view') {
    return {
      challengeId,
      backToChallengeDetail,
      goToWriteupEditor: () => {
        void push(platformChallengeWriteupEditorRoute(challengeId.value))
      },
    }
  }

  return {
    challengeId,
    backToChallengeDetail,
  }
}
