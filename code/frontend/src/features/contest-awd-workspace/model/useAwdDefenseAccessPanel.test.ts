import { computed, ref } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { AWDDefenseSSHAccessData, ContestAWDWorkspaceServiceData } from '@/api/contracts'
import { useAwdDefenseAccessPanel } from './useAwdDefenseAccessPanel'

const toastMocks = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
}))

vi.mock('@/composables/useToast', () => ({
  useToast: () => toastMocks,
}))

describe('useAwdDefenseAccessPanel', () => {
  beforeEach(() => {
    toastMocks.success.mockReset()
    toastMocks.error.mockReset()
    Object.defineProperty(globalThis.navigator, 'clipboard', {
      configurable: true,
      value: {
        writeText: vi.fn().mockResolvedValue(undefined),
      },
    })
  })

  it('应暴露当前选中服务的 access 与复制态', async () => {
    const selectedServiceId = ref('7009')
    const servicesByServiceId = ref(
      new Map<string, ContestAWDWorkspaceServiceData>([
        [
          '7009',
          {
            service_id: '7009',
            awd_challenge_id: '101',
            instance_id: '900',
            instance_status: 'running',
            service_status: 'up',
            attack_received: 0,
            sla_score: 1,
            defense_score: 2,
            attack_score: 0,
          },
        ],
      ])
    )
    const sshAccessByServiceId = ref<Record<string, AWDDefenseSSHAccessData>>({
      '7009': {
        host: '127.0.0.1',
        port: 2222,
        username: 'student+1+7009',
        password: 'ticket-secret',
        command: 'ssh student+1+7009@127.0.0.1 -p 2222',
        expires_at: '2026-04-12T08:15:00Z',
      },
    })

    const state = useAwdDefenseAccessPanel({
      selectedServiceId: computed(() => selectedServiceId.value),
      servicesByServiceId: computed(() => servicesByServiceId.value),
      sshAccessByServiceId: computed(() => sshAccessByServiceId.value),
      openService: vi.fn(),
    })

    await state.copySSHCommand('7009')

    expect(state.selectedDefenseAccess.value?.password).toBe('ticket-secret')
    expect(state.selectedDefenseCopiedCommand.value).toBe(true)
    expect(state.getSSHCommand('7009')).toBe('ssh student+1+7009@127.0.0.1 -p 2222')
  })

  it('打开本队服务时应桥接实例 access owner', () => {
    const openService = vi.fn()
    const state = useAwdDefenseAccessPanel({
      selectedServiceId: computed(() => '7009'),
      servicesByServiceId: computed(
        () =>
          new Map<string, ContestAWDWorkspaceServiceData>([
            [
              '7009',
              {
                service_id: '7009',
                awd_challenge_id: '101',
                instance_id: '900',
                instance_status: 'running',
                service_status: 'up',
                attack_received: 0,
                sla_score: 1,
                defense_score: 2,
                attack_score: 0,
              },
            ],
          ])
      ),
      sshAccessByServiceId: computed(() => ({})),
      openService,
    })

    state.openDefenseService('7009')

    expect(openService).toHaveBeenCalledWith('900')
  })

  it('复制 SSH 密码失败时应提示手动复制', async () => {
    Object.defineProperty(globalThis.navigator, 'clipboard', {
      configurable: true,
      value: undefined,
    })

    const state = useAwdDefenseAccessPanel({
      selectedServiceId: computed(() => '7009'),
      servicesByServiceId: computed(() => new Map<string, ContestAWDWorkspaceServiceData>()),
      sshAccessByServiceId: computed(
        () => ({
          '7009': {
            host: '127.0.0.1',
            port: 2222,
            username: 'student+1+7009',
            password: 'ticket-secret',
            command: 'ssh student+1+7009@127.0.0.1 -p 2222',
            expires_at: '2026-04-12T08:15:00Z',
          },
        })
      ),
      openService: vi.fn(),
    })

    await state.copySSHPassword('7009')

    expect(toastMocks.error).toHaveBeenCalledWith('复制失败，请手动选择文本')
    expect(state.selectedDefenseCopiedPassword.value).toBe(false)
  })
})
