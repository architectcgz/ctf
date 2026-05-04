import { describe, expect, it } from 'vitest'

import type {
  ContestAWDWorkspaceServiceData,
  ContestChallengeItem,
} from '@/api/contracts'
import { toDefenseServiceCards } from './awdDefensePresentation'

function challenge(
  id: string,
  title: string,
  serviceId = `service-${id}`
): ContestChallengeItem {
  return {
    id,
    challenge_id: `challenge-${id}`,
    awd_challenge_id: `awd-${id}`,
    awd_service_id: serviceId,
    title,
    category: 'web',
    difficulty: 'medium',
    points: 100,
    solved_count: 0,
    is_solved: false,
  }
}

function service(
  serviceId: string,
  overrides: Partial<ContestAWDWorkspaceServiceData>
): ContestAWDWorkspaceServiceData {
  return {
    service_id: serviceId,
    awd_challenge_id: `awd-${serviceId}`,
    instance_id: `instance-${serviceId}`,
    instance_status: 'running',
    service_status: 'up',
    attack_received: 0,
    sla_score: 100,
    defense_score: 100,
    attack_score: 0,
    ...overrides,
  }
}

describe('awdDefensePresentation', () => {
  it('按防守风险排序服务卡', () => {
    const challenges = [
      challenge('stable', 'Stable'),
      challenge('attacked', 'Attacked'),
      challenge('down', 'Down'),
      challenge('compromised', 'Compromised'),
      challenge('creating', 'Creating'),
    ]
    const services = [
      service('service-stable', { service_status: 'up' }),
      service('service-attacked', { attack_received: 3 }),
      service('service-down', { service_status: 'down', instance_status: 'failed' }),
      service('service-compromised', { service_status: 'compromised', attack_received: 1 }),
      service('service-creating', { instance_status: 'creating' }),
    ]

    expect(
      toDefenseServiceCards({ challenges, services }).map((card) => card.title)
    ).toEqual(['Compromised', 'Down', 'Attacked', 'Creating', 'Stable'])
  })

  it('生成服务卡状态、操作能力和风险原因', () => {
    const [card] = toDefenseServiceCards({
      challenges: [challenge('web', 'Web')],
      services: [
        service('service-web', {
          service_status: 'compromised',
          attack_received: 2,
        }),
      ],
    })

    expect(card).toMatchObject({
      serviceId: 'service-web',
      challengeId: 'awd-web',
      title: 'Web',
      riskLevel: 'critical',
      riskReasons: ['服务已失陷', '检测到 2 次攻击'],
      serviceStatusLabel: '失陷',
      instanceStatusLabel: '平台代理已就绪',
      canOpenService: true,
      canRequestSSH: true,
      canRestart: true,
    })
  })

  it('稳定服务保持原始顺序排在最后', () => {
    const challenges = [challenge('a', 'A'), challenge('b', 'B')]
    const services = [
      service('service-a', { service_status: 'up' }),
      service('service-b', { service_status: 'up' }),
    ]

    expect(toDefenseServiceCards({ challenges, services }).map((card) => card.title)).toEqual([
      'A',
      'B',
    ])
  })
})
