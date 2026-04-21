import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import PlatformClassManagement from '../ClassManage.vue'
import adminClassManageSource from '../ClassManage.vue?raw'

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
    expect(adminClassManageSource).toContain("from '@/components/common/WorkspaceDataTable.vue'")
    expect(adminClassManageSource).toContain(
      "from '@/components/common/WorkspaceDirectoryPagination.vue'"
    )
    expect(adminClassManageSource).toContain("from '@/api/teacher'")
    expect(adminClassManageSource).not.toContain("from '@/api/admin'")
    expect(adminClassManageSource).not.toContain('getAdminClasses')
    expect(adminClassManageSource).toContain("from '@/components/common/WorkspaceDataTable.vue'")
    expect(adminClassManageSource).toContain('<WorkspaceDataTable')
    expect(adminClassManageSource).toContain('<WorkspaceDirectoryPagination')
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
