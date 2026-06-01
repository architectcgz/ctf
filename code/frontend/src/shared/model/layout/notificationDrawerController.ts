import type { ComponentPublicInstance } from 'vue'
import type { Ref } from 'vue'

import type { NotificationItem } from '@/api/contracts'
import type { NotificationTypeMeta } from '@/entities/notification'

/**
 * 通知抽屉控制器的公共接口。
 * 实现在 features/layout/ 层，通过 props 传入 shared UI 组件，
 * 避免 shared 直接依赖 features。
 */
export interface NotificationDrawerController {
  open: Ref<boolean>
  setTriggerRef: (element: Element | ComponentPublicInstance | null) => void
  unreadCount: Ref<number>
  isMarkingAllRead: Ref<boolean>
  items: Ref<NotificationItem[]>
  typeMeta: (type: string) => NotificationTypeMeta
  close: () => void
  toggleOpen: () => void
  goToNotifications: () => void
  goToNotificationDetail: (id: string) => void
  markAllRead: () => Promise<void>
}
