import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import UserManage from '../UserManage.vue'

describe('UserManage', () => {
  it('应该渲染降级态并明确说明后端接口缺口', async () => {
    const wrapper = mount(UserManage)

    await wrapper.vm.$nextTick()

    expect(wrapper.exists()).toBe(true)
    expect(wrapper.text()).toContain('用户管理')
    expect(wrapper.text()).toContain('后端接口待补齐')
    expect(wrapper.text()).toContain('暂时无法展示用户列表')
    expect(wrapper.text()).toContain('/api/v1/admin/users')
    expect(wrapper.text()).not.toContain('张三')
  })
})
