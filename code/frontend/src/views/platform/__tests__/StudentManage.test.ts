import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import PlatformStudentManagement from '../StudentManage.vue'
import adminStudentManageSource from '../StudentManage.vue?raw'
import studentManageHeroPanelSource from '@/components/platform/student/StudentManageHeroPanel.vue?raw'
import studentManageWorkspacePanelSource from '@/components/platform/student/StudentManageWorkspacePanel.vue?raw'

const pushMock = vi.fn()

const teacherApiMocks = vi.hoisted(() => ({
  getClasses: vi.fn(),
  getStudentsDirectory: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
  }
})

vi.mock('@/api/teacher', () => teacherApiMocks)

describe('PlatformStudentManagement', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    pushMock.mockReset()
    teacherApiMocks.getClasses.mockReset()
    teacherApiMocks.getStudentsDirectory.mockReset()

    teacherApiMocks.getClasses.mockResolvedValue([
      { name: 'Class A', student_count: 2 },
      { name: 'Class B', student_count: 1 },
    ])
    teacherApiMocks.getStudentsDirectory.mockImplementation(async (params) => {
      const all = [
        {
          id: 'stu-1',
          username: 'alice',
          name: 'Alice Zhang',
          student_no: '2024001',
          total_score: 320,
          recent_event_count: 5,
          class_name: 'Class A',
        },
        {
          id: 'stu-2',
          username: 'bob',
          name: 'Bob Li',
          student_no: '2024002',
          total_score: 180,
          recent_event_count: 0,
          class_name: 'Class A',
        },
        {
          id: 'stu-3',
          username: 'charlie',
          name: 'Charlie Wang',
          student_no: '2024011',
          total_score: 60,
          recent_event_count: 2,
          class_name: 'Class B',
        },
      ]

      const filtered = all.filter((item) => {
        const classMatched = !params?.class_name || item.class_name === params.class_name
        const keywordMatched =
          !params?.keyword ||
          item.username.includes(params.keyword) ||
          (item.name ?? '').includes(params.keyword) ||
          (item.student_no ?? '').includes(params.keyword)
        return classMatched && keywordMatched
      })

      return {
        list: filtered,
        total: filtered.length,
        page: params?.page ?? 1,
        page_size: params?.page_size ?? 20,
      }
    })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('应复用后台工作台目录组件和 teacher 学生目录接口', async () => {
    expect(adminStudentManageSource).toContain(
      "import StudentManageWorkspacePanel from '@/components/platform/student/StudentManageWorkspacePanel.vue'"
    )
    expect(adminStudentManageSource).toContain("from '@/api/teacher'")
    expect(adminStudentManageSource).toContain("from '@/composables/useStudentDirectoryQuery'")
    expect(adminStudentManageSource).toContain(
      "import StudentManageHeroPanel from '@/components/platform/student/StudentManageHeroPanel.vue'"
    )
    expect(adminStudentManageSource).not.toContain("from '@/composables/usePlatformStudentDirectory'")
    expect(adminStudentManageSource).toContain('<StudentManageHeroPanel')
    expect(adminStudentManageSource).toContain('<StudentManageWorkspacePanel')
    expect(studentManageHeroPanelSource).toContain('刷新目录')
    expect(studentManageHeroPanelSource).toContain(
      'class="admin-summary-grid admin-student-manage-shell__summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(studentManageWorkspacePanelSource).toContain(
      "from '@/components/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(studentManageWorkspacePanelSource).toContain(
      "from '@/components/common/WorkspaceDataTable.vue'"
    )
    expect(studentManageWorkspacePanelSource).toContain(
      "from '@/components/common/WorkspaceDirectoryPagination.vue'"
    )
    expect(studentManageWorkspacePanelSource).toContain('<WorkspaceDirectoryToolbar')
    expect(studentManageWorkspacePanelSource).toContain('<WorkspaceDataTable')
    expect(studentManageWorkspacePanelSource).toContain('<WorkspaceDirectoryPagination')

    const wrapper = mount(PlatformStudentManagement)
    await flushPromises()

    expect(teacherApiMocks.getClasses).toHaveBeenCalledTimes(1)
    expect(teacherApiMocks.getStudentsDirectory).toHaveBeenCalledWith({
      class_name: undefined,
      keyword: undefined,
      student_no: undefined,
      sort_key: 'name',
      sort_order: 'asc',
      page: 1,
      page_size: 20,
    })
    expect(wrapper.text()).toContain('学生管理')
    expect(wrapper.text()).toContain('Alice Zhang')
    expect(wrapper.text()).toContain('Bob Li')
    expect(wrapper.text()).toContain('Charlie Wang')
    expect(wrapper.text()).toContain('学生姓名')
    expect(wrapper.text()).toContain('用户名')
    expect(wrapper.text()).toContain('学号')
    expect(wrapper.text()).toContain('班级')
    expect(wrapper.text()).toContain('查看学员')
  })

  it('应支持检索、班级筛选和进入学员分析页', async () => {
    const wrapper = mount(PlatformStudentManagement, {
      attachTo: document.body,
    })
    await flushPromises()

    const searchInput = wrapper.get('input[placeholder="检索姓名、用户名或学号..."]')
    await searchInput.setValue('Alice')
    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(wrapper.text()).toContain('Alice Zhang')
    expect(wrapper.text()).not.toContain('Bob Li')
    expect(teacherApiMocks.getStudentsDirectory).toHaveBeenLastCalledWith({
      class_name: undefined,
      keyword: 'Alice',
      student_no: undefined,
      sort_key: 'name',
      sort_order: 'asc',
      page: 1,
      page_size: 20,
    })

    await searchInput.setValue('')
    vi.advanceTimersByTime(250)
    await flushPromises()

    await wrapper.get('.workspace-directory-toolbar__filter-toggle').trigger('click')
    const classSelect = wrapper.get('select')
    await classSelect.setValue('Class B')
    await flushPromises()

    expect(wrapper.text()).toContain('Charlie Wang')
    expect(wrapper.text()).not.toContain('Alice Zhang')
    expect(teacherApiMocks.getStudentsDirectory).toHaveBeenLastCalledWith({
      class_name: 'Class B',
      keyword: undefined,
      student_no: undefined,
      sort_key: 'name',
      sort_order: 'asc',
      page: 1,
      page_size: 20,
    })

    await wrapper.get('.workspace-directory-toolbar__filter-reset').trigger('click')
    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(teacherApiMocks.getStudentsDirectory).toHaveBeenLastCalledWith({
      class_name: undefined,
      keyword: undefined,
      student_no: undefined,
      sort_key: 'name',
      sort_order: 'asc',
      page: 1,
      page_size: 20,
    })

    await wrapper
      .findAll('button')
      .find((node) => node.text().includes('查看学员'))
      ?.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'PlatformStudentAnalysis',
      params: { className: 'Class A', studentId: 'stu-1' },
    })

    wrapper.unmount()
  })
})
