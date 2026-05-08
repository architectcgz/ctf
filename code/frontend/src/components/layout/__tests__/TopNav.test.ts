import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'

import TopNav from '../TopNav.vue'
import topNavSource from '../TopNav.vue?raw'
import { useBackofficeBreadcrumbDetail } from '@/composables/useBackofficeBreadcrumbDetail'
import { useAuthStore } from '@/stores/auth'

const authMocks = vi.hoisted(() => ({
  logout: vi.fn(),
}))

vi.mock('@/features/auth', () => ({
  useAuth: () => authMocks,
}))

vi.mock('@/components/layout/NotificationDrawer.vue', () => ({
  default: {
    name: 'NotificationDrawer',
    props: ['realtimeStatus'],
    template: '<div class="notification-drawer-stub" />',
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
        path: '/academy/classes',
        component: { template: '<div>academy classes</div>' },
        meta: { title: '班级管理' },
      },
      {
        path: '/academy/classes/:className',
        name: 'TeacherClassStudents',
        component: { template: '<div>academy class students</div>' },
        meta: { title: '班级学生' },
      },
      {
        path: '/academy/classes/:className/students/:studentId',
        name: 'TeacherStudentAnalysis',
        component: { template: '<div>teacher student analysis</div>' },
        meta: { title: '学员分析' },
      },
      {
        path: '/academy/awd-reviews',
        component: { template: '<div>teacher awd reviews</div>' },
        meta: { title: 'AWD复盘' },
      },
      {
        path: '/academy/awd-reviews/:contestId',
        name: 'TeacherAWDReviewDetail',
        component: { template: '<div>teacher awd review detail</div>' },
        meta: { title: 'AWD复盘详情' },
      },
      {
        path: '/platform/overview',
        component: { template: '<div>platform</div>' },
        meta: { title: '系统概览' },
      },
      {
        path: '/platform/classes',
        component: { template: '<div>classes</div>' },
        meta: { title: '班级管理' },
      },
      {
        path: '/platform/classes/:className',
        name: 'PlatformClassStudents',
        component: { template: '<div>class students</div>' },
        meta: { title: '班级学生' },
      },
      {
        path: '/platform/classes/:className/students/:studentId',
        name: 'PlatformStudentAnalysis',
        component: { template: '<div>student analysis</div>' },
        meta: { title: '学员分析' },
      },
      {
        path: '/platform/awd-reviews',
        component: { template: '<div>awd reviews</div>' },
        meta: { title: 'AWD复盘' },
      },
      {
        path: '/platform/awd-reviews/:contestId',
        name: 'PlatformAwdReviewDetail',
        component: { template: '<div>awd review detail</div>' },
        meta: { title: 'AWD复盘详情' },
      },
      {
        path: '/platform/challenges/:id',
        name: 'PlatformChallengeDetail',
        component: { template: '<div>challenge detail</div>' },
        meta: { title: '题目详情' },
      },
      {
        path: '/platform/challenges/:id/writeup',
        name: 'PlatformChallengeWriteup',
        component: { template: '<div>challenge writeup</div>' },
        meta: { title: '题解管理' },
      },
      {
        path: '/platform/challenges/imports/:importId',
        name: 'PlatformChallengeImportPreview',
        component: { template: '<div>challenge import preview</div>' },
        meta: { title: '导入预览' },
      },
      {
        path: '/platform/challenges',
        component: { template: '<div>challenges</div>' },
        meta: { title: '题目管理' },
      },
      {
        path: '/platform/contests',
        component: { template: '<div>contests</div>' },
        meta: { title: '竞赛目录' },
      },
      {
        path: '/platform/contests/:id/edit',
        name: 'ContestEdit',
        component: { template: '<div>contest edit</div>' },
        meta: { title: '竞赛工作室' },
      },
      {
        path: '/platform/contests/:id/manage',
        name: 'ContestOperations',
        component: { template: '<div>contest operations</div>' },
        meta: { title: '运维指挥中心' },
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
  authStore.setAuth({
    id: 'student-1',
    username: 'alice',
    name: 'Alice',
    role: 'student',
  })

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

async function mountBackofficeTopNav(
  path = '/platform/overview',
  role: 'admin' | 'teacher' = 'admin'
) {
  setActivePinia(createPinia())
  localStorage.clear()
  document.documentElement.removeAttribute('data-brand')
  document.documentElement.removeAttribute('data-theme')
  authMocks.logout.mockReset()

  const authStore = useAuthStore()
  authStore.setAuth({
    id: 'admin-1',
    username: 'admin',
    name: 'Admin',
    role,
  })

  const router = createTestRouter()
  await router.push(path)
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
    useBackofficeBreadcrumbDetail().setBreadcrumbDetailTitle()
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
      /\.topnav-actions\s*:deep\(\.notification-trigger\)\s*\{[\s\S]*border:\s*1px solid var\(--topnav-line\);/s
    )
  })

  it('uses shared shell classes instead of page-level arbitrary widths and type sizes', () => {
    expect(topNavSource).not.toContain('max-w-[1600px]')
    expect(topNavSource).not.toContain('md:text-[15px]')
  })

  it('supports the shared admin workspace treatment across all role shells', () => {
    expect(topNavSource).toContain('useWorkspaceShellNavigation')
    expect(topNavSource).toContain('topnav-shell--admin')
    expect(topNavSource).toContain('Workspace')
  })

  it('tokenizes backoffice shell surfaces so dark theme does not fall back to white chrome', () => {
    expect(topNavSource).toContain('--topnav-surface')
    expect(topNavSource).toContain('--topnav-line')
    expect(topNavSource).toContain(":global([data-theme='dark']) .topnav-shell--admin")
    expect(topNavSource).toContain(":global([data-theme='dark']) .topnav-tool-cluster--admin")
  })

  it('renders backoffice breadcrumbs from sidebar module and submenu instead of the removed horizontal subnav', () => {
    expect(topNavSource).toContain('useWorkspaceShellNavigation')
    expect(topNavSource).toContain('backofficeBreadcrumb')
    expect(topNavSource).toContain('backofficeBreadcrumb.workspacePath')
    expect(topNavSource).toContain('backofficeBreadcrumb.moduleLabel')
    expect(topNavSource).toContain('backofficeBreadcrumb.modulePath')
    expect(topNavSource).toContain('backofficeBreadcrumb.secondaryLabel')
    expect(topNavSource).toContain('backofficeBreadcrumb.secondaryPath')
    expect(topNavSource).toContain('backofficeBreadcrumb.detailLabel')
    expect(topNavSource).toContain('backofficeBreadcrumb.detailPath')
    expect(topNavSource).toContain('useBackofficeBreadcrumbDetail')
    expect(topNavSource).toContain('class="topnav-breadcrumb__link')
    expect(topNavSource).toContain('@click="navigateBreadcrumb(')
    expect(topNavSource).not.toContain('Backoffice Workspace / {{ pageTitle }}')
  })

  it('后台路径面包屑应支持点击跳转', async () => {
    const { wrapper } = await mountBackofficeTopNav('/platform/overview')
    const breadcrumbLinks = wrapper.findAll('.topnav-breadcrumb button')

    expect(breadcrumbLinks).toHaveLength(3)
    expect(breadcrumbLinks.map((link) => link.text())).toEqual(['Workspace', '总览', '系统概览'])
    expect(breadcrumbLinks[2].attributes('aria-current')).toBe('page')

    wrapper.unmount()
  })

  it('点击后台路径段应跳转到对应配置页面', async () => {
    const { wrapper } = await mountBackofficeTopNav('/platform/classes')
    const breadcrumbButtons = wrapper.findAll('.topnav-breadcrumb button')

    await breadcrumbButtons[0].trigger('click')
    await flushPromises()
    expect(wrapper.vm.$route.path).toBe('/platform/overview')

    await wrapper.vm.$router.push('/platform/classes')
    await flushPromises()
    await breadcrumbButtons[1].trigger('click')
    await flushPromises()
    expect(wrapper.vm.$route.path).toBe('/platform/classes')

    wrapper.unmount()
  })

  it('题目详情页面包屑应追加当前题目编号而不是只停留在题目管理', async () => {
    const { wrapper } = await mountBackofficeTopNav('/platform/challenges/3')
    const breadcrumbButtons = wrapper.findAll('.topnav-breadcrumb button')

    expect(breadcrumbButtons.map((button) => button.text())).toEqual([
      'Workspace',
      '题库与资源',
      '题目管理',
      '题目 #3',
    ])
    expect(breadcrumbButtons[2].attributes('aria-current')).toBeUndefined()
    expect(breadcrumbButtons[3].attributes('aria-current')).toBe('page')

    await breadcrumbButtons[2].trigger('click')
    await flushPromises()
    expect(wrapper.vm.$route.path).toBe('/platform/challenges')

    wrapper.unmount()
  })

  it('题目详情页面包屑应优先显示详情页写入的题目名称', async () => {
    const { wrapper } = await mountBackofficeTopNav('/platform/challenges/3')

    useBackofficeBreadcrumbDetail().setBreadcrumbDetailTitle('双节点演练')
    await flushPromises()

    const breadcrumbButtons = wrapper.findAll('.topnav-breadcrumb button')
    expect(breadcrumbButtons.map((button) => button.text())).toEqual([
      'Workspace',
      '题库与资源',
      '题目管理',
      '双节点演练',
    ])

    wrapper.unmount()
  })

  it('参数化题目子页面也应追加题目详情标题', async () => {
    const { wrapper } = await mountBackofficeTopNav('/platform/challenges/3/writeup')
    const breadcrumbButtons = wrapper.findAll('.topnav-breadcrumb button')

    expect(breadcrumbButtons.map((button) => button.text())).toEqual([
      'Workspace',
      '题库与资源',
      '题目管理',
      '题目 #3',
    ])

    wrapper.unmount()
  })

  it('题目导入预览页应追加导入详情标题', async () => {
    const { wrapper } = await mountBackofficeTopNav('/platform/challenges/imports/import-1')
    const breadcrumbButtons = wrapper.findAll('.topnav-breadcrumb button')

    expect(breadcrumbButtons.map((button) => button.text())).toEqual([
      'Workspace',
      '题库与资源',
      '题目管理',
      '导入 #import-1',
    ])

    wrapper.unmount()
  })

  it('班级列表进入班级详情后应追加班级名称', async () => {
    const { wrapper } = await mountBackofficeTopNav('/platform/classes/Class%20A')
    const breadcrumbButtons = wrapper.findAll('.topnav-breadcrumb button')

    expect(breadcrumbButtons.map((button) => button.text())).toEqual([
      'Workspace',
      '教学运营',
      '学生管理',
      'Class A',
    ])

    wrapper.unmount()
  })

  it('学生详情页应追加学生详情标题', async () => {
    const { wrapper } = await mountBackofficeTopNav('/platform/classes/Class%20A/students/stu-1')
    const breadcrumbButtons = wrapper.findAll('.topnav-breadcrumb button')

    expect(breadcrumbButtons.map((button) => button.text())).toEqual([
      'Workspace',
      '教学运营',
      '学生管理',
      '学生 stu-1',
    ])

    wrapper.unmount()
  })

  it('竞赛详情和运维详情页应追加竞赛详情标题', async () => {
    const { wrapper } = await mountBackofficeTopNav('/platform/contests/contest-1/edit')

    expect(wrapper.findAll('.topnav-breadcrumb button').map((button) => button.text())).toEqual([
      'Workspace',
      '系统治理',
      '竞赛目录',
      '竞赛 #contest-1',
    ])

    await wrapper.vm.$router.push('/platform/contests/contest-1/manage')
    await flushPromises()

    expect(wrapper.findAll('.topnav-breadcrumb button').map((button) => button.text())).toEqual([
      'Workspace',
      '赛事运维',
      '竞赛列表',
      '竞赛 #contest-1',
    ])

    wrapper.unmount()
  })

  it('教师侧详情页也应追加对象详情标题', async () => {
    const { wrapper } = await mountBackofficeTopNav('/academy/awd-reviews/contest-1', 'teacher')

    expect(wrapper.findAll('.topnav-breadcrumb button').map((button) => button.text())).toEqual([
      'Workspace',
      '教学运营',
      'AWD复盘',
      '赛事 #contest-1',
    ])

    await wrapper.vm.$router.push('/academy/classes/Class%20A/students/stu-1')
    await flushPromises()

    expect(wrapper.findAll('.topnav-breadcrumb button').map((button) => button.text())).toEqual([
      'Workspace',
      '教学运营',
      '学生管理',
      '学生 stu-1',
    ])

    wrapper.unmount()
  })

  it('removes the duplicate desktop backoffice sidebar toggle from the global topnav', async () => {
    expect(topNavSource).toContain('v-if="isMobile"')

    const originalWidth = window.innerWidth
    Object.defineProperty(window, 'innerWidth', {
      configurable: true,
      writable: true,
      value: 1280,
    })

    const { wrapper } = await mountBackofficeTopNav()
    const toggleButton = wrapper.find('button[aria-label="折叠导航"]')

    expect(toggleButton.exists()).toBe(false)

    wrapper.unmount()
    Object.defineProperty(window, 'innerWidth', {
      configurable: true,
      writable: true,
      value: originalWidth,
    })
  })

  it('renders student routes with the same topnav shell and workspace breadcrumb treatment', async () => {
    const { wrapper } = await mountTopNav()

    expect(wrapper.find('.topnav-shell--admin').exists()).toBe(true)
    expect(wrapper.find('.topnav-breadcrumb').exists()).toBe(true)
    expect(wrapper.findAll('.topnav-breadcrumb button').map((button) => button.text())).toEqual([
      'Workspace',
      '训练',
      '仪表盘',
    ])

    wrapper.unmount()
  })
})
