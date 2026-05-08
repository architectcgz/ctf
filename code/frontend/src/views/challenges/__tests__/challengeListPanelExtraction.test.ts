import { describe, expect, it } from 'vitest'

import challengeListSource from '@/views/challenges/ChallengeList.vue?raw'
import challengeDirectoryPanelSource from '@/components/challenge/ChallengeDirectoryPanel.vue?raw'

describe('ChallengeList panel extraction', () => {
  it('应将题目目录工作区抽到独立 challenge 组件', () => {
    expect(challengeListSource).toContain(
      "import ChallengeDirectoryPanel from '@/components/challenge/ChallengeDirectoryPanel.vue'"
    )
    expect(challengeListSource).toContain('<ChallengeDirectoryPanel')
  })

  it('题目目录面板应继续把单行题目展示下沉到 challenge entity ui', () => {
    expect(challengeDirectoryPanelSource).toContain(
      "import { ChallengeDirectoryRow } from '@/entities/challenge'"
    )
    expect(challengeDirectoryPanelSource).toContain('<ChallengeDirectoryRow')
    expect(challengeDirectoryPanelSource).not.toContain('class="challenge-row"')
  })
})
