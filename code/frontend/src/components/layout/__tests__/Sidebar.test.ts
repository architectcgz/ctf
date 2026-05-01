import { beforeEach, describe, expect, it } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'

import sidebarSource from '../Sidebar.vue?raw'
import Sidebar from '../Sidebar.vue'
import { useAuthStore } from '@/stores/auth'

describe('Sidebar desktop layout', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('stretches the desktop nav to align its bottom edge with the content area', () => {
    expect(sidebarSource).toMatch(
      /<aside[\s\S]*class="[^"]*backoffice-sidebar[^"]*backoffice-sidebar--desktop[^"]*sticky[^"]*top-0[^"]*z-\[60\][^"]*hidden[^"]*min-h-screen[^"]*shrink-0[^"]*flex-col[^"]*md:flex"/s
    )
    expect(sidebarSource).toContain(
      ":class=\"collapsed ? 'w-20' : 'backoffice-sidebar--expanded'\""
    )
    expect(sidebarSource).toContain('.backoffice-sidebar--expanded')
    expect(sidebarSource).toMatch(
      /<nav[\s\S]*class="[^"]*flex-1[^"]*space-y-1\.5[^"]*overflow-x-hidden[^"]*"/s
    )
  })

  it('matches the admin example sidebar shell structure instead of a custom console variant', () => {
    expect(sidebarSource).toContain('backoffice-sidebar__collapse')
    expect(sidebarSource).toContain('backoffice-sidebar__header')
    expect(sidebarSource).toContain('Workspace')
    expect(sidebarSource).toContain('backoffice-sidebar__children')
    expect(sidebarSource).toContain('backoffice-sidebar__child')
    expect(sidebarSource).toContain('backoffice-sidebar__child-indicator')
  })

  it('avoids low-information arbitrary Tailwind values in the shared backoffice shell', () => {
    expect(sidebarSource).not.toContain('w-[260px]')
    expect(sidebarSource).not.toContain('text-[10px]')
    expect(sidebarSource).not.toContain('text-[13px]')
    expect(sidebarSource).not.toContain('h-[18px]')
    expect(sidebarSource).not.toContain('w-[18px]')
    expect(sidebarSource).not.toContain('ml-[22px]')
    expect(sidebarSource).not.toContain('-left-[14px]')
    expect(sidebarSource).not.toContain('w-[3px]')
  })

  it('uses the shared workspace shell identity composable across student, academy and platform routes', () => {
    expect(sidebarSource).toContain('useWorkspaceShellNavigation')
    expect(sidebarSource).not.toContain("route.path.startsWith('/admin/')")
    expect(sidebarSource).toContain('brandKicker')
    expect(sidebarSource).toContain('shell.value.brandKicker')
  })

  it('tokenizes backoffice sidebar surfaces for dark theme instead of keeping white utility backgrounds', () => {
    expect(sidebarSource).toContain('backoffice-sidebar')
    expect(sidebarSource).toContain('--backoffice-shell-surface')
    expect(sidebarSource).toContain(":global([data-theme='dark']) .backoffice-sidebar")
    expect(sidebarSource).toContain('backoffice-sidebar__collapse')
  })

  it('uses unified backoffice modules instead of raw main/teacher/admin route buckets', () => {
    expect(sidebarSource).toContain('shell.value.modules')
    expect(sidebarSource).toContain('backofficeModuleIconMap')
    expect(sidebarSource).toContain('currentBackofficeModuleKey')
    expect(sidebarSource).toContain('useWorkspaceShellNavigation')
    expect(sidebarSource).toContain('backofficeNavGroups')
    expect(sidebarSource).toContain('defaultNavGroups')
  })

  it('renders student users inside the same backoffice sidebar chrome', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/student/dashboard', component: { template: '<div>dashboard</div>' } },
        { path: '/challenges', component: { template: '<div>challenges</div>' } },
        { path: '/student/instances', component: { template: '<div>instances</div>' } },
      ],
    })

    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'student-1',
      username: 'alice',
      role: 'student',
      name: 'Alice',
    })

    await router.push('/student/dashboard')
    await router.isReady()

    const wrapper = mount(Sidebar, {
      props: {
        collapsed: false,
        mobileOpen: false,
      },
      global: {
        plugins: [router],
      },
    })

    expect(wrapper.find('.backoffice-sidebar--desktop').exists()).toBe(true)
    expect(wrapper.find('.sidebar-shell-desktop').exists()).toBe(false)
    expect(wrapper.text()).toContain('Student Space')
    expect(wrapper.text()).toContain('训练')
    expect(wrapper.text()).toContain('仪表盘')
    expect(wrapper.text()).toContain('题目')

    wrapper.unmount()
  })

  it('shows the four primary backoffice modules for admin users', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/platform/overview', component: { template: '<div>admin</div>' } },
        { path: '/academy/classes', component: { template: '<div>classes</div>' } },
        { path: '/platform/challenges', component: { template: '<div>challenges</div>' } },
        { path: '/platform/contest-ops/contests', component: { template: '<div>event ops</div>' } },
        { path: '/platform/contests', component: { template: '<div>contests</div>' } },
      ],
    })

    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'admin-1',
      username: 'admin',
      role: 'admin',
      name: 'Admin',
    })

    await router.push('/platform/overview')
    await router.isReady()

    const wrapper = mount(Sidebar, {
      props: {
        collapsed: false,
        mobileOpen: false,
      },
      global: {
        plugins: [router],
      },
    })

    expect(wrapper.text()).toContain('总览')
    expect(wrapper.text()).toContain('教学运营')
    expect(wrapper.text()).toContain('题库与资源')
    expect(wrapper.text()).toContain('赛事运维')
    expect(wrapper.text()).toContain('系统治理')

    wrapper.unmount()
  })

  it('expands the contest operations module and renders its secondary entries for admin users', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/platform/overview', component: { template: '<div>admin</div>' } },
        {
          path: '/platform/contest-ops/contests',
          component: { template: '<div>contest management</div>' },
        },
        {
          path: '/platform/contest-ops/projector',
          component: { template: '<div>projector</div>' },
        },
        { path: '/platform/contests/:id/manage', component: { template: '<div>manage</div>' } },
      ],
    })

    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'admin-1',
      username: 'admin',
      role: 'admin',
      name: 'Admin',
    })

    await router.push('/platform/contest-ops/contests')
    await router.isReady()

    const wrapper = mount(Sidebar, {
      props: {
        collapsed: false,
        mobileOpen: false,
      },
      global: {
        plugins: [router],
      },
    })

    expect(wrapper.text()).toContain('赛事运维')
    expect(wrapper.text()).toContain('竞赛列表')
    expect(wrapper.text()).toContain('大屏展示')
    expect(wrapper.text()).not.toContain('环境管理')
    expect(wrapper.text()).not.toContain('流量监控')
    expect(wrapper.text()).not.toContain('大屏投射')
    expect(wrapper.text()).not.toContain('排行榜')

    wrapper.unmount()
  })

  it('keeps contest manage pages highlighted under 赛事运维 instead of 系统治理', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/platform/overview', component: { template: '<div>admin</div>' } },
        {
          path: '/platform/contest-ops/contests',
          component: { template: '<div>contest management</div>' },
        },
        { path: '/platform/contests/:id/manage', component: { template: '<div>manage</div>' } },
        { path: '/platform/contests', component: { template: '<div>contests</div>' } },
      ],
    })

    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'admin-1',
      username: 'admin',
      role: 'admin',
      name: 'Admin',
    })

    await router.push('/platform/contests/contest-1/manage')
    await router.isReady()

    const wrapper = mount(Sidebar, {
      props: {
        collapsed: false,
        mobileOpen: false,
      },
      global: {
        plugins: [router],
      },
    })

    const activeChildren = wrapper
      .findAll('.backoffice-sidebar__child--active')
      .map((node) => node.text().trim())

    expect(activeChildren).toContain('竞赛列表')
    expect(activeChildren).not.toContain('竞赛目录')

    wrapper.unmount()
  })

  it('navigates admin users to the canonical teaching module entry', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/platform/overview', component: { template: '<div>admin</div>' } },
        { path: '/platform/classes', component: { template: '<div>platform classes</div>' } },
        { path: '/academy/classes', component: { template: '<div>academy classes</div>' } },
        { path: '/platform/challenges', component: { template: '<div>challenges</div>' } },
        { path: '/platform/contests', component: { template: '<div>contests</div>' } },
      ],
    })

    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'admin-1',
      username: 'admin',
      role: 'admin',
      name: 'Admin',
    })

    await router.push('/platform/overview')
    await router.isReady()

    const wrapper = mount(Sidebar, {
      props: {
        collapsed: false,
        mobileOpen: false,
      },
      global: {
        plugins: [router],
      },
    })

    const operationsButton = wrapper
      .findAll('button')
      .find((node) => node.text().includes('教学运营'))

    expect(operationsButton).toBeTruthy()

    await operationsButton!.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.fullPath).toBe('/platform/classes')

    wrapper.unmount()
  })

  it('expands the active backoffice module and renders its secondary entries inside the sidebar', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/academy/overview', component: { template: '<div>overview</div>' } },
        { path: '/academy/classes', component: { template: '<div>classes</div>' } },
        { path: '/academy/students', component: { template: '<div>students</div>' } },
        { path: '/academy/awd-reviews', component: { template: '<div>awd reviews</div>' } },
        { path: '/academy/instances', component: { template: '<div>instances</div>' } },
        { path: '/platform/challenges', component: { template: '<div>challenges</div>' } },
      ],
    })

    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'teacher-1',
      username: 'teacher',
      role: 'teacher',
      name: 'Teacher',
    })

    await router.push('/academy/classes')
    await router.isReady()

    const wrapper = mount(Sidebar, {
      props: {
        collapsed: false,
        mobileOpen: false,
      },
      global: {
        plugins: [router],
      },
    })

    expect(wrapper.text()).toContain('总览')
    expect(wrapper.text()).toContain('教学运营')
    expect(wrapper.text()).toContain('题库与资源')
    expect(wrapper.text()).toContain('班级管理')
    expect(wrapper.text()).toContain('学生管理')
    expect(wrapper.text()).toContain('AWD复盘')
    expect(wrapper.text()).toContain('实例管理')

    wrapper.unmount()
  })

  it('highlights only the matched secondary entry inside the expanded backoffice module', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/academy/overview', component: { template: '<div>overview</div>' } },
        { path: '/academy/classes', component: { template: '<div>classes</div>' } },
        { path: '/academy/students', component: { template: '<div>students</div>' } },
        { path: '/academy/awd-reviews', component: { template: '<div>awd reviews</div>' } },
        { path: '/academy/instances', component: { template: '<div>instances</div>' } },
      ],
    })

    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'teacher-1',
      username: 'teacher',
      role: 'teacher',
      name: 'Teacher',
    })

    await router.push('/academy/classes')
    await router.isReady()

    const wrapper = mount(Sidebar, {
      props: {
        collapsed: false,
        mobileOpen: false,
      },
      global: {
        plugins: [router],
      },
    })

    const classButtons = wrapper
      .findAll('button')
      .filter((node) => node.text().includes('班级管理'))
    const studentButtons = wrapper
      .findAll('button')
      .filter((node) => node.text().includes('学生管理'))
    const reviewButtons = wrapper
      .findAll('button')
      .filter((node) => node.text().includes('AWD复盘'))

    expect(classButtons.length).toBeGreaterThan(0)
    expect(studentButtons.length).toBeGreaterThan(0)
    expect(reviewButtons.length).toBeGreaterThan(0)
    expect(
      classButtons.every((node) => node.classes().includes('backoffice-sidebar__child--active'))
    ).toBe(true)
    expect(
      studentButtons.every((node) => !node.classes().includes('backoffice-sidebar__child--active'))
    ).toBe(true)
    expect(
      reviewButtons.every((node) => !node.classes().includes('backoffice-sidebar__child--active'))
    ).toBe(true)

    wrapper.unmount()
  })

  it('navigates teacher users when clicking operations and resources secondary entries', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/academy/overview', component: { template: '<div>overview</div>' } },
        { path: '/academy/classes', component: { template: '<div>classes</div>' } },
        { path: '/academy/students', component: { template: '<div>students</div>' } },
        { path: '/academy/awd-reviews', component: { template: '<div>awd reviews</div>' } },
        { path: '/academy/instances', component: { template: '<div>instances</div>' } },
        { path: '/platform/challenges', component: { template: '<div>challenges</div>' } },
        { path: '/platform/awd-challenges', component: { template: '<div>awd challenges</div>' } },
        { path: '/platform/images', component: { template: '<div>images</div>' } },
      ],
    })

    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'teacher-1',
      username: 'teacher',
      role: 'teacher',
      name: 'Teacher',
    })

    await router.push('/academy/overview')
    await router.isReady()

    const wrapper = mount(Sidebar, {
      props: {
        collapsed: false,
        mobileOpen: false,
      },
      global: {
        plugins: [router],
      },
    })

    const desktopAside = wrapper.findAll('aside').at(-1)

    expect(desktopAside).toBeTruthy()

    async function clickSidebarEntry(label: string, expectedPath: string): Promise<void> {
      const button = desktopAside!.findAll('button').find((node) => node.text().trim() === label)

      expect(button).toBeTruthy()

      await button!.trigger('click')
      await flushPromises()

      expect(router.currentRoute.value.path).toBe(expectedPath)
    }

    await clickSidebarEntry('教学运营', '/academy/classes')
    await clickSidebarEntry('学生管理', '/academy/students')
    await clickSidebarEntry('AWD复盘', '/academy/awd-reviews')
    await clickSidebarEntry('实例管理', '/academy/instances')
    await clickSidebarEntry('题库与资源', '/platform/challenges')
    await clickSidebarEntry('AWD题库', '/platform/awd-challenges')
    await clickSidebarEntry('镜像管理', '/platform/images')

    wrapper.unmount()
  })

  it('allows the active backoffice module to collapse after the user toggles it closed', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/academy/overview', component: { template: '<div>overview</div>' } },
        { path: '/academy/classes', component: { template: '<div>classes</div>' } },
        { path: '/academy/students', component: { template: '<div>students</div>' } },
        { path: '/academy/awd-reviews', component: { template: '<div>awd reviews</div>' } },
        { path: '/academy/instances', component: { template: '<div>instances</div>' } },
      ],
    })

    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'teacher-1',
      username: 'teacher',
      role: 'teacher',
      name: 'Teacher',
    })

    await router.push('/academy/classes')
    await router.isReady()

    const wrapper = mount(Sidebar, {
      props: {
        collapsed: false,
        mobileOpen: false,
      },
      global: {
        plugins: [router],
      },
    })

    const desktopAside = wrapper.findAll('aside').at(-1)

    expect(desktopAside).toBeTruthy()
    expect(desktopAside!.text()).toContain('学生管理')
    expect(desktopAside!.text()).toContain('AWD复盘')

    const operationsButton = desktopAside!
      .findAll('button')
      .find((node) => node.text().trim() === '教学运营')

    expect(operationsButton).toBeTruthy()

    const toggleIcon = operationsButton!
      .findAll('svg')
      .find((node) => node.classes().some((className) => className.includes('chevron-down')))

    expect(toggleIcon).toBeTruthy()

    await toggleIcon!.trigger('click')
    await flushPromises()

    expect(desktopAside!.text()).not.toContain('学生管理')
    expect(desktopAside!.text()).not.toContain('AWD复盘')

    wrapper.unmount()
  })

  it('hides governance from teacher users while keeping overview, operations and resources', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/academy/overview', component: { template: '<div>overview</div>' } },
        { path: '/platform/challenges', component: { template: '<div>challenges</div>' } },
      ],
    })

    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'teacher-1',
      username: 'teacher',
      role: 'teacher',
      name: 'Teacher',
    })

    await router.push('/academy/overview')
    await router.isReady()

    const wrapper = mount(Sidebar, {
      props: {
        collapsed: false,
        mobileOpen: false,
      },
      global: {
        plugins: [router],
      },
    })

    expect(wrapper.text()).toContain('总览')
    expect(wrapper.text()).toContain('教学运营')
    expect(wrapper.text()).toContain('题库与资源')
    expect(wrapper.text()).not.toContain('系统治理')

    wrapper.unmount()
  })
})
