import { beforeEach, describe, expect, it, vi } from 'vitest'
import {
  configureChallengeFlag,
  createChallenge,
  createChallengePublishRequest,
  createEnvironmentTemplate,
  deleteChallengeTopology,
  deleteChallengeWriteup,
  deleteEnvironmentTemplate,
  deleteImage,
  getChallengeDetail,
  getChallengeTopology,
  getChallengeWriteup,
  getChallenges,
  getEnvironmentTemplates,
  getImages,
  getLatestChallengePublishRequest,
  listChallengeImports,
  recommendChallengeWriteup,
  saveChallengeTopology,
  saveChallengeWriteup,
  unrecommendChallengeWriteup,
} from '@/api/admin/authoring'

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

describe('admin authoring api contract', () => {
  beforeEach(() => {
    requestMock.mockReset()
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
    })
  })

  it('应该透传删除拓扑与环境模板请求', async () => {
    requestMock.mockResolvedValue(undefined)

    await deleteChallengeTopology('12')
    expect(requestMock).toHaveBeenNthCalledWith(1, {
      method: 'DELETE',
      url: '/authoring/challenges/12/topology',
    })

    await deleteEnvironmentTemplate('7')
    expect(requestMock).toHaveBeenNthCalledWith(2, {
      method: 'DELETE',
      url: '/authoring/environment-templates/7',
    })
  })

  it('应该透传删除镜像请求', async () => {
    requestMock.mockResolvedValue(undefined)

    await deleteImage('9')

    expect(requestMock).toHaveBeenCalledWith({
      method: 'DELETE',
      url: '/authoring/images/9',
    })
  })

  it('应该把环境模板列表与创建结果归一化', async () => {
    requestMock.mockResolvedValueOnce([
      {
        id: 3,
        name: '双节点AWD 题目',
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
      name: '双节点AWD 题目',
      usage_count: 4,
      nodes: [{ key: 'web', name: 'Web', image_id: '8', network_keys: ['default'] }],
    })

    requestMock.mockResolvedValueOnce({
      id: 4,
      name: '三层AWD 题目',
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
      name: '三层AWD 题目',
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
        name: '三层AWD 题目',
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
