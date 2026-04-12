import { beforeEach, describe, expect, it, vi } from 'vitest'

const requestMock = vi.hoisted(() => vi.fn())

vi.mock('@/api/request', () => ({
  request: requestMock,
}))

import {
  destroyTeacherInstance,
  exportTeacherAWDReviewArchive,
  exportTeacherAWDReviewReport,
  getClasses,
  getTeacherAWDReview,
  getTeacherWriteupSubmissions,
  listTeacherAWDReviews,
} from '@/api/teacher'

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

  it('销毁教师实例时应关闭全局错误提示，交给页面展示具体原因', async () => {
    requestMock.mockResolvedValue(undefined)

    await destroyTeacherInstance('inst-3')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'DELETE',
      url: '/teacher/instances/inst-3',
      suppressErrorToast: true,
    })
  })

  it('获取题解投稿列表时应保留 student_no 并标准化标识字段', async () => {
    requestMock.mockResolvedValue({
      list: [
        {
          id: 9,
          user_id: 12,
          student_username: 'alice_01',
          student_name: '张三',
          student_no: '20240001',
          class_name: '网安 1 班',
          challenge_id: 5,
          challenge_title: '双节点演练',
          title: '利用链梳理',
          content_preview: 'preview',
          submission_status: 'published',
          visibility_status: 'visible',
          is_recommended: true,
          published_at: '2026-04-10T10:00:00Z',
          updated_at: '2026-04-10T10:30:00Z',
        },
      ],
      total: 1,
      page: 2,
      page_size: 6,
    })

    const result = await getTeacherWriteupSubmissions({ challenge_id: '5', page: 2, page_size: 6 })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/teacher/writeup-submissions',
      params: {
        student_id: undefined,
        challenge_id: '5',
        class_name: undefined,
        submission_status: undefined,
        visibility_status: undefined,
        page: 2,
        page_size: 6,
      },
    })
    expect(result).toEqual({
      list: [
        {
          id: '9',
          user_id: '12',
          student_username: 'alice_01',
          student_name: '张三',
          student_no: '20240001',
          class_name: '网安 1 班',
          challenge_id: '5',
          challenge_title: '双节点演练',
          title: '利用链梳理',
          content_preview: 'preview',
          submission_status: 'published',
          visibility_status: 'visible',
          is_recommended: true,
          published_at: '2026-04-10T10:00:00Z',
          updated_at: '2026-04-10T10:30:00Z',
        },
      ],
      total: 1,
      page: 2,
      page_size: 6,
    })
  })

  it('获取教师 AWD 复盘列表时应透传筛选参数', async () => {
    requestMock.mockResolvedValue({
      contests: [],
    })

    await listTeacherAWDReviews({ status: 'running', keyword: '春季' })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/teacher/awd/reviews',
      params: {
        status: 'running',
        keyword: '春季',
      },
    })
  })

  it('获取教师 AWD 复盘详情时应透传轮次和队伍筛选', async () => {
    requestMock.mockResolvedValue({
      generated_at: '2026-04-12T10:00:00Z',
      scope: {
        snapshot_type: 'live',
        requested_by: 3,
        requested_id: 9,
      },
      contest: {
        id: 9,
        title: '春季 AWD',
        mode: 'awd',
        status: 'running',
        round_count: 3,
        team_count: 6,
        export_ready: false,
      },
      rounds: [],
    })

    const result = await getTeacherAWDReview('9', { round: 2, team_id: 'team-2' })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/teacher/awd/reviews/9',
      params: {
        round: 2,
        team_id: 'team-2',
      },
    })
    expect(result.contest.id).toBe('9')
  })

  it('创建教师 AWD 复盘归档导出时应标准化 report_id', async () => {
    requestMock.mockResolvedValue({
      report_id: 21,
      status: 'processing',
    })

    const result = await exportTeacherAWDReviewArchive('9', {
      round_number: 2,
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/teacher/awd/reviews/9/export/archive',
      data: {
        round_number: 2,
      },
    })
    expect(result.report_id).toBe('21')
    expect(result.status).toBe('processing')
  })

  it('创建教师 AWD 复盘报告导出时应标准化 report_id', async () => {
    requestMock.mockResolvedValue({
      report_id: 22,
      status: 'processing',
    })

    const result = await exportTeacherAWDReviewReport('9', {
      round_number: 3,
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/teacher/awd/reviews/9/export/report',
      data: {
        round_number: 3,
      },
    })
    expect(result.report_id).toBe('22')
    expect(result.status).toBe('processing')
  })
})
