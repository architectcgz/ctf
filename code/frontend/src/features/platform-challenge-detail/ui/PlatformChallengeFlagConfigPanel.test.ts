import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import PlatformChallengeFlagConfigPanel from './PlatformChallengeFlagConfigPanel.vue'
import type { PlatformChallengeFlagDraft } from '../model'

function createDraft(overrides: Partial<PlatformChallengeFlagDraft> = {}): PlatformChallengeFlagDraft {
  return {
    flagConfigSummary: '静态 Flag',
    flagDraftSummary: '静态 Flag',
    flagPrefix: '',
    flagRegex: '',
    flagType: 'static',
    flagValue: 'flag{demo}',
    isSharedInstanceChallenge: false,
    saving: false,
    ...overrides,
  }
}

describe('PlatformChallengeFlagConfigPanel', () => {
  it('应在切换判题模式时抛出 draft patch', async () => {
    const wrapper = mount(PlatformChallengeFlagConfigPanel, {
      props: {
        draft: createDraft(),
      },
    })

    await wrapper.get('select.flag-field-input').setValue('dynamic')

    expect(wrapper.emitted('update:draft')).toBeTruthy()
    expect(wrapper.emitted('update:draft')?.[0]).toEqual([{ flagType: 'dynamic' }])
  })

  it('应按草稿状态展示共享实例与人工审核提示', async () => {
    const wrapper = mount(PlatformChallengeFlagConfigPanel, {
      props: {
        draft: createDraft({
          isSharedInstanceChallenge: true,
          flagType: 'manual_review',
        }),
      },
    })

    expect(wrapper.text()).toContain('共享实例只适用于无状态题')
    expect(wrapper.text()).toContain('学生提交的答案将进入教师审核队列')

    await wrapper.setProps({
      draft: createDraft({
        isSharedInstanceChallenge: false,
        flagType: 'static',
      }),
    })

    expect(wrapper.text()).not.toContain('共享实例只适用于无状态题')
    expect(wrapper.text()).not.toContain('学生提交的答案将进入教师审核队列')
  })

  it('应在点击保存按钮时抛出 save 事件', async () => {
    const wrapper = mount(PlatformChallengeFlagConfigPanel, {
      props: {
        draft: createDraft({
          saving: false,
        }),
      },
    })

    await wrapper.get('button.ui-btn--primary').trigger('click')

    expect(wrapper.emitted('save')).toBeTruthy()
    expect(wrapper.emitted('save')?.length).toBe(1)
  })
})
