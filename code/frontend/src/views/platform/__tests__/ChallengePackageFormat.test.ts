import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { vi } from 'vitest'
import challengePackageFormatSource from '../ChallengePackageFormat.vue?raw'
import challengePackageFormatGuidePanelSource from '@/components/platform/challenge/ChallengePackageFormatGuidePanel.vue?raw'

import ChallengePackageFormat from '../ChallengePackageFormat.vue'

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: vi.fn() }),
  }
})

describe('ChallengePackageFormat', () => {
  it('应该展示题目包结构与 challenge.yml 示例', () => {
    const wrapper = mount(ChallengePackageFormat)

    expect(wrapper.text()).toContain('题目包示例')
    expect(wrapper.text()).toContain('challenge.yml')
    expect(wrapper.text()).toContain('statement.md')
    expect(wrapper.text()).toContain('Dockerfile')
    expect(wrapper.text()).toContain('app.py')
    expect(wrapper.text()).toContain('不要写 # 题目名')
    expect(wrapper.text()).toContain('不要写 ## 题目描述')
    expect(wrapper.text()).toContain('api_version: v1')
    expect(wrapper.text()).toContain('flag:')
    expect(wrapper.text()).toContain('checker:')
    expect(wrapper.text()).toContain('http_standard')
    expect(wrapper.text()).toContain('X-AWD-Checker-Token')
  })

  it('应使用共享 workspace overline 作为上传示例页的 hero 标记', () => {
    expect(challengePackageFormatGuidePanelSource).toContain(
      '<div class="workspace-overline">Uploader Guide</div>'
    )
    expect(challengePackageFormatGuidePanelSource).not.toContain(
      '<div class="journal-eyebrow">Uploader Guide</div>'
    )
    expect(challengePackageFormatSource).toContain(
      "import ChallengePackageFormatGuidePanel from '@/components/platform/challenge/ChallengePackageFormatGuidePanel.vue'"
    )
  })
})
