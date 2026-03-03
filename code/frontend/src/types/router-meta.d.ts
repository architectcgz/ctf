import 'vue-router'

import type { UserRole } from '@/utils/constants'

declare module 'vue-router' {
  interface RouteMeta {
    requiresAuth?: boolean
    roles?: UserRole[]
    title?: string
    icon?: string
  }
}
