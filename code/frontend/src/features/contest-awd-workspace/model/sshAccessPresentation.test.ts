import { describe, expect, it } from 'vitest'

import { getVSCodeSSHCommand } from './sshAccessPresentation'

describe('sshAccessPresentation', () => {
  it('SSH 凭证展示层应直接使用后端返回的 ssh 命令', () => {
    const command = getVSCodeSSHCommand({
      host: '127.0.0.1',
      port: 2222,
      username: 'student+8+7009',
      password: 'ticket-secret',
      command: 'ssh student+8+7009@127.0.0.1 -p 2222',
      expires_at: '2026-04-12T08:15:00Z',
    })

    expect(command).toBe('ssh student+8+7009@127.0.0.1 -p 2222')
  })
})
