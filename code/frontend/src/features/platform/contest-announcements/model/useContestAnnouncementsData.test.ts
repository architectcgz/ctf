import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, ref } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'

import { useContestAnnouncementsData } from './useContestAnnouncementsData'

const adminApiMocks = vi.hoisted(() => ({
  getContest: vi.fn(),
}))

vi.mock('@/api/admin/contest-manage', async () => {
  const actual =
    await vi.importActual<typeof import('@/api/admin/contest-manage')>('@/api/admin/contest-manage')
  return {
    ...actual,
    getContest: adminApiMocks.getContest,
  }
})

describe('useContestAnnouncementsData', () => {
  beforeEach(() => {
    adminApiMocks.getContest.mockReset()
  })

  it('应在初始化时加载竞赛详情并触发公告列表加载', async () => {
    adminApiMocks.getContest.mockResolvedValue({
      id: 'contest-1',
      title: '2026 春季赛',
      description: '公告运营',
      mode: 'jeopardy',
      status: 'running',
      starts_at: '2026-04-22T08:00:00.000Z',
      ends_at: '2026-04-22T18:00:00.000Z',
      scoreboard_frozen: false,
    })
    const loadAnnouncements = vi.fn().mockResolvedValue([])

    let composable!: ReturnType<typeof useContestAnnouncementsData>
    const Harness = defineComponent({
      setup() {
        composable = useContestAnnouncementsData(ref('contest-1'), loadAnnouncements)
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    expect(adminApiMocks.getContest).toHaveBeenCalledWith('contest-1')
    expect(loadAnnouncements).toHaveBeenCalledTimes(1)
    expect(composable.contest.value?.title).toBe('2026 春季赛')
    expect(composable.loadError.value).toBe('')

    wrapper.unmount()
  })

  it('缺少竞赛编号时应暴露页面错误状态', async () => {
    const loadAnnouncements = vi.fn().mockResolvedValue([])

    let composable!: ReturnType<typeof useContestAnnouncementsData>
    const Harness = defineComponent({
      setup() {
        composable = useContestAnnouncementsData(ref(''), loadAnnouncements)
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    expect(adminApiMocks.getContest).not.toHaveBeenCalled()
    expect(loadAnnouncements).not.toHaveBeenCalled()
    expect(composable.loadError.value).toBe('缺少竞赛编号。')

    wrapper.unmount()
  })

  it('加载失败时应暴露竞赛公告页面错误状态', async () => {
    adminApiMocks.getContest.mockRejectedValue(new Error('boom'))
    const loadAnnouncements = vi.fn().mockResolvedValue([])

    let composable!: ReturnType<typeof useContestAnnouncementsData>
    const Harness = defineComponent({
      setup() {
        composable = useContestAnnouncementsData(ref('contest-1'), loadAnnouncements)
        return () => null
      },
    })

    const wrapper = mount(Harness)
    await flushPromises()

    expect(composable.loadError.value).toBe('boom')
    expect(loadAnnouncements).not.toHaveBeenCalled()

    wrapper.unmount()
  })
})
