import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import ScoreboardView from '../ScoreboardView.vue'

describe('ScoreboardView', () => {
  it('应该渲染排行榜页面', async () => {
    const wrapper = mount(ScoreboardView)

    await wrapper.vm.$nextTick()

    expect(wrapper.exists()).toBe(true)
    expect(wrapper.text()).toContain('排行榜')
    expect(wrapper.text()).toContain('Binary Wizards')
  })
})
