import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { vi } from 'vitest'

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
    expect(wrapper.text()).toContain('不要写 # 题目名')
    expect(wrapper.text()).toContain('不要写 ## 题目描述')
    expect(wrapper.text()).toContain('api_version: v1')
    expect(wrapper.text()).toContain('flag:')
  })
})
