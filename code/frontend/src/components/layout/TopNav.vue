<template>
  <header class="topnav-shell sticky top-0 z-50">
    <div class="mx-auto flex min-h-16 w-full max-w-[1600px] items-center justify-between gap-4 px-4 py-3 md:px-6 xl:px-8">
      <div class="flex min-w-0 items-center gap-2 md:gap-3">
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
          <div class="topnav-kicker truncate">
            route://{{ routeSection }}
          </div>
          <div class="truncate text-sm font-semibold text-text-primary">{{ pageTitle }}</div>
        </div>
      </div>

      <div class="flex shrink-0 items-center gap-2">
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

        <div class="topnav-user-card flex items-center gap-2 px-2.5 py-1.5 sm:gap-3 sm:px-3">
          <div class="topnav-user-mark">
            {{ userInitial }}
          </div>
          <div class="hidden min-w-0 sm:block">
            <div class="topnav-user-role truncate">
              {{ roleCaption }}
            </div>
            <div class="truncate text-sm font-semibold text-text-primary">{{ userDisplayName }}</div>
            <div v-if="userMetaLine" class="topnav-user-meta truncate">
              {{ userMetaLine }}
            </div>
          </div>
          <button
            type="button"
            class="topnav-icon-button h-9 w-9"
            aria-label="退出登录"
            @click="logout"
          >
            <LogOut class="h-4 w-4" />
          </button>
        </div>
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
const routeSection = computed(() => {
  const path = route.path
  if (path.startsWith('/admin/')) return 'admin'
  if (path.startsWith('/teacher/')) return 'teaching'
  if (path.startsWith('/profile') || path.startsWith('/settings')) return 'profile'
  if (path.startsWith('/dashboard')) return 'dashboard'
  if (path.startsWith('/challenges')) return 'challenges'
  if (path.startsWith('/contests')) return 'contests'
  if (path.startsWith('/notifications')) return 'notifications'
  return 'workspace'
})
const userLabel = computed(() => authStore.user?.username || '未登录')
const roleLabel = computed(() => {
  const role = authStore.user?.role
  if (role === 'admin') return '管理员'
  if (role === 'teacher') return '教师'
  return '学员'
})
const roleCaption = computed(() => {
  const role = authStore.user?.role
  if (role === 'admin') return '系统管理'
  if (role === 'teacher') return '教学空间'
  return '学生空间'
})
const userDisplayName = computed(() => authStore.user?.name || authStore.user?.username || '未登录')
const userMetaLine = computed(() => {
  const className = authStore.user?.class_name
  const username = authStore.user?.username
  const name = authStore.user?.name

  if (className) return className
  if (name && username) return `@${username}`
  return ''
})
const userInitial = computed(() => userDisplayName.value.slice(0, 1).toUpperCase())
</script>

<style scoped>
.topnav-shell {
  border-bottom: 1px solid color-mix(in srgb, var(--color-border-default) 76%, transparent);
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--color-bg-base) 96%, var(--color-bg-surface)), color-mix(in srgb, var(--color-bg-base) 98%, var(--color-bg-surface))),
    radial-gradient(circle at top left, color-mix(in srgb, var(--color-primary) 10%, transparent), transparent 18rem);
}

.topnav-icon-button {
  display: inline-flex;
  height: 2.5rem;
  width: 2.5rem;
  align-items: center;
  justify-content: center;
  border-radius: 14px;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 78%, transparent);
  background: color-mix(in srgb, var(--color-bg-surface) 70%, var(--color-bg-base));
  color: var(--color-text-secondary);
  transition: all 0.2s ease;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.04);
}

.topnav-icon-button:hover {
  color: var(--color-text-primary);
  border-color: color-mix(in srgb, var(--color-primary) 34%, var(--color-border-default));
  box-shadow: 0 0 18px color-mix(in srgb, var(--color-primary) 14%, transparent);
}

.topnav-title-block {
  min-width: 0;
}

.topnav-kicker {
  font-family: "JetBrains Mono", "Fira Code", "SFMono-Regular", monospace;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.topnav-user-card {
  border: 1px solid color-mix(in srgb, var(--color-border-default) 76%, transparent);
  border-radius: 18px;
  background: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.04);
}

.topnav-user-mark {
  display: inline-flex;
  height: 2.25rem;
  width: 2.25rem;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  border: 1px solid color-mix(in srgb, var(--color-primary) 24%, var(--color-border-default));
  background: color-mix(in srgb, var(--color-primary) 12%, var(--color-bg-surface));
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--color-primary);
}

.topnav-user-role {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.topnav-user-meta {
  margin-top: 0.15rem;
  font-family: "JetBrains Mono", "Fira Code", "SFMono-Regular", monospace;
  font-size: 11px;
  color: var(--color-text-muted);
}

:global([data-theme="light"]) .topnav-shell {
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 96%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle, var(--color-bg-elevated)) 94%, var(--color-bg-base))),
    radial-gradient(circle at top left, rgba(99, 102, 241, 0.08), transparent 18rem);
}
</style>
