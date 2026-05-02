import { describe, expect, it } from 'vitest'

import source from './useChallengePackageImport.ts?raw'

describe('useChallengePackageImport boundary', () => {
  it('应组合上传流程与错误归一化子模块，避免主组合器内联上传编排细节', () => {
    expect(source).toContain("from './challengeImportUploadFlow'")
    expect(source).not.toContain('function normalizeUploadError(')
    expect(source).not.toContain('function buildFriendlyUploadMessage(')
    expect(source).not.toContain('async function selectPackages(')
    expect(source).not.toContain('async function refreshQueue(')
  })
})
