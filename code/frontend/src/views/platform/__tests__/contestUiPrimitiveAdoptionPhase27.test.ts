import { describe, expect, it } from 'vitest'

import contestOperationsHubSource from '../ContestOperationsHub.vue?raw'
import contestOrchestrationSource from '@/components/platform/contest/ContestOrchestrationPage.vue?raw'

describe('contest ui primitive adoption phase 27', () => {
  it('contest operations hub shell should not keep legacy journal-eyebrow-text marker once workspace overline owns the hero copy', () => {
    expect(contestOperationsHubSource).not.toContain('journal-eyebrow-text')
  })

  it('contest orchestration page shell should not keep legacy journal-eyebrow-text marker once tabs and workspace overlines own the page structure', () => {
    expect(contestOrchestrationSource).not.toContain('journal-eyebrow-text')
  })
})
