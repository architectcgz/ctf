import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'

import PlatformClassManagement from '@/pages/platform/ClassManageRoutePage.vue'
import adminClassManageSource from '@/pages/platform/ClassManageRoutePage.vue?raw'
import appRouteLinkSource from '@/shared/ui/navigation/AppRouteLink.vue?raw'
import classManageHeroPanelSource from '@/features/platform/class-management/ui/ClassManageHeroPanel.vue?raw'
import classManageWorkspacePanelSource from '@/features/platform/class-management/ui/ClassManageWorkspacePanel.vue?raw'
import platformClassManagementPageSource from '@/features/platform/class-management/model/usePlatformClassManagementPage.ts?raw'

const adminTeachingApiMocks = vi.hoisted(() => ({
  getClasses: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  getClasses: adminTeachingApiMocks.getClasses,
}))

function createTestRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/platform/classes', component: PlatformClassManagement },
      {
        path: '/platform/classes/:className',
        name: 'PlatformClassStudents',
        component: { template: '<div>class students</div>' },
      },
    ],
  })
}

async function mountPage() {
  const router = createTestRouter()
  await router.push('/platform/classes')
  await router.isReady()

  const wrapper = mount(PlatformClassManagement, {
    global: {
      plugins: [router],
    },
  })

  await flushPromises()
  return { wrapper, router }
}

describe('PlatformClassManagement', () => {
  beforeEach(() => {
    adminTeachingApiMocks.getClasses.mockReset()
    adminTeachingApiMocks.getClasses.mockResolvedValue({
      list: [
        { name: 'Class A', student_count: 2 },
        { name: 'Class B', student_count: 0 },
      ],
      total: 2,
      page: 1,
      page_size: 20,
    })
  })

  it('应使用后台工作台目录组件而不是教师端班级目录壳层', async () => {
    expect(adminClassManageSource).toContain("from '@/features/platform/class-management'")
    expect(adminClassManageSource).toContain('usePlatformClassManagementPage')
    expect(adminClassManageSource).not.toContain("from '@/api/teacher'")
    expect(adminClassManageSource).not.toContain("from '@/api/admin'")
    expect(adminClassManageSource).not.toContain('getAdminClasses')
    expect(adminClassManageSource).toContain('ClassManageHeroPanel')
    expect(adminClassManageSource).toContain('ClassManageWorkspacePanel')
    expect(adminClassManageSource).toContain('<ClassManageHeroPanel')
    expect(adminClassManageSource).toContain('<ClassManageWorkspacePanel')
    expect(adminClassManageSource).toContain(
      'class="workspace-shell journal-shell journal-shell-admin journal-hero admin-class-manage-shell"'
    )
    expect(classManageHeroPanelSource).toContain('刷新目录')
    expect(classManageHeroPanelSource).toContain(
      'class="admin-summary-grid admin-class-manage-shell__summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(classManageWorkspacePanelSource).toContain("from '@/shared/ui/common/WorkspaceDataTable.vue'")
    expect(classManageWorkspacePanelSource).toContain(
      "from '@/shared/ui/common/WorkspaceDirectoryPagination.vue'"
    )
    expect(classManageWorkspacePanelSource).toContain(
      "from '@/shared/ui/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(classManageWorkspacePanelSource).toContain(
      "from '@/shared/ui/navigation/AppRouteLink.vue'"
    )
    expect(classManageWorkspacePanelSource).toContain('<WorkspaceDirectoryToolbar')
    expect(classManageWorkspacePanelSource).toContain('<WorkspaceDataTable')
    expect(classManageWorkspacePanelSource).toContain('<WorkspaceDirectoryPagination')
    expect(classManageWorkspacePanelSource).toContain('<AppRouteLink')
    expect(classManageWorkspacePanelSource).toContain('search-placeholder="搜索班级名称..."')
    expect(classManageWorkspacePanelSource).toContain('filter-panel-title="班级筛选"')
    expect(classManageWorkspacePanelSource).toContain('class="ui-btn ui-btn--primary ui-btn--sm"')
    expect(classManageWorkspacePanelSource).not.toContain('class="ui-btn ui-btn--ghost"')
    expect(appRouteLinkSource).toContain("from 'vue-router'")
    expect(platformClassManagementPageSource).not.toContain("from 'vue-router'")
    expect(platformClassManagementPageSource).toContain('function buildClassRoute')
    expect(adminClassManageSource).not.toContain('teacher-management-shell')
    expect(adminClassManageSource).not.toContain('teacher-directory-row')

    const { wrapper } = await mountPage()

    expect(adminTeachingApiMocks.getClasses).toHaveBeenCalledWith({ page: 1, page_size: 20 })
    expect(wrapper.text()).toContain('班级管理')
    expect(wrapper.text()).toContain('Class A')
    expect(wrapper.text()).toContain('Class B')
    expect(wrapper.text()).toContain('班级名称')
    expect(wrapper.text()).toContain('学生人数')
    expect(wrapper.text()).toContain('查看班级')
    expect(wrapper.text()).toContain('共 2 个班级')
  })

  it('应支持按班级名称筛选目录', async () => {
    const { wrapper } = await mountPage()

    await wrapper.get('.workspace-directory-toolbar__search-input').setValue('Class A')

    expect(wrapper.text()).toContain('Class A')
    expect(wrapper.text()).not.toContain('Class B')
  })

  it('应支持进入班级详情', async () => {
    const { wrapper, router } = await mountPage()
    const classLink = wrapper.findAll('a').find((node) => node.text().includes('查看班级'))

    expect(classLink).toBeTruthy()

    await classLink!.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.name).toBe('PlatformClassStudents')
    expect(router.currentRoute.value.params).toEqual({ className: 'Class A' })
  })
})
