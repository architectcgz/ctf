import { flushPromises } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useStudentListQuery } from './useStudentListQuery'

function deferred<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void
  const promise = new Promise<T>((nextResolve) => {
    resolve = nextResolve
  })
  return { promise, resolve }
}

describe('useStudentListQuery', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('应该加载学生列表并暴露状态', async () => {
    const request = vi.fn().mockResolvedValue([
      {
        id: 'stu-1',
        username: 'alice',
      },
    ])

    const query = useStudentListQuery({
      request,
      errorMessage: '加载学生列表失败，请稍后重试',
      getParams: () => ({
        keyword: 'Alice',
      }),
    })

    const loadingPromise = query.loadStudents('Class A')

    expect(query.loading.value).toBe(true)

    await loadingPromise

    expect(request).toHaveBeenCalledWith('Class A', {
      keyword: 'Alice',
    })
    expect(query.students.value).toEqual([
      {
        id: 'stu-1',
        username: 'alice',
      },
    ])
    expect(query.error.value).toBeNull()
    expect(query.loading.value).toBe(false)
  })

  it('应该对计划查询做防抖，只触发最后一次输入', async () => {
    const request = vi.fn().mockResolvedValue([])

    const query = useStudentListQuery({
      request,
      errorMessage: '加载学生列表失败，请稍后重试',
      debounceMs: 250,
    })

    query.scheduleLoadStudents('Class A')
    query.scheduleLoadStudents('Class B')

    expect(request).not.toHaveBeenCalled()

    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(request).toHaveBeenCalledTimes(1)
    expect(request).toHaveBeenCalledWith('Class B', undefined)
  })

  it('应该忽略过期请求返回的数据', async () => {
    const slowRequest = deferred<Array<{ id: string; username: string }>>()
    const fastRequest = deferred<Array<{ id: string; username: string }>>()
    const request = vi
      .fn()
      .mockImplementationOnce(() => slowRequest.promise)
      .mockImplementationOnce(() => fastRequest.promise)

    const query = useStudentListQuery({
      request,
      errorMessage: '加载学生列表失败，请稍后重试',
    })

    const slowPromise = query.loadStudents('Class A')
    const fastPromise = query.loadStudents('Class A')

    fastRequest.resolve([{ id: 'stu-1', username: 'alice' }])
    await fastPromise

    slowRequest.resolve([{ id: 'stu-2', username: 'bob' }])
    await slowPromise

    expect(query.students.value).toEqual([{ id: 'stu-1', username: 'alice' }])
  })

  it('应该在班级为空时清空学生列表', async () => {
    const request = vi.fn().mockResolvedValue([{ id: 'stu-1', username: 'alice' }])

    const query = useStudentListQuery({
      request,
      errorMessage: '加载学生列表失败，请稍后重试',
    })

    await query.loadStudents('Class A')
    await query.loadStudents('')

    expect(query.students.value).toEqual([])
    expect(query.error.value).toBeNull()
    expect(query.loading.value).toBe(false)
    expect(request).toHaveBeenCalledTimes(1)
  })
})
