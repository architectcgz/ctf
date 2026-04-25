import { describe, expect, it } from 'vitest'

import challengePackageFormatSource from '../ChallengePackageFormat.vue?raw'
import challengePackageFormatGuidePanelSource from '@/components/platform/challenge/ChallengePackageFormatGuidePanel.vue?raw'

describe('ChallengePackageFormat guide extraction', () => {
  it('应将题目包示例页的 guide 内容抽到独立 platform challenge 组件', () => {
    expect(challengePackageFormatSource).toContain(
      "import ChallengePackageFormatGuidePanel from '@/components/platform/challenge/ChallengePackageFormatGuidePanel.vue'"
    )
    expect(challengePackageFormatSource).toContain('<ChallengePackageFormatGuidePanel')
    expect(challengePackageFormatGuidePanelSource).toContain('<div class="workspace-overline">Uploader Guide</div>')
    expect(challengePackageFormatGuidePanelSource).toContain('challenge.yml')
    expect(challengePackageFormatGuidePanelSource).toContain('statement.md')
  })
})
