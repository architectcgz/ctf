<template>
  <header class="topnav-shell sticky top-0 z-50">
    <div class="topnav-inner mx-auto flex min-h-16 w-full max-w-[1600px] items-center justify-between gap-4 px-4 py-3 md:px-6 xl:px-8">
      <div class="topnav-main flex min-w-0 items-center gap-3 md:gap-4">
        <button
          type="button"
          class="topnav-icon-button"
          :aria-label="sidebarCollapsed ? '展开导航' : '折叠导航'"
          @click="isMobile ? $emit('toggleSidebar') : $emit('toggleCollapse')"
        >
          <Menu class="h-4 w-4 md:hidden" />
          <PanelLeftClose v-if="!sidebarCollapsed" class="hidden h-4 w-4 md:block" />
          <PanelLeftOpen v-else class="hidden h-4 w-4 md:block" />
        </button>

        <div class="topnav-title-block min-w-0">
          <div class="topnav-page-title truncate text-sm font-semibold text-text-primary md:text-[15px]">
            {{ pageTitle }}
          </div>
        </div>
      </div>

      <div class="topnav-actions flex shrink-0 items-center gap-3">
        <div class="topnav-tool-cluster">
          <button
            type="button"
            class="topnav-icon-button"
            :aria-label="theme === 'light' ? '切换到深色模式' : '切换到浅色模式'"
            @click="toggleTheme"
          >
            <Sun v-if="theme === 'dark'" class="h-4 w-4" />
            <Moon v-else class="h-4 w-4" />
          </button>

          <NotificationDropdown :realtime-status="notificationStatus" />
        </div>

        <div class="topnav-user-card flex items-center gap-3 px-2.5 py-1.5 sm:px-3">
          <div class="topnav-user-mark">
            {{ userInitial }}
          </div>
          <div class="topnav-user-identity hidden min-w-0 sm:block">
            <div class="topnav-user-name truncate text-sm font-semibold text-text-primary">
              {{ userDisplayName }}
            </div>
            <div class="topnav-user-role truncate">
              {{ roleCaption }}
            </div>
          </div>
        </div>

        <button
          type="button"
          class="topnav-icon-button topnav-icon-button--quiet topnav-logout h-9 w-9"
          aria-label="退出登录"
          @click="logout"
        >
          <LogOut class="h-4 w-4" />
        </button>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { LogOut, Menu, Moon, PanelLeftClose, PanelLeftOpen, Sun } from 'lucide-vue-next'

import NotificationDropdown from '@/components/layout/NotificationDropdown.vue'
import { useAuth } from '@/composables/useAuth'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/composables/useTheme'
import type { WebSocketStatus } from '@/composables/useWebSocket'
import { resolveRouteTitle } from '@/utils/routeTitle'

defineProps<{
  sidebarCollapsed: boolean
  notificationStatus: WebSocketStatus
}>()

defineEmits<{
  toggleSidebar: []
  toggleCollapse: []
}>()

const route = useRoute()
const authStore = useAuthStore()

const isMobile = ref(window.innerWidth < 768)
function onResize() { isMobile.value = window.innerWidth < 768 }
onMounted(() => window.addEventListener('resize', onResize))
onUnmounted(() => window.removeEventListener('resize', onResize))
const { logout } = useAuth()
const { theme, toggleTheme } = useTheme()

const pageTitle = computed(() => resolveRouteTitle(route))
const roleCaption = computed(() => {
  const role = authStore.user?.role
  if (role === 'admin') return '系统管理'
  if (role === 'teacher') return '教学空间'
  return '学生空间'
})
const userDisplayName = computed(() => authStore.user?.name || authStore.user?.username || '未登录')
const userInitial = computed(() => userDisplayName.value.slice(0, 1).toUpperCase())
</script>

<style scoped>
.topnav-shell {
  border-bottom: 1px solid color-mix(in srgb, var(--color-border-default) 76%, transparent);
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--color-bg-base) 97%, var(--color-bg-surface)), color-mix(in srgb, var(--color-bg-base) 99%, var(--color-bg-surface))),
    radial-gradient(circle at top left, color-mix(in srgb, var(--color-primary) 8%, transparent), transparent 18rem);
  backdrop-filter: blur(14px);
}

.topnav-inner {
  position: relative;
}

.topnav-main,
.topnav-actions {
  min-height: 2.75rem;
}

.topnav-tool-cluster {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.25rem;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 72%, transparent);
  border-radius: 16px;
  background: color-mix(in srgb, var(--color-bg-surface) 54%, var(--color-bg-base));
}

.topnav-actions :deep(.notification-trigger) {
  height: 2.5rem;
  width: 2.5rem;
  border-radius: 12px;
  border: 1px solid transparent;
  background: transparent;
  box-shadow: none;
}

.topnav-actions :deep(.notification-trigger:hover) {
  border-color: color-mix(in srgb, var(--color-primary) 20%, var(--color-border-default));
  background: color-mix(in srgb, var(--color-primary) 5%, var(--color-bg-surface));
}

.topnav-icon-button {
  display: inline-flex;
  height: 2.5rem;
  width: 2.5rem;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 74%, transparent);
  background: color-mix(in srgb, var(--color-bg-surface) 58%, var(--color-bg-base));
  color: var(--color-text-secondary);
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease,
    transform 0.2s ease;
}

.topnav-icon-button:hover {
  color: var(--color-text-primary);
  border-color: color-mix(in srgb, var(--color-primary) 24%, var(--color-border-default));
  background: color-mix(in srgb, var(--color-primary) 5%, var(--color-bg-surface));
  transform: translateY(-1px);
}

.topnav-icon-button--quiet {
  height: 2.25rem;
  width: 2.25rem;
}

.topnav-title-block {
  min-width: 0;
}

.topnav-page-title {
  letter-spacing: -0.01em;
}

.topnav-user-card {
  border: 1px solid color-mix(in srgb, var(--color-border-default) 72%, transparent);
  border-radius: 16px;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base)), color-mix(in srgb, var(--color-bg-surface) 58%, var(--color-bg-base)));
  min-height: 2.75rem;
}

.topnav-user-mark {
  display: inline-flex;
  height: 2.25rem;
  width: 2.25rem;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  border: 1px solid color-mix(in srgb, var(--color-primary) 18%, var(--color-border-default));
  background: color-mix(in srgb, var(--color-primary) 10%, var(--color-bg-surface));
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--color-primary);
}

.topnav-user-identity {
  min-width: 0;
}

.topnav-user-role {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.topnav-logout {
  border-color: color-mix(in srgb, var(--color-danger) 16%, var(--color-border-default));
}

.topnav-logout:hover {
  border-color: color-mix(in srgb, var(--color-danger) 28%, var(--color-border-default));
  background: color-mix(in srgb, var(--color-danger) 6%, var(--color-bg-surface));
}

:global([data-theme="light"]) .topnav-shell {
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 97%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle, var(--color-bg-elevated)) 95%, var(--color-bg-base))),
    radial-gradient(circle at top left, rgba(99, 102, 241, 0.06), transparent 18rem);
}

.topnav-icon-button:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--color-primary) 44%, white);
  outline-offset: 3px;
}

@media (max-width: 767px) {
  .topnav-actions {
    gap: 0.5rem;
  }

  .topnav-tool-cluster {
    gap: 0.2rem;
    padding: 0.2rem;
  }

  .topnav-user-card {
    padding-left: 0.45rem;
    padding-right: 0.45rem;
  }
}
</style>
