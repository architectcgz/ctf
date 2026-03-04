<template>
  <header class="sticky top-0 z-40 border-b border-border bg-surface/85 backdrop-blur">
    <div class="mx-auto flex h-14 max-w-7xl items-center justify-between px-4 md:px-6">
      <div class="flex items-center gap-3">
        <div class="text-sm font-semibold">{{ pageTitle }}</div>
      </div>

      <div class="flex items-center gap-2">
        <button
          @click="toggleTheme"
          class="flex h-8 w-8 items-center justify-center rounded-lg border border-border bg-elevated text-text-secondary transition-colors hover:bg-surface hover:text-text-primary"
          :aria-label="theme === 'light' ? '切换到深色模式' : '切换到浅色模式'"
        >
          <Sun v-if="theme === 'dark'" class="h-4 w-4" />
          <Moon v-else class="h-4 w-4" />
        </button>

        <NotificationDropdown />

        <div class="flex items-center gap-2 rounded-lg border border-border bg-elevated px-3 py-1.5">
          <div class="min-w-0">
            <div class="truncate text-xs text-text-muted">当前账号</div>
            <div class="truncate text-sm font-semibold">{{ userLabel }}</div>
          </div>

          <ElButton size="small" type="primary" plain @click="logout">退出</ElButton>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Sun, Moon } from 'lucide-vue-next'

import NotificationDropdown from '@/components/layout/NotificationDropdown.vue'
import { useAuth } from '@/composables/useAuth'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/composables/useTheme'

const route = useRoute()
const authStore = useAuthStore()
const { logout } = useAuth()
const { theme, toggleTheme } = useTheme()

const pageTitle = computed(() => (typeof route.meta?.title === 'string' ? route.meta.title : ''))
const userLabel = computed(() => authStore.user?.username || '未登录')
</script>

