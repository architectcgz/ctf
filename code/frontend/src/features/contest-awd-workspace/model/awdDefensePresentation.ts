import type {
  AWDDefenseConnectionData,
  ContestAWDWorkspaceServiceData,
  ContestChallengeItem,
  ID,
} from '@/api/contracts'

export type AWDDefenseRiskLevel = 'critical' | 'warning' | 'watch' | 'stable'

export interface AWDDefenseServiceCard {
  serviceId: ID
  challengeId: ID
  instanceId?: ID
  title: string
  riskLevel: AWDDefenseRiskLevel
  riskReasons: string[]
  serviceStatusLabel: string
  instanceStatusLabel: string
  defenseConnection?: AWDDefenseConnectionData
  canOpenService: boolean
  canRequestSSH: boolean
  canRestart: boolean
}

interface ToDefenseServiceCardsOptions {
  challenges: ContestChallengeItem[]
  services: ContestAWDWorkspaceServiceData[]
}

const ACTIVE_OPERATION_STATUSES = new Set(['requested', 'provisioning', 'recovering'])

const ACTIVE_INSTANCE_STATUSES = new Set(['pending', 'creating'])

const RISK_ORDER: Record<AWDDefenseRiskLevel, number> = {
  critical: 0,
  warning: 1,
  watch: 2,
  stable: 3,
}

export function toDefenseServiceCards({
  challenges,
  services,
}: ToDefenseServiceCardsOptions): AWDDefenseServiceCard[] {
  const serviceById = new Map(
    services
      .filter((service): service is ContestAWDWorkspaceServiceData & { service_id: ID } =>
        Boolean(service.service_id)
      )
      .map((service) => [service.service_id, service])
  )

  return challenges
    .filter((challenge): challenge is ContestChallengeItem & { awd_service_id: ID } =>
      Boolean(challenge.awd_service_id)
    )
    .map((challenge, index) => {
      const service = serviceById.get(challenge.awd_service_id)
      const risk = getDefenseServiceRisk(service)
      return {
        serviceId: challenge.awd_service_id,
        challengeId: challenge.awd_challenge_id || challenge.challenge_id,
        instanceId: service?.instance_id,
        title: challenge.title,
        riskLevel: risk.level,
        riskReasons: risk.reasons,
        serviceStatusLabel: getDisplayedServiceStatus(service).label,
        instanceStatusLabel: getDefenseInstanceStatusLabel(service),
        defenseConnection: service?.defense_connection,
        canOpenService: canOpenDefenseService(service),
        canRequestSSH: canRequestDefenseSSH(service),
        canRestart: Boolean(service?.service_id),
        sortIndex: index,
      }
    })
    .sort((left, right) => {
      const riskDelta = RISK_ORDER[left.riskLevel] - RISK_ORDER[right.riskLevel]
      if (riskDelta !== 0) return riskDelta
      return left.sortIndex - right.sortIndex
    })
    .map((card) => {
      const { sortIndex, ...result } = card
      void sortIndex
      return result
    })
}

export function getDisplayedServiceStatus(service?: ContestAWDWorkspaceServiceData): {
  status: 'up' | 'down' | 'compromised' | 'pending'
  label: string
} {
  if (service?.instance_status === 'running' && service.service_status === 'down') {
    return {
      status: 'pending',
      label: '待同步',
    }
  }

  if (
    service?.service_status === 'up' ||
    service?.service_status === 'down' ||
    service?.service_status === 'compromised'
  ) {
    return {
      status: service.service_status,
      label: getDefenseServiceStatusLabel(service.service_status),
    }
  }

  return {
    status: 'pending',
    label: '待命',
  }
}

export function getDefenseServiceStatusLabel(status?: string): string {
  switch (status) {
    case 'up':
      return '正常'
    case 'down':
      return '离线'
    case 'compromised':
      return '失陷'
    default:
      return '待命'
  }
}

export function getDefenseInstanceStatusLabel(service?: ContestAWDWorkspaceServiceData): string {
  const workspaceStatus = service?.defense_connection?.workspace_status
  switch (service?.instance_status) {
    case 'pending':
      return workspaceStatus === 'running' ? '工作区可用，服务重启中' : '重启队列中'
    case 'creating':
      return workspaceStatus === 'running' ? '工作区可用，服务启动中' : '正在启动'
    case 'running':
      if (service.service_status === 'down') {
        return '平台代理已就绪，等待状态同步'
      }
      return '平台代理已就绪'
    case 'failed':
      return workspaceStatus === 'running' ? '工作区可用，服务启动失败' : '启动失败'
    default:
      if (workspaceStatus === 'running') {
        return '工作区已就绪'
      }
      return service?.instance_id ? '已通过平台代理就绪' : '等待分配实例'
  }
}

export function canOpenDefenseService(service?: ContestAWDWorkspaceServiceData): boolean {
  return Boolean(
    service?.instance_id && (!service.instance_status || service.instance_status === 'running')
  )
}

export function canRequestDefenseSSH(service?: ContestAWDWorkspaceServiceData): boolean {
  return Boolean(
    service?.service_id &&
    service.defense_connection?.entry_mode === 'ssh' &&
    service.defense_connection?.workspace_status === 'running'
  )
}

function getDefenseServiceRisk(service?: ContestAWDWorkspaceServiceData): {
  level: AWDDefenseRiskLevel
  reasons: string[]
} {
  if (!service) {
    return {
      level: 'watch',
      reasons: ['等待服务分配'],
    }
  }

  const reasons: string[] = []

  if (service.service_status === 'compromised') {
    reasons.push('服务已失陷')
    if ((service.attack_received ?? 0) > 0) {
      reasons.push(`检测到 ${service.attack_received} 次攻击`)
    }
    return {
      level: 'critical',
      reasons,
    }
  }

  if (service.service_status === 'down') {
    reasons.push(service.instance_status === 'running' ? '等待状态同步' : '服务离线')
    return {
      level: 'warning',
      reasons,
    }
  }

  if ((service.attack_received ?? 0) > 0) {
    return {
      level: 'watch',
      reasons: [`检测到 ${service.attack_received} 次攻击`],
    }
  }

  if (
    ACTIVE_INSTANCE_STATUSES.has(service.instance_status || '') ||
    ACTIVE_OPERATION_STATUSES.has(service.operation_status || '')
  ) {
    return {
      level: 'watch',
      reasons: ['服务操作进行中'],
    }
  }

  return {
    level: 'stable',
    reasons: ['服务正常'],
  }
}
