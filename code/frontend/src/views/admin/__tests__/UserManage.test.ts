import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import UserManage from '../UserManage.vue'

describe('UserManage', () => {
  it('应该渲染用户管理页面', async () => {
    const wrapper = mount(UserManage)

    await wrapper.vm.$nextTick()

    expect(wrapper.exists()).toBe(true)
    expect(wrapper.text()).toContain('用户管理')
    expect(wrapper.text()).toContain('张三')
  })
})
