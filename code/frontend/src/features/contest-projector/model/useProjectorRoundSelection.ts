import type { Ref } from 'vue'

import type { AWDRoundData } from '@/api/contracts'

interface UseProjectorRoundSelectionOptions {
  roundAutoFollow: Ref<boolean>
  selectedRoundId: Ref<string>
}

export function useProjectorRoundSelection({
  roundAutoFollow,
  selectedRoundId,
}: UseProjectorRoundSelectionOptions) {
  function chooseLiveRound(nextRounds: AWDRoundData[]): AWDRoundData | null {
    return nextRounds.find((item) => item.status === 'running') ?? nextRounds[nextRounds.length - 1] ?? null
  }

  function chooseDisplayRound(nextRounds: AWDRoundData[]): AWDRoundData | null {
    if (!roundAutoFollow.value && selectedRoundId.value) {
      const manualRound = nextRounds.find((item) => item.id === selectedRoundId.value)
      if (manualRound) {
        return manualRound
      }
      roundAutoFollow.value = true
    }
    return chooseLiveRound(nextRounds)
  }

  function enableAutoFollow(): boolean {
    if (roundAutoFollow.value) {
      return false
    }
    roundAutoFollow.value = true
    return true
  }

  return {
    chooseLiveRound,
    chooseDisplayRound,
    enableAutoFollow,
  }
}
