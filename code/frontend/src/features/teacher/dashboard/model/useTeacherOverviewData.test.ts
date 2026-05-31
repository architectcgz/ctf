import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

import { useTeacherOverviewData } from './useTeacherOverviewData'

const teacherApiMocks = vi.hoisted(() => ({
  getTeacherOverview: vi.fn(),
}))

vi.mock('@/api/teacher', () => teacherApiMocks)

describe('useTeacherOverviewData', () => {
  beforeEach(() => {
    teacherApiMocks.getTeacherOverview.mockReset()
  })

  it('应在初始化时加载教师概览数据', async () => {
    teacherApiMocks.getTeacherOverview.mockResolvedValue({
      summary: {
        class_count: 2,
        student_count: 5,
        active_student_count: 3,
        active_rate: 60,
        average_solved: 3.4,
        recent_event_count: 12,
        risk_student_count: 1,
      },
      trend: { points: [] },
      focus_classes: [],
      focus_students: [],
      spotlight_student: null,
      weak_dimensions: [],
    })

    let composable!: ReturnType<typeof useTeacherOverviewData>
    const Harness = defineComponent({
      setup() {
        composable = useTeacherOverviewData()
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    expect(teacherApiMocks.getTeacherOverview).toHaveBeenCalledTimes(1)
    expect(composable.overview.value?.summary.class_count).toBe(2)
    expect(composable.error.value).toBeNull()

    wrapper.unmount()
  })

  it('初始化失败时应暴露教师概览错误状态', async () => {
    teacherApiMocks.getTeacherOverview.mockRejectedValue(new Error('boom'))

    let composable!: ReturnType<typeof useTeacherOverviewData>
    const Harness = defineComponent({
      setup() {
        composable = useTeacherOverviewData()
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    expect(composable.error.value).toBe('加载教师概览失败，请稍后重试')
    expect(composable.overview.value).toBeNull()

    wrapper.unmount()
  })
})
