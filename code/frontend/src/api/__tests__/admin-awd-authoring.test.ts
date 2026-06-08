import { beforeEach, describe, expect, it, vi } from 'vitest'
import {
  commitAdminAwdChallengeImport,
  createAdminAwdChallenge,
  deleteAdminAwdChallenge,
  listAdminAwdChallengeImports,
  listAdminAwdChallenges,
  previewAdminAwdChallengeImport,
  updateAdminAwdChallenge,
} from '@/api/admin/awd-authoring'

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

describe('admin AWD authoring api contract', () => {
  beforeEach(() => {
    requestMock.mockReset()
  })

  it('应该把 AWD 题目分页结果归一化', async () => {
    requestMock.mockResolvedValueOnce({
      items: [
        {
          id: 5,
          name: 'Bank Portal AWD',
          slug: 'bank-portal-awd',
          category: 'web',
          difficulty: 'hard',
          description: 'multi-step banking target',
          service_type: 'web_http',
          deployment_mode: 'single_container',
          version: 'v1',
          status: 'draft',
          readiness_status: 'pending',
          created_by: 9,
          last_verified_at: null,
          created_at: '2026-04-17T08:00:00.000Z',
          updated_at: '2026-04-17T09:00:00.000Z',
        },
      ],
      total: 1,
      page: 2,
      size: 10,
    })

    const page = await listAdminAwdChallenges({
      page: 2,
      page_size: 10,
      keyword: 'bank',
      service_type: 'web_http',
      status: 'draft',
    })

    expect(requestMock).toHaveBeenCalledWith({
      method: 'GET',
      url: '/authoring/awd-challenges',
      params: {
        page: 2,
        page_size: 10,
        keyword: 'bank',
        service_type: 'web_http',
        status: 'draft',
      },
    })
    expect(page).toEqual({
      list: [
        {
          id: '5',
          name: 'Bank Portal AWD',
          slug: 'bank-portal-awd',
          category: 'web',
          difficulty: 'hard',
          description: 'multi-step banking target',
          service_type: 'web_http',
          deployment_mode: 'single_container',
          version: 'v1',
          status: 'draft',
          readiness_status: 'pending',
          created_by: '9',
          last_verified_at: undefined,
          created_at: '2026-04-17T08:00:00.000Z',
          updated_at: '2026-04-17T09:00:00.000Z',
        },
      ],
      total: 1,
      page: 2,
      page_size: 10,
    })
  })

  it('应该把 AWD 题目创建更新删除请求转换成后台接口格式', async () => {
    requestMock.mockResolvedValueOnce({
      id: 5,
      name: 'Bank Portal AWD',
      slug: 'bank-portal-awd',
      category: 'web',
      difficulty: 'hard',
      description: 'desc',
      service_type: 'web_http',
      deployment_mode: 'single_container',
      version: 'v1',
      status: 'draft',
      readiness_status: 'pending',
      created_at: '2026-04-17T08:00:00.000Z',
      updated_at: '2026-04-17T09:00:00.000Z',
    })

    await createAdminAwdChallenge({
      name: 'Bank Portal AWD',
      slug: 'bank-portal-awd',
      category: 'web',
      difficulty: 'hard',
      description: 'desc',
      service_type: 'web_http',
      deployment_mode: 'single_container',
    })

    expect(requestMock).toHaveBeenNthCalledWith(1, {
      method: 'POST',
      url: '/authoring/awd-challenges',
      data: {
        name: 'Bank Portal AWD',
        slug: 'bank-portal-awd',
        category: 'web',
        difficulty: 'hard',
        description: 'desc',
        service_type: 'web_http',
        deployment_mode: 'single_container',
      },
    })

    requestMock.mockResolvedValueOnce({
      id: 5,
      name: 'Bank Portal AWD v2',
      slug: 'bank-portal-awd',
      category: 'web',
      difficulty: 'hard',
      description: 'desc',
      service_type: 'web_http',
      deployment_mode: 'single_container',
      version: 'v1',
      status: 'published',
      readiness_status: 'passed',
      created_at: '2026-04-17T08:00:00.000Z',
      updated_at: '2026-04-17T10:00:00.000Z',
    })

    await updateAdminAwdChallenge('5', {
      name: 'Bank Portal AWD v2',
      status: 'published',
    })

    expect(requestMock).toHaveBeenNthCalledWith(2, {
      method: 'PUT',
      url: '/authoring/awd-challenges/5',
      data: {
        name: 'Bank Portal AWD v2',
        status: 'published',
      },
    })

    requestMock.mockResolvedValueOnce(undefined)
    await deleteAdminAwdChallenge('5')

    expect(requestMock).toHaveBeenNthCalledWith(3, {
      method: 'DELETE',
      url: '/authoring/awd-challenges/5',
    })
  })

  it('应该把 AWD 题目包导入预览与确认接口转换成后台格式', async () => {
    const file = new File(['zip'], 'awd-bank-portal-01.zip', { type: 'application/zip' })

    requestMock.mockResolvedValueOnce({
      id: 'imp-1',
      file_name: 'awd-bank-portal-01.zip',
      slug: 'awd-bank-portal-01',
      title: 'Bank Portal AWD',
      category: 'web',
      difficulty: 'hard',
      description: 'multi-step banking target',
      service_type: 'web_http',
      deployment_mode: 'single_container',
      version: 'v2026.04',
      checker_type: 'http_standard',
      checker_config: {
        put_flag: { method: 'PUT', path: '/api/flag' },
      },
      flag_mode: 'dynamic_team',
      flag_config: { flag_prefix: 'awd' },
      defense_entry_mode: 'http',
      access_config: { service_port: 8080 },
      runtime_config: { image_ref: 'registry.example.edu/ctf/awd-bank-portal:v1' },
      warnings: ['meta.points 仅作为建议分值，不会直接写入AWD 题目。'],
      created_at: '2026-04-21T08:00:00.000Z',
    })

    const preview = await previewAdminAwdChallengeImport(file)

    expect(requestMock).toHaveBeenNthCalledWith(1, {
      method: 'POST',
      url: '/authoring/awd-challenge-imports',
      data: expect.any(FormData),
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    expect(preview).toEqual({
      id: 'imp-1',
      file_name: 'awd-bank-portal-01.zip',
      slug: 'awd-bank-portal-01',
      title: 'Bank Portal AWD',
      category: 'web',
      difficulty: 'hard',
      description: 'multi-step banking target',
      service_type: 'web_http',
      deployment_mode: 'single_container',
      version: 'v2026.04',
      checker_type: 'http_standard',
      checker_config: {
        put_flag: { method: 'PUT', path: '/api/flag' },
      },
      flag_mode: 'dynamic_team',
      flag_config: { flag_prefix: 'awd' },
      defense_entry_mode: 'http',
      access_config: { service_port: 8080 },
      runtime_config: { image_ref: 'registry.example.edu/ctf/awd-bank-portal:v1' },
      warnings: ['meta.points 仅作为建议分值，不会直接写入AWD 题目。'],
      created_at: '2026-04-21T08:00:00.000Z',
    })

    requestMock.mockResolvedValueOnce([
      {
        id: 'imp-1',
        file_name: 'awd-bank-portal-01.zip',
        slug: 'awd-bank-portal-01',
        title: 'Bank Portal AWD',
        category: 'web',
        difficulty: 'hard',
        description: 'multi-step banking target',
        service_type: 'web_http',
        deployment_mode: 'single_container',
        version: 'v2026.04',
        checker_type: 'http_standard',
        checker_config: {},
        flag_mode: 'dynamic_team',
        flag_config: {},
        defense_entry_mode: 'http',
        access_config: {},
        runtime_config: {},
        warnings: [],
        created_at: '2026-04-21T08:00:00.000Z',
      },
    ])

    await listAdminAwdChallengeImports()

    expect(requestMock).toHaveBeenNthCalledWith(2, {
      method: 'GET',
      url: '/authoring/awd-challenge-imports',
    })

    requestMock.mockResolvedValueOnce({
      challenge: {
        id: 5,
        name: 'Bank Portal AWD',
        slug: 'awd-bank-portal-01',
        category: 'web',
        difficulty: 'hard',
        description: 'multi-step banking target',
        service_type: 'web_http',
        deployment_mode: 'single_container',
        version: 'v2026.04',
        status: 'published',
        readiness_status: 'pending',
        created_at: '2026-04-21T08:00:00.000Z',
        updated_at: '2026-04-21T08:05:00.000Z',
      },
    })

    await commitAdminAwdChallengeImport('imp-1')

    expect(requestMock).toHaveBeenNthCalledWith(3, {
      method: 'POST',
      url: '/authoring/awd-challenge-imports/imp-1/commit',
    })
  })
})
