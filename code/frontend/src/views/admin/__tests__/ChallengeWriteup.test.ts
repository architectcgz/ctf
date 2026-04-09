import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import ChallengeWriteupEditorPage from '@/components/admin/writeup/ChallengeWriteupEditorPage.vue'

const adminApiMocks = vi.hoisted(() => ({
  getChallengeDetail: vi.fn().mockResolvedValue({
    id: '11',
    title: '双节点演练',
    category: 'web',
    difficulty: 'easy',
    status: 'draft',
    points: 100,
    created_at: '2026-03-10T00:00:00.000Z',
  }),
  getChallengeWriteup: vi.fn().mockResolvedValue({
    id: '21',
    challenge_id: '11',
    title: '官方题解',
    content: '## Step 1',
    visibility: 'public',
    is_recommended: true,
    recommended_at: '2026-03-10T03:00:00.000Z',
    created_at: '2026-03-10T00:00:00.000Z',
    updated_at: '2026-03-10T02:00:00.000Z',
  }),
  saveChallengeWriteup: vi.fn(),
  deleteChallengeWriteup: vi.fn(),
  recommendChallengeWriteup: vi.fn().mockResolvedValue({
    id: '21',
    challenge_id: '11',
    title: '官方题解',
    content: '## Step 1',
    visibility: 'public',
    is_recommended: true,
    recommended_at: '2026-03-10T03:00:00.000Z',
    created_at: '2026-03-10T00:00:00.000Z',
    updated_at: '2026-03-10T03:00:00.000Z',
  }),
  unrecommendChallengeWriteup: vi.fn().mockResolvedValue({
    id: '21',
    challenge_id: '11',
    title: '官方题解',
    content: '## Step 1',
    visibility: 'public',
    is_recommended: false,
    created_at: '2026-03-10T00:00:00.000Z',
    updated_at: '2026-03-10T04:00:00.000Z',
  }),
}))

vi.mock('@/api/admin', () => adminApiMocks)

describe('ChallengeWriteupEditorPage', () => {
  it('应该渲染已保存题解的核心信息', async () => {
    const wrapper = mount(ChallengeWriteupEditorPage, {
      props: {
        challengeId: '11',
      },
      global: {
        stubs: {
          AppCard: { template: '<section><slot name="header" /><slot /></section>' },
          AppEmpty: { template: '<div><slot /></div>' },
          AppLoading: { template: '<div><slot /></div>' },
          PageHeader: { template: '<div><slot /></div>' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('双节点演练')
    expect(wrapper.text()).toContain('官方题解')
    expect(wrapper.text()).toContain('public')
    expect(wrapper.text()).not.toContain('scheduled')
    expect(wrapper.text()).toContain('推荐题解')
    expect(wrapper.text()).toContain('取消推荐')
    expect(wrapper.text()).toContain('恢复已保存版本')
    expect(wrapper.text()).toContain('2026-03-10T02:00:00.000Z')
  })

  it('应该支持取消推荐官方题解', async () => {
    const wrapper = mount(ChallengeWriteupEditorPage, {
      props: {
        challengeId: '11',
      },
      global: {
        stubs: {
          AppCard: { template: '<section><slot name="header" /><slot /></section>' },
          AppEmpty: { template: '<div><slot /></div>' },
          AppLoading: { template: '<div><slot /></div>' },
          PageHeader: { template: '<div><slot /></div>' },
        },
      },
    })

    await flushPromises()

    const actionButton = wrapper.findAll('button').find((button) => button.text().includes('取消推荐'))
    expect(actionButton).toBeTruthy()

    await actionButton!.trigger('click')
    await flushPromises()

    expect(adminApiMocks.unrecommendChallengeWriteup).toHaveBeenCalledWith('11')
  })
})
