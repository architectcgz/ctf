import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { ElButton, ElTable, ElTableColumn } from 'element-plus'

import TeacherClassStudents from '../TeacherClassStudents.vue'

const pushMock = vi.fn()
const routeMock = {
  params: {
    className: 'Class A',
  },
}

const teacherApiMocks = vi.hoisted(() => ({
  getClasses: vi.fn(),
  getClassStudents: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
    useRoute: () => routeMock,
  }
})

vi.mock('@/api/teacher', () => teacherApiMocks)

function deferred<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void
  const promise = new Promise<T>((nextResolve) => {
    resolve = nextResolve
  })
  return { promise, resolve }
}

describe('TeacherClassStudents', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    pushMock.mockReset()
    routeMock.params.className = 'Class A'
    teacherApiMocks.getClasses.mockReset()
    teacherApiMocks.getClassStudents.mockReset()

    teacherApiMocks.getClasses.mockResolvedValue([{ name: 'Class A', student_count: 2 }])
    teacherApiMocks.getClassStudents.mockResolvedValue([
      { id: 'stu-1', username: 'alice' },
      { id: 'stu-2', username: 'bob' },
    ])
  })

  it('应该展示班级学生列表并支持进入学员分析页', async () => {
    const wrapper = mount(TeacherClassStudents, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('学生列表')
    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).toContain('bob')

    wrapper.findComponent({ name: 'ClassStudentsPage' }).vm.$emit('openStudent', 'stu-1')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'TeacherStudentAnalysis',
      params: { className: 'Class A', studentId: 'stu-1' },
    })
  })

  it('应该保留已解码的班级名并使用原值请求学生列表', async () => {
    routeMock.params.className = '100% 班级'

    mount(TeacherClassStudents, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
      },
    })

    await flushPromises()

    expect(teacherApiMocks.getClassStudents).toHaveBeenCalledWith('100% 班级', {
      student_no: undefined,
    })
  })

  it('应该忽略过期学号搜索请求的返回结果', async () => {
    const slowRequest = deferred<Array<{ id: string; username: string; name?: string }>>()
    const fastRequest = deferred<Array<{ id: string; username: string; name?: string }>>()

    teacherApiMocks.getClassStudents.mockReset()
    teacherApiMocks.getClassStudents
      .mockResolvedValueOnce([
        { id: 'stu-1', username: 'alice', name: 'Alice Zhang' },
        { id: 'stu-2', username: 'bob' },
      ])
      .mockImplementationOnce(() => slowRequest.promise)
      .mockImplementationOnce(() => fastRequest.promise)

    const wrapper = mount(TeacherClassStudents, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
      },
    })

    await flushPromises()

    const studentNoInput = wrapper.find('input[placeholder="输入学号后实时查询"]')
    await studentNoInput.setValue('20260001')
    await studentNoInput.setValue('20260002')

    fastRequest.resolve([{ id: 'stu-1', username: 'alice', name: 'Alice Zhang' }])
    await flushPromises()

    slowRequest.resolve([{ id: 'stu-2', username: 'bob' }])
    await flushPromises()

    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).not.toContain('bob')
  })
})
