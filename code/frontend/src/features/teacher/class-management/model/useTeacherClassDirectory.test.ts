import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

import { useTeacherClassDirectory } from './useTeacherClassDirectory'

const teacherApiMocks = vi.hoisted(() => ({
  getClasses: vi.fn(),
}))

vi.mock('@/api/teacher', () => teacherApiMocks)

describe('useTeacherClassDirectory', () => {
  beforeEach(() => {
    teacherApiMocks.getClasses.mockReset()
  })

  it('应在初始化时加载教师班级目录', async () => {
    teacherApiMocks.getClasses.mockResolvedValue({
      list: [{ name: 'Class A', student_count: 2 }],
      total: 1,
      page: 1,
      page_size: 20,
    })

    let composable!: ReturnType<typeof useTeacherClassDirectory>
    const Harness = defineComponent({
      setup() {
        composable = useTeacherClassDirectory()
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    expect(teacherApiMocks.getClasses).toHaveBeenCalledWith({ page: 1, page_size: 20 })
    expect(composable.classes.value).toEqual([{ name: 'Class A', student_count: 2 }])
    expect(composable.total.value).toBe(1)

    wrapper.unmount()
  })

  it('翻页后应按目标页码刷新班级目录', async () => {
    teacherApiMocks.getClasses
      .mockResolvedValueOnce({
        list: Array.from({ length: 20 }, (_, index) => ({
          name: `Class ${index + 1}`,
          student_count: index + 1,
        })),
        total: 21,
        page: 1,
        page_size: 20,
      })
      .mockResolvedValueOnce({
        list: [{ name: 'Class 21', student_count: 21 }],
        total: 21,
        page: 2,
        page_size: 20,
      })

    let composable!: ReturnType<typeof useTeacherClassDirectory>
    const Harness = defineComponent({
      setup() {
        composable = useTeacherClassDirectory()
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    composable.handlePageChange(2)
    await flushPromises()

    expect(teacherApiMocks.getClasses).toHaveBeenNthCalledWith(2, { page: 2, page_size: 20 })
    expect(composable.page.value).toBe(2)
    expect(composable.classes.value).toEqual([{ name: 'Class 21', student_count: 21 }])

    wrapper.unmount()
  })

  it('初始化失败时应暴露目录错误状态', async () => {
    teacherApiMocks.getClasses.mockRejectedValue(new Error('boom'))

    let composable!: ReturnType<typeof useTeacherClassDirectory>
    const Harness = defineComponent({
      setup() {
        composable = useTeacherClassDirectory()
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    expect(composable.error.value).toBe('加载班级列表失败，请稍后重试')
    expect(composable.classes.value).toEqual([])
    expect(composable.total.value).toBe(0)

    wrapper.unmount()
  })
})
