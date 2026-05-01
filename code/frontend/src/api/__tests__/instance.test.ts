import { beforeEach, describe, expect, it, vi } from 'vitest'

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

import { createInstance } from '@/api/challenge'
import { destroyInstance, getMyInstances, requestInstanceAccess } from '@/api/instance'

describe('instance api contract', () => {
  beforeEach(() => {
    requestMock.mockReset()
  })

  it('应该把实例列表中的延时计数归一化为剩余次数', async () => {
    requestMock.mockResolvedValue([
      {
        id: 9,
        challenge_id: 3,
        challenge_title: 'SQL Injection Login Bypass',
        category: 'web',
        difficulty: 'easy',
        status: 'running',
        access_url: 'http://127.0.0.1:30000',
        flag_type: 'static',
        share_scope: 'shared',
        contest_mode: 'awd',
        expires_at: '2099-01-01T00:00:00Z',
        extend_count: 1,
        max_extends: 3,
        created_at: '2026-03-12T00:00:00.000Z',
      },
    ])

    const result = await getMyInstances()

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/instances',
    })
    expect(result).toEqual([
      {
        id: '9',
        challenge_id: '3',
        challenge_title: 'SQL Injection Login Bypass',
        category: 'web',
        difficulty: 'easy',
        status: 'running',
        access_url: 'http://127.0.0.1:30000',
        flag_type: 'static',
        share_scope: 'shared',
        contest_mode: 'awd',
        expires_at: '2099-01-01T00:00:00Z',
        remaining_extends: 2,
        created_at: '2026-03-12T00:00:00.000Z',
      },
    ])
  })

  it('应该把创建实例响应中的延时计数归一化为剩余次数', async () => {
    requestMock.mockResolvedValue({
      id: 5,
      challenge_id: 3,
      status: 'running',
      access_url: 'http://127.0.0.1:30000',
      flag_type: 'static',
      share_scope: 'shared',
      expires_at: '2099-01-01T00:00:00Z',
      extend_count: 0,
      max_extends: 2,
      created_at: '2026-03-12T00:00:00.000Z',
    })

    const result = await createInstance('3')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/challenges/3/instances',
    })
    expect(result).toEqual({
      id: '5',
      challenge_id: '3',
      status: 'running',
      access_url: 'http://127.0.0.1:30000',
      flag_type: 'static',
      share_scope: 'shared',
      expires_at: '2099-01-01T00:00:00Z',
      remaining_extends: 2,
      created_at: '2026-03-12T00:00:00.000Z',
    })
  })

  it('销毁实例时应保持 API 契约简洁', async () => {
    requestMock.mockResolvedValue(undefined)

    await destroyInstance('inst-9')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'DELETE',
      url: '/instances/inst-9',
    })
  })

  it('请求实例访问入口时应保持 API 契约简洁', async () => {
    requestMock.mockResolvedValue({
      access_url: 'http://instance.ready.test',
    })

    await requestInstanceAccess('inst-9')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/instances/inst-9/access',
    })
  })
})
