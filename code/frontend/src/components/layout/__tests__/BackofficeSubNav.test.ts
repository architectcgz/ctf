import { beforeEach, describe, expect, it } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'

import BackofficeSubNav from '../BackofficeSubNav.vue'
import backofficeSubNavSource from '../BackofficeSubNav.vue?raw'
import { useAuthStore } from '@/stores/auth'

async function mountWithRoute(path: string, role: 'teacher' | 'admin') {
  setActivePinia(createPinia())

  const authStore = useAuthStore()
  authStore.setAuth(
    {
      id: `${role}-1`,
      username: role,
      role,
      name: role,
    })

  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/academy/classes', component: { template: '<div>classes</div>' } },
      { path: '/academy/students', component: { template: '<div>students</div>' } },
      { path: '/academy/awd-reviews', component: { template: '<div>reviews</div>' } },
      { path: '/academy/instances', component: { template: '<div>instances</div>' } },
      { path: '/platform/challenges', component: { template: '<div>challenges</div>' } },
      { path: '/platform/challenges/:id/writeup', component: { template: '<div>writeup</div>' } },
      { path: '/platform/awd-challenges', component: { template: '<div>awd</div>' } },
      { path: '/platform/images', component: { template: '<div>images</div>' } },
      { path: '/platform/contests', component: { template: '<div>contests</div>' } },
      { path: '/platform/users', component: { template: '<div>users</div>' } },
      { path: '/platform/integrity', component: { template: '<div>integrity</div>' } },
      { path: '/platform/audit', component: { template: '<div>audit</div>' } },
    ],
  })

  await router.push(path)
  await router.isReady()

  const wrapper = mount(BackofficeSubNav, {
    global: {
      plugins: [router],
    },
  })

  await flushPromises()

  return { wrapper }
}

describe('BackofficeSubNav', () => {
  beforeEach(() => {
    document.body.innerHTML = ''
  })

  it('shows teaching operation entries for teacher academy routes', async () => {
    const { wrapper } = await mountWithRoute('/academy/classes', 'teacher')

    expect(wrapper.text()).toContain('班级管理')
    expect(wrapper.text()).toContain('学生管理')
    expect(wrapper.text()).toContain('AWD复盘')
    expect(wrapper.text()).toContain('实例管理')
    expect(wrapper.text()).not.toContain('用户管理')
  })

  it('shows governance entries only for admin routes', async () => {
    const { wrapper } = await mountWithRoute('/platform/contests', 'admin')

    expect(wrapper.text()).toContain('竞赛目录')
    expect(wrapper.text()).toContain('用户管理')
    expect(wrapper.text()).toContain('作弊检测')
    expect(wrapper.text()).toContain('审计日志')
  })

  it('keeps the owning secondary entry active for detail routes', async () => {
    const { wrapper } = await mountWithRoute('/platform/challenges/11/writeup', 'admin')

    const activeButton = wrapper.find('.backoffice-subnav__item--active')
    expect(activeButton.exists()).toBe(true)
    expect(activeButton.text()).toContain('Jeopardy题库')
  })

  it('uses shared theme tokens instead of hardcoded light navigation colors', () => {
    expect(backofficeSubNavSource).toContain('var(--color-border-default)')
    expect(backofficeSubNavSource).toContain('var(--color-bg-surface)')
    expect(backofficeSubNavSource).toContain('var(--color-text-secondary)')
    expect(backofficeSubNavSource).toContain('var(--color-primary)')
    expect(backofficeSubNavSource).not.toContain('#e2e8f0')
    expect(backofficeSubNavSource).not.toContain('#64748b')
    expect(backofficeSubNavSource).not.toContain('#0f172a')
    expect(backofficeSubNavSource).not.toContain('#2563eb')
    expect(backofficeSubNavSource).not.toContain('background: white;')
  })
})
