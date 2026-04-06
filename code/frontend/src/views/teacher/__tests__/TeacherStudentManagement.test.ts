import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { ElButton, ElTable, ElTableColumn } from 'element-plus'

import TeacherStudentManagement from '../TeacherStudentManagement.vue'
import { useAuthStore } from '@/stores/auth'

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

function deferred<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void
  const promise = new Promise<T>((nextResolve) => {
    resolve = nextResolve
  })
  return { promise, resolve }
}

describe('TeacherStudentManagement', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    pushMock.mockReset()
    teacherApiMocks.getClasses.mockReset()
    teacherApiMocks.getClassStudents.mockReset()

    teacherApiMocks.getClasses.mockResolvedValue([{ name: 'Class A', student_count: 2 }])
    teacherApiMocks.getClassStudents.mockImplementation(async (_className, params) => {
      if (params?.keyword === 'alice') {
        return [{ id: 'stu-1', username: 'alice', name: 'Alice Zhang', recent_event_count: 0 }]
      }
      if (params?.keyword === 'Alice') {
        return [{ id: 'stu-1', username: 'alice', name: 'Alice Zhang', recent_event_count: 0 }]
      }
      return [
        { id: 'stu-1', username: 'alice', name: 'Alice Zhang', recent_event_count: 0 },
        { id: 'stu-2', username: 'bob', recent_event_count: 2 },
      ]
    })

    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'teacher-1',
        username: 'teacher',
        role: 'teacher',
        class_name: 'Class A',
      },
      'token'
    )
  })

  it('应该支持搜索学生并进入学员分析页', async () => {
    const wrapper = mount(TeacherStudentManagement, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('学生管理')
    expect(wrapper.find('.teacher-directory-head').exists()).toBe(true)
    expect(wrapper.findAll('.teacher-directory-row')).toHaveLength(2)
    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).toContain('bob')

    const searchInput = wrapper.find('input[placeholder="搜索姓名或用户名"]')
    await searchInput.setValue('Alice')
    await flushPromises()

    expect(teacherApiMocks.getClassStudents).toHaveBeenLastCalledWith('Class A', {
      keyword: 'Alice',
      student_no: undefined,
    })
    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).not.toContain('bob')

    wrapper.findComponent({ name: 'StudentManagementPage' }).vm.$emit('openStudent', 'stu-1')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'TeacherStudentAnalysis',
      params: { className: 'Class A', studentId: 'stu-1' },
    })
  })

  it('应该忽略过期搜索请求的返回结果', async () => {
    const slowRequest =
      deferred<
        Array<{ id: string; username: string; name?: string; recent_event_count?: number }>
      >()
    const fastRequest =
      deferred<
        Array<{ id: string; username: string; name?: string; recent_event_count?: number }>
      >()

    teacherApiMocks.getClassStudents.mockReset()
    teacherApiMocks.getClassStudents
      .mockResolvedValueOnce([
        { id: 'stu-1', username: 'alice', name: 'Alice Zhang', recent_event_count: 0 },
        { id: 'stu-2', username: 'bob', recent_event_count: 2 },
      ])
      .mockImplementationOnce(() => slowRequest.promise)
      .mockImplementationOnce(() => fastRequest.promise)

    const wrapper = mount(TeacherStudentManagement, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
      },
    })

    await flushPromises()

    const searchInput = wrapper.find('input[placeholder="搜索姓名或用户名"]')
    await searchInput.setValue('A')
    await searchInput.setValue('Ali')

    fastRequest.resolve([
      { id: 'stu-1', username: 'alice', name: 'Alice Zhang', recent_event_count: 0 },
    ])
    await flushPromises()

    slowRequest.resolve([{ id: 'stu-2', username: 'bob', recent_event_count: 2 }])
    await flushPromises()

    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).not.toContain('bob')
  })
})
