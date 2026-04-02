import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import ChallengeManage from '../ChallengeManage.vue'

vi.mock('@/api/admin', () => ({
  getChallenges: vi.fn().mockResolvedValue({
    list: [
      {
        id: '1',
        title: 'Test Challenge',
        category: 'web',
        difficulty: 'easy',
        status: 'draft',
        points: 100,
        created_at: '2024-01-01T00:00:00Z',
      },
    ],
    total: 1,
    page: 1,
    page_size: 20,
  }),
  getImages: vi.fn().mockResolvedValue({
    list: [
      {
        id: 'img-1',
        name: 'php-sqli',
        tag: 'latest',
        status: 'available',
        created_at: '2024-01-01T00:00:00Z',
      },
    ],
    total: 1,
    page: 1,
    page_size: 20,
  }),
  getChallengeDetail: vi.fn().mockResolvedValue({
    id: '1',
    title: 'Test Challenge',
    description: 'Test description',
    category: 'web',
    difficulty: 'easy',
    status: 'draft',
    points: 100,
    image_id: 'img-1',
    attachment_url: 'https://example.com/files/test.zip',
    hints: [
      {
        id: 'hint-1',
        level: 1,
        title: '入口提示',
        cost_points: 0,
        content: '从登录页开始排查',
      },
    ],
    created_at: '2024-01-01T00:00:00Z',
  }),
  previewChallengeImport: vi.fn(),
  getChallengeImport: vi.fn(),
  commitChallengeImport: vi.fn(),
  createChallenge: vi.fn(),
  updateChallenge: vi.fn(),
  configureChallengeFlag: vi.fn(),
  getChallengeFlagConfig: vi.fn().mockResolvedValue({
    flag_type: 'static',
    configured: false,
  }),
  createChallengePublishRequest: vi.fn(),
  getLatestChallengePublishRequest: vi.fn().mockResolvedValue({
    id: 'req-1',
    challenge_id: '1',
    status: 'failed',
    active: false,
    failure_summary: 'Flag 未配置',
    created_at: '2026-04-01T08:00:00Z',
    updated_at: '2026-04-01T08:01:00Z',
  }),
  deleteChallenge: vi.fn(),
}))

describe('ChallengeManage', () => {
  it('应该渲染挑战管理页面', async () => {
    const wrapper = mount(ChallengeManage, {
      global: {
        stubs: {
          ElTable: true,
          ElTableColumn: true,
          ElButton: true,
          ElPagination: true,
          ElDialog: true,
          ElForm: true,
          ElFormItem: true,
          ElInput: true,
          ElInputNumber: true,
          ElSelect: true,
          ElOption: true,
        },
      },
    })

    await wrapper.vm.$nextTick()
    await new Promise((resolve) => setTimeout(resolve, 100))

    expect(wrapper.element.tagName).toBe('SECTION')
    expect(wrapper.classes()).toContain('journal-shell')
    expect(wrapper.classes()).toContain('journal-hero')
    expect(wrapper.classes()).toContain('min-h-full')
    expect(wrapper.text()).toContain('挑战管理')
    expect(wrapper.text()).toContain('导入题目包')
    expect(wrapper.text()).toContain('检查失败')
    expect(wrapper.text()).toContain('Flag 未配置')
    expect(wrapper.text()).not.toContain('创建挑战')
    expect(wrapper.findAll('button').some((button) => button.text() === '提交发布检查')).toBe(true)
    expect(wrapper.findAll('button').some((button) => button.text() === '发布')).toBe(false)
  })
})
