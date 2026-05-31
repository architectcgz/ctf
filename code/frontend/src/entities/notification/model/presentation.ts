import type { NotificationType } from './types'

export type NotificationAccent = 'primary' | 'success' | 'warning' | 'violet'

const notificationTypeLabels: Record<NotificationType, string> = {
  system: '系统',
  contest: '竞赛',
  challenge: '训练',
  team: '团队',
}

const notificationTypeAccents: Record<NotificationType, NotificationAccent> = {
  system: 'primary',
  contest: 'warning',
  challenge: 'success',
  team: 'violet',
}

const notificationAccentColors: Record<NotificationAccent, string> = {
  primary: 'var(--color-primary)',
  success: 'var(--color-success)',
  warning: 'var(--color-warning)',
  violet: 'var(--color-cat-reverse)',
}

function isNotificationType(value: string): value is NotificationType {
  return value in notificationTypeLabels
}

export function getNotificationTypeLabel(type: NotificationType | string): string {
  return isNotificationType(type) ? notificationTypeLabels[type] : notificationTypeLabels.system
}

export function getNotificationTypeAccent(type: NotificationType | string): NotificationAccent {
  return isNotificationType(type) ? notificationTypeAccents[type] : notificationTypeAccents.system
}

export function getNotificationAccentColor(accent: NotificationAccent): string {
  return notificationAccentColors[accent]
}

export function getNotificationTypeAccentColor(type: NotificationType | string): string {
  return getNotificationAccentColor(getNotificationTypeAccent(type))
}

export function getNotificationReadStateLabel(unread: boolean): string {
  return unread ? '未读' : '已读'
}
