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

import {
  createAdminContestChallenge,
  createEnvironmentTemplate,
  createChallenge,
  createChallengePublishRequest,
  createContest,
  configureChallengeFlag,
  deleteChallengeTopology,
  deleteEnvironmentTemplate,
  deleteImage,
  deleteChallengeWriteup,
  getAdminContestLiveScoreboard,
  getChallengeTopology,
  getChallengeDetail,
  getLatestChallengePublishRequest,
  getContestAWDRoundSummary,
  getContestAWDRoundTrafficSummary,
  getChallengeWriteup,
  getChallenges,
  getCheatDetection,
  getContests,
  getEnvironmentTemplates,
  getImages,
  getUsers,
  listChallengeImports,
  listAdminContestChallenges,
  listContestAWDRoundAttacks,
  listContestAWDRoundTrafficEvents,
  publishAdminNotification,
  recommendChallengeWriteup,
  runContestAWDRoundCheck,
  saveChallengeTopology,
  saveChallengeWriteup,
  unrecommendChallengeWriteup,
  updateAdminContestChallenge,
  updateContest,
} from '@/api/admin'
import * as adminApi from '@/api/admin'

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
        page_size: 5,
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

  it('应该在导入记录接口返回空值时兜底为空数组', async () => {
    requestMock.mockResolvedValue(null)

    const result = await listChallengeImports()

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/authoring/challenge-imports',
    })
    expect(result).toEqual([])
  })

  it('应该归一化管理员竞赛题目列表中的题目元信息', async () => {
    requestMock.mockResolvedValue([
      {
        id: 31,
        contest_id: 7,
        challenge_id: 11,
        title: 'SQL Injection 101',
        category: 'web',
        difficulty: 'easy',
        points: 120,
        order: 2,
        is_visible: true,
        awd_checker_type: 'http_standard',
        awd_checker_config: {
          put_flag: { method: 'PUT', path: '/api/flag' },
          get_flag: { method: 'GET', path: '/api/flag' },
        },
        awd_sla_score: 18,
        awd_defense_score: 28,
        awd_checker_validation_state: 'passed',
        awd_checker_last_preview_at: '2026-03-12T00:10:00.000Z',
        awd_checker_last_preview_result: {
          checker_type: 'http_standard',
          service_status: 'up',
          check_result: {
            checker_type: 'http_standard',
            check_source: 'checker_preview',
            status_reason: 'healthy',
          },
          preview_context: {
            access_url: 'http://preview.internal',
            preview_flag: 'flag{preview}',
            round_number: 0,
            team_id: 0,
            challenge_id: 11,
          },
        },
        created_at: '2026-03-12T00:00:00.000Z',
      },
    ])

    const result = await listAdminContestChallenges('7')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/admin/contests/7/challenges',
    })
    expect(result).toEqual([
      {
        id: '31',
        contest_id: '7',
        challenge_id: '11',
        title: 'SQL Injection 101',
        category: 'web',
        difficulty: 'easy',
        points: 120,
        order: 2,
        is_visible: true,
        awd_checker_type: 'http_standard',
        awd_checker_config: {
          put_flag: { method: 'PUT', path: '/api/flag' },
          get_flag: { method: 'GET', path: '/api/flag' },
        },
        awd_sla_score: 18,
        awd_defense_score: 28,
        awd_checker_validation_state: 'passed',
        awd_checker_last_preview_at: '2026-03-12T00:10:00.000Z',
        awd_checker_last_preview_result: {
          checker_type: 'http_standard',
          service_status: 'up',
          check_result: {
            checker_type: 'http_standard',
            check_source: 'checker_preview',
            status_reason: 'healthy',
          },
          preview_context: {
            access_url: 'http://preview.internal',
            preview_flag: 'flag{preview}',
            round_number: 0,
            team_id: '0',
            challenge_id: '11',
          },
        },
        created_at: '2026-03-12T00:00:00.000Z',
      },
    ])
  })

  it('应该按后端契约创建带 AWD 配置的竞赛题目', async () => {
    requestMock.mockResolvedValue({
      id: 31,
      contest_id: 7,
      challenge_id: 11,
      title: 'SQL Injection 101',
      category: 'web',
      difficulty: 'easy',
      points: 120,
      order: 2,
      is_visible: true,
      awd_checker_type: 'http_standard',
      awd_checker_config: {
        put_flag: { method: 'PUT', path: '/api/flag' },
      },
      awd_sla_score: 18,
      awd_defense_score: 28,
      created_at: '2026-03-12T00:00:00.000Z',
    })

    const result = await createAdminContestChallenge('7', {
      challenge_id: 11,
      points: 120,
      order: 2,
      is_visible: true,
      awd_checker_type: 'http_standard',
      awd_checker_config: {
        put_flag: { method: 'PUT', path: '/api/flag' },
      },
      awd_sla_score: 18,
      awd_defense_score: 28,
      awd_checker_preview_token: 'preview-token-1',
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/admin/contests/7/challenges',
      data: {
        challenge_id: 11,
        points: 120,
        order: 2,
        is_visible: true,
        awd_checker_type: 'http_standard',
        awd_checker_config: {
          put_flag: { method: 'PUT', path: '/api/flag' },
        },
        awd_sla_score: 18,
        awd_defense_score: 28,
        awd_checker_preview_token: 'preview-token-1',
      },
    })
    expect(result).toEqual({
      id: '31',
      contest_id: '7',
      challenge_id: '11',
      title: 'SQL Injection 101',
      category: 'web',
      difficulty: 'easy',
      points: 120,
      order: 2,
      is_visible: true,
      awd_checker_type: 'http_standard',
      awd_checker_config: {
        put_flag: { method: 'PUT', path: '/api/flag' },
      },
      awd_sla_score: 18,
      awd_defense_score: 28,
      awd_checker_validation_state: 'pending',
      awd_checker_last_preview_at: undefined,
      awd_checker_last_preview_result: undefined,
      created_at: '2026-03-12T00:00:00.000Z',
    })
  })

  it('应该按后端契约更新竞赛题目的 AWD 配置', async () => {
    requestMock.mockResolvedValue(null)

    await updateAdminContestChallenge('7', '11', {
      points: 150,
      order: 3,
      is_visible: false,
      awd_checker_type: 'legacy_probe',
      awd_checker_config: {
        health_path: '/healthz',
      },
      awd_sla_score: 10,
      awd_defense_score: 20,
      awd_checker_preview_token: 'preview-token-2',
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'PUT',
      url: '/admin/contests/7/challenges/11',
      data: {
        points: 150,
        order: 3,
        is_visible: false,
        awd_checker_type: 'legacy_probe',
        awd_checker_config: {
          health_path: '/healthz',
        },
        awd_sla_score: 10,
        awd_defense_score: 20,
        awd_checker_preview_token: 'preview-token-2',
      },
    })
  })

  it('应该请求指定 AWD 轮次巡检接口并归一化返回值', async () => {
    requestMock.mockResolvedValue({
      round: {
        id: 41,
        contest_id: 7,
        round_number: 3,
        status: 'finished',
        started_at: '2026-03-12T10:00:00.000Z',
        ended_at: '2026-03-12T10:05:00.000Z',
        attack_score: 80,
        defense_score: 45,
        created_at: '2026-03-12T10:00:00.000Z',
        updated_at: '2026-03-12T10:06:00.000Z',
      },
      services: [
        {
          id: 91,
          round_id: 41,
          team_id: 12,
          team_name: 'Blue',
          challenge_id: 101,
          service_status: 'up',
          check_result: { status_reason: 'healthy' },
          checker_type: 'http_standard',
          attack_received: 0,
          sla_score: 18,
          defense_score: 45,
          attack_score: 0,
          updated_at: '2026-03-12T10:06:00.000Z',
        },
      ],
    })

    const result = await runContestAWDRoundCheck('7', '41')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/admin/contests/7/awd/rounds/41/check',
    })
    expect(result).toEqual({
      round: {
        id: '41',
        contest_id: '7',
        round_number: 3,
        status: 'finished',
        started_at: '2026-03-12T10:00:00.000Z',
        ended_at: '2026-03-12T10:05:00.000Z',
        attack_score: 80,
        defense_score: 45,
        created_at: '2026-03-12T10:00:00.000Z',
        updated_at: '2026-03-12T10:06:00.000Z',
      },
      services: [
        {
          id: '91',
          round_id: '41',
          team_id: '12',
          team_name: 'Blue',
          challenge_id: '101',
          service_status: 'up',
          check_result: { status_reason: 'healthy' },
          checker_type: 'http_standard',
          attack_received: 0,
          sla_score: 18,
          defense_score: 45,
          attack_score: 0,
          updated_at: '2026-03-12T10:06:00.000Z',
        },
      ],
    })
  })

  it('应该请求 AWD checker 试跑接口并归一化返回值', async () => {
    requestMock.mockResolvedValue({
      checker_type: 'http_standard',
      service_status: 'up',
      check_result: {
        checker_type: 'http_standard',
        check_source: 'checker_preview',
        status_reason: 'healthy',
      },
      preview_context: {
        access_url: 'http://preview.internal',
        preview_flag: 'flag{preview}',
        round_number: 0,
        team_id: 0,
        challenge_id: 101,
      },
      preview_token: 'preview-token-9',
    })

    const previewFn = (
      adminApi as typeof adminApi & {
        runContestAWDCheckerPreview: (
          contestId: string,
          data: Record<string, unknown>
        ) => Promise<Record<string, unknown>>
      }
    ).runContestAWDCheckerPreview

    const result = await previewFn('7', {
      challenge_id: 101,
      checker_type: 'http_standard',
      checker_config: {
        put_flag: { method: 'PUT', path: '/api/flag' },
        get_flag: { method: 'GET', path: '/api/flag' },
      },
      access_url: 'http://preview.internal',
      preview_flag: 'flag{preview}',
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/admin/contests/7/awd/checker-preview',
      data: {
        challenge_id: 101,
        checker_type: 'http_standard',
        checker_config: {
          put_flag: { method: 'PUT', path: '/api/flag' },
          get_flag: { method: 'GET', path: '/api/flag' },
        },
        access_url: 'http://preview.internal',
        preview_flag: 'flag{preview}',
      },
    })
    expect(result).toEqual({
      checker_type: 'http_standard',
      service_status: 'up',
      check_result: {
        checker_type: 'http_standard',
        check_source: 'checker_preview',
        status_reason: 'healthy',
      },
      preview_context: {
        access_url: 'http://preview.internal',
        preview_flag: 'flag{preview}',
        round_number: 0,
        team_id: '0',
        challenge_id: '101',
      },
      preview_token: 'preview-token-9',
    })
  })

  it('应该归一化 AWD 轮次汇总中的运维指标', async () => {
    requestMock.mockResolvedValue({
      round: {
        id: 41,
        contest_id: 7,
        round_number: 3,
        status: 'finished',
        started_at: '2026-03-12T10:00:00.000Z',
        ended_at: '2026-03-12T10:05:00.000Z',
        attack_score: 80,
        defense_score: 45,
        created_at: '2026-03-12T10:00:00.000Z',
        updated_at: '2026-03-12T10:06:00.000Z',
      },
      metrics: {
        total_service_count: 6,
        service_up_count: 4,
        service_down_count: 1,
        service_compromised_count: 1,
        attacked_service_count: 2,
        defense_success_count: 1,
        total_attack_count: 5,
        successful_attack_count: 3,
        failed_attack_count: 2,
        scheduler_check_count: 4,
        manual_current_round_check_count: 0,
        manual_selected_round_check_count: 1,
        manual_service_check_count: 1,
        submission_attack_count: 3,
        manual_attack_log_count: 2,
        legacy_attack_log_count: 0,
      },
      items: [
        {
          team_id: 12,
          team_name: 'Blue',
          service_up_count: 1,
          service_down_count: 0,
          service_compromised_count: 0,
          sla_score: 18,
          defense_score: 45,
          attack_score: 0,
          successful_attack_count: 0,
          successful_breach_count: 0,
          unique_attackers_against: 0,
          total_score: 63,
        },
      ],
    })

    const result = await getContestAWDRoundSummary('7', '41')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/admin/contests/7/awd/rounds/41/summary',
    })
    expect(result).toEqual({
      round: {
        id: '41',
        contest_id: '7',
        round_number: 3,
        status: 'finished',
        started_at: '2026-03-12T10:00:00.000Z',
        ended_at: '2026-03-12T10:05:00.000Z',
        attack_score: 80,
        defense_score: 45,
        created_at: '2026-03-12T10:00:00.000Z',
        updated_at: '2026-03-12T10:06:00.000Z',
      },
      metrics: {
        total_service_count: 6,
        service_up_count: 4,
        service_down_count: 1,
        service_compromised_count: 1,
        attacked_service_count: 2,
        defense_success_count: 1,
        total_attack_count: 5,
        successful_attack_count: 3,
        failed_attack_count: 2,
        scheduler_check_count: 4,
        manual_current_round_check_count: 0,
        manual_selected_round_check_count: 1,
        manual_service_check_count: 1,
        submission_attack_count: 3,
        manual_attack_log_count: 2,
        legacy_attack_log_count: 0,
      },
      items: [
        {
          team_id: '12',
          team_name: 'Blue',
          service_up_count: 1,
          service_down_count: 0,
          service_compromised_count: 0,
          sla_score: 18,
          defense_score: 45,
          attack_score: 0,
          successful_attack_count: 0,
          successful_breach_count: 0,
          unique_attackers_against: 0,
          total_score: 63,
        },
      ],
    })
  })

  it('应该归一化 AWD 攻击流量摘要数据', async () => {
    requestMock.mockResolvedValue({
      round: {
        id: 41,
        contest_id: 7,
        round_number: 3,
        status: 'running',
        started_at: '2026-03-12T10:00:00.000Z',
        ended_at: null,
        attack_score: 50,
        defense_score: 40,
        created_at: '2026-03-12T09:59:00.000Z',
        updated_at: '2026-03-12T10:01:11.000Z',
      },
      contest_id: 7,
      round_id: 41,
      total_request_count: 18,
      active_attacker_team_count: 3,
      victim_team_count: 2,
      error_request_count: 4,
      unique_path_count: 7,
      latest_event_at: '2026-03-12T10:01:11.000Z',
      trend_buckets: [
        {
          bucket_start_at: '2026-03-12T10:00:00.000Z',
          bucket_end_at: '2026-03-12T10:01:00.000Z',
          request_count: 6,
          error_count: 1,
        },
      ],
      top_victims: [
        {
          team_id: 12,
          team_name: 'Blue',
          request_count: 9,
          error_count: 2,
        },
      ],
      top_attackers: [
        {
          team_id: 11,
          team_name: 'Red',
          request_count: 10,
          error_count: 3,
        },
      ],
      top_challenges: [
        {
          challenge_id: 101,
          challenge_title: 'Web 1',
          request_count: 11,
          error_count: 4,
        },
      ],
      top_paths: [
        {
          path: '/api/flag',
          request_count: 8,
          error_count: 3,
          last_status_code: 500,
        },
      ],
      top_error_paths: [
        {
          path: '/api/flag',
          request_count: 8,
          error_count: 3,
          last_status_code: 500,
        },
      ],
    })

    const result = await getContestAWDRoundTrafficSummary('7', '41')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/admin/contests/7/awd/rounds/41/traffic/summary',
    })
    expect(result).toEqual({
      round: {
        id: '41',
        contest_id: '7',
        round_number: 3,
        status: 'running',
        started_at: '2026-03-12T10:00:00.000Z',
        ended_at: undefined,
        attack_score: 50,
        defense_score: 40,
        created_at: '2026-03-12T09:59:00.000Z',
        updated_at: '2026-03-12T10:01:11.000Z',
      },
      contest_id: '7',
      round_id: '41',
      total_request_count: 18,
      active_attacker_team_count: 3,
      victim_team_count: 2,
      error_request_count: 4,
      unique_path_count: 7,
      latest_event_at: '2026-03-12T10:01:11.000Z',
      trend_buckets: [
        {
          bucket_start_at: '2026-03-12T10:00:00.000Z',
          bucket_end_at: '2026-03-12T10:01:00.000Z',
          request_count: 6,
          error_count: 1,
        },
      ],
      top_victims: [
        {
          team_id: '12',
          team_name: 'Blue',
          request_count: 9,
          error_count: 2,
        },
      ],
      top_attackers: [
        {
          team_id: '11',
          team_name: 'Red',
          request_count: 10,
          error_count: 3,
        },
      ],
      top_challenges: [
        {
          challenge_id: '101',
          challenge_title: 'Web 1',
          request_count: 11,
          error_count: 4,
        },
      ],
      top_paths: [
        {
          path: '/api/flag',
          request_count: 8,
          error_count: 3,
          last_status_code: 500,
        },
      ],
      top_error_paths: [
        {
          path: '/api/flag',
          request_count: 8,
          error_count: 3,
          last_status_code: 500,
        },
      ],
    })
  })

  it('应该归一化 AWD 攻击流量事件分页数据', async () => {
    requestMock.mockResolvedValue({
      list: [
        {
          id: 301,
          contest_id: 7,
          round_id: 41,
          occurred_at: '2026-03-12T10:00:30.000Z',
          attacker_team_id: 11,
          attacker_team_name: 'Red',
          victim_team_id: 12,
          victim_team_name: 'Blue',
          challenge_id: 101,
          challenge_title: 'Web 1',
          method: 'GET',
          path: '/api/flag',
          status_code: 500,
          status_group: 'server_error',
          is_error: true,
          source: 'proxy_audit',
          request_id: 'req-1',
        },
      ],
      total: 23,
      page: 2,
      page_size: 20,
    })

    const result = await listContestAWDRoundTrafficEvents('7', '41', {
      page: 2,
      page_size: 20,
      attacker_team_id: '11',
      status_group: 'server_error',
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/admin/contests/7/awd/rounds/41/traffic/events',
      params: {
        page: 2,
        page_size: 20,
        attacker_team_id: '11',
        status_group: 'server_error',
      },
    })
    expect(result).toEqual({
      list: [
        {
          id: '301',
          contest_id: '7',
          round_id: '41',
          occurred_at: '2026-03-12T10:00:30.000Z',
          attacker_team_id: '11',
          attacker_team_name: 'Red',
          victim_team_id: '12',
          victim_team_name: 'Blue',
          challenge_id: '101',
          challenge_title: 'Web 1',
          method: 'GET',
          path: '/api/flag',
          status_code: 500,
          status_group: 'server_error',
          is_error: true,
          source: 'proxy_audit',
          request_id: 'req-1',
        },
      ],
      total: 23,
      page: 2,
      page_size: 20,
    })
  })

  it('应该请求管理员实时排行榜接口并归一化返回值', async () => {
    requestMock.mockResolvedValue({
      contest: {
        id: 7,
        title: '春季赛',
        status: 'frozen',
        started_at: '2026-03-12T09:00:00.000Z',
        ends_at: '2026-03-12T12:00:00.000Z',
      },
      scoreboard: {
        list: [
          {
            rank: 1,
            team_id: 11,
            team_name: 'Blue',
            score: 350,
            solved_count: 4,
            last_submission_at: '2026-03-12T11:40:00.000Z',
          },
        ],
        total: 1,
        page: 1,
        page_size: 10,
      },
      frozen: false,
    })

    const result = await getAdminContestLiveScoreboard('7', { page: 1, page_size: 10 })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/admin/contests/7/scoreboard/live',
      params: { page: 1, page_size: 10 },
    })
    expect(result).toEqual({
      contest: {
        id: '7',
        title: '春季赛',
        status: 'frozen',
        started_at: '2026-03-12T09:00:00.000Z',
        ends_at: '2026-03-12T12:00:00.000Z',
      },
      scoreboard: {
        list: [
          {
            rank: 1,
            team_id: '11',
            team_name: 'Blue',
            score: 350,
            solved_count: 4,
            last_submission_at: '2026-03-12T11:40:00.000Z',
          },
        ],
        total: 1,
        page: 1,
        page_size: 10,
      },
      frozen: false,
    })
  })

  it('应该归一化 AWD 攻击日志来源字段', async () => {
    requestMock.mockResolvedValue([
      {
        id: 71,
        round_id: 41,
        attacker_team_id: 11,
        attacker_team: 'Red',
        victim_team_id: 12,
        victim_team: 'Blue',
        challenge_id: 101,
        attack_type: 'flag_capture',
        source: 'submission',
        submitted_flag: 'flag{demo}',
        is_success: true,
        score_gained: 80,
        created_at: '2026-03-12T10:07:00.000Z',
      },
    ])

    const result = await listContestAWDRoundAttacks('7', '41')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/admin/contests/7/awd/rounds/41/attacks',
    })
    expect(result).toEqual([
      {
        id: '71',
        round_id: '41',
        attacker_team_id: '11',
        attacker_team: 'Red',
        victim_team_id: '12',
        victim_team: 'Blue',
        challenge_id: '101',
        attack_type: 'flag_capture',
        source: 'submission',
        submitted_flag: 'flag{demo}',
        is_success: true,
        score_gained: 80,
        created_at: '2026-03-12T10:07:00.000Z',
      },
    ])
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

  it('应该把管理员挑战列表响应归一化', async () => {
    requestMock.mockResolvedValue({
      list: [
        {
          id: 11,
          title: 'SQL 注入训练',
          description: '基础注入题',
          category: 'web',
          difficulty: 'easy',
          points: 150,
          image_id: 9,
          status: 'draft',
          created_at: '2026-03-10T09:00:00.000Z',
          updated_at: '2026-03-10T09:10:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const result = await getChallenges({ page: 1, page_size: 20 })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/authoring/challenges',
      params: { page: 1, page_size: 20 },
    })
    expect(result.list[0]).toEqual({
      id: '11',
      title: 'SQL 注入训练',
      description: '基础注入题',
      category: 'web',
      difficulty: 'easy',
      points: 150,
      instance_sharing: 'per_user',
      created_by: undefined,
      image_id: '9',
      attachment_url: undefined,
      hints: undefined,
      status: 'draft',
      created_at: '2026-03-10T09:00:00.000Z',
      updated_at: '2026-03-10T09:10:00.000Z',
      flag_config: undefined,
    })
  })

  it('应该把管理员挑战详情和 Flag 配置合并', async () => {
    requestMock
      .mockResolvedValueOnce({
        id: 12,
        title: 'RCE 入门',
        description: '命令执行',
        category: 'web',
        difficulty: 'medium',
        points: 200,
        image_id: 15,
        attachment_url: 'https://example.com/files/rce.zip',
        hints: [
          {
            id: 31,
            level: 1,
            title: '入口提示',
            content: '先观察回显位置',
          },
        ],
        status: 'published',
        created_at: '2026-03-10T10:00:00.000Z',
        updated_at: '2026-03-10T10:05:00.000Z',
      })
      .mockResolvedValueOnce({
        flag_type: 'regex',
        flag_regex: '^flag\\{demo-[0-9]+\\}$',
        flag_prefix: 'flag',
        configured: true,
      })

    const result = await getChallengeDetail('12')

    expect(requestMock).toHaveBeenNthCalledWith(1, {
      method: 'GET',
      url: '/authoring/challenges/12',
    })
    expect(requestMock).toHaveBeenNthCalledWith(2, {
      method: 'GET',
      url: '/authoring/challenges/12/flag',
    })
    expect(result.flag_config).toEqual({
      flag_type: 'regex',
      flag_regex: '^flag\\{demo-[0-9]+\\}$',
      flag_prefix: 'flag',
      configured: true,
    })
    expect(result.attachment_url).toBe('https://example.com/files/rce.zip')
    expect(result.hints).toEqual([
      {
        id: '31',
        level: 1,
        title: '入口提示',
        content: '先观察回显位置',
      },
    ])
  })

  it('应该提交发布检查请求并归一化最新请求状态', async () => {
    requestMock
      .mockResolvedValueOnce({
        id: 41,
        challenge_id: 12,
        requested_by: 7,
        status: 'running',
        request_source: 'admin_publish',
        active: true,
        failure_summary: '',
        started_at: '2026-04-01T08:00:01.000Z',
        created_at: '2026-04-01T08:00:00.000Z',
        updated_at: '2026-04-01T08:00:05.000Z',
      })
      .mockResolvedValueOnce({
        id: 41,
        challenge_id: 12,
        requested_by: 7,
        status: 'failed',
        request_source: 'admin_publish',
        active: false,
        failure_summary: 'Flag 未配置',
        started_at: '2026-04-01T08:00:01.000Z',
        finished_at: '2026-04-01T08:01:00.000Z',
        created_at: '2026-04-01T08:00:00.000Z',
        updated_at: '2026-04-01T08:01:00.000Z',
        result: {
          challenge_id: 12,
          precheck: {
            passed: true,
            started_at: '2026-04-01T08:00:01.000Z',
            ended_at: '2026-04-01T08:00:03.000Z',
            steps: [{ name: 'flag', passed: true, message: 'ok' }],
          },
          runtime: {
            passed: false,
            started_at: '2026-04-01T08:00:03.000Z',
            ended_at: '2026-04-01T08:01:00.000Z',
            access_url: 'http://127.0.0.1:18080',
            container_count: 1,
            network_count: 1,
            steps: [{ name: 'http', passed: false, message: '503' }],
          },
        },
      })

    const created = await createChallengePublishRequest('12')
    const latest = await getLatestChallengePublishRequest('12')

    expect(requestMock).toHaveBeenNthCalledWith(1, {
      method: 'POST',
      url: '/authoring/challenges/12/publish-requests',
    })
    expect(requestMock).toHaveBeenNthCalledWith(2, {
      method: 'GET',
      url: '/authoring/challenges/12/publish-requests/latest',
      suppressErrorToast: true,
    })
    expect(created).toEqual({
      id: '41',
      challenge_id: '12',
      requested_by: '7',
      status: 'running',
      active: true,
      request_source: 'admin_publish',
      failure_summary: '',
      started_at: '2026-04-01T08:00:01.000Z',
      finished_at: undefined,
      published_at: undefined,
      result: undefined,
      created_at: '2026-04-01T08:00:00.000Z',
      updated_at: '2026-04-01T08:00:05.000Z',
    })
    expect(latest).toEqual({
      id: '41',
      challenge_id: '12',
      requested_by: '7',
      status: 'failed',
      active: false,
      request_source: 'admin_publish',
      failure_summary: 'Flag 未配置',
      started_at: '2026-04-01T08:00:01.000Z',
      finished_at: '2026-04-01T08:01:00.000Z',
      published_at: undefined,
      result: {
        challenge_id: '12',
        precheck: {
          passed: true,
          started_at: '2026-04-01T08:00:01.000Z',
          ended_at: '2026-04-01T08:00:03.000Z',
          steps: [{ name: 'flag', passed: true, message: 'ok' }],
        },
        runtime: {
          passed: false,
          started_at: '2026-04-01T08:00:03.000Z',
          ended_at: '2026-04-01T08:01:00.000Z',
          access_url: 'http://127.0.0.1:18080',
          container_count: 1,
          network_count: 1,
          steps: [{ name: 'http', passed: false, message: '503' }],
        },
      },
      created_at: '2026-04-01T08:00:00.000Z',
      updated_at: '2026-04-01T08:01:00.000Z',
    })
  })

  it('应该发送 manual review Flag 配置载荷', async () => {
    requestMock.mockResolvedValue({ message: 'ok' })

    await configureChallengeFlag('12', {
      flag_type: 'manual_review',
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'PUT',
      url: '/authoring/challenges/12/flag',
      data: {
        flag_type: 'manual_review',
      },
    })
  })

  it('应该按后端当前挑战创建契约发送请求并归一化返回值', async () => {
    requestMock.mockResolvedValue({
      id: 21,
      title: '文件包含',
      description: 'LFI 训练',
      category: 'web',
      difficulty: 'hard',
      points: 300,
      image_id: 6,
      status: 'draft',
      created_at: '2026-03-10T11:00:00.000Z',
      updated_at: '2026-03-10T11:00:30.000Z',
    })

    const result = await createChallenge({
      title: '文件包含',
      description: 'LFI 训练',
      category: 'web',
      difficulty: 'hard',
      points: 300,
      image_id: 6,
      attachment_url: 'https://example.com/files/lfi.zip',
      hints: [
        {
          level: 1,
          title: '提示一',
          content: '检查文件包含点',
        },
      ],
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/authoring/challenges',
      data: {
        title: '文件包含',
        description: 'LFI 训练',
        category: 'web',
        difficulty: 'hard',
        points: 300,
        image_id: 6,
        attachment_url: 'https://example.com/files/lfi.zip',
        hints: [
          {
            level: 1,
            title: '提示一',
            content: '检查文件包含点',
          },
        ],
      },
    })
    expect(result.challenge).toEqual({
      id: '21',
      title: '文件包含',
      description: 'LFI 训练',
      category: 'web',
      difficulty: 'hard',
      points: 300,
      instance_sharing: 'per_user',
      created_by: undefined,
      image_id: '6',
      attachment_url: undefined,
      hints: undefined,
      status: 'draft',
      created_at: '2026-03-10T11:00:00.000Z',
      updated_at: '2026-03-10T11:00:30.000Z',
      flag_config: undefined,
    })
  })

  it('应该把镜像列表响应归一化为当前后端状态枚举', async () => {
    requestMock.mockResolvedValue({
      list: [
        {
          id: 5,
          name: 'php-sqli',
          tag: 'latest',
          description: 'SQL 注入环境',
          size: 1048576,
          status: 'available',
          created_at: '2026-03-10T08:00:00.000Z',
          updated_at: '2026-03-10T08:05:00.000Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
    })

    const result = await getImages({ page: 1, page_size: 20 })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/authoring/images',
      params: { page: 1, page_size: 20 },
    })
    expect(result.list[0]).toEqual({
      id: '5',
      name: 'php-sqli',
      tag: 'latest',
      description: 'SQL 注入环境',
      size_bytes: 1048576,
      status: 'available',
      created_at: '2026-03-10T08:00:00.000Z',
      updated_at: '2026-03-10T08:05:00.000Z',
    })
  })

  it('应该把挑战拓扑响应归一化，并在 404 时返回 null', async () => {
    requestMock.mockResolvedValueOnce({
      id: 15,
      challenge_id: 11,
      template_id: 7,
      entry_node_key: 'web',
      networks: [{ key: 'public', name: 'Public', internal: false }],
      nodes: [
        {
          key: 'web',
          name: 'Web',
          image_id: 9,
          service_port: 8080,
          inject_flag: true,
          tier: 'public',
          network_keys: ['public'],
          env: { FLAG: 'flag{demo}' },
        },
      ],
      links: [{ from_node_key: 'web', to_node_key: 'web' }],
      policies: [{ source_node_key: 'web', target_node_key: 'web', action: 'deny' }],
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T01:00:00.000Z',
    })

    const result = await getChallengeTopology('11')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/authoring/challenges/11/topology',
      suppressErrorToast: true,
    })
    expect(result).toEqual({
      id: '15',
      challenge_id: '11',
      template_id: '7',
      entry_node_key: 'web',
      networks: [{ key: 'public', name: 'Public', internal: false }],
      nodes: [
        {
          key: 'web',
          name: 'Web',
          image_id: '9',
          service_port: 8080,
          inject_flag: true,
          tier: 'public',
          network_keys: ['public'],
          env: { FLAG: 'flag{demo}' },
          resources: undefined,
        },
      ],
      links: [{ from_node_key: 'web', to_node_key: 'web' }],
      policies: [
        {
          source_node_key: 'web',
          target_node_key: 'web',
          action: 'deny',
          protocol: undefined,
          ports: undefined,
        },
      ],
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T01:00:00.000Z',
    })

    requestMock.mockRejectedValueOnce(
      Object.assign(new Error('not found'), { name: 'ApiError', status: 404 })
    )
    expect(await getChallengeTopology('12')).toBeNull()
  })

  it('应该把挑战拓扑保存请求透传到后端字段', async () => {
    requestMock.mockResolvedValue({
      id: 18,
      challenge_id: 11,
      entry_node_key: 'web',
      networks: [{ key: 'default', name: '默认网络' }],
      nodes: [{ key: 'web', name: 'Web', service_port: 8080, network_keys: ['default'] }],
      links: [],
      policies: [],
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T02:00:00.000Z',
    })

    await saveChallengeTopology('11', {
      entry_node_key: 'web',
      networks: [{ key: 'default', name: '默认网络' }],
      nodes: [{ key: 'web', name: 'Web', service_port: 8080, network_keys: ['default'] }],
      links: [],
      policies: [],
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'PUT',
      url: '/authoring/challenges/11/topology',
      data: {
        entry_node_key: 'web',
        networks: [{ key: 'default', name: '默认网络' }],
        nodes: [{ key: 'web', name: 'Web', service_port: 8080, network_keys: ['default'] }],
        links: [],
        policies: [],
      },
    })
  })

  it('应该把挑战题解查询与保存请求归一化', async () => {
    requestMock.mockResolvedValueOnce({
      id: 5,
      challenge_id: 11,
      title: '官方题解',
      content: '## Step 1',
      visibility: 'public',
      created_by: 9,
      is_recommended: true,
      recommended_at: '2026-03-10T04:00:00.000Z',
      recommended_by: 9,
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T02:00:00.000Z',
    })

    const detail = await getChallengeWriteup('11')
    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/authoring/challenges/11/writeup',
      suppressErrorToast: true,
    })
    expect(detail).toEqual({
      id: '5',
      challenge_id: '11',
      title: '官方题解',
      content: '## Step 1',
      visibility: 'public',
      created_by: '9',
      is_recommended: true,
      recommended_at: '2026-03-10T04:00:00.000Z',
      recommended_by: '9',
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T02:00:00.000Z',
    })

    requestMock.mockResolvedValueOnce({
      id: 5,
      challenge_id: 11,
      title: '官方题解',
      content: '## Updated',
      visibility: 'public',
      created_by: 9,
      is_recommended: false,
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T03:00:00.000Z',
    })

    await saveChallengeWriteup('11', {
      title: '官方题解',
      content: '## Updated',
      visibility: 'public',
    })

    expect(requestMock).toHaveBeenLastCalledWith({
      method: 'PUT',
      url: '/authoring/challenges/11/writeup',
      data: {
        title: '官方题解',
        content: '## Updated',
        visibility: 'public',
      },
    })
  })

  it('应该透传官方题解推荐与取消推荐请求', async () => {
    requestMock.mockResolvedValueOnce({
      id: 5,
      challenge_id: 11,
      title: '官方题解',
      content: '## Step 1',
      visibility: 'public',
      created_by: 9,
      is_recommended: true,
      recommended_at: '2026-03-10T04:00:00.000Z',
      recommended_by: 9,
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T04:00:00.000Z',
    })

    const recommended = await recommendChallengeWriteup('11')
    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/authoring/challenges/11/writeup/recommend',
    })
    expect(recommended.is_recommended).toBe(true)
    expect(recommended.recommended_by).toBe('9')

    requestMock.mockResolvedValueOnce({
      id: 5,
      challenge_id: 11,
      title: '官方题解',
      content: '## Step 1',
      visibility: 'public',
      created_by: 9,
      is_recommended: false,
      recommended_at: null,
      recommended_by: null,
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T05:00:00.000Z',
    })

    const unrecommended = await unrecommendChallengeWriteup('11')
    expect(requestMock).toHaveBeenLastCalledWith({
      method: 'DELETE',
      url: '/authoring/challenges/11/writeup/recommend',
    })
    expect(unrecommended.is_recommended).toBe(false)
    expect(unrecommended.recommended_by).toBeUndefined()
  })

  it('应该在题解不存在时返回 null，并透传删除请求', async () => {
    requestMock.mockRejectedValueOnce(
      Object.assign(new Error('not found'), { name: 'ApiError', status: 404 })
    )
    expect(await getChallengeWriteup('12')).toBeNull()

    requestMock.mockResolvedValueOnce(undefined)
    await deleteChallengeWriteup('12')
    expect(requestMock).toHaveBeenLastCalledWith({
      method: 'DELETE',
      url: '/authoring/challenges/12/writeup',
      suppressErrorToast: true,
    })
  })

  it('应该在删除拓扑与环境模板时关闭全局错误提示', async () => {
    requestMock.mockResolvedValue(undefined)

    await deleteChallengeTopology('12')
    expect(requestMock).toHaveBeenNthCalledWith(1, {
      method: 'DELETE',
      url: '/authoring/challenges/12/topology',
      suppressErrorToast: true,
    })

    await deleteEnvironmentTemplate('7')
    expect(requestMock).toHaveBeenNthCalledWith(2, {
      method: 'DELETE',
      url: '/authoring/environment-templates/7',
      suppressErrorToast: true,
    })
  })

  it('应该在删除镜像时关闭全局错误提示', async () => {
    requestMock.mockResolvedValue(undefined)

    await deleteImage('9')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'DELETE',
      url: '/authoring/images/9',
      suppressErrorToast: true,
    })
  })

  it('应该把环境模板列表与创建结果归一化', async () => {
    requestMock.mockResolvedValueOnce([
      {
        id: 3,
        name: '双节点模板',
        description: 'web + db',
        entry_node_key: 'web',
        networks: [{ key: 'default', name: '默认网络' }],
        nodes: [{ key: 'web', name: 'Web', image_id: 8, network_keys: ['default'] }],
        links: [],
        policies: [],
        usage_count: 4,
        created_at: '2026-03-10T00:00:00.000Z',
        updated_at: '2026-03-10T02:00:00.000Z',
      },
    ])

    const list = await getEnvironmentTemplates('双节点')
    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/authoring/environment-templates',
      params: { keyword: '双节点' },
    })
    expect(list[0]).toMatchObject({
      id: '3',
      name: '双节点模板',
      usage_count: 4,
      nodes: [{ key: 'web', name: 'Web', image_id: '8', network_keys: ['default'] }],
    })

    requestMock.mockResolvedValueOnce({
      id: 4,
      name: '三层模板',
      description: 'web app db',
      entry_node_key: 'web',
      networks: [{ key: 'default', name: '默认网络' }],
      nodes: [{ key: 'web', name: 'Web', network_keys: ['default'] }],
      links: [],
      policies: [],
      usage_count: 0,
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T03:00:00.000Z',
    })

    await createEnvironmentTemplate({
      name: '三层模板',
      description: 'web app db',
      entry_node_key: 'web',
      networks: [{ key: 'default', name: '默认网络' }],
      nodes: [{ key: 'web', name: 'Web', network_keys: ['default'] }],
      links: [],
      policies: [],
    })

    expect(requestMock).toHaveBeenLastCalledWith({
      method: 'POST',
      url: '/authoring/environment-templates',
      data: {
        name: '三层模板',
        description: 'web app db',
        entry_node_key: 'web',
        networks: [{ key: 'default', name: '默认网络' }],
        nodes: [{ key: 'web', name: 'Web', network_keys: ['default'] }],
        links: [],
        policies: [],
      },
    })
  })
})
