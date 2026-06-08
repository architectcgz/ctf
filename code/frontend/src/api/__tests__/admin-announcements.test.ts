import { beforeEach, describe, expect, it, vi } from 'vitest'
import {
  createAdminContestAnnouncement,
  createAdminContestChallenge,
  createContest,
  createContestAWDService,
  createContestAWDRound,
  deleteAdminContestAnnouncement,
  deleteAdminContestChallenge,
  deleteContestAWDService,
  getAdminContestAnnouncements,
  getAdminContestLiveScoreboard,
  getContestAWDReadiness,
  getContestAWDRoundSummary,
  getContestAWDRoundTrafficSummary,
  getContests,
  listAdminContestChallenges,
  listContestAWDServices,
  listContestAWDRoundAttacks,
  listContestAWDRoundTrafficEvents,
  prewarmContestAWDInstances,
  runContestAWDCheckerPreview,
  runContestAWDCurrentRoundCheck,
  runContestAWDRoundCheck,
  startContestAWDTeamServiceInstance,
  updateAdminContestChallenge,
  updateContest,
  updateContestAWDService,
} from '@/api/admin/contests'

const requestMock = vi.hoisted(() => vi.fn())

vi.mock('@/api/request', () => ({
  request: requestMock,
  ApiError: class ApiError extends Error {
    status?: number

    constructor(message: string, opts?: { status?: number }) {
      super(message)
      this.name = 'ApiError'
      this.status = opts?.status
    }
  },
}))

describe('admin contest api contract', () => {
  beforeEach(() => {
    requestMock.mockReset()
  })

  it('应该读取管理员竞赛公告列表并归一化公告 id', async () => {
    requestMock.mockResolvedValue([
      {
        id: 12,
        title: '报名提醒',
        content: '请尽快完成组队。',
        created_at: '2026-04-22T09:00:00.000Z',
      },
    ])

    const result = await getAdminContestAnnouncements('7')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/admin/contests/7/announcements',
    })
    expect(result).toEqual([
      {
        id: '12',
        title: '报名提醒',
        content: '请尽快完成组队。',
        created_at: '2026-04-22T09:00:00.000Z',
      },
    ])
  })

  it('应该按后端契约创建管理员竞赛公告', async () => {
    requestMock.mockResolvedValue({
      id: 13,
      title: '开赛通知',
      content: '比赛将于十分钟后开始。',
      created_at: '2026-04-22T09:10:00.000Z',
    })

    const result = await createAdminContestAnnouncement('7', {
      title: '开赛通知',
      content: '比赛将于十分钟后开始。',
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/admin/contests/7/announcements',
      data: {
        title: '开赛通知',
        content: '比赛将于十分钟后开始。',
      },
    })
    expect(result).toEqual({
      id: '13',
      title: '开赛通知',
      content: '比赛将于十分钟后开始。',
      created_at: '2026-04-22T09:10:00.000Z',
    })
  })

  it('应该按后端契约删除管理员竞赛公告', async () => {
    requestMock.mockResolvedValue(undefined)

    await deleteAdminContestAnnouncement('7', '13')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'DELETE',
      url: '/admin/contests/7/announcements/13',
    })
  })
})
