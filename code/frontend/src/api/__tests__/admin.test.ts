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
  createEnvironmentTemplate,
  createChallenge,
  createContest,
  deleteChallengeWriteup,
  getAdminContestLiveScoreboard,
  getChallengeTopology,
  getChallengeDetail,
  getContestAWDRoundSummary,
  getChallengeWriteup,
  getChallenges,
  getCheatDetection,
  getContests,
  getEnvironmentTemplates,
  getImages,
  getUsers,
  listAdminContestChallenges,
  listContestAWDRoundAttacks,
  runContestAWDRoundCheck,
  saveChallengeTopology,
  saveChallengeWriteup,
  updateContest,
} from '@/api/admin'

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
        created_at: '2026-03-12T00:00:00.000Z',
      },
    ])
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
          attack_received: 0,
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
          attack_received: 0,
          defense_score: 45,
          attack_score: 0,
          updated_at: '2026-03-12T10:06:00.000Z',
        },
      ],
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
      items: [],
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
      items: [],
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
        size: 20,
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

    const result = await getChallenges({ page: 1, size: 20 })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/admin/challenges',
      params: { page: 1, size: 20 },
    })
    expect(result.list[0]).toEqual({
      id: '11',
      title: 'SQL 注入训练',
      description: '基础注入题',
      category: 'web',
      difficulty: 'easy',
      points: 150,
      image_id: '9',
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
            cost_points: 0,
            content: '先观察回显位置',
          },
        ],
        status: 'published',
        created_at: '2026-03-10T10:00:00.000Z',
        updated_at: '2026-03-10T10:05:00.000Z',
      })
      .mockResolvedValueOnce({
        flag_type: 'dynamic',
        flag_prefix: 'ctf',
        configured: true,
      })

    const result = await getChallengeDetail('12')

    expect(requestMock).toHaveBeenNthCalledWith(1, {
      method: 'GET',
      url: '/admin/challenges/12',
    })
    expect(requestMock).toHaveBeenNthCalledWith(2, {
      method: 'GET',
      url: '/admin/challenges/12/flag',
    })
    expect(result.flag_config).toEqual({
      flag_type: 'dynamic',
      flag_prefix: 'ctf',
      configured: true,
    })
    expect(result.attachment_url).toBe('https://example.com/files/rce.zip')
    expect(result.hints).toEqual([
      {
        id: '31',
        level: 1,
        title: '入口提示',
        cost_points: 0,
        content: '先观察回显位置',
      },
    ])
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
          cost_points: 0,
          content: '检查文件包含点',
        },
      ],
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'POST',
      url: '/admin/challenges',
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
            cost_points: 0,
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
      image_id: '6',
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

    const result = await getImages({ page: 1, size: 20 })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/admin/images',
      params: { page: 1, size: 20 },
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
      url: '/admin/challenges/11/topology',
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
      url: '/admin/challenges/11/topology',
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
      visibility: 'scheduled',
      release_at: '2026-03-12T08:00:00.000Z',
      created_by: 9,
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T02:00:00.000Z',
    })

    const detail = await getChallengeWriteup('11')
    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/admin/challenges/11/writeup',
      suppressErrorToast: true,
    })
    expect(detail).toEqual({
      id: '5',
      challenge_id: '11',
      title: '官方题解',
      content: '## Step 1',
      visibility: 'scheduled',
      release_at: '2026-03-12T08:00:00.000Z',
      created_by: '9',
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T02:00:00.000Z',
    })

    requestMock.mockResolvedValueOnce({
      id: 5,
      challenge_id: 11,
      title: '官方题解',
      content: '## Updated',
      visibility: 'public',
      release_at: null,
      created_by: 9,
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
      url: '/admin/challenges/11/writeup',
      data: {
        title: '官方题解',
        content: '## Updated',
        visibility: 'public',
      },
    })
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
      url: '/admin/challenges/12/writeup',
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
      url: '/admin/environment-templates',
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
      url: '/admin/environment-templates',
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
