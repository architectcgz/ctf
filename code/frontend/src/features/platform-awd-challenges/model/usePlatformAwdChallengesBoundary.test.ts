import { describe, expect, it } from 'vitest'

import source from './usePlatformAwdChallenges.ts?raw'

describe('usePlatformAwdChallenges boundary', () => {
  it('应组合导入流程子模块，避免在主组合器内联导入队列逻辑', () => {
    expect(source).toContain("import { useAwdChallengeImportFlow } from './useAwdChallengeImportFlow'")
    expect(source).not.toContain('async function refreshImportQueue()')
    expect(source).not.toContain('async function selectImportPackages(')
    expect(source).not.toContain('async function commitImportPreview(')
  })
})
