import { describe, expect, it } from 'vitest'

import challengeDetailSource from '../ChallengeDetail.vue?raw'
import challengeActionAsideSource from '@/components/challenge/ChallengeActionAside.vue?raw'
import challengeDetailInteractionsSource from '@/composables/useChallengeDetailInteractions.ts?raw'
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

  it('题目详情提交态应通过语义类承接状态样式，而不是从 composable 返回任意主题类', () => {
    expect(challengeDetailPresentationSource).toContain('flag-input-wrap--success')
    expect(challengeActionAsideSource).toContain('.flag-input-wrap--success')
    expect(challengeDetailPresentationSource).not.toContain('border-[var(--color-success)]')
    expect(challengeDetailPresentationSource).not.toContain('bg-[var(--color-success)]')
    expect(challengeDetailPresentationSource).not.toContain('border-[var(--color-warning)]')
    expect(challengeDetailPresentationSource).not.toContain('bg-[var(--color-warning)]')
    expect(challengeDetailInteractionsSource).not.toContain('text-[var(--color-success)]')
    expect(challengeDetailInteractionsSource).not.toContain('text-[var(--color-warning)]')
    expect(challengeDetailInteractionsSource).not.toContain('text-[var(--color-danger)]')
    expect(challengeActionAsideSource).not.toContain('text-[var(--color-success)]')
  })
})
