import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import TeacherAWDReviewIndex from '../TeacherAWDReviewIndex.vue'

const pushMock = vi.fn()

const teacherApiMocks = vi.hoisted(() => ({
  listTeacherAWDReviews: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
  }
})

vi.mock('@/api/teacher', () => teacherApiMocks)

describe('TeacherAWDReviewIndex', () => {
  beforeEach(() => {
    pushMock.mockReset()
    Object.values(teacherApiMocks).forEach((mock) => mock.mockReset())

    teacherApiMocks.listTeacherAWDReviews.mockResolvedValue([
      {
        id: 'contest-1',
        title: '春季 AWD 联训',
        mode: 'awd',
        status: 'running',
        current_round: 2,
        round_count: 6,
        team_count: 8,
        export_ready: false,
      },
      {
        id: 'contest-2',
        title: '期末 AWD 复盘',
        mode: 'awd',
        status: 'ended',
        current_round: 8,
        round_count: 8,
        team_count: 10,
        export_ready: true,
      },
    ])
  })

  it('应加载 AWD 赛事目录并渲染进入复盘入口', async () => {
    const wrapper = mount(TeacherAWDReviewIndex)

    await flushPromises()

    expect(teacherApiMocks.listTeacherAWDReviews).toHaveBeenCalled()
    expect(wrapper.text()).toContain('AWD复盘')
    expect(wrapper.text()).toContain('春季 AWD 联训')
    expect(wrapper.text()).toContain('进入复盘')
  })
})
