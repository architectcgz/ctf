import { beforeEach, describe, expect, it, vi } from 'vitest'
import {
  createContestAWDRound,
  getAdminContestLiveScoreboard,
  getContestAWDReadiness,
  getContestAWDRoundSummary,
  getContestAWDRoundTrafficSummary,
  listContestAWDRoundAttacks,
  listContestAWDRoundTrafficEvents,
  prewarmContestAWDInstances,
  runContestAWDCurrentRoundCheck,
  startContestAWDTeamServiceInstance,
  updateContest,
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

describe('admin contest AWD runtime api contract', () => {
  beforeEach(() => {
    requestMock.mockReset()
  })

  it('应该把更新竞赛状态转换成后端枚举', async () => {
    requestMock.mockResolvedValue({
      id: 9,
      title: '春季赛',
      description: '测试竞赛',
      mode: 'awd',
      start_time: '2026-03-12T09:00:00.000Z',
      end_time: '2026-03-12T12:00:00.000Z',
      freeze_time: null,
      status: 'running',
      created_at: '2026-03-01T00:00:00.000Z',
      updated_at: '2026-03-02T00:00:00.000Z',
    })

    await updateContest('9', {
      status: 'registering',
      ends_at: '2026-03-12T12:00:00.000Z',
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'PUT',
      url: '/admin/contests/9',
      data: {
        title: undefined,
        description: undefined,
        mode: undefined,
        start_time: undefined,
        end_time: '2026-03-12T12:00:00.000Z',
        status: 'registration',
      },
    })
  })

  it('应该在更新赛事时透传 override 字段', async () => {
    requestMock.mockResolvedValue({
      id: 9,
      title: '春季赛',
      description: '测试竞赛',
      mode: 'awd',
      start_time: '2026-03-12T09:00:00.000Z',
      end_time: '2026-03-12T12:00:00.000Z',
      freeze_time: null,
      status: 'running',
      created_at: '2026-03-01T00:00:00.000Z',
      updated_at: '2026-03-02T00:00:00.000Z',
    })

    await updateContest('9', {
      status: 'running',
      force_override: true,
      override_reason: 'teacher drill',
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'PUT',
      url: '/admin/contests/9',
      data: {
        title: undefined,
        description: undefined,
        mode: undefined,
        start_time: undefined,
        end_time: undefined,
        status: 'running',
        force_override: true,
        override_reason: 'teacher drill',
      },
    })
  })

  it('不应在 running 状态更新时附带额外请求层展示参数', async () => {
    requestMock.mockResolvedValue({
      id: 9,
      title: '春季赛',
      description: '测试竞赛',
      mode: 'awd',
      start_time: '2026-03-12T09:00:00.000Z',
      end_time: '2026-03-12T12:00:00.000Z',
      freeze_time: null,
      status: 'running',
      created_at: '2026-03-01T00:00:00.000Z',
      updated_at: '2026-03-02T00:00:00.000Z',
    })

    await updateContest('9', {
      status: 'running',
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'PUT',
      url: '/admin/contests/9',
      data: {
        title: undefined,
        description: undefined,
        mode: undefined,
        start_time: undefined,
        end_time: undefined,
        status: 'running',
        force_override: undefined,
        override_reason: undefined,
      },
    })
  })
})
