import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import ContestManage from '../ContestManage.vue'

describe('ContestManage', () => {
  it('应该渲染竞赛管理页面', async () => {
    const wrapper = mount(ContestManage)

    await wrapper.vm.$nextTick()

    expect(wrapper.exists()).toBe(true)
    expect(wrapper.text()).toContain('竞赛管理')
  })
})
