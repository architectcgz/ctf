export const notificationTypes = ['system', 'contest', 'challenge', 'team'] as const

export type NotificationType = (typeof notificationTypes)[number]
