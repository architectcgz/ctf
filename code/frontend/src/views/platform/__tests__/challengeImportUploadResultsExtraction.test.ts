import { describe, expect, it } from 'vitest'

import challengeImportManageSource from '../ChallengeImportManage.vue?raw'
import challengeImportUploadResultsPanelSource from '@/components/platform/challenge/ChallengeImportUploadResultsPanel.vue?raw'

describe('ChallengeImportManage upload results extraction', () => {
  it('应将最近上传结果工作区抽到独立 platform challenge 组件', () => {
    expect(challengeImportManageSource).toContain(
      "import ChallengeImportUploadResultsPanel from '@/components/platform/challenge/ChallengeImportUploadResultsPanel.vue'"
    )
    expect(challengeImportManageSource).toContain('<ChallengeImportUploadResultsPanel')
    expect(challengeImportUploadResultsPanelSource).toContain('Upload Receipt')
    expect(challengeImportUploadResultsPanelSource).toContain('最近上传结果')
    expect(challengeImportUploadResultsPanelSource).toContain('challenge-upload-result--success')
    expect(challengeImportUploadResultsPanelSource).toContain('challenge-upload-result--error')
  })
})
