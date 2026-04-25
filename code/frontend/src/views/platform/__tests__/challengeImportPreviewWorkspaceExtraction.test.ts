import { describe, expect, it } from 'vitest'

import challengeImportPreviewSource from '../ChallengeImportPreview.vue?raw'

describe('ChallengeImportPreview workspace extraction', () => {
  it('应将导入预览页工作区壳层抽到独立 platform challenge 组件', () => {
    expect(challengeImportPreviewSource).toContain(
      "import ChallengeImportPreviewWorkspacePanel from '@/components/platform/challenge/ChallengeImportPreviewWorkspacePanel.vue'"
    )
    expect(challengeImportPreviewSource).toContain('<ChallengeImportPreviewWorkspacePanel')
  })
})
