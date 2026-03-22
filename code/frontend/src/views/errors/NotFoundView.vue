<script setup lang="ts">
import { computed } from 'vue'
import { Home, LifeBuoy, MapPinned } from 'lucide-vue-next'
import { RouterLink } from 'vue-router'

import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

const homePath = computed(() => {
  if (!authStore.isLoggedIn) return '/login'
  if (authStore.isAdmin) return '/admin/dashboard'
  if (authStore.isTeacher) return '/teacher/dashboard'
  return '/dashboard'
})
</script>

<template>
  <section class="not-found-view">
    <div class="not-found-kicker">
      <MapPinned class="h-4 w-4" />
      <span>404 · Not Found</span>
    </div>

    <h1 class="not-found-title">
      页面不存在
    </h1>
    <p class="not-found-text">
      你访问的地址不存在，或已被移动。可以返回主页继续操作，或联系管理员确认入口链接是否已更新。
    </p>

    <div class="not-found-actions">
      <RouterLink
        :to="homePath"
        class="not-found-action not-found-action-primary"
      >
        <Home class="h-4 w-4" />
        返回仪表盘
      </RouterLink>
      <RouterLink
        to="/notifications"
        class="not-found-action not-found-action-secondary"
      >
        <LifeBuoy class="h-4 w-4" />
        打开通知中心
      </RouterLink>
    </div>
  </section>
</template>

<style scoped>
.not-found-view {
  min-height: calc(100vh - 11rem);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: flex-start;
  margin-inline: auto;
  width: min(56rem, 100%);
  padding: 2.75rem 1rem 3.2rem;
}

.not-found-kicker {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding-left: 0.72rem;
  border-left: 2px solid color-mix(in srgb, var(--color-warning) 55%, transparent);
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--color-warning) 82%, var(--color-text-primary));
}

.not-found-title {
  margin-top: 1rem;
  font-size: clamp(1.55rem, 2.8vw, 2.1rem);
  font-weight: 700;
  line-height: 1.2;
  letter-spacing: -0.01em;
  color: var(--color-text-primary);
}

.not-found-text {
  margin-top: 0.75rem;
  max-width: 52ch;
  font-size: 0.92rem;
  line-height: 1.8;
  color: var(--color-text-secondary);
}

.not-found-actions {
  margin-top: 1.2rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.65rem;
}

.not-found-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  border-radius: 10px;
  border: 1px solid transparent;
  padding: 0.54rem 0.82rem;
  font-size: 0.85rem;
  font-weight: 600;
  transition: all 180ms ease;
}

.not-found-action-primary {
  border-color: color-mix(in srgb, var(--color-primary) 45%, transparent);
  background: color-mix(in srgb, var(--color-primary) 92%, #0b4f60);
  color: #f8feff;
}

.not-found-action-primary:hover {
  transform: translateY(-1px);
  filter: brightness(1.03);
}

.not-found-action-secondary {
  border-color: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  background: color-mix(in srgb, var(--color-bg-surface) 72%, transparent);
  color: var(--color-text-primary);
}

.not-found-action-secondary:hover {
  border-color: color-mix(in srgb, var(--color-primary) 40%, transparent);
}

@media (max-width: 767px) {
  .not-found-view {
    min-height: calc(100vh - 9.5rem);
    padding-top: 1.8rem;
    justify-content: flex-start;
  }
}

:global([data-theme='light']) .not-found-action-secondary {
  background: color-mix(in srgb, var(--color-bg-surface) 92%, #f1f5f9);
}
</style>
