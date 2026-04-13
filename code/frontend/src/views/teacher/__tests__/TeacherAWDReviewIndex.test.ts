import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import TeacherAWDReviewIndex from '../TeacherAWDReviewIndex.vue'
import teacherAwdReviewIndexSource from '../TeacherAWDReviewIndex.vue?raw'

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

  it('筛选区应保持平铺，不应继续在页面局部做成独立卡片壳', () => {
    expect(teacherAwdReviewIndexSource).not.toMatch(
      /\.teacher-controls\s*\{[\s\S]*border:\s*1px solid var\(--teacher-card-border\);/s
    )
    expect(teacherAwdReviewIndexSource).not.toMatch(
      /\.teacher-controls\s*\{[\s\S]*background:\s*color-mix\(in srgb,\s*var\(--journal-surface-subtle\)\s*84%,\s*transparent\);/s
    )
    expect(teacherAwdReviewIndexSource).not.toMatch(
      /\.teacher-controls\s*\{[\s\S]*box-shadow:\s*0 10px 24px var\(--color-shadow-soft\);/s
    )
  })
})
