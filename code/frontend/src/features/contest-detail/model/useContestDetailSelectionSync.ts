import { toValue, type MaybeRefOrGetter } from 'vue'

interface UseContestDetailSelectionSyncOptions {
  selectedChallengeId?: MaybeRefOrGetter<string | null | Array<string | null> | undefined>
  syncSelectedChallengeById: (challengeId: string) => void
}

export function normalizeContestChallengeId(
  value: string | null | Array<string | null> | undefined
): string {
  if (Array.isArray(value)) {
    return value.find((item): item is string => typeof item === 'string' && item.length > 0) ?? ''
  }
  return typeof value === 'string' ? value : ''
}

export function useContestDetailSelectionSync({
  selectedChallengeId,
  syncSelectedChallengeById,
}: UseContestDetailSelectionSyncOptions) {
  function requestedChallengeId(): string {
    return normalizeContestChallengeId(toValue(selectedChallengeId))
  }

  function syncSelectedChallengeFromQuery() {
    syncSelectedChallengeById(requestedChallengeId())
  }

  return {
    requestedChallengeId,
    syncSelectedChallengeFromQuery,
  }
}
