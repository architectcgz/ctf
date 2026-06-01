import type { Component } from 'vue'

export const notificationTypes = ['system', 'contest', 'challenge', 'team'] as const

export type NotificationType = (typeof notificationTypes)[number]

/** 通知类型元数据——icon、标签文案、点缀色及背景/边框样式。entities/notification 为唯一真相源。 */
export interface NotificationTypeMeta {
  icon: Component
  label: string
  accentColor: string
  iconWrapStyle: Record<string, string>
  badgeStyle: Record<string, string>
}
