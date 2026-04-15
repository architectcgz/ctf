import { beforeEach, describe, expect, it } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'

import BackofficeSubNav from '../BackofficeSubNav.vue'
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
    },
    'token'
  )

  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/academy/classes', component: { template: '<div>classes</div>' } },
      { path: '/academy/students', component: { template: '<div>students</div>' } },
      { path: '/academy/awd-reviews', component: { template: '<div>reviews</div>' } },
      { path: '/academy/instances', component: { template: '<div>instances</div>' } },
      { path: '/platform/challenges', component: { template: '<div>challenges</div>' } },
      { path: '/platform/challenges/:id/writeup', component: { template: '<div>writeup</div>' } },
      { path: '/platform/environment-templates', component: { template: '<div>env</div>' } },
      { path: '/platform/images', component: { template: '<div>images</div>' } },
      { path: '/admin/contests', component: { template: '<div>contests</div>' } },
      { path: '/admin/users', component: { template: '<div>users</div>' } },
      { path: '/admin/integrity', component: { template: '<div>integrity</div>' } },
      { path: '/admin/audit', component: { template: '<div>audit</div>' } },
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
    const { wrapper } = await mountWithRoute('/admin/contests', 'admin')

    expect(wrapper.text()).toContain('竞赛管理')
    expect(wrapper.text()).toContain('用户管理')
    expect(wrapper.text()).toContain('作弊检测')
    expect(wrapper.text()).toContain('审计日志')
  })

  it('keeps the owning secondary entry active for detail routes', async () => {
    const { wrapper } = await mountWithRoute('/platform/challenges/11/writeup', 'admin')

    const activeButton = wrapper.find('.backoffice-subnav__item--active')
    expect(activeButton.exists()).toBe(true)
    expect(activeButton.text()).toContain('题目管理')
  })
})
