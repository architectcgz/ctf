import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'

import PlatformStudentManagement from '@/pages/platform/StudentManageRoutePage.vue'
import adminStudentManageSource from '@/pages/platform/StudentManageRoutePage.vue?raw'
import appRouteLinkSource from '@/shared/ui/navigation/AppRouteLink.vue?raw'
import studentManageHeroPanelSource from '@/features/platform/student-management/ui/StudentManageHeroPanel.vue?raw'
import studentManageWorkspacePanelSource from '@/features/platform/student-management/ui/StudentManageWorkspacePanel.vue?raw'
import platformStudentManagementPageSource from '@/features/platform/student-management/model/usePlatformStudentManagementPage.ts?raw'
import platformStudentDirectorySource from '@/features/platform/student-management/model/usePlatformStudentDirectory.ts?raw'

const adminTeachingApiMocks = vi.hoisted(() => ({
  getClasses: vi.fn(),
  getStudentsDirectory: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  getClasses: adminTeachingApiMocks.getClasses,
  getStudentsDirectory: adminTeachingApiMocks.getStudentsDirectory,
}))

function createTestRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/platform/students', component: PlatformStudentManagement },
      {
        path: '/platform/students/:className/:studentId',
        name: 'PlatformStudentAnalysis',
        component: { template: '<div>student analysis</div>' },
      },
    ],
  })
}

async function mountPage() {
  const router = createTestRouter()
  await router.push('/platform/students')
  await router.isReady()

  const wrapper = mount(PlatformStudentManagement, {
    global: {
      plugins: [router],
    },
  })

  await flushPromises()
  return { wrapper, router }
}

describe('PlatformStudentManagement', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    adminTeachingApiMocks.getClasses.mockReset()
    adminTeachingApiMocks.getStudentsDirectory.mockReset()

    adminTeachingApiMocks.getClasses.mockResolvedValue([
      { name: 'Class A', student_count: 2 },
      { name: 'Class B', student_count: 1 },
    ])
    adminTeachingApiMocks.getStudentsDirectory.mockImplementation(async (params) => {
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

  it('应复用后台工作台目录组件和 admin 教学目录接口 owner', async () => {
    expect(adminStudentManageSource).toContain("from '@/features/platform/student-management'")
    expect(adminStudentManageSource).toContain('usePlatformStudentManagementPage')
    expect(adminStudentManageSource).not.toContain("from '@/api/teacher'")
    expect(adminStudentManageSource).not.toContain("from '@/features/student-directory'")
    expect(adminStudentManageSource).not.toContain("from '@/composables/usePlatformStudentDirectory'")
    expect(studentManageWorkspacePanelSource).toContain(
      "from '@/shared/ui/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(studentManageWorkspacePanelSource).toContain(
      "from '@/shared/ui/common/WorkspaceDataTable.vue'"
    )
    expect(studentManageWorkspacePanelSource).toContain(
      "from '@/shared/ui/common/WorkspaceDirectoryPagination.vue'"
    )
    expect(platformStudentManagementPageSource).not.toContain("from 'vue-router'")
    expect(platformStudentManagementPageSource).toContain("from './usePlatformStudentDirectory'")
    expect(platformStudentManagementPageSource).not.toContain("from '@/api/admin'")
    expect(platformStudentDirectorySource).toContain("from '@/api/admin'")
    expect(platformStudentManagementPageSource).toContain('function buildStudentRoute')

    const { wrapper } = await mountPage()

    expect(adminTeachingApiMocks.getClasses).toHaveBeenCalledTimes(1)
    expect(adminTeachingApiMocks.getStudentsDirectory).toHaveBeenCalledWith({
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
    const { wrapper, router } = await mountPage()

    const searchInput = wrapper.get('input[placeholder="检索姓名、用户名或学号..."]')
    await searchInput.setValue('Alice')
    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(wrapper.text()).toContain('Alice Zhang')
    expect(wrapper.text()).not.toContain('Bob Li')
    expect(adminTeachingApiMocks.getStudentsDirectory).toHaveBeenLastCalledWith({
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
    expect(adminTeachingApiMocks.getStudentsDirectory).toHaveBeenLastCalledWith({
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

    expect(adminTeachingApiMocks.getStudentsDirectory).toHaveBeenLastCalledWith({
      class_name: undefined,
      keyword: undefined,
      student_no: undefined,
      sort_key: 'name',
      sort_order: 'asc',
      page: 1,
      page_size: 20,
    })

    const studentLink = wrapper.findAll('a').find((node) => node.text().includes('查看学员'))
    expect(studentLink).toBeTruthy()

    await studentLink!.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.name).toBe('PlatformStudentAnalysis')
    expect(router.currentRoute.value.params).toEqual({ className: 'Class A', studentId: 'stu-1' })
  })
})
