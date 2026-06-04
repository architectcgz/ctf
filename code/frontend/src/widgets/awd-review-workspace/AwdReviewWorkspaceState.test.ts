import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import AwdReviewWorkspaceState from './AwdReviewWorkspaceState.vue'

describe('AwdReviewWorkspaceState', () => {
  it('错误态应透传重试事件', async () => {
    const wrapper = mount(AwdReviewWorkspaceState, {
      props: {
        loading: false,
        error: '加载失败',
        hasReview: false,
      },
    })

    const reloadButton = wrapper.get('button')
    expect(reloadButton.text()).toContain('重新加载')

    await reloadButton.trigger('click')

    expect(wrapper.emitted('loadReview')).toBeTruthy()
  })

  it('有数据时应渲染默认插槽', () => {
    const wrapper = mount(AwdReviewWorkspaceState, {
      props: {
        loading: false,
        error: null,
        hasReview: true,
      },
      slots: {
        default: '<div>loaded</div>',
      },
    })

    expect(wrapper.text()).toContain('loaded')
  })
})
