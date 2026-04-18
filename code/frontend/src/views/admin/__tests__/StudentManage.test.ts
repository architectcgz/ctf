import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import AdminStudentManagement from '../StudentManage.vue'
import adminStudentManageSource from '../StudentManage.vue?raw'

const pushMock = vi.fn()

const teacherApiMocks = vi.hoisted(() => ({
  getClasses: vi.fn(),
  getClassStudents: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
  }
})

vi.mock('@/api/teacher', () => teacherApiMocks)

describe('AdminStudentManagement', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    pushMock.mockReset()
    teacherApiMocks.getClasses.mockReset()
    teacherApiMocks.getClassStudents.mockReset()

    teacherApiMocks.getClasses.mockResolvedValue([
      { name: 'Class A', student_count: 2 },
      { name: 'Class B', student_count: 1 },
    ])
    teacherApiMocks.getClassStudents.mockImplementation(async (className, params) => {
      const all = {
        'Class A': [
          {
            id: 'stu-1',
            username: 'alice',
            name: 'Alice Zhang',
            student_no: '2024001',
            solved_count: 5,
            total_score: 320,
            weak_dimension: 'Web',
            class_name: 'Class A',
          },
          {
            id: 'stu-2',
            username: 'bob',
            name: 'Bob Li',
            student_no: '2024002',
            solved_count: 3,
            total_score: 180,
            weak_dimension: 'Pwn',
            class_name: 'Class A',
          },
        ],
        'Class B': [
          {
            id: 'stu-3',
            username: 'charlie',
            name: 'Charlie Wang',
            student_no: '2024011',
            solved_count: 1,
            total_score: 60,
            weak_dimension: 'Crypto',
            class_name: 'Class B',
          },
        ],
      } as const

      const merged = className ? (all[className as keyof typeof all] ?? []) : []
      if (!params?.keyword && !params?.student_no) {
        return merged
      }

      return merged.filter((item) => {
        const keywordMatched =
          !params?.keyword ||
          item.username.includes(params.keyword) ||
          (item.name ?? '').includes(params.keyword)
        const studentNoMatched = !params?.student_no || item.student_no?.includes(params.student_no)
        return keywordMatched && studentNoMatched
      })
    })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('应使用后台工作台目录组件而不是教师端学生目录壳层', async () => {
    expect(adminStudentManageSource).toContain(
      "from '@/components/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(adminStudentManageSource).toContain("from '@/components/common/WorkspaceDataTable.vue'")
    expect(adminStudentManageSource).toContain(
      "from '@/components/common/WorkspaceDirectoryPagination.vue'"
    )
    expect(adminStudentManageSource).toContain('class="workspace-shell admin-student-manage-shell"')
    expect(adminStudentManageSource).toContain('<WorkspaceDirectoryToolbar')
    expect(adminStudentManageSource).toContain('<WorkspaceDataTable')
    expect(adminStudentManageSource).toContain('<WorkspaceDirectoryPagination')
    expect(adminStudentManageSource).toMatch(
      /\.admin-student-manage-directory\s*\{[\s\S]*display:\s*grid;[\s\S]*gap:\s*var\(--space-4\);/s
    )
    expect(adminStudentManageSource).toMatch(
      /\.admin-student-manage-directory :deep\(\.workspace-directory-toolbar\)\s*\{[\s\S]*margin-bottom:\s*0;/s
    )
    expect(adminStudentManageSource).not.toContain('teacher-management-shell')
    expect(adminStudentManageSource).not.toContain('teacher-directory-row')

    const wrapper = mount(AdminStudentManagement)
    await flushPromises()

    expect(wrapper.text()).toContain('学生管理')
    expect(wrapper.text()).toContain('Alice Zhang')
    expect(wrapper.text()).toContain('Bob Li')
    expect(wrapper.text()).toContain('学生名称')
    expect(wrapper.text()).toContain('学号')
    expect(wrapper.text()).toContain('班级')
    expect(wrapper.text()).toContain('得分')
  })

  it('应支持检索并可进入学员分析页', async () => {
    const wrapper = mount(AdminStudentManagement)
    await flushPromises()

    const searchInput = wrapper.get('input[placeholder="检索姓名、用户名或学号..."]')
    await searchInput.setValue('Alice')
    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(wrapper.text()).toContain('Alice Zhang')
    expect(wrapper.text()).not.toContain('Bob Li')

    await wrapper
      .findAll('button')
      .find((node) => node.text().includes('查看学员'))
      ?.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'AdminStudentAnalysis',
      params: { className: 'Class A', studentId: 'stu-1' },
    })
  })
})
