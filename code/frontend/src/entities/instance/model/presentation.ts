import { getUserDisplayName, getUserUsernameHandle } from '@/entities/user'

export type InstanceStatusTone = 'success' | 'warning' | 'danger' | 'muted'

type KnownInstanceStatus =
  | 'pending'
  | 'creating'
  | 'running'
  | 'expired'
  | 'destroying'
  | 'destroyed'
  | 'failed'
  | 'crashed'

type InstancePresentationStatus = KnownInstanceStatus | string

interface InstanceWaitingPresentationInput {
  status: InstancePresentationStatus
  queue_position?: number | null
  eta_seconds?: number | null
  progress?: number | null
}

interface InstanceAccessPresentationInput {
  access_url?: string | null
  access?: {
    command?: string | null
  } | null
  ssh_info?: {
    host: string
    port: number
  } | null
}

interface InstanceStudentPresentationInput {
  student_name?: string | null
  student_username?: string | null
  student_no?: string | null
  class_name?: string | null
}

const instanceStatusLabels: Record<KnownInstanceStatus, string> = {
  pending: '等待中',
  creating: '创建中',
  running: '运行中',
  expired: '已过期',
  destroying: '销毁中',
  destroyed: '已销毁',
  failed: '启动失败',
  crashed: '运行异常',
}

const instanceStatusTones: Record<KnownInstanceStatus, InstanceStatusTone> = {
  pending: 'warning',
  creating: 'warning',
  running: 'success',
  expired: 'muted',
  destroying: 'warning',
  destroyed: 'muted',
  failed: 'danger',
  crashed: 'danger',
}

const instanceStatusDotClasses: Record<InstanceStatusTone, string> = {
  success: 'instance-status-dot--success',
  warning: 'instance-status-dot--warning',
  danger: 'instance-status-dot--danger',
  muted: 'instance-status-dot--muted',
}

const instanceStatusPillClasses: Record<InstanceStatusTone, string> = {
  success: 'instance-status-pill--running',
  warning: 'instance-status-pill--pending',
  danger: 'instance-status-pill--danger',
  muted: 'instance-status-pill--inactive',
}

function normalizeText(value: string | null | undefined): string {
  return typeof value === 'string' ? value.trim() : ''
}

function isInstanceStatus(value: string): value is KnownInstanceStatus {
  return value in instanceStatusLabels
}

function formatEtaSeconds(seconds?: number | null): string {
  if (typeof seconds !== 'number' || seconds <= 0) {
    return '预计时间计算中'
  }

  const minutes = Math.floor(seconds / 60)
  const restSeconds = seconds % 60
  if (minutes <= 0) {
    return `${restSeconds} 秒`
  }
  return `${minutes} 分 ${restSeconds} 秒`
}

export function getInstanceStatusLabel(status: InstancePresentationStatus): string {
  const normalizedStatus = normalizeText(status)
  if (!normalizedStatus) {
    return '--'
  }

  return isInstanceStatus(normalizedStatus)
    ? instanceStatusLabels[normalizedStatus]
    : normalizedStatus
}

export function getInstanceStatusTone(status: InstancePresentationStatus): InstanceStatusTone {
  const normalizedStatus = normalizeText(status)
  if (!normalizedStatus || !isInstanceStatus(normalizedStatus)) {
    return 'muted'
  }

  return instanceStatusTones[normalizedStatus]
}

export function getInstanceStatusDotClass(status: InstancePresentationStatus): string {
  return instanceStatusDotClasses[getInstanceStatusTone(status)]
}

export function getInstanceStatusPillClass(status: InstancePresentationStatus): string {
  return instanceStatusPillClasses[getInstanceStatusTone(status)]
}

export function getInstanceRemainingSeconds(expiresAt: string, nowMs = Date.now()): number {
  const expiresAtMs = new Date(expiresAt).getTime()
  if (!Number.isFinite(expiresAtMs)) {
    return 0
  }

  return Math.max(0, Math.floor((expiresAtMs - nowMs) / 1000))
}

export function getInstanceRemainingTone(
  seconds: number,
  options?: { warningSeconds?: number; dangerSeconds?: number }
): InstanceStatusTone {
  const normalizedSeconds = Number.isFinite(seconds) ? Math.max(0, Math.floor(seconds)) : 0
  const warningSeconds = options?.warningSeconds ?? 600
  const dangerSeconds = options?.dangerSeconds ?? 300

  if (normalizedSeconds <= 0) {
    return 'muted'
  }
  if (normalizedSeconds < dangerSeconds) {
    return 'danger'
  }
  if (normalizedSeconds < warningSeconds) {
    return 'warning'
  }
  return 'success'
}

export function formatInstanceRemainingTime(
  seconds: number,
  options?: { expiredLabel?: string }
): string {
  const normalizedSeconds = Number.isFinite(seconds) ? Math.max(0, Math.floor(seconds)) : 0

  if (normalizedSeconds <= 0 && options?.expiredLabel) {
    return options.expiredLabel
  }

  const hours = Math.floor(normalizedSeconds / 3600)
  const minutes = Math.floor((normalizedSeconds % 3600) / 60)
  const restSeconds = normalizedSeconds % 60

  return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(restSeconds).padStart(2, '0')}`
}

export function getInstanceWaitingHint(instance: InstanceWaitingPresentationInput): string {
  if (instance.status === 'failed') {
    return '启动失败，当前目标不可访问'
  }
  if (instance.status === 'crashed') {
    return '实例运行异常，当前目标不可访问'
  }
  if (instance.status !== 'pending' && instance.status !== 'creating') {
    return ''
  }

  const details: string[] = ['实例正在排队创建']

  if (typeof instance.queue_position === 'number' && instance.queue_position > 0) {
    details.push(`队列第 ${instance.queue_position} 位`)
  }

  details.push(`预计等待 ${formatEtaSeconds(instance.eta_seconds)}`)

  if (typeof instance.progress === 'number') {
    const progress = Math.max(0, Math.min(100, Math.round(instance.progress)))
    details.push(`进度 ${progress}%`)
  }

  return details.join('，')
}

export function getInstanceWaitingQueueLabel(
  instance: InstanceWaitingPresentationInput | null | undefined
): string {
  if (!instance || (instance.status !== 'pending' && instance.status !== 'creating')) {
    return ''
  }

  if (typeof instance.queue_position === 'number' && instance.queue_position > 0) {
    return `当前排队：第 ${instance.queue_position} 位`
  }

  return '当前排队：排队信息同步中'
}

export function getInstanceWaitingEtaLabel(
  instance: InstanceWaitingPresentationInput | null | undefined
): string {
  if (!instance || (instance.status !== 'pending' && instance.status !== 'creating')) {
    return ''
  }

  return `预计等待：${formatEtaSeconds(instance.eta_seconds)}`
}

export function getInstanceWaitingProgressLabel(
  instance: InstanceWaitingPresentationInput | null | undefined
): string {
  if (
    !instance ||
    (instance.status !== 'pending' && instance.status !== 'creating') ||
    typeof instance.progress !== 'number'
  ) {
    return ''
  }

  const normalized = Math.max(0, Math.min(100, Math.round(instance.progress)))
  return `创建进度：${normalized}%`
}

export function formatInstanceAccessDisplay(instance: InstanceAccessPresentationInput): string {
  return (
    normalizeText(instance.access?.command) ||
    normalizeText(instance.access_url) ||
    (instance.ssh_info ? `${instance.ssh_info.host}:${instance.ssh_info.port}` : '')
  )
}

export function getInstanceStudentDisplayName(
  student: InstanceStudentPresentationInput | null | undefined,
  fallback = '--'
): string {
  return getUserDisplayName(
    {
      name: student?.student_name,
      username: student?.student_username,
    },
    fallback
  )
}

export function getInstanceStudentIdentityLabel(
  student: InstanceStudentPresentationInput | null | undefined,
  fallback = '--'
): string {
  const studentNo = normalizeText(student?.student_no)
  if (studentNo) {
    return studentNo
  }

  return getUserUsernameHandle({ username: student?.student_username }, fallback)
}

export function getInstanceStudentSecondaryLabel(
  student: InstanceStudentPresentationInput | null | undefined,
  fallback = '--'
): string {
  const usernameHandle = getUserUsernameHandle({ username: student?.student_username }, '')
  const className = normalizeText(student?.class_name)

  if (usernameHandle && className) {
    return `${usernameHandle} · ${className}`
  }

  return usernameHandle || className || fallback
}
