import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import ChallengeWriteupManagePanel from '@/components/admin/writeup/ChallengeWriteupManagePanel.vue'

const pushMock = vi.fn()

const adminApiMocks = vi.hoisted(() => ({
  getChallengeWriteup: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
  }
})

vi.mock('@/api/admin', () => adminApiMocks)

describe('ChallengeWriteupManagePanel', () => {
  it('存在题解时应显示目录行并支持进入独立编辑页', async () => {
    adminApiMocks.getChallengeWriteup.mockResolvedValue({
      id: '21',
      challenge_id: '11',
      title: '官方题解',
      content: '## Step 1',
      visibility: 'public',
      is_recommended: true,
      recommended_at: '2026-03-10T03:00:00.000Z',
      created_at: '2026-03-10T00:00:00.000Z',
      updated_at: '2026-03-10T02:00:00.000Z',
    })

    const wrapper = mount(ChallengeWriteupManagePanel, {
      props: {
        challengeId: '11',
        challengeTitle: '双节点演练',
      },
      global: {
        stubs: {
          AppEmpty: { template: '<div><slot /></div>' },
          AppLoading: { template: '<div><slot /></div>' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('官方题解')
    expect(wrapper.text()).toContain('public')
    expect(wrapper.text()).toContain('推荐题解')

    const editButton = wrapper.findAll('button').find((button) => button.text().includes('查看 / 编辑'))
    expect(editButton).toBeTruthy()

    await editButton!.trigger('click')

    expect(pushMock).toHaveBeenCalledWith('/platform/challenges/11/writeup')
  })

  it('没有题解时应显示空状态并支持编写题解', async () => {
    adminApiMocks.getChallengeWriteup.mockResolvedValue(null)

    const wrapper = mount(ChallengeWriteupManagePanel, {
      props: {
        challengeId: '11',
        challengeTitle: '双节点演练',
      },
      global: {
        stubs: {
          AppEmpty: {
            template:
              '<div><div>{{ title }}</div><div>{{ description }}</div><slot /></div>',
            props: ['title', 'description'],
          },
          AppLoading: { template: '<div><slot /></div>' },
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('当前还没有题解')
    expect(wrapper.text()).toContain('双节点演练')

    const createButton = wrapper.findAll('button').find((button) => button.text().includes('编写题解'))
    expect(createButton).toBeTruthy()

    await createButton!.trigger('click')

    expect(pushMock).toHaveBeenCalledWith('/platform/challenges/11/writeup')
  })
})
