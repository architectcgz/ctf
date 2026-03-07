import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import CheatDetection from '../CheatDetection.vue'

const pushMock = vi.fn()

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
  }
})

describe('CheatDetection', () => {
  beforeEach(() => {
    pushMock.mockReset()
  })

  it('应该提示当前能力边界并支持跳转到审计日志', async () => {
    const wrapper = mount(CheatDetection)

    expect(wrapper.text()).toContain('尚未接入独立的作弊检测 API')

    const quickAction = wrapper.findAll('button').find((button) => button.text().includes('查看提交记录'))
    expect(quickAction).toBeTruthy()

    await quickAction!.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'AuditLog',
      query: { action: 'submit' },
    })
  })
})
