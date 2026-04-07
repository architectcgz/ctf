import { beforeEach, describe, expect, it, vi } from 'vitest'

const requestMock = vi.hoisted(() => vi.fn())

vi.mock('@/api/request', () => ({
  request: requestMock,
}))

import { getClasses } from '@/api/teacher'

describe('teacher api contract', () => {
  beforeEach(() => {
    requestMock.mockReset()
  })

  it('不传分页参数时应继续返回班级数组', async () => {
    requestMock.mockResolvedValue({
      list: [{ name: 'Class A', student_count: 2 }],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const result = await getClasses()

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/teacher/classes',
      params: {
        page: undefined,
        page_size: undefined,
      },
    })
    expect(result).toEqual([{ name: 'Class A', student_count: 2 }])
  })

  it('传分页参数时应返回分页结构', async () => {
    requestMock.mockResolvedValue({
      list: [{ name: 'Class B', student_count: 3 }],
      total: 21,
      page: 2,
      page_size: 20,
    })

    const result = await getClasses({ page: 2, page_size: 20 })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/teacher/classes',
      params: {
        page: 2,
        page_size: 20,
      },
    })
    expect(result).toEqual({
      list: [{ name: 'Class B', student_count: 3 }],
      total: 21,
      page: 2,
      page_size: 20,
    })
  })
})
