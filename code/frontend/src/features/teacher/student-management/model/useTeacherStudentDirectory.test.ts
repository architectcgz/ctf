import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

import { useAuthStore } from '@/stores/auth'

import { useTeacherStudentDirectory } from './useTeacherStudentDirectory'

const teacherApiMocks = vi.hoisted(() => ({
  getClasses: vi.fn(),
  getStudentsDirectory: vi.fn(),
}))

vi.mock('@/api/teacher', () => teacherApiMocks)

describe('useTeacherStudentDirectory', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    setActivePinia(createPinia())
    teacherApiMocks.getClasses.mockReset()
    teacherApiMocks.getStudentsDirectory.mockReset()
    useAuthStore().setAuth({
      id: 'teacher-1',
      username: 'teacher',
      role: 'teacher',
      class_name: 'Class A',
    })
  })

  it('应在初始化时按教师默认班级加载学生目录', async () => {
    teacherApiMocks.getClasses.mockResolvedValue([{ name: 'Class A', student_count: 2 }])
    teacherApiMocks.getStudentsDirectory.mockResolvedValue({
      list: [{ id: 'stu-1', username: 'alice', class_name: 'Class A' }],
      total: 1,
      page: 1,
      page_size: 20,
    })

    let composable!: ReturnType<typeof useTeacherStudentDirectory>
    const Harness = defineComponent({
      setup() {
        composable = useTeacherStudentDirectory()
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    expect(teacherApiMocks.getClasses).toHaveBeenCalledTimes(1)
    expect(teacherApiMocks.getStudentsDirectory).toHaveBeenCalledWith({
      class_name: 'Class A',
      keyword: undefined,
      student_no: undefined,
      sort_key: 'solved_count',
      sort_order: 'desc',
      page: 1,
      page_size: 20,
    })
    expect(composable.selectedClassName.value).toBe('Class A')
    expect(composable.totalStudents.value).toBe(2)

    wrapper.unmount()
  })

  it('搜索值看起来像学号时应改走 student_no 查询', async () => {
    teacherApiMocks.getClasses.mockResolvedValue([{ name: 'Class A', student_count: 2 }])
    teacherApiMocks.getStudentsDirectory.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })

    let composable!: ReturnType<typeof useTeacherStudentDirectory>
    const Harness = defineComponent({
      setup() {
        composable = useTeacherStudentDirectory()
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    composable.updateSearchQuery('2024001')
    await flushPromises()
    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(teacherApiMocks.getStudentsDirectory).toHaveBeenLastCalledWith({
      class_name: 'Class A',
      keyword: undefined,
      student_no: '2024001',
      sort_key: 'solved_count',
      sort_order: 'desc',
      page: 1,
      page_size: 20,
    })

    wrapper.unmount()
  })

  it('初始化失败时应暴露页面错误状态', async () => {
    teacherApiMocks.getClasses.mockRejectedValue(new Error('boom'))

    let composable!: ReturnType<typeof useTeacherStudentDirectory>
    const Harness = defineComponent({
      setup() {
        composable = useTeacherStudentDirectory()
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    expect(composable.error.value).toBe('加载学生管理失败，请稍后重试')

    wrapper.unmount()
  })
})
