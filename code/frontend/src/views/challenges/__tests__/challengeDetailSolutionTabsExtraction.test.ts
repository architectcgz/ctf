import { describe, expect, it } from 'vitest'

import challengeDetailSource from '../ChallengeDetail.vue?raw'
import challengeActionAsideSource from '@/components/challenge/ChallengeActionAside.vue?raw'
import challengeDetailInteractionsSource from '@/features/challenge-detail/model/useChallengeDetailInteractions.ts?raw'
import challengeDetailPageSource from '@/features/challenge-detail/model/useChallengeDetailPage.ts?raw'
import challengeDetailPresentationSource from '@/features/challenge-detail/model/useChallengeDetailPresentation.ts?raw'

describe('challenge detail solution tabs extraction', () => {
  it('ChallengeDetail 的题解子标签键盘逻辑应由 page-level feature 承接，presentation composable 保持纯派生', () => {
    expect(challengeDetailSource).toContain(
      "import { useChallengeDetailPage } from '@/features/challenge-detail'"
    )
    expect(challengeDetailPageSource).toContain(
      "import { useTabKeyboardNavigation } from '@/composables/useTabKeyboardNavigation'"
    )
    expect(challengeDetailPageSource).toContain('useTabKeyboardNavigation<ChallengeSolutionTab>({')
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
