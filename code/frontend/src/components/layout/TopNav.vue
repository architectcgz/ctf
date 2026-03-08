<template>
  <header class="sticky top-0 z-50 border-b border-border bg-base/88 backdrop-blur-xl">
    <div class="mx-auto flex h-14 w-full max-w-[1600px] items-center justify-between gap-4 px-4 md:px-6 xl:px-8">
      <div class="flex min-w-0 items-center gap-2 md:gap-3">
        <button
          type="button"
          class="inline-flex h-10 w-10 items-center justify-center rounded-xl border border-border bg-surface text-text-secondary transition hover:border-primary/45 hover:bg-elevated hover:text-text-primary md:hidden"
          aria-label="打开导航"
          @click="$emit('toggleSidebar')"
        >
          <Menu class="h-4 w-4" />
        </button>
        <button
          type="button"
          class="hidden h-10 w-10 items-center justify-center rounded-xl border border-border bg-surface text-text-secondary transition hover:border-primary/45 hover:bg-elevated hover:text-text-primary md:inline-flex"
          :aria-label="sidebarCollapsed ? '展开导航' : '折叠导航'"
          @click="$emit('toggleCollapse')"
        >
          <PanelLeftClose v-if="!sidebarCollapsed" class="h-4 w-4" />
          <PanelLeftOpen v-else class="h-4 w-4" />
        </button>

        <div class="min-w-0">
          <div class="truncate text-[11px] font-semibold uppercase tracking-[0.22em] text-text-muted">
            CTF Platform
          </div>
          <div class="truncate text-sm font-semibold text-text-primary">{{ pageTitle }}</div>
        </div>
      </div>

      <div class="flex shrink-0 items-center gap-2">
        <button
          type="button"
          class="inline-flex h-10 w-10 items-center justify-center rounded-xl border border-border bg-surface text-text-secondary transition hover:border-primary/45 hover:bg-elevated hover:text-text-primary"
          :aria-label="theme === 'light' ? '切换到深色模式' : '切换到浅色模式'"
          @click="toggleTheme"
        >
          <Sun v-if="theme === 'dark'" class="h-4 w-4" />
          <Moon v-else class="h-4 w-4" />
        </button>

        <NotificationDropdown />

        <div class="flex items-center gap-2 rounded-2xl border border-border bg-surface px-2.5 py-1.5 shadow-[0_10px_24px_var(--color-shadow-soft)] sm:gap-3 sm:px-3">
          <div class="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/12 text-xs font-semibold text-primary">
            {{ userInitial }}
          </div>
          <div class="hidden min-w-0 sm:block">
            <div class="truncate text-[11px] font-semibold uppercase tracking-[0.16em] text-text-muted">
              {{ roleLabel }}
            </div>
            <div class="truncate text-sm font-semibold text-text-primary">{{ userLabel }}</div>
          </div>
          <button
            type="button"
            class="inline-flex h-9 w-9 items-center justify-center rounded-xl border border-border bg-elevated text-text-secondary transition hover:border-primary/45 hover:text-text-primary sm:hidden"
            aria-label="退出登录"
            @click="logout"
          >
            <LogOut class="h-4 w-4" />
          </button>
          <ElButton class="hidden sm:inline-flex" size="small" type="primary" plain @click="logout">退出</ElButton>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { LogOut, Menu, Moon, PanelLeftClose, PanelLeftOpen, Sun } from 'lucide-vue-next'

import NotificationDropdown from '@/components/layout/NotificationDropdown.vue'
import { useAuth } from '@/composables/useAuth'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/composables/useTheme'
import { resolveRouteTitle } from '@/utils/routeTitle'

defineProps<{
  sidebarCollapsed: boolean
}>()

defineEmits<{
  toggleSidebar: []
  toggleCollapse: []
}>()

const route = useRoute()
const authStore = useAuthStore()
const { logout } = useAuth()
const { theme, toggleTheme } = useTheme()

const pageTitle = computed(() => resolveRouteTitle(route))
const userLabel = computed(() => authStore.user?.username || '未登录')
const roleLabel = computed(() => {
  const role = authStore.user?.role
  if (role === 'admin') return '管理员'
  if (role === 'teacher') return '教师'
  return '学员'
})
const userInitial = computed(() => userLabel.value.slice(0, 1).toUpperCase())
</script>
