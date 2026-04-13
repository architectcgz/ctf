import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { mount, RouterLinkStub } from '@vue/test-utils'

import ForbiddenView from '../ForbiddenView.vue'
import { useAuthStore } from '@/stores/auth'

describe('ForbiddenView', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it('未登录时应引导回登录页', () => {
    const wrapper = mount(ForbiddenView, {
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
        },
      },
    })

    const links = wrapper.findAllComponents(RouterLinkStub)

    expect(wrapper.text()).toContain('403')
    expect(wrapper.text()).toContain('你当前没有访问这个区域的权限')
    expect(wrapper.text()).toContain('返回上一页')
    expect(wrapper.text()).toContain('返回登录页')
    expect(wrapper.find('aside').exists()).toBe(false)
    expect(links[0]?.props('to')).toBe('/login')
    expect(links).toHaveLength(1)
    expect(wrapper.get('button.error-status-action-primary').text()).toContain('返回上一页')
  })

  it('管理员登录时应引导回管理工作台', () => {
    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'admin-1',
        username: 'root',
        role: 'admin',
      },
      'token'
    )

    const wrapper = mount(ForbiddenView, {
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
        },
      },
    })

    const links = wrapper.findAllComponents(RouterLinkStub)

    expect(wrapper.find('aside').exists()).toBe(false)
    expect(wrapper.text()).toContain('返回上一页')
    expect(wrapper.text()).toContain('返回管理工作台')
    expect(links[0]?.props('to')).toBe('/admin/dashboard')
    expect(links).toHaveLength(1)
    expect(wrapper.get('button.error-status-action-primary').text()).toContain('返回上一页')
  })
})
