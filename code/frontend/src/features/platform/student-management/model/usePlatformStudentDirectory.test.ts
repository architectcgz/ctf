import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

import { usePlatformStudentDirectory } from './usePlatformStudentDirectory'

const adminApiMocks = vi.hoisted(() => ({
  getClasses: vi.fn(),
  getStudentsDirectory: vi.fn(),
}))

vi.mock('@/api/admin', () => adminApiMocks)

describe('usePlatformStudentDirectory', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    adminApiMocks.getClasses.mockReset()
    adminApiMocks.getStudentsDirectory.mockReset()
  })

  it('应在初始化时加载平台班级与学生目录', async () => {
    adminApiMocks.getClasses.mockResolvedValue([{ name: 'Class A', student_count: 2 }])
    adminApiMocks.getStudentsDirectory.mockResolvedValue({
      list: [{ id: 'stu-1', username: 'alice', class_name: 'Class A', recent_event_count: 1 }],
      total: 1,
      page: 1,
      page_size: 20,
    })

    let composable!: ReturnType<typeof usePlatformStudentDirectory>
    const Harness = defineComponent({
      setup() {
        composable = usePlatformStudentDirectory()
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    expect(adminApiMocks.getClasses).toHaveBeenCalledTimes(1)
    expect(adminApiMocks.getStudentsDirectory).toHaveBeenCalledWith({
      class_name: undefined,
      keyword: undefined,
      student_no: undefined,
      sort_key: 'name',
      sort_order: 'asc',
      page: 1,
      page_size: 20,
    })
    expect(composable.assignedClassCount.value).toBe(1)
    expect(composable.activeStudents.value).toBe(1)

    wrapper.unmount()
  })

  it('关键词变更后应 debounce 刷新学生目录', async () => {
    adminApiMocks.getClasses.mockResolvedValue([])
    adminApiMocks.getStudentsDirectory.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })

    let composable!: ReturnType<typeof usePlatformStudentDirectory>
    const Harness = defineComponent({
      setup() {
        composable = usePlatformStudentDirectory()
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    composable.handleKeywordChange('Alice')
    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(adminApiMocks.getStudentsDirectory).toHaveBeenLastCalledWith({
      class_name: undefined,
      keyword: 'Alice',
      student_no: undefined,
      sort_key: 'name',
      sort_order: 'asc',
      page: 1,
      page_size: 20,
    })

    wrapper.unmount()
  })

  it('班级切换后应立即按所选班级刷新目录', async () => {
    adminApiMocks.getClasses.mockResolvedValue([
      { name: 'Class A', student_count: 2 },
      { name: 'Class B', student_count: 1 },
    ])
    adminApiMocks.getStudentsDirectory.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })

    let composable!: ReturnType<typeof usePlatformStudentDirectory>
    const Harness = defineComponent({
      setup() {
        composable = usePlatformStudentDirectory()
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    await composable.handleClassFilterChange('Class B')
    await flushPromises()

    expect(adminApiMocks.getStudentsDirectory).toHaveBeenLastCalledWith({
      class_name: 'Class B',
      keyword: undefined,
      student_no: undefined,
      sort_key: 'name',
      sort_order: 'asc',
      page: 1,
      page_size: 20,
    })

    wrapper.unmount()
  })
})
