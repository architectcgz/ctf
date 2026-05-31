import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

import { usePlatformClassDirectory } from './usePlatformClassDirectory'

const adminApiMocks = vi.hoisted(() => ({
  getClasses: vi.fn(),
}))

vi.mock('@/api/admin', () => adminApiMocks)

describe('usePlatformClassDirectory', () => {
  beforeEach(() => {
    adminApiMocks.getClasses.mockReset()
  })

  it('应在初始化时加载平台班级目录', async () => {
    adminApiMocks.getClasses.mockResolvedValue({
      list: [{ name: 'Class A', student_count: 2 }],
      total: 1,
      page: 1,
      page_size: 20,
    })

    let composable!: ReturnType<typeof usePlatformClassDirectory>
    const Harness = defineComponent({
      setup() {
        composable = usePlatformClassDirectory()
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    expect(adminApiMocks.getClasses).toHaveBeenCalledWith({ page: 1, page_size: 20 })
    expect(composable.list.value).toEqual([{ name: 'Class A', student_count: 2 }])
    expect(composable.total.value).toBe(1)

    wrapper.unmount()
  })

  it('数组响应时应兼容旧的班级目录契约', async () => {
    adminApiMocks.getClasses.mockResolvedValue([{ name: 'Class A', student_count: 2 }])

    let composable!: ReturnType<typeof usePlatformClassDirectory>
    const Harness = defineComponent({
      setup() {
        composable = usePlatformClassDirectory()
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    expect(composable.list.value).toEqual([{ name: 'Class A', student_count: 2 }])
    expect(composable.total.value).toBe(1)
    expect(composable.page.value).toBe(1)

    wrapper.unmount()
  })

  it('翻页后应按目标页码刷新平台班级目录', async () => {
    adminApiMocks.getClasses
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

    let composable!: ReturnType<typeof usePlatformClassDirectory>
    const Harness = defineComponent({
      setup() {
        composable = usePlatformClassDirectory()
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    composable.handlePageChange(2)
    await flushPromises()

    expect(adminApiMocks.getClasses).toHaveBeenNthCalledWith(2, { page: 2, page_size: 20 })
    expect(composable.page.value).toBe(2)
    expect(composable.list.value).toEqual([{ name: 'Class 21', student_count: 21 }])

    wrapper.unmount()
  })

  it('初始化失败时应暴露目录错误状态', async () => {
    adminApiMocks.getClasses.mockRejectedValue(new Error('boom'))

    let composable!: ReturnType<typeof usePlatformClassDirectory>
    const Harness = defineComponent({
      setup() {
        composable = usePlatformClassDirectory()
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    expect(composable.error.value).toBe('加载班级列表失败，请稍后重试')
    expect(composable.list.value).toEqual([])
    expect(composable.total.value).toBe(0)

    wrapper.unmount()
  })
})
