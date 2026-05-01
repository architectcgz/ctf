import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, ref } from 'vue'

import { useContestAnnouncementManagement } from '@/composables/useContestAnnouncementManagement'
import type { ContestDetailData } from '@/api/contracts'

const adminApiMocks = vi.hoisted(() => ({
  getAdminContestAnnouncements: vi.fn(),
  createAdminContestAnnouncement: vi.fn(),
  deleteAdminContestAnnouncement: vi.fn(),
}))

const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
}))

vi.mock('@/api/admin/contests', () => adminApiMocks)
vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))

function buildContest(overrides: Partial<ContestDetailData> = {}): ContestDetailData {
  return {
    id: 'contest-1',
    title: '2026 春季赛',
    description: '公告运营',
    mode: 'jeopardy',
    status: 'running',
    starts_at: '2026-04-22T08:00:00.000Z',
    ends_at: '2026-04-22T18:00:00.000Z',
    scoreboard_frozen: false,
    ...overrides,
  }
}

function withSetup(contest = ref<ContestDetailData | null>(buildContest())) {
  let result!: ReturnType<typeof useContestAnnouncementManagement>

  const wrapper = mount(
    defineComponent({
      setup() {
        result = useContestAnnouncementManagement(contest)
        return () => null
      },
    })
  )

  return { result, wrapper, contest }
}

describe('useContestAnnouncementManagement', () => {
  afterEach(() => {
    vi.useRealTimers()
  })

  beforeEach(() => {
    adminApiMocks.getAdminContestAnnouncements.mockReset()
    adminApiMocks.createAdminContestAnnouncement.mockReset()
    adminApiMocks.deleteAdminContestAnnouncement.mockReset()
    toastMocks.success.mockReset()
    toastMocks.error.mockReset()
    toastMocks.warning.mockReset()

    adminApiMocks.getAdminContestAnnouncements.mockResolvedValue([
      {
        id: 'announcement-1',
        title: '报名提醒',
        content: '请在今晚前完成组队。',
        created_at: '2026-04-22T09:00:00.000Z',
      },
    ])
    adminApiMocks.createAdminContestAnnouncement.mockResolvedValue({
      id: 'announcement-2',
      title: '开赛通知',
      content: '比赛将于十分钟后开始。',
      created_at: '2026-04-22T09:10:00.000Z',
    })
    adminApiMocks.deleteAdminContestAnnouncement.mockResolvedValue(undefined)
  })

  it('加载公告列表成功后应写入 announcements', async () => {
    const { result, wrapper } = withSetup()

    await result.loadAnnouncements()
    await flushPromises()

    expect(adminApiMocks.getAdminContestAnnouncements).toHaveBeenCalledWith('contest-1')
    expect(result.announcements.value).toEqual([
      expect.objectContaining({
        id: 'announcement-1',
        title: '报名提醒',
      }),
    ])
    expect(result.loadError.value).toBe('')
    expect(result.loading.value).toBe(false)

    wrapper.unmount()
  })

  it('发布成功后应清空表单并刷新列表', async () => {
    adminApiMocks.getAdminContestAnnouncements
      .mockResolvedValueOnce([
        {
          id: 'announcement-1',
          title: '报名提醒',
          content: '请在今晚前完成组队。',
          created_at: '2026-04-22T09:00:00.000Z',
        },
      ])
      .mockResolvedValueOnce([
        {
          id: 'announcement-1',
          title: '报名提醒',
          content: '请在今晚前完成组队。',
          created_at: '2026-04-22T09:00:00.000Z',
        },
        {
          id: 'announcement-2',
          title: '开赛通知',
          content: '比赛将于十分钟后开始。',
          created_at: '2026-04-22T09:10:00.000Z',
        },
      ])

    const { result, wrapper } = withSetup()
    await result.loadAnnouncements()

    result.form.title = '  开赛通知  '
    result.form.content = '  比赛将于十分钟后开始。  '

    await expect(result.publishAnnouncement()).resolves.toEqual(
      expect.objectContaining({ id: 'announcement-2' })
    )
    await flushPromises()

    expect(adminApiMocks.createAdminContestAnnouncement).toHaveBeenCalledWith('contest-1', {
      title: '开赛通知',
      content: '比赛将于十分钟后开始。',
    })
    expect(result.form.title).toBe('')
    expect(result.form.content).toBe('')
    expect(result.announcements.value).toHaveLength(2)
    expect(toastMocks.success).toHaveBeenCalledWith('公告已发布')

    wrapper.unmount()
  })

  it('删除成功后应刷新列表', async () => {
    adminApiMocks.getAdminContestAnnouncements
      .mockResolvedValueOnce([
        {
          id: 'announcement-1',
          title: '报名提醒',
          content: '请在今晚前完成组队。',
          created_at: '2026-04-22T09:00:00.000Z',
        },
        {
          id: 'announcement-2',
          title: '开赛通知',
          content: '比赛将于十分钟后开始。',
          created_at: '2026-04-22T09:10:00.000Z',
        },
      ])
      .mockResolvedValueOnce([
        {
          id: 'announcement-2',
          title: '开赛通知',
          content: '比赛将于十分钟后开始。',
          created_at: '2026-04-22T09:10:00.000Z',
        },
      ])

    const { result, wrapper } = withSetup()
    await result.loadAnnouncements()

    await expect(result.deleteAnnouncement('announcement-1')).resolves.toBe(true)
    await flushPromises()

    expect(adminApiMocks.deleteAdminContestAnnouncement).toHaveBeenCalledWith(
      'contest-1',
      'announcement-1'
    )
    expect(result.announcements.value).toEqual([expect.objectContaining({ id: 'announcement-2' })])
    expect(toastMocks.success).toHaveBeenCalledWith('公告已删除')

    wrapper.unmount()
  })

  it('ended 状态下不应允许管理公告', () => {
    const { result, wrapper } = withSetup(ref(buildContest({ status: 'ended' })))

    expect(result.canManageAnnouncements.value).toBe(false)

    wrapper.unmount()
  })

  it('发布和删除失败时应在本地消费异常', async () => {
    adminApiMocks.createAdminContestAnnouncement.mockRejectedValueOnce(new Error('publish failed'))
    adminApiMocks.deleteAdminContestAnnouncement.mockRejectedValueOnce(new Error('delete failed'))

    const { result, wrapper } = withSetup()
    await result.loadAnnouncements()

    result.form.title = '开赛通知'
    result.form.content = '比赛将于十分钟后开始。'

    await expect(result.publishAnnouncement()).resolves.toBeNull()
    await expect(result.deleteAnnouncement('announcement-1')).resolves.toBe(false)
    await flushPromises()

    expect(toastMocks.error).toHaveBeenCalledWith('publish failed')
    expect(toastMocks.error).toHaveBeenCalledWith('delete failed')

    wrapper.unmount()
  })
})
