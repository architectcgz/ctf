<template>
  <AppLayout
    :notification-status="notificationStatus"
    :logout="logout"
    :notification-drawer-controller="notificationDrawer"
  />
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'

import {
  useLayoutNotificationDrawerBridge,
  useLayoutNotificationRealtimeBridge,
  useLayoutSessionActionsBridge,
} from '@/features/layout'
import AppLayout from '@/shared/ui/layout/AppLayout.vue'

/**
 * App shell 的组装根页面。
 * 在这里将 features/layout/ 层的 bridge composable 接线到
 * shared/ui/layout/ 的纯展示组件上，保持 shared 层对 features 层的零依赖。
 * 路由导航回调也在此注入，使 feature model 不直接依赖 vue-router。
 */

const router = useRouter()

const { start, status: notificationStatus } = useLayoutNotificationRealtimeBridge()
const notificationDrawer = useLayoutNotificationDrawerBridge(
  () => notificationStatus.value,
  () => { void router.push('/notifications') },
  (id: string) => { void router.push(`/notifications/${encodeURIComponent(id)}`) }
)
const { logout } = useLayoutSessionActionsBridge(
  () => { void router.push('/login') }
)

onMounted(() => {
  void start()
})
</script>
