import { beforeEach, describe, expect, it, vi } from 'vitest'
import adminContestReviewsSource from '@/api/admin/contest-reviews.ts?raw'
import adminTeachingApiSource from '@/api/admin/teaching.ts?raw'
import teachingAwdReviewApiSource from '@/api/teaching/awd-reviews.ts?raw'
import teachingInstanceApiSource from '@/api/teaching/instances.ts?raw'

import { getCheatDetection, publishAdminNotification } from '@/api/admin/platform'
import { getUsers } from '@/api/admin/users'

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

describe('admin platform api contract', () => {
  beforeEach(() => {
    requestMock.mockReset()
  })

  it('AWD review platform wrapper 与 teaching 实现 owner 应保持中性分层', () => {
    expect(teachingAwdReviewApiSource).toContain('export async function listAwdReviews')
    expect(teachingAwdReviewApiSource).toContain('export async function getAwdReview')
    expect(teachingAwdReviewApiSource).toContain('export async function exportAwdReviewArchive')
    expect(teachingAwdReviewApiSource).toContain('export async function exportAwdReviewReport')
    expect(teachingAwdReviewApiSource).not.toContain('export async function listTeacherAWDReviews')
    expect(teachingAwdReviewApiSource).not.toContain('export async function getTeacherAWDReview')
    expect(teachingAwdReviewApiSource).not.toContain(
      'export async function exportTeacherAWDReviewArchive'
    )
    expect(teachingAwdReviewApiSource).not.toContain(
      'export async function exportTeacherAWDReviewReport'
    )

    expect(adminContestReviewsSource).toContain('export async function listPlatformAWDReviews')
    expect(adminContestReviewsSource).toContain('export async function getPlatformAWDReview')
    expect(adminContestReviewsSource).toContain(
      'export async function exportPlatformAWDReviewArchive'
    )
    expect(adminContestReviewsSource).toContain(
      'export async function exportPlatformAWDReviewReport'
    )
    expect(adminContestReviewsSource).not.toContain(
      'listTeacherAWDReviews as listPlatformAWDReviews'
    )
    expect(adminContestReviewsSource).not.toContain('getTeacherAWDReview as getPlatformAWDReview')
    expect(adminContestReviewsSource).not.toContain(
      'exportTeacherAWDReviewArchive as exportPlatformAWDReviewArchive'
    )
    expect(adminContestReviewsSource).not.toContain(
      'exportTeacherAWDReviewReport as exportPlatformAWDReviewReport'
    )
  })

  it('实例目录 platform wrapper 与 teaching 实现 owner 应保持中性分层', () => {
    expect(teachingInstanceApiSource).toContain('export async function getInstanceDirectory')
    expect(teachingInstanceApiSource).toContain('export async function destroyManagedInstance')
    expect(teachingInstanceApiSource).not.toContain('export async function getTeacherInstances')
    expect(teachingInstanceApiSource).not.toContain('export async function destroyTeacherInstance')

    expect(adminTeachingApiSource).toContain('export async function getPlatformInstances')
    expect(adminTeachingApiSource).toContain('export async function destroyPlatformInstance')
    expect(adminTeachingApiSource).toContain('return getInstanceDirectory(params, options)')
    expect(adminTeachingApiSource).toContain('return destroyManagedInstance(id)')
    expect(adminTeachingApiSource).not.toContain('getTeacherInstances as getPlatformInstances')
    expect(adminTeachingApiSource).not.toContain(
      'destroyTeacherInstance as destroyPlatformInstance'
    )
  })

  it('应该把用户列表参数和返回值归一化', async () => {
    requestMock.mockResolvedValue({
      list: [
        {
          id: 3,
          username: 'alice',
          email: 'alice@example.com',
          student_no: null,
          teacher_no: 'T-1001',
          class_name: 'Class A',
          status: 'active',
          roles: ['teacher'],
          created_at: '2026-03-01T00:00:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const result = await getUsers({
      page: 1,
      page_size: 20,
      keyword: 'alice',
      student_no: '20240001',
      teacher_no: 'T-1001',
      role: 'teacher',
      status: 'active',
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/admin/users',
      params: {
        page: 1,
        page_size: 20,
        keyword: 'alice',
        student_no: '20240001',
        teacher_no: 'T-1001',
        role: 'teacher',
        status: 'active',
        class_name: undefined,
      },
    })
    expect(result.list[0]).toEqual({
      id: '3',
      username: 'alice',
      email: 'alice@example.com',
      student_no: undefined,
      teacher_no: 'T-1001',
      class_name: 'Class A',
      status: 'active',
      roles: ['teacher'],
      created_at: '2026-03-01T00:00:00.000Z',
    })
  })

  it('应该请求管理员通知发布接口并归一化批次回执', async () => {
    requestMock.mockResolvedValue({
      batch_id: 88,
      recipient_count: 56,
    })

    const result = await publishAdminNotification({
      type: 'system',
      title: '维护通知',
      content: '今晚 23:00 进行维护。',
      audience_rules: {
        mode: 'union',
        rules: [{ type: 'all' }],
      },
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/admin/notifications',
      data: {
        type: 'system',
        title: '维护通知',
        content: '今晚 23:00 进行维护。',
        audience_rules: {
          mode: 'union',
          rules: [{ type: 'all' }],
        },
      },
    })
    expect(result).toEqual({
      batch_id: '88',
      recipient_count: 56,
    })
  })

  it('应该把作弊检测响应中的用户 ID 归一化', async () => {
    requestMock.mockResolvedValue({
      generated_at: '2026-03-07T06:00:00.000Z',
      summary: {
        submit_burst_users: 1,
        shared_ip_groups: 1,
        affected_users: 2,
      },
      suspects: [
        {
          user_id: 8,
          username: 'alice',
          submit_count: 9,
          last_seen_at: '2026-03-07T05:58:00.000Z',
          reason: '短时间内提交次数异常偏高',
        },
      ],
      shared_ips: [
        {
          ip: '10.0.0.1',
          user_count: 2,
          usernames: ['alice', 'bob'],
        },
      ],
    })

    const result = await getCheatDetection()

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/admin/cheat-detection',
    })
    expect(result.suspects[0].user_id).toBe('8')
  })
})
