import { beforeEach, describe, expect, it, vi } from 'vitest'

const requestMock = vi.hoisted(() => vi.fn())

vi.mock('@/api/request', () => ({
  request: requestMock,
}))

import { createContest, getContests, updateContest } from '@/api/admin'

describe('admin contest api contract', () => {
  beforeEach(() => {
    requestMock.mockReset()
  })

  it('应该把竞赛列表参数和返回值归一化', async () => {
    requestMock.mockResolvedValue({
      list: [
        {
          id: 7,
          title: '春季赛',
          description: '测试竞赛',
          mode: 'jeopardy',
          start_time: '2026-03-10T09:00:00.000Z',
          end_time: '2026-03-10T12:00:00.000Z',
          freeze_time: null,
          status: 'registration',
          created_at: '2026-03-01T00:00:00.000Z',
          updated_at: '2026-03-02T00:00:00.000Z',
        },
      ],
      total: 1,
      page: 2,
      page_size: 5,
    })

    const result = await getContests({ page: 2, page_size: 5, status: 'registering' })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/admin/contests',
      params: {
        page: 2,
        size: 5,
        status: 'registration',
      },
    })
    expect(result).toEqual({
      list: [
        {
          id: '7',
          title: '春季赛',
          description: '测试竞赛',
          mode: 'jeopardy',
          status: 'registering',
          starts_at: '2026-03-10T09:00:00.000Z',
          ends_at: '2026-03-10T12:00:00.000Z',
          scoreboard_frozen: false,
        },
      ],
      total: 1,
      page: 2,
      page_size: 5,
    })
  })

  it('应该把创建竞赛请求转换成后端字段', async () => {
    requestMock.mockResolvedValue({
      id: 9,
      title: '春季赛',
      description: '测试竞赛',
      mode: 'awd',
      start_time: '2026-03-12T09:00:00.000Z',
      end_time: '2026-03-12T12:00:00.000Z',
      freeze_time: '2026-03-12T11:30:00.000Z',
      status: 'draft',
      created_at: '2026-03-01T00:00:00.000Z',
      updated_at: '2026-03-02T00:00:00.000Z',
    })

    const result = await createContest({
      title: '春季赛',
      description: '测试竞赛',
      mode: 'awd',
      starts_at: '2026-03-12T09:00:00.000Z',
      ends_at: '2026-03-12T12:00:00.000Z',
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/admin/contests',
      data: {
        title: '春季赛',
        description: '测试竞赛',
        mode: 'awd',
        start_time: '2026-03-12T09:00:00.000Z',
        end_time: '2026-03-12T12:00:00.000Z',
        status: undefined,
      },
    })
    expect(result).toEqual({
      contest: {
        id: '9',
        title: '春季赛',
        description: '测试竞赛',
        mode: 'awd',
        status: 'draft',
        starts_at: '2026-03-12T09:00:00.000Z',
        ends_at: '2026-03-12T12:00:00.000Z',
        scoreboard_frozen: true,
      },
    })
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
})
