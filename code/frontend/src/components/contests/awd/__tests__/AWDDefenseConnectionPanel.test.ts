import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import type { AWDDefenseSSHAccessData } from '@/api/contracts'
import AWDDefenseConnectionPanel from '../AWDDefenseConnectionPanel.vue'

function access(overrides: Partial<AWDDefenseSSHAccessData> = {}): AWDDefenseSSHAccessData {
  return {
    host: '127.0.0.1',
    port: 2222,
    username: 'student+8+21',
    password: 'ticket-secret',
    command: 'ssh student+8+21@127.0.0.1 -p 2222',
    expires_at: '2026-04-12T08:15:00Z',
    ...overrides,
  }
}

describe('AWDDefenseConnectionPanel', () => {
  it('展示 SSH 凭证、命令和票据过期时间', () => {
    const wrapper = mount(AWDDefenseConnectionPanel, {
      props: {
        access: access(),
        serviceId: '21',
        copiedCommand: false,
      },
    })

    expect(wrapper.text()).toContain('SSH 连接')
    expect(wrapper.text()).toContain('ssh student+8+21@127.0.0.1 -p 2222')
    expect(wrapper.text()).toContain('127.0.0.1')
    expect(wrapper.text()).toContain('2222')
    expect(wrapper.text()).toContain('student+8+21')
    expect(wrapper.text()).toContain('ticket-secret')
    expect(wrapper.text()).toContain('票据将在')
    expect(wrapper.text()).toContain('过期')
  })

  it('向父组件上抛复制命令事件', async () => {
    const wrapper = mount(AWDDefenseConnectionPanel, {
      props: {
        access: access(),
        serviceId: '21',
        copiedCommand: false,
      },
    })

    await wrapper.find('button').trigger('click')

    expect(wrapper.emitted('copyCommand')).toEqual([['21']])
  })
})
