import { beforeEach, describe, expect, it, vi } from 'vitest'

const teacherApiMocks = vi.hoisted(() => ({
  getTeacherInstances: vi.fn(),
  destroyTeacherInstance: vi.fn(),
}))

const adminApiMocks = vi.hoisted(() => ({
  getPlatformInstances: vi.fn(),
  destroyPlatformInstance: vi.fn(),
}))

vi.mock('@/api/teacher', () => teacherApiMocks)
vi.mock('@/api/admin', () => adminApiMocks)

import {
  destroyManagedInstanceByRole,
  getInstanceDirectoryByRole,
} from '@/api/instances'

describe('instances api role-aware access owner', () => {
  beforeEach(() => {
    Object.values(teacherApiMocks).forEach((mock) => mock.mockReset())
    Object.values(adminApiMocks).forEach((mock) => mock.mockReset())
  })

  it('admin role 应委托到 platform instance owner', async () => {
    const response = {
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
      summary: {
        total_count: 0,
        running_count: 0,
        expiring_soon_count: 0,
        warning_count: 0,
      },
    }
    adminApiMocks.getPlatformInstances.mockResolvedValue(response)
    adminApiMocks.destroyPlatformInstance.mockResolvedValue(undefined)

    const signal = AbortSignal.timeout(1000)
    const result = await getInstanceDirectoryByRole(
      'admin',
      {
        class_name: 'Class A',
        keyword: 'alice',
        student_no: '20240001',
        status: 'running',
        page: 2,
        page_size: 15,
      },
      { signal }
    )

    await destroyManagedInstanceByRole('admin', 'inst-1')

    expect(adminApiMocks.getPlatformInstances).toHaveBeenCalledWith(
      {
        class_name: 'Class A',
        keyword: 'alice',
        student_no: '20240001',
        status: 'running',
        page: 2,
        page_size: 15,
      },
      { signal }
    )
    expect(adminApiMocks.destroyPlatformInstance).toHaveBeenCalledWith('inst-1')
    expect(teacherApiMocks.getTeacherInstances).not.toHaveBeenCalled()
    expect(teacherApiMocks.destroyTeacherInstance).not.toHaveBeenCalled()
    expect(result).toBe(response)
  })

  it('teacher role 与空 role 应继续委托到 teacher instance owner', async () => {
    const response = {
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
      summary: {
        total_count: 0,
        running_count: 0,
        expiring_soon_count: 0,
        warning_count: 0,
      },
    }
    teacherApiMocks.getTeacherInstances.mockResolvedValue(response)
    teacherApiMocks.destroyTeacherInstance.mockResolvedValue(undefined)

    const teacherResult = await getInstanceDirectoryByRole('teacher', {
      keyword: 'bob',
    })
    const defaultResult = await getInstanceDirectoryByRole(undefined, {
      keyword: 'carol',
    })
    await destroyManagedInstanceByRole(undefined, 'inst-2')

    expect(teacherApiMocks.getTeacherInstances).toHaveBeenNthCalledWith(
      1,
      {
        keyword: 'bob',
      },
      undefined
    )
    expect(teacherApiMocks.getTeacherInstances).toHaveBeenNthCalledWith(
      2,
      {
        keyword: 'carol',
      },
      undefined
    )
    expect(teacherApiMocks.destroyTeacherInstance).toHaveBeenCalledWith('inst-2')
    expect(adminApiMocks.getPlatformInstances).not.toHaveBeenCalled()
    expect(adminApiMocks.destroyPlatformInstance).not.toHaveBeenCalled()
    expect(teacherResult).toBe(response)
    expect(defaultResult).toBe(response)
  })
})
