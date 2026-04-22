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
    expect(wrapper.text()).toContain('返回上一页')
    expect(wrapper.text()).toContain('返回登录页')
    expect(wrapper.text()).not.toContain('通知中心')
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

    const wrapper = mount(NotFoundView, {
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
        },
      },
    })

    const links = wrapper.findAllComponents(RouterLinkStub)

    expect(wrapper.text()).toContain('返回上一页')
    expect(wrapper.text()).toContain('返回管理工作台')
    expect(wrapper.text()).not.toContain('通知中心')
    expect(links[0]?.props('to')).toBe('/platform/overview')
    expect(links).toHaveLength(1)
    expect(wrapper.get('button.error-status-action-primary').text()).toContain('返回上一页')
  })

  it('连续点击状态区域后应显示路径试探附注', async () => {
    const wrapper = mount(NotFoundView, {
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
        },
      },
    })

    expect(wrapper.text()).not.toContain('路径枚举记录已写入：热情可嘉，命中率一般。')

    const kicker = wrapper.get('.error-status-kicker')
    await kicker.trigger('click')
    await kicker.trigger('click')
    await kicker.trigger('click')
    await kicker.trigger('click')

    expect(wrapper.text()).toContain('路径枚举记录已写入：热情可嘉，命中率一般。')
    expect(wrapper.get('button.error-status-action-primary').text()).toContain('返回上一页')
    expect(wrapper.findAllComponents(RouterLinkStub)[0]?.props('to')).toBe('/login')
  })
})
