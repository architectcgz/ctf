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
      /<aside[\s\S]*class="[^"]*relative[^"]*z-30[^"]*hidden[^"]*h-screen[^"]*shrink-0[^"]*flex-col[^"]*border-r[^"]*border-slate-200[^"]*bg-white[^"]*md:flex"/s
    )
    expect(sidebarSource).toMatch(
      /<nav[\s\S]*class="[^"]*flex-1[^"]*space-y-1\.5[^"]*overflow-x-hidden[^"]*"/s
    )
  })

  it('matches the admin example sidebar shell structure instead of a custom console variant', () => {
    expect(sidebarSource).toContain(
      "absolute -right-3.5 top-6 bg-white border border-slate-200 rounded-full p-1.5 text-slate-400 hover:text-blue-600 hover:border-blue-300 hover:shadow-md shadow-sm z-50 transition-all cursor-pointer"
    )
    expect(sidebarSource).toContain(
      'class="h-16 flex items-center px-5 border-b border-slate-100 overflow-hidden whitespace-nowrap"'
    )
    expect(sidebarSource).toContain('Main Navigation')
    expect(sidebarSource).toContain(
      "mt-1.5 flex flex-col gap-1 pl-11 pr-2 animate-in slide-in-from-top-2 duration-200"
    )
  })

  it('uses the same ChallengeOps shell identity across academy and platform backoffice routes', () => {
    expect(sidebarSource).toContain('const isBackofficeRoute = computed(')
    expect(sidebarSource).toContain("route.path.startsWith('/academy/')")
    expect(sidebarSource).toContain("route.path.startsWith('/platform/')")
    expect(sidebarSource).toContain("route.path.startsWith('/admin/')")
    expect(sidebarSource).toContain('sidebar-shell--admin')
    expect(sidebarSource).toContain('ChallengeOps')
    expect(sidebarSource).not.toContain('Academic Ops')
  })

  it('uses unified backoffice modules instead of raw main/teacher/admin route buckets', () => {
    expect(sidebarSource).toContain('getVisibleBackofficeModules')
    expect(sidebarSource).toContain('getVisibleBackofficeSecondaryItems')
    expect(sidebarSource).toContain('backofficeModuleIconMap')
    expect(sidebarSource).toContain('currentBackofficeModuleKey')
    expect(sidebarSource).toContain('isBackofficeRoute')
    expect(sidebarSource).toContain('backofficeNavGroups')
    expect(sidebarSource).toContain('defaultNavGroups')
  })

  it('shows the four primary backoffice modules for admin users', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/admin/dashboard', component: { template: '<div>admin</div>' } },
        { path: '/academy/classes', component: { template: '<div>classes</div>' } },
        { path: '/platform/challenges', component: { template: '<div>challenges</div>' } },
        { path: '/admin/contest-ops/environment', component: { template: '<div>event ops</div>' } },
        { path: '/admin/contests', component: { template: '<div>contests</div>' } },
      ],
    })

    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'admin-1',
        username: 'admin',
        role: 'admin',
        name: 'Admin',
      },
      'token'
    )

    await router.push('/admin/dashboard')
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
        { path: '/admin/dashboard', component: { template: '<div>admin</div>' } },
        { path: '/admin/contest-ops/environment', component: { template: '<div>environment</div>' } },
        { path: '/admin/contest-ops/traffic', component: { template: '<div>traffic</div>' } },
        { path: '/admin/contest-ops/projector', component: { template: '<div>projector</div>' } },
        { path: '/admin/contest-ops/scoreboard', component: { template: '<div>scoreboard</div>' } },
      ],
    })

    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'admin-1',
        username: 'admin',
        role: 'admin',
        name: 'Admin',
      },
      'token'
    )

    await router.push('/admin/contest-ops/environment')
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
    expect(wrapper.text()).toContain('环境管理')
    expect(wrapper.text()).toContain('流量监控')
    expect(wrapper.text()).toContain('大屏投射')
    expect(wrapper.text()).toContain('排行榜')

    wrapper.unmount()
  })

  it('navigates admin users to the canonical teaching module entry', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/admin/dashboard', component: { template: '<div>admin</div>' } },
        { path: '/academy/classes', component: { template: '<div>academy classes</div>' } },
        { path: '/platform/challenges', component: { template: '<div>challenges</div>' } },
        { path: '/admin/contests', component: { template: '<div>contests</div>' } },
      ],
    })

    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'admin-1',
        username: 'admin',
        role: 'admin',
        name: 'Admin',
      },
      'token'
    )

    await router.push('/admin/dashboard')
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

    const operationsButton = wrapper.findAll('button').find((node) => node.text().includes('教学运营'))

    expect(operationsButton).toBeTruthy()

    await operationsButton!.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.fullPath).toBe('/academy/classes')

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
    authStore.setAuth(
      {
        id: 'teacher-1',
        username: 'teacher',
        role: 'teacher',
        name: 'Teacher',
      },
      'token'
    )

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
    authStore.setAuth(
      {
        id: 'teacher-1',
        username: 'teacher',
        role: 'teacher',
        name: 'Teacher',
      },
      'token'
    )

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

    const classButtons = wrapper.findAll('button').filter((node) => node.text().includes('班级管理'))
    const studentButtons = wrapper.findAll('button').filter((node) => node.text().includes('学生管理'))
    const reviewButtons = wrapper.findAll('button').filter((node) => node.text().includes('AWD复盘'))

    expect(classButtons.length).toBeGreaterThan(0)
    expect(studentButtons.length).toBeGreaterThan(0)
    expect(reviewButtons.length).toBeGreaterThan(0)
    expect(classButtons.every((node) => node.classes().includes('text-blue-700'))).toBe(true)
    expect(studentButtons.every((node) => !node.classes().includes('text-blue-700'))).toBe(true)
    expect(reviewButtons.every((node) => !node.classes().includes('text-blue-700'))).toBe(true)

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
    authStore.setAuth(
      {
        id: 'teacher-1',
        username: 'teacher',
        role: 'teacher',
        name: 'Teacher',
      },
      'token'
    )

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
