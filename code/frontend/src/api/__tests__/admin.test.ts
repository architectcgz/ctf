import { beforeEach, describe, expect, it, vi } from 'vitest'

const requestMock = vi.hoisted(() => vi.fn())

vi.mock('@/api/request', () => ({
  request: requestMock,
}))

import {
  createChallenge,
  createContest,
  getChallengeDetail,
  getChallenges,
  getCheatDetection,
  getContests,
  getImages,
  getUsers,
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
      image_id: '6',
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
        image_id: '6',
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
})
