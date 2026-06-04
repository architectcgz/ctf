import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import ReviewArchiveState from './ReviewArchiveState.vue'

describe('ReviewArchiveState', () => {
  it('错误态应透传 reload 事件', async () => {
    const wrapper = mount(ReviewArchiveState, {
      props: {
        loading: false,
        error: '加载失败',
        hasArchive: false,
      },
    })

    const reloadButton = wrapper.get('button')
    expect(reloadButton.text()).toContain('重新加载')

    await reloadButton.trigger('click')

    expect(wrapper.emitted('reload')).toBeTruthy()
  })

  it('有数据时应渲染默认插槽', () => {
    const wrapper = mount(ReviewArchiveState, {
      props: {
        loading: false,
        error: null,
        hasArchive: true,
      },
      slots: {
        default: '<div>loaded</div>',
      },
    })

    expect(wrapper.text()).toContain('loaded')
  })
})
