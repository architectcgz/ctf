import { beforeEach, describe, expect, it, vi } from 'vitest'

const requestMock = vi.hoisted(() => vi.fn())

vi.mock('@/api/request', () => ({
  request: requestMock,
}))

import { getNotifications, markAsRead } from '@/api/notification'

describe('notification api contract', () => {
  beforeEach(() => {
    requestMock.mockReset()
  })

  it('应按通知列表接口读取消息分页数据', async () => {
    requestMock.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
    })

    await getNotifications({ page: 1 })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/notifications',
      params: { page: 1 },
    })
  })

  it('标记已读时应保持 API 契约简洁', async () => {
    requestMock.mockResolvedValue(undefined)

    await markAsRead('9')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'PUT',
      url: '/notifications/9/read',
    })
  })
})
