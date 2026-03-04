import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import InstanceList from '../InstanceList.vue'

describe('InstanceList', () => {
  it('应该渲染实例列表页面', async () => {
    const wrapper = mount(InstanceList)

    await wrapper.vm.$nextTick()

    expect(wrapper.exists()).toBe(true)
    expect(wrapper.text()).toContain('我的实例')
    expect(wrapper.text()).toContain('SQL 注入基础')
  })
})
