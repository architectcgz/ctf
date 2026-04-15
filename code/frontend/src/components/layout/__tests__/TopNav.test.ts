import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'

import TopNav from '../TopNav.vue'
import topNavSource from '../TopNav.vue?raw'
import { useAuthStore } from '@/stores/auth'

const authMocks = vi.hoisted(() => ({
  logout: vi.fn(),
}))

vi.mock('@/composables/useAuth', () => ({
  useAuth: () => authMocks,
}))

vi.mock('@/components/layout/NotificationDropdown.vue', () => ({
  default: {
    name: 'NotificationDropdown',
    props: ['realtimeStatus'],
    template: '<div class="notification-dropdown-stub" />',
  },
}))

function createTestRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      {
        path: '/student/dashboard',
        component: { template: '<div>dashboard</div>' },
        meta: { title: '仪表盘' },
      },
      {
        path: '/academy/overview',
        component: { template: '<div>academy</div>' },
        meta: { title: '教学概览' },
      },
      {
        path: '/admin/dashboard',
        component: { template: '<div>admin</div>' },
        meta: { title: '系统概览' },
      },
    ],
  })
}

async function mountTopNav() {
  setActivePinia(createPinia())
  localStorage.clear()
  document.documentElement.removeAttribute('data-brand')
  document.documentElement.removeAttribute('data-theme')
  authMocks.logout.mockReset()

  const authStore = useAuthStore()
  authStore.setAuth(
    {
      id: 'student-1',
      username: 'alice',
      name: 'Alice',
      role: 'student',
    },
    'token'
  )

  const router = createTestRouter()
  await router.push('/student/dashboard')
  await router.isReady()

  const wrapper = mount(TopNav, {
    attachTo: document.body,
    props: {
      sidebarCollapsed: false,
      notificationStatus: 'open',
    },
    global: {
      plugins: [router],
    },
  })

  await flushPromises()

  return { wrapper }
}

describe('TopNav', () => {
  beforeEach(() => {
    document.body.innerHTML = ''
  })

  it('保持紧凑头部布局并展示当前用户信息', async () => {
    const { wrapper } = await mountTopNav()

    expect(wrapper.find('.topnav-main').exists()).toBe(true)
    expect(wrapper.find('.topnav-actions').exists()).toBe(true)
    expect(wrapper.find('.topnav-user-name').text()).toBe('Alice')
    expect(wrapper.find('.topnav-user-role').text()).toBe('学生空间')

    wrapper.unmount()
  })

  it('点击调色盘后会弹出 4 个主题色圆点并完成切换', async () => {
    const { wrapper } = await mountTopNav()
    const paletteButton = wrapper.find('button[aria-label="切换主题色"]')

    expect(paletteButton.attributes('aria-expanded')).toBe('false')

    await paletteButton.trigger('click')
    await flushPromises()

    expect(paletteButton.attributes('aria-expanded')).toBe('true')

    const brandOptions = wrapper.findAll('button[role="menuitemradio"]')
    expect(brandOptions).toHaveLength(4)
    expect(wrapper.find('#topnav-brand-picker-panel').exists()).toBe(true)

    const orangeOption = wrapper.find('button[aria-label="切换到橙色主题"]')
    await orangeOption.trigger('click')
    await flushPromises()

    expect(localStorage.getItem('theme-brand')).toBe('orange')
    expect(document.documentElement.getAttribute('data-brand')).toBe('orange')
    expect(wrapper.find('#topnav-brand-picker-panel').exists()).toBe(false)

    wrapper.unmount()
  })

  it('支持点击外部和按 Esc 关闭主题色面板', async () => {
    const { wrapper } = await mountTopNav()
    const paletteButton = wrapper.find('button[aria-label="切换主题色"]')

    await paletteButton.trigger('click')
    await flushPromises()
    expect(wrapper.find('#topnav-brand-picker-panel').exists()).toBe(true)

    document.body.dispatchEvent(new MouseEvent('mousedown', { bubbles: true }))
    await flushPromises()
    expect(wrapper.find('#topnav-brand-picker-panel').exists()).toBe(false)

    await paletteButton.trigger('click')
    await flushPromises()
    expect(wrapper.find('#topnav-brand-picker-panel').exists()).toBe(true)

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await flushPromises()
    expect(wrapper.find('#topnav-brand-picker-panel').exists()).toBe(false)

    wrapper.unmount()
  })

  it('通知按钮应当显式保留和相邻工具按钮一致的外边框', () => {
    expect(topNavSource).toMatch(
      /\.topnav-actions\s*:deep\(\.notification-trigger\)\s*\{[\s\S]*border:\s*1px solid #f1f5f9;/s
    )
  })

  it('supports a dedicated admin workspace treatment on platform routes', () => {
    expect(topNavSource).toContain('const isBackofficeRoute = computed(() =>')
    expect(topNavSource).toContain("import { isBackofficeRoute as checkBackofficeRoute }")
    expect(topNavSource).toContain('checkBackofficeRoute(route.path)')
    expect(topNavSource).toContain('topnav-shell--admin')
    expect(topNavSource).toContain('Workspace')
  })

  it('renders backoffice breadcrumbs from sidebar module and submenu instead of the removed horizontal subnav', () => {
    expect(topNavSource).toContain('getBackofficeModuleByPath')
    expect(topNavSource).toContain('getVisibleBackofficeSecondaryItems')
    expect(topNavSource).toContain('backofficeBreadcrumb')
    expect(topNavSource).toContain('backofficeBreadcrumb.moduleLabel')
    expect(topNavSource).toContain('backofficeBreadcrumb.secondaryLabel')
    expect(topNavSource).not.toContain('Backoffice Workspace / {{ pageTitle }}')
  })
})
