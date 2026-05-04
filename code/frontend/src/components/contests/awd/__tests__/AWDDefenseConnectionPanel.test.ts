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
    ssh_profile: {
      alias: 'ctf-awd-8-21',
      host_name: '127.0.0.1',
      port: 2222,
      user: 'student+8+21',
    },
    expires_at: '2026-04-12T08:15:00Z',
    ...overrides,
  }
}

describe('AWDDefenseConnectionPanel', () => {
  it('分开展示 VS Code 命令、OpenSSH 配置和票据过期时间', () => {
    const wrapper = mount(AWDDefenseConnectionPanel, {
      props: {
        access: access(),
        serviceId: '21',
        copiedCommand: false,
        copiedConfig: false,
      },
    })

    expect(wrapper.text()).toContain('VS Code Remote-SSH')
    expect(wrapper.text()).toContain('ssh student+8+21@127.0.0.1 -p 2222')
    expect(wrapper.text()).toContain('OpenSSH 配置')
    expect(wrapper.text()).toContain('Host ctf-awd-8-21')
    expect(wrapper.text()).toContain('票据将在')
    expect(wrapper.text()).toContain('过期')
  })

  it('向父组件上抛复制命令和复制配置事件', async () => {
    const wrapper = mount(AWDDefenseConnectionPanel, {
      props: {
        access: access(),
        serviceId: '21',
        copiedCommand: false,
        copiedConfig: false,
      },
    })

    await wrapper.findAll('button')[0].trigger('click')
    await wrapper.find('summary').trigger('click')
    await wrapper.findAll('button')[1].trigger('click')

    expect(wrapper.emitted('copyCommand')).toEqual([['21']])
    expect(wrapper.emitted('copyConfig')).toEqual([['21']])
  })
})
