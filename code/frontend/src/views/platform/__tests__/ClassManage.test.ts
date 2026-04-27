import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import PlatformClassManagement from '../ClassManage.vue'
import adminClassManageSource from '../ClassManage.vue?raw'
import classManageHeroPanelSource from '@/components/platform/class/ClassManageHeroPanel.vue?raw'
import classManageWorkspacePanelSource from '@/components/platform/class/ClassManageWorkspacePanel.vue?raw'

const pushMock = vi.fn()

const teacherApiMocks = vi.hoisted(() => ({
  getClasses: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
  }
})

vi.mock('@/api/teacher', () => teacherApiMocks)

describe('PlatformClassManagement', () => {
  beforeEach(() => {
    pushMock.mockReset()
    teacherApiMocks.getClasses.mockReset()
    teacherApiMocks.getClasses.mockResolvedValue({
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
    expect(adminClassManageSource).toContain("from '@/api/teacher'")
    expect(adminClassManageSource).not.toContain("from '@/api/admin'")
    expect(adminClassManageSource).not.toContain('getAdminClasses')
    expect(adminClassManageSource).toContain(
      "import ClassManageHeroPanel from '@/components/platform/class/ClassManageHeroPanel.vue'"
    )
    expect(adminClassManageSource).toContain(
      "import ClassManageWorkspacePanel from '@/components/platform/class/ClassManageWorkspacePanel.vue'"
    )
    expect(adminClassManageSource).toContain('<ClassManageHeroPanel')
    expect(adminClassManageSource).toContain('<ClassManageWorkspacePanel')
    expect(adminClassManageSource).toContain(
      'class="workspace-shell journal-shell journal-shell-admin journal-hero admin-class-manage-shell"'
    )
    expect(classManageHeroPanelSource).toContain('刷新目录')
    expect(classManageHeroPanelSource).toContain(
      'class="admin-summary-grid admin-class-manage-shell__summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(classManageWorkspacePanelSource).toContain("from '@/components/common/WorkspaceDataTable.vue'")
    expect(classManageWorkspacePanelSource).toContain(
      "from '@/components/common/WorkspaceDirectoryPagination.vue'"
    )
    expect(classManageWorkspacePanelSource).toContain(
      "from '@/components/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(classManageWorkspacePanelSource).toContain('<WorkspaceDirectoryToolbar')
    expect(classManageWorkspacePanelSource).toContain('<WorkspaceDataTable')
    expect(classManageWorkspacePanelSource).toContain('<WorkspaceDirectoryPagination')
    expect(classManageWorkspacePanelSource).toContain('search-placeholder="搜索班级名称..."')
    expect(classManageWorkspacePanelSource).toContain('filter-panel-title="班级筛选"')
    expect(classManageWorkspacePanelSource).toContain('class="ui-btn ui-btn--primary ui-btn--sm"')
    expect(classManageWorkspacePanelSource).not.toContain('class="ui-btn ui-btn--ghost"')
    expect(adminClassManageSource).not.toContain('teacher-management-shell')
    expect(adminClassManageSource).not.toContain('teacher-directory-row')

    const wrapper = mount(PlatformClassManagement)
    await flushPromises()

    expect(teacherApiMocks.getClasses).toHaveBeenCalledWith({ page: 1, page_size: 20 })
    expect(wrapper.text()).toContain('班级管理')
    expect(wrapper.text()).toContain('Class A')
    expect(wrapper.text()).toContain('Class B')
    expect(wrapper.text()).toContain('班级名称')
    expect(wrapper.text()).toContain('学生人数')
    expect(wrapper.text()).toContain('查看班级')
    expect(wrapper.text()).toContain('共 2 个班级')
  })

  it('应支持按班级名称筛选目录', async () => {
    const wrapper = mount(PlatformClassManagement)
    await flushPromises()

    await wrapper.get('.workspace-directory-toolbar__search-input').setValue('Class A')

    expect(wrapper.text()).toContain('Class A')
    expect(wrapper.text()).not.toContain('Class B')
  })

  it('应支持进入班级详情', async () => {
    const wrapper = mount(PlatformClassManagement)
    await flushPromises()

    await wrapper
      .findAll('button')
      .find((node) => node.text().includes('查看班级'))
      ?.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'PlatformClassStudents',
      params: { className: 'Class A' },
    })
  })
})
