<template>
  <header class="topnav-shell topnav-shell--admin sticky top-0 z-50 h-16 shrink-0">
    <div
      class="topnav-inner topnav-inner-shell mx-auto flex h-full w-full items-center justify-between gap-4 px-4 md:px-6 xl:px-8"
      :class="{ 'topnav-inner-shell--sidebar-collapsed': sidebarCollapsed && !isMobile }"
    >
      <div class="topnav-main flex min-w-0 items-center gap-3 md:gap-4">
        <TopNavMobileToggle
          v-if="isMobile"
          :sidebar-collapsed="sidebarCollapsed"
          @toggle="emit('toggleSidebar')"
        />

        <TopNavBreadcrumbs
          :breadcrumb="backofficeBreadcrumb"
          @navigate="navigateBreadcrumb"
        />
      </div>

      <div class="topnav-actions flex shrink-0 items-center gap-3">
        <div
          class="topnav-tool-cluster"
          :class="{ 'topnav-tool-cluster--admin': true }"
        >
          <button
            type="button"
            class="topnav-icon-button"
            :aria-label="theme === 'light' ? '切换到深色模式' : '切换到浅色模式'"
            @click="toggleTheme"
          >
            <Sun
              v-if="theme === 'dark'"
              class="h-4 w-4"
            />
            <Moon
              v-else
              class="h-4 w-4"
            />
          </button>

          <div ref="brandPickerRef" class="topnav-brand-picker">
            <TopNavBrandPicker
              :open="brandPickerOpen"
              :brand="brand"
              :current-brand-label="currentBrandLabel"
              :available-brands="availableBrands"
              @toggle="toggleBrandPicker"
              @select="selectBrand"
            />
          </div>

          <NotificationDrawer :realtime-status="props.notificationStatus" :controller="props.notificationDrawerController">
            <template #trigger="{ open, toggle, hasUnread, unreadBadgeLabel, setTriggerRef }">
              <TopNavNotificationTrigger
                :open="open"
                :has-unread="hasUnread"
                :unread-badge-label="unreadBadgeLabel"
                :set-trigger-ref="setTriggerRef"
                @toggle="toggle()"
              />
            </template>
          </NotificationDrawer>
        </div>

        <TopNavUserCard
          :user-initial="userInitial"
          :user-display-name="userDisplayName"
          :role-caption="roleCaption"
        />

        <button
          type="button"
          class="topnav-icon-button topnav-icon-button--quiet topnav-logout h-9 w-9"
          aria-label="退出登录"
          @click="props.logout"
        >
          <LogOut class="h-4 w-4" />
        </button>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { LogOut, Moon, Sun } from 'lucide-vue-next'

import NotificationDrawer from '@/shared/ui/layout/NotificationDrawer.vue'
import TopNavBrandPicker from '@/shared/ui/layout/topnav/TopNavBrandPicker.vue'
import TopNavBreadcrumbs from '@/shared/ui/layout/topnav/TopNavBreadcrumbs.vue'
import TopNavMobileToggle from '@/shared/ui/layout/topnav/TopNavMobileToggle.vue'
import TopNavNotificationTrigger from '@/shared/ui/layout/topnav/TopNavNotificationTrigger.vue'
import TopNavUserCard from '@/shared/ui/layout/topnav/TopNavUserCard.vue'
import type { WebSocketStatus } from '@/shared/model/realtime/useWebSocket'
import { useTopNavViewState } from '@/shared/model/layout/topnav/useTopNavViewState'
import '@/shared/ui/layout/topnav/topNavShell.css'

const props = defineProps<{
  sidebarCollapsed: boolean
  notificationStatus: WebSocketStatus
  logout: () => Promise<void>
  notificationDrawerController: import('@/shared/model/layout/notificationDrawerController').NotificationDrawerController
}>()

const emit = defineEmits<{
  toggleSidebar: []
  toggleCollapse: []
}>()

const {
  isMobile,
  brandPickerRef,
  brandPickerOpen,
  backofficeBreadcrumb,
  roleCaption,
  currentBrandLabel,
  userDisplayName,
  userInitial,
  availableBrands,
  brand,
  theme,
  toggleTheme,
  toggleBrandPicker,
  selectBrand,
  navigateBreadcrumb,
} = useTopNavViewState(props.logout)
</script>
