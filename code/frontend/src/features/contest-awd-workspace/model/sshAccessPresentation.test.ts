import { describe, expect, it } from 'vitest'

import { buildOpenSSHConfig, getVSCodeSSHCommand } from './sshAccessPresentation'

describe('sshAccessPresentation', () => {
  it('为 OpenSSH 配置生成 Host 块', () => {
    expect(
      buildOpenSSHConfig({
        alias: 'ctf-awd-8-7009',
        host_name: '127.0.0.1',
        port: 2222,
        user: 'student+8+7009',
      })
    ).toBe('Host ctf-awd-8-7009\n  HostName 127.0.0.1\n  Port 2222\n  User student+8+7009\n')
  })

  it('VS Code 新增 SSH 主机应复制 ssh 命令而不是 Host 配置块', () => {
    const command = getVSCodeSSHCommand({
      host: '127.0.0.1',
      port: 2222,
      username: 'student+8+7009',
      password: 'ticket-secret',
      command: 'ssh student+8+7009@127.0.0.1 -p 2222',
      ssh_profile: {
        alias: 'ctf-awd-8-7009',
        host_name: '127.0.0.1',
        port: 2222,
        user: 'student+8+7009',
      },
      expires_at: '2026-04-12T08:15:00Z',
    })

    expect(command).toBe('ssh student+8+7009@127.0.0.1 -p 2222')
    expect(command).not.toMatch(/^Host\s+/)
  })
})
