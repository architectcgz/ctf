import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import PlatformClassStudents from '../PlatformClassStudents.vue'
import platformClassStudentsSource from '../PlatformClassStudents.vue?raw'
import classStudentsPageSource from '@/components/teacher/class-management/ClassStudentsPage.vue?raw'
import { useAuthStore } from '@/stores/auth'

const ElTable = { template: '<div><slot /></div>' }
const ElTableColumn = { template: '<div><slot /></div>' }
const ElButton = { template: '<button><slot /></button>' }

const pushMock = vi.fn()
const replaceMock = vi.fn()
const routeMock = {
  params: {
    className: 'Class A',
  },
  query: {
    panel: 'students',
    from_date: undefined as string | undefined,
    to_date: undefined as string | undefined,
  },
}

const teachingApiMocks = vi.hoisted(() => ({
  getClassStudents: vi.fn(),
  getClassReview: vi.fn(),
  getClassSummary: vi.fn(),
  getClassTrend: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock, replace: replaceMock }),
    useRoute: () => routeMock,
  }
})

vi.mock('@/api/teaching', () => teachingApiMocks)

describe('PlatformClassStudents', () => {
  const reportDialogStub = {
    name: 'TeacherClassReportExportDialog',
    props: ['modelValue', 'defaultClassName', 'defaultFromDate', 'defaultToDate'],
    template:
      '<div data-testid="class-report-dialog" :data-open="String(modelValue)" :data-default-class-name="defaultClassName || \'\'" :data-default-from-date="defaultFromDate || \'\'" :data-default-to-date="defaultToDate || \'\'" />',
  }

  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    pushMock.mockReset()
    replaceMock.mockReset()
    routeMock.params.className = 'Class A'
    routeMock.query.panel = 'students'
    delete routeMock.query.from_date
    delete routeMock.query.to_date
    teachingApiMocks.getClassStudents.mockReset()
    teachingApiMocks.getClassReview.mockReset()
    teachingApiMocks.getClassSummary.mockReset()
    teachingApiMocks.getClassTrend.mockReset()

    teachingApiMocks.getClassStudents.mockResolvedValue([
      {
        id: 'stu-1',
        username: 'alice',
        name: 'Alice Zhang',
        solved_count: 3,
        total_score: 280,
        recent_event_count: 0,
        weak_dimension: 'crypto',
      },
      {
        id: 'stu-2',
        username: 'bob',
        solved_count: 1,
        total_score: 100,
        recent_event_count: 2,
        weak_dimension: 'pwn',
      },
    ])
    teachingApiMocks.getClassReview.mockResolvedValue({
      class_name: 'Class A',
      items: [],
    })
    teachingApiMocks.getClassSummary.mockResolvedValue({
      class_name: 'Class A',
      student_count: 2,
      average_solved: 2,
      active_student_count: 1,
      active_rate: 50,
      recent_event_count: 6,
    })
    teachingApiMocks.getClassTrend.mockResolvedValue({
      class_name: 'Class A',
      points: [
        { date: '2026-03-05', active_student_count: 1, event_count: 2, solve_count: 1 },
        { date: '2026-03-06', active_student_count: 1, event_count: 3, solve_count: 1 },
      ],
    })

    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'admin-1',
      username: 'admin',
      role: 'admin',
      class_name: '',
    })
  })

  it('应通过平台 route view 复用中性班级工作台 feature', async () => {
    expect(platformClassStudentsSource).toContain(
      "import { useClassStudentsPage } from '@/features/class-students-workspace'"
    )
    expect(platformClassStudentsSource).not.toContain("from '@/views/teacher/TeacherClassStudents.vue'")
    expect(platformClassStudentsSource).not.toContain("from '@/api/teacher'")
    expect(classStudentsPageSource).toContain('学生列表')

    const wrapper = mount(PlatformClassStudents, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
        stubs: {
          LineChart: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()
    await flushPromises()

    expect(wrapper.text()).toContain('学生列表')
    expect(wrapper.text()).toContain('Alice Zhang')
    expect(wrapper.text()).toContain('bob')

    await wrapper
      .findAll('button')
      .find((node) => node.attributes('aria-label')?.includes('查看学员分析'))
      ?.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'PlatformStudentAnalysis',
      params: {
        className: 'Class A',
        studentId: 'stu-1',
      },
    })
  })
})
