import type { Component, ComponentPublicInstance } from 'vue'
import type { Ref } from 'vue'

import type { NotificationItem } from '@/api/contracts'

/** 通知类型元数据的展示信息（纯展示，来自 features/layout bridge） */
export interface NotificationDrawerTypeMeta {
  icon: Component
  label: string
  accentColor: string
}

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
  typeMeta: (type: string) => NotificationDrawerTypeMeta
  close: () => void
  toggleOpen: () => void
  goToNotifications: () => void
  goToNotificationDetail: (id: string) => void
  markAllRead: () => Promise<void>
}
