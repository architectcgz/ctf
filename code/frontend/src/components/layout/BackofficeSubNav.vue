<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getVisibleBackofficeSecondaryItems } from '@/config/backofficeNavigation'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const items = computed(() =>
  getVisibleBackofficeSecondaryItems(route.path, authStore.user?.role ?? null)
)

async function navigate(path: string): Promise<void> {
  if (route.path === path) {
    return
  }

  await router.push(path)
}
</script>

<template>
  <nav
    v-if="items.length > 0"
    class="backoffice-subnav"
    role="tablist"
    aria-label="后台模块导航"
  >
    <button
      v-for="item in items"
      :key="item.routeName"
      type="button"
      class="backoffice-subnav__item"
      :class="{ 'backoffice-subnav__item--active': item.active }"
      :aria-selected="item.active ? 'true' : 'false'"
      @click="navigate(item.path)"
    >
      {{ item.label }}
    </button>
  </nav>
</template>

<style scoped>
.backoffice-subnav {
  display: flex;
  align-items: center;
  gap: 2rem;
  overflow-x: auto;
  padding: 0 2rem;
  border-bottom: 1px solid #e2e8f0;
  background: white;
}

.backoffice-subnav__item {
  position: relative;
  display: inline-flex;
  align-items: center;
  min-height: 3rem;
  padding: 0 0 0.875rem;
  border: none;
  background: transparent;
  color: #64748b;
  font-size: 0.875rem;
  font-weight: 700;
  white-space: nowrap;
  transition: color 0.2s ease;
}

.backoffice-subnav__item::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 2px;
  border-radius: 999px;
  background: transparent;
}

.backoffice-subnav__item:hover {
  color: #0f172a;
}

.backoffice-subnav__item--active {
  color: #2563eb;
}

.backoffice-subnav__item--active::after {
  background: #2563eb;
}
</style>
