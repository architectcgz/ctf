import { describe, expect, it } from 'vitest'

import challengeDetailSource from '@/views/challenges/ChallengeDetail.vue?raw'
import challengeWorkspaceShellSource from '@/components/challenge/ChallengeWorkspaceShell.vue?raw'
import challengeQuestionPanelSource from '@/components/challenge/ChallengeQuestionPanel.vue?raw'

describe('ChallengeDetail panel extraction', () => {
  it('应将工作区壳层与五块主要装配区抽到独立 challenge 组件', () => {
    expect(challengeDetailSource).toContain(
      "import { useChallengeDetailPage } from '@/features/challenge-detail'"
    )
    expect(challengeDetailSource).not.toContain("from '@/api/challenge'")
    expect(challengeDetailSource).toContain(
      "import ChallengeWorkspaceShell from '@/components/challenge/ChallengeWorkspaceShell.vue'"
    )
    expect(challengeDetailSource).toContain('<ChallengeWorkspaceShell')

    expect(challengeWorkspaceShellSource).toContain(
      "import ChallengeQuestionPanel from '@/components/challenge/ChallengeQuestionPanel.vue'"
    )
    expect(challengeWorkspaceShellSource).toContain(
      "import ChallengeSolutionsPanel from '@/components/challenge/ChallengeSolutionsPanel.vue'"
    )
    expect(challengeWorkspaceShellSource).toContain(
      "import ChallengeSubmissionRecordsPanel from '@/components/challenge/ChallengeSubmissionRecordsPanel.vue'"
    )
    expect(challengeWorkspaceShellSource).toContain(
      "import ChallengeWriteupPanel from '@/components/challenge/ChallengeWriteupPanel.vue'"
    )
    expect(challengeWorkspaceShellSource).toContain(
      "import ChallengeActionAside from '@/components/challenge/ChallengeActionAside.vue'"
    )
    expect(challengeWorkspaceShellSource).toContain('<ChallengeQuestionPanel')
    expect(challengeWorkspaceShellSource).toContain('<ChallengeSolutionsPanel')
    expect(challengeWorkspaceShellSource).toContain('<ChallengeSubmissionRecordsPanel')
    expect(challengeWorkspaceShellSource).toContain('<ChallengeWriteupPanel')
    expect(challengeWorkspaceShellSource).toContain('<ChallengeActionAside')
  })

  it('题目面板应把题目 meta 展示继续下沉到 challenge entity ui', () => {
    expect(challengeQuestionPanelSource).toContain("ChallengeMetaStrip } from '@/entities/challenge'")
    expect(challengeQuestionPanelSource).toContain('<ChallengeMetaStrip :challenge="challenge" />')
    expect(challengeQuestionPanelSource).not.toContain(':build-meta-pill-style=')
    expect(challengeQuestionPanelSource).not.toContain(':get-category-label=')
  })
})
