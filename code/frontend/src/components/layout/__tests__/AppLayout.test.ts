import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { createPinia, setActivePinia } from 'pinia'

import appLayoutSource from '../AppLayout.vue?raw'
import AppLayout from '../AppLayout.vue'
import routerSource from '../../../router/index.ts?raw'
import platformRoutesSource from '../../../router/routes/platformRoutes.ts?raw'
import studentRoutesSource from '../../../router/routes/studentRoutes.ts?raw'
import teacherRoutesSource from '../../../router/routes/teacherRoutes.ts?raw'
import { platformRoutes } from '@/router/routes/platformRoutes'
import { teacherRoutes } from '@/router/routes/teacherRoutes'
import { useAuthStore } from '@/stores/auth'

const teacherApiMocks = vi.hoisted(() => ({
  getClasses: vi.fn(),
  getStudentsDirectory: vi.fn(),
  listTeacherAWDReviews: vi.fn(),
  getTeacherInstances: vi.fn(),
}))

const adminApiMocks = vi.hoisted(() => ({
  getChallenges: vi.fn(),
  getImages: vi.fn(),
  getLatestChallengePublishRequest: vi.fn(),
  listChallengeImports: vi.fn(),
}))

vi.mock('@/api/teacher', async () => {
  const actual = await vi.importActual<typeof import('@/api/teacher')>('@/api/teacher')
  return {
    ...actual,
    ...teacherApiMocks,
  }
})
vi.mock('@/api/admin/authoring', async () => {
  const actual =
    await vi.importActual<typeof import('@/api/admin/authoring')>('@/api/admin/authoring')
  return {
    ...actual,
    ...adminApiMocks,
  }
})
vi.mock('@/composables/useNotificationRealtime', async () => {
  const { ref } = await vi.importActual<typeof import('vue')>('vue')
  return {
    useNotificationRealtime: () => ({
      start: vi.fn(),
      status: ref('idle'),
    }),
  }
})

const routeSources = [
  routerSource,
  platformRoutesSource,
  studentRoutesSource,
  teacherRoutesSource,
].join('\n')

describe('AppLayout workspace shell', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    Object.values(teacherApiMocks).forEach((mock) => mock.mockReset())
    Object.values(adminApiMocks).forEach((mock) => mock.mockReset())

    teacherApiMocks.getClasses.mockImplementation((params?: unknown) => {
      if (params && typeof params === 'object' && 'page' in params) {
        return Promise.resolve({
          list: [{ name: 'Class A', student_count: 2 }],
          total: 1,
          page: 1,
          page_size: 20,
        })
      }
      return Promise.resolve([{ name: 'Class A', student_count: 2 }])
    })
    teacherApiMocks.getStudentsDirectory.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })
    teacherApiMocks.listTeacherAWDReviews.mockResolvedValue([])
    teacherApiMocks.getTeacherInstances.mockResolvedValue([])
    adminApiMocks.getChallenges.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })
    adminApiMocks.getImages.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })
    adminApiMocks.getLatestChallengePublishRequest.mockResolvedValue(null)
    adminApiMocks.listChallengeImports.mockResolvedValue([])
  })

  it('switches academy and platform routes into a shared backoffice shell while leaving student routes on the default layout', () => {
    expect(appLayoutSource).toContain('isBackofficeRoute(route.path)')
    expect(appLayoutSource).toContain('workspace-main--backoffice')
    expect(appLayoutSource).not.toContain('BackofficeSubNav')
  })

  it('suppresses route-level workspace topbars in backoffice mode so pages do not duplicate the global topnav', () => {
    expect(appLayoutSource).toContain(
      '.workspace-main--backoffice :deep(.workspace-shell > .workspace-topbar)'
    )
    expect(appLayoutSource).toContain('display: none;')
  })

  it('owns full-bleed page spacing and drives it from route meta', () => {
    expect(appLayoutSource).toContain('<RouterView v-slot="{ Component, route: resolvedRoute }">')
    expect(appLayoutSource).toContain('workspace-page')
    expect(appLayoutSource).toContain('workspace-page--bleed')
    expect(appLayoutSource).toContain('pageShellClass')
    expect(appLayoutSource).toContain('workspace-route-root')
    expect(appLayoutSource).toContain('workspace-route-root--bleed')
    expect(routeSources).toContain("contentLayout: 'bleed'")
    expect((routeSources.match(/contentLayout: 'bleed'/g) ?? []).length).toBeGreaterThanOrEqual(30)
  })

  it('wraps workspace routes in the shared page transition without animating query-only changes', () => {
    expect(appLayoutSource).toContain('<Transition')
    expect(appLayoutSource).toContain('name="workspace-route"')
    expect(appLayoutSource).toContain('mode="out-in"')
    expect(appLayoutSource).toContain(':key="resolvedRoute.path"')
  })

  it('stretches full-bleed route roots so wide screens do not expose main shell gaps', () => {
    expect(appLayoutSource).toContain('.workspace-page--bleed :deep(.workspace-route-root--bleed)')
    expect(appLayoutSource).toContain('flex: 1 1 auto;')
  })

  it('removes vertical main padding for full-bleed routes instead of canceling it with negative top margins', () => {
    expect(appLayoutSource).toContain('mainShellClass')
    expect(appLayoutSource).toContain('workspace-main--bleed')
    expect(appLayoutSource).toContain('padding-block: 0;')
    expect(appLayoutSource).toContain('padding-inline: 0;')
    expect(appLayoutSource).toContain('max-width: none;')
  })

  it('makes the topnav content column a flex stack so main can consume the remaining height', () => {
    expect(appLayoutSource).toContain('<div class="min-w-0 flex flex-1 flex-col">')
    expect(appLayoutSource).toContain('.workspace-main {')
    expect(appLayoutSource).toContain('flex: 1 1 auto;')
    expect(appLayoutSource).toContain('min-height: 0;')
  })

  it('loads target teacher pages after sidebar secondary navigation inside the app layout', async () => {
    const toTopLevelRoutes = [...teacherRoutes, ...platformRoutes].map((route) => ({
      ...route,
      path: route.path.startsWith('/') ? route.path : `/${route.path}`,
    }))
    const router = createRouter({
      history: createMemoryHistory(),
      routes: toTopLevelRoutes,
    })
    const pushSpy = vi.spyOn(router, 'push')
    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'teacher-1',
      username: 'teacher',
      role: 'teacher',
      name: 'Teacher',
      class_name: 'Class A',
    })

    await router.push('/academy/classes')
    await router.isReady()

    const wrapper = mount(AppLayout, {
      global: {
        plugins: [router],
        stubs: {
          Transition: false,
        },
      },
    })
    await flushPromises()

    const desktopAside = wrapper.find('.backoffice-sidebar--desktop')

    expect(desktopAside).toBeTruthy()

    expect(teacherApiMocks.getClasses).toHaveBeenCalled()

    const studentsButton = desktopAside!
      .findAll('button')
      .find((node) => node.text().trim() === '学生管理')

    expect(studentsButton).toBeTruthy()

    await studentsButton!.trigger('click')
    await flushPromises()

    expect(pushSpy).toHaveBeenCalledWith({ path: '/academy/students', query: {} })
    await vi.waitFor(() => {
      expect(router.currentRoute.value.path).toBe('/academy/students')
    })
    await vi.waitFor(() => {
      expect(teacherApiMocks.getStudentsDirectory).toHaveBeenCalled()
    })

    const resourcesButton = desktopAside!
      .findAll('button')
      .find((node) => node.text().trim() === '题库与资源')

    expect(resourcesButton).toBeTruthy()

    await resourcesButton!.trigger('click')
    await flushPromises()

    await vi.waitFor(() => {
      expect(router.currentRoute.value.path).toBe('/platform/challenges')
    })
    await vi.waitFor(() => {
      expect(adminApiMocks.getChallenges).toHaveBeenCalled()
    })

    wrapper.unmount()
  })
})
