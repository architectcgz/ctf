import { describe, expect, it } from 'vitest'

import challengeDetailSource from '../ChallengeDetail.vue?raw'
import challengeDetailPresentationSource from '@/composables/useChallengeDetailPresentation.ts?raw'

describe('challenge detail solution tabs extraction', () => {
  it('ChallengeDetail 应复用 useTabKeyboardNavigation，且 presentation composable 不再内置题解子标签键盘逻辑', () => {
    expect(challengeDetailSource).toContain(
      "import { useTabKeyboardNavigation } from '@/composables/useTabKeyboardNavigation'"
    )
    expect(challengeDetailSource).toContain('useTabKeyboardNavigation<ChallengeSolutionTab>({')
    expect(challengeDetailSource).not.toContain('function focusTab(id: string): void {')
    expect(challengeDetailSource).not.toContain(
      'function handleSolutionTabKeydown(event: KeyboardEvent, currentTab: ChallengeSolutionTab): void {'
    )

    expect(challengeDetailPresentationSource).not.toContain('function handleSolutionTabKeydown(')
    expect(challengeDetailPresentationSource).not.toContain('focusTab: (tabId: string) => void')
    expect(challengeDetailPresentationSource).not.toContain('handleSolutionTabKeydown,')
  })
})
