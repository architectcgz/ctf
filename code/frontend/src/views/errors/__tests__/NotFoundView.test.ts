import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { mount, RouterLinkStub } from '@vue/test-utils'

import NotFoundView from '../NotFoundView.vue'
import { useAuthStore } from '@/stores/auth'

describe('NotFoundView', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it('未登录时应引导回登录页', () => {
    const wrapper = mount(NotFoundView, {
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
        },
      },
    })

    const links = wrapper.findAllComponents(RouterLinkStub)

    expect(wrapper.text()).toContain('404')
    expect(wrapper.text()).toContain('页面不存在')
    expect(wrapper.text()).toContain('返回登录页')
    expect(wrapper.text()).toContain('返回上一页')
    expect(wrapper.text()).not.toContain('通知中心')
    expect(links[0]?.props('to')).toBe('/login')
    expect(links).toHaveLength(1)
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

    const wrapper = mount(NotFoundView, {
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
        },
      },
    })

    const links = wrapper.findAllComponents(RouterLinkStub)

    expect(wrapper.text()).toContain('返回管理工作台')
    expect(wrapper.text()).toContain('返回上一页')
    expect(wrapper.text()).not.toContain('通知中心')
    expect(links[0]?.props('to')).toBe('/admin/dashboard')
    expect(links).toHaveLength(1)
  })
})
