import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { mount, RouterLinkStub } from '@vue/test-utils'

vi.mock('@/utils/browser', () => ({
  redirectTo: vi.fn(),
  getNavigationType: vi.fn(() => null),
  reloadPage: vi.fn(),
}))

import UnauthorizedView from '../UnauthorizedView.vue'
import TooManyRequestsView from '../TooManyRequestsView.vue'
import InternalServerErrorView from '../InternalServerErrorView.vue'
import BadGatewayView from '../BadGatewayView.vue'
import ServiceUnavailableView from '../ServiceUnavailableView.vue'
import GatewayTimeoutView from '../GatewayTimeoutView.vue'
import { getNavigationType, redirectTo, reloadPage } from '@/utils/browser'

describe('additional error views', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    vi.clearAllMocks()
    window.history.replaceState({}, '', '/')
  })

  it('renders 401 with login-oriented recovery action', () => {
    const wrapper = mount(UnauthorizedView, {
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
        },
      },
    })

    const links = wrapper.findAllComponents(RouterLinkStub)

    expect(wrapper.text()).toContain('401')
    expect(wrapper.text()).toContain('登录状态已失效')
    expect(wrapper.text()).toContain('返回上一页')
    expect(wrapper.text()).not.toContain('通知中心')
    expect(links[0]?.props('to')).toBe('/login')
    expect(links).toHaveLength(1)
  })

  it('renders 429 with a safe fallback and back action', () => {
    const wrapper = mount(TooManyRequestsView, {
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
        },
      },
    })

    const links = wrapper.findAllComponents(RouterLinkStub)

    expect(wrapper.text()).toContain('429')
    expect(wrapper.text()).toContain('请求过于频繁')
    expect(wrapper.text()).toContain('返回上一页')
    expect(wrapper.text()).toContain('返回登录页')
    expect(wrapper.text()).not.toContain('通知中心')
    expect(links[0]?.props('to')).toBe('/login')
    expect(links).toHaveLength(1)
    expect(wrapper.get('button.error-status-action-primary').text()).toContain('返回上一页')
  })

  it('renders server-side failure pages with retry-first and back recovery actions', () => {
    const pages = [
      { component: InternalServerErrorView, code: '500', text: '系统内部错误' },
      { component: BadGatewayView, code: '502', text: '上游服务响应异常' },
      { component: ServiceUnavailableView, code: '503', text: '服务暂时不可用' },
      { component: GatewayTimeoutView, code: '504', text: '服务响应超时' },
    ]

    for (const page of pages) {
      const wrapper = mount(page.component, {
        global: {
          stubs: {
            RouterLink: RouterLinkStub,
          },
        },
      })

      const links = wrapper.findAllComponents(RouterLinkStub)

      expect(wrapper.text()).toContain(page.code)
      expect(wrapper.text()).toContain(page.text)
      expect(wrapper.text()).toContain('刷新页面')
      expect(wrapper.text()).toContain('返回上一页')
      expect(wrapper.text()).not.toContain('通知中心')
      expect(links).toHaveLength(0)
      expect(wrapper.get('button.error-status-action-primary').text()).toContain('刷新页面')
      expect(wrapper.get('button.error-status-action-secondary').text()).toContain('返回上一页')
    }
  })

  it('prefers navigating back to the recorded source route when retrying from /500', async () => {
    window.history.replaceState({}, '', '/500?from=%2Fchallenges%2F5')

    const wrapper = mount(InternalServerErrorView, {
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
        },
      },
    })

    await wrapper.get('button.error-status-action-primary').trigger('click')

    expect(redirectTo).toHaveBeenCalledWith('/challenges/5')
    expect(reloadPage).not.toHaveBeenCalled()
  })

  it('automatically retries the recorded source route when the browser reloads /500', () => {
    vi.mocked(getNavigationType).mockReturnValue('reload')
    window.history.replaceState({}, '', '/500?from=%2Fchallenges%2F5')

    mount(InternalServerErrorView, {
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
        },
      },
    })

    expect(redirectTo).toHaveBeenCalledWith('/challenges/5')
    expect(reloadPage).not.toHaveBeenCalled()
  })
})
