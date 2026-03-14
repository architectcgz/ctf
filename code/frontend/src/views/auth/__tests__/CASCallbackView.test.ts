import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import CASCallbackView from '@/views/auth/CASCallbackView.vue'

const casMocks = vi.hoisted(() => ({
  finishCASLogin: vi.fn(),
}))

const routerMocks = vi.hoisted(() => ({
  replace: vi.fn(),
}))

const routeState = vi.hoisted(() => ({
  query: {
    ticket: 'ST-2026',
  } as Record<string, unknown>,
}))

vi.mock('@/composables/useCASAuth', () => ({
  useCASAuth: () => casMocks,
}))
vi.mock('vue-router', () => ({
  useRoute: () => routeState,
  useRouter: () => routerMocks,
}))

describe('CASCallbackView', () => {
  beforeEach(() => {
    casMocks.finishCASLogin.mockReset()
    casMocks.finishCASLogin.mockResolvedValue(undefined)
    routerMocks.replace.mockReset()
    routeState.query = { ticket: 'ST-2026' }
  })

  it('应该在带 ticket 时触发 CAS 回调完成逻辑', async () => {
    mount(CASCallbackView, {
      global: {
        stubs: {
          AppLoading: { template: '<div><slot /></div>' },
          AppEmpty: { template: '<div><slot name="action" /></div>' },
        },
      },
    })

    await flushPromises()

    expect(casMocks.finishCASLogin).toHaveBeenCalledWith('ST-2026', undefined)
  })

  it('应该在缺少 ticket 时展示错误态', async () => {
    routeState.query = {}

    const wrapper = mount(CASCallbackView, {
      global: {
        stubs: {
          AppLoading: { template: '<div><slot /></div>' },
          AppEmpty: {
            props: ['title', 'description'],
            template: '<div><div>{{ title }}</div><div>{{ description }}</div><slot name="action" /></div>',
          },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('CAS 登录未完成')
    expect(wrapper.text()).toContain('未收到 CAS ticket')
    expect(casMocks.finishCASLogin).not.toHaveBeenCalled()
  })
})
