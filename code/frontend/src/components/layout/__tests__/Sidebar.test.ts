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
      /<aside\s+class="[^"]*sidebar-shell-desktop[^"]*min-h-screen[^"]*self-stretch[^"]*"/s
    )
    expect(sidebarSource).toMatch(
      /<nav class="[^"]*flex[^"]*min-h-full[^"]*flex-col[^"]*space-y-7[^"]*">/s
    )
  })

  it('uses a flatter console navigation system instead of stacked card buttons', () => {
    expect(sidebarSource).toContain('class="sidebar-brand-button flex min-w-0 items-center gap-3 px-2.5 py-2 text-left transition"')
    expect(sidebarSource).toContain('sidebar-nav-scroll')
    expect(sidebarSource).toContain('sidebar-group-title--collapsed')
    expect(sidebarSource).toContain('.sidebar-item-active::before,')
    expect(sidebarSource).toContain('.sidebar-item-button--collapsed::before')
  })

  it('uses a role-neutral profile url when admin opens the profile page from sidebar', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/admin/dashboard', component: { template: '<div>admin</div>' } },
        { path: '/profile', component: { template: '<div>profile</div>' } },
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

    const profileButton = wrapper.findAll('.sidebar-shell-desktop button').find((node) => node.text().includes('个人资料'))

    expect(profileButton).toBeTruthy()

    await profileButton!.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.fullPath).toBe('/profile')

    wrapper.unmount()
  })

  it('uses academy paths for shared teaching entries when admin navigates from the sidebar', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/admin/dashboard', component: { template: '<div>admin</div>' } },
        { path: '/academy/classes', component: { template: '<div>academy classes</div>' } },
        { path: '/teacher/classes', component: { template: '<div>legacy teacher classes</div>' } },
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

    const classButton = wrapper.findAll('.sidebar-shell-desktop button').find((node) => node.text().includes('班级管理'))

    expect(classButton).toBeTruthy()

    await classButton!.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.fullPath).toBe('/academy/classes')

    wrapper.unmount()
  })
})
