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
  getStudentAttackSessions,
  getStudentEvidence,
  getStudentsDirectory,
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

  it('获取学生目录分页时应透传筛选和排序参数，并标准化标识字段', async () => {
    requestMock.mockResolvedValue({
      list: [
        {
          id: 9,
          username: 'alice',
          student_no: '20240001',
          name: '张三',
          class_name: 'Class A',
          solved_count: 5,
          total_score: 320,
          recent_event_count: 2,
          weak_dimension: 'Web',
        },
      ],
      total: 31,
      page: 2,
      page_size: 10,
    })

    const result = await getStudentsDirectory({
      class_name: 'Class A',
      keyword: 'alice',
      student_no: '20240001',
      sort_key: 'total_score',
      sort_order: 'desc',
      page: 2,
      page_size: 10,
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/teacher/students',
      params: {
        class_name: 'Class A',
        keyword: 'alice',
        student_no: '20240001',
        sort_key: 'total_score',
        sort_order: 'desc',
        page: 2,
        page_size: 10,
      },
    })
    expect(result).toEqual({
      list: [
        {
          id: '9',
          username: 'alice',
          student_no: '20240001',
          name: '张三',
          class_name: 'Class A',
          solved_count: 5,
          total_score: 320,
          recent_event_count: 2,
          weak_dimension: 'Web',
        },
      ],
      total: 31,
      page: 2,
      page_size: 10,
    })
  })

  it('销毁教师实例时应保持 API 契约简洁', async () => {
    requestMock.mockResolvedValue(undefined)

    await destroyTeacherInstance('inst-3')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'DELETE',
      url: '/teacher/instances/inst-3',
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

  it('获取学员攻击会话时应透传筛选参数并标准化标识字段', async () => {
    requestMock.mockResolvedValue({
      summary: {
        total_sessions: 1,
        success_count: 1,
        failed_count: 0,
        in_progress_count: 0,
        unknown_count: 0,
        event_count: 2,
        capture_available_count: 0,
      },
      sessions: [
        {
          id: 9,
          mode: 'awd',
          student_id: 12,
          team_id: 5,
          challenge_id: 18,
          contest_id: 3,
          round_id: 7,
          service_id: 11,
          victim_team_id: 19,
          title: 'bank-portal',
          started_at: '2026-05-03T10:00:00Z',
          ended_at: '2026-05-03T10:05:00Z',
          result: 'success',
          event_count: 2,
          capture_count: 0,
          events: [
            {
              id: 21,
              session_id: 9,
              type: 'awd_attack_submission',
              stage: 'exploit',
              source: 'awd_attack_logs',
              occurred_at: '2026-05-03T10:04:00Z',
              actor: {
                user_id: 12,
                team_id: 5,
              },
              target: {
                challenge_id: 18,
                contest_id: 3,
                round_id: 7,
                service_id: 11,
                victim_team_id: 19,
              },
              summary: 'AWD 攻击提交成功',
              capture_available: false,
            },
          ],
        },
      ],
    })

    const result = await getStudentAttackSessions('stu-1', {
      mode: 'awd',
      contest_id: '3',
      round_id: '7',
      result: 'success',
      with_events: true,
      limit: 10,
      offset: 20,
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/teacher/students/stu-1/attack-sessions',
      params: {
        mode: 'awd',
        challenge_id: undefined,
        contest_id: '3',
        round_id: '7',
        result: 'success',
        with_events: true,
        limit: 10,
        offset: 20,
      },
    })
    expect(result.sessions[0]).toMatchObject({
      id: '9',
      student_id: '12',
      team_id: '5',
      challenge_id: '18',
      contest_id: '3',
      round_id: '7',
      service_id: '11',
      victim_team_id: '19',
    })
    expect(result.sessions[0].events?.[0]).toMatchObject({
      id: '21',
      session_id: '9',
      actor: {
        user_id: '12',
        team_id: '5',
      },
      target: {
        challenge_id: '18',
        contest_id: '3',
        round_id: '7',
        service_id: '11',
        victim_team_id: '19',
      },
    })
  })

  it('获取学员证据时应透传筛选参数并标准化标识字段', async () => {
    requestMock.mockResolvedValue({
      summary: {
        total_events: 2,
        proxy_request_count: 1,
        submit_count: 1,
        success_count: 1,
        challenge_id: 5,
      },
      events: [
        {
          type: 'instance_proxy_request',
          challenge_id: 5,
          title: 'web-1',
          detail: 'POST /login 200',
          timestamp: '2026-05-03T10:00:00Z',
          meta: {
            request_method: 'POST',
          },
        },
      ],
    })

    const result = await getStudentEvidence('stu-1', {
      challenge_id: '5',
      event_type: 'instance_proxy_request',
      from: '2026-05-03T09:00:00Z',
      to: '2026-05-03T11:00:00Z',
      limit: 10,
      offset: 20,
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/teacher/students/stu-1/evidence',
      params: {
        challenge_id: '5',
        contest_id: undefined,
        round_id: undefined,
        event_type: 'instance_proxy_request',
        from: '2026-05-03T09:00:00Z',
        to: '2026-05-03T11:00:00Z',
        limit: 10,
        offset: 20,
      },
    })
    expect(result).toEqual({
      summary: {
        total_events: 2,
        proxy_request_count: 1,
        submit_count: 1,
        success_count: 1,
        challenge_id: '5',
      },
      events: [
        {
          type: 'instance_proxy_request',
          challenge_id: '5',
          title: 'web-1',
          detail: 'POST /login 200',
          timestamp: '2026-05-03T10:00:00Z',
          meta: {
            request_method: 'POST',
          },
        },
      ],
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
      selected_round: {
        round: {
          id: 41,
          contest_id: 9,
          round_number: 2,
          status: 'running',
          attack_score: 60,
          defense_score: 40,
          service_count: 4,
          attack_count: 3,
          traffic_count: 18,
        },
        teams: [
          {
            team_id: 13,
            team_name: 'Red',
            captain_id: 101,
            total_score: 118,
            member_count: 3,
          },
        ],
        services: [
          {
            id: 501,
            round_id: 41,
            team_id: 13,
            team_name: 'Red',
            service_id: 7009,
            challenge_id: 9,
            challenge_title: 'Bank Portal',
            service_status: 'up',
            attack_received: 0,
            sla_score: 18,
            defense_score: 40,
            attack_score: 0,
            updated_at: '2026-04-12T10:02:00Z',
          },
        ],
        attacks: [
          {
            id: 601,
            round_id: 41,
            attacker_team_id: 13,
            attacker_team_name: 'Red',
            victim_team_id: 14,
            victim_team_name: 'Blue',
            service_id: 7009,
            challenge_id: 9,
            challenge_title: 'Bank Portal',
            attack_type: 'flag_capture',
            source: 'submission',
            is_success: true,
            score_gained: 60,
            created_at: '2026-04-12T10:03:00Z',
          },
        ],
        traffic: [],
      },
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
    expect(result.selected_round?.services[0].service_id).toBe('7009')
    expect(result.selected_round?.attacks[0].service_id).toBe('7009')
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
