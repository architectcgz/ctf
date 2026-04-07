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

          <div ref="brandPickerRef" class="topnav-brand-picker">
            <button
              type="button"
              class="topnav-icon-button"
              aria-label="切换主题色"
              aria-controls="topnav-brand-picker-panel"
              :aria-expanded="brandPickerOpen ? 'true' : 'false'"
              :title="`当前主题色：${currentBrandLabel}`"
              @click="toggleBrandPicker"
            >
              <Palette class="h-4 w-4" />
            </button>

            <div
              v-if="brandPickerOpen"
              id="topnav-brand-picker-panel"
              class="topnav-brand-picker-panel"
              role="menu"
              aria-label="主题色选择"
            >
              <button
                v-for="option in availableBrands"
                :key="option.value"
                type="button"
                class="topnav-brand-dot"
                :class="{ 'topnav-brand-dot--active': option.value === brand }"
                role="menuitemradio"
                :aria-checked="option.value === brand"
                :aria-label="`切换到${option.label}主题`"
                :data-brand="option.value"
                :title="option.label"
                @click="selectBrand(option.value)"
              >
                <span class="sr-only">{{ option.label }}</span>
              </button>
            </div>
          </div>

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
import { LogOut, Menu, Moon, Palette, PanelLeftClose, PanelLeftOpen, Sun } from 'lucide-vue-next'

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
const brandPickerRef = ref<HTMLElement | null>(null)
const brandPickerOpen = ref(false)

function onResize() { isMobile.value = window.innerWidth < 768 }
const { logout } = useAuth()
const { availableBrands, brand, setBrand, theme, toggleTheme } = useTheme()

const pageTitle = computed(() => resolveRouteTitle(route))
const roleCaption = computed(() => {
  const role = authStore.user?.role
  if (role === 'admin') return '系统管理'
  if (role === 'teacher') return '教学空间'
  return '学生空间'
})
const currentBrandLabel = computed(
  () => availableBrands.find((option) => option.value === brand.value)?.label || '绿色'
)
const userDisplayName = computed(() => authStore.user?.name || authStore.user?.username || '未登录')
const userInitial = computed(() => userDisplayName.value.slice(0, 1).toUpperCase())

function toggleBrandPicker(): void {
  brandPickerOpen.value = !brandPickerOpen.value
}

function closeBrandPicker(): void {
  brandPickerOpen.value = false
}

function selectBrand(nextBrand: (typeof availableBrands)[number]['value']): void {
  setBrand(nextBrand)
  closeBrandPicker()
}

function handleDocumentPointerDown(event: MouseEvent): void {
  if (!brandPickerOpen.value) return

  const target = event.target
  if (!(target instanceof Node)) return
  if (brandPickerRef.value?.contains(target)) return

  closeBrandPicker()
}

function handleDocumentKeydown(event: KeyboardEvent): void {
  if (event.key !== 'Escape' || !brandPickerOpen.value) return
  closeBrandPicker()
}

onMounted(() => {
  window.addEventListener('resize', onResize)
  document.addEventListener('mousedown', handleDocumentPointerDown)
  document.addEventListener('keydown', handleDocumentKeydown)
})

onUnmounted(() => {
  window.removeEventListener('resize', onResize)
  document.removeEventListener('mousedown', handleDocumentPointerDown)
  document.removeEventListener('keydown', handleDocumentKeydown)
})
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
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.25rem;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 72%, transparent);
  border-radius: 16px;
  background: color-mix(in srgb, var(--color-bg-surface) 54%, var(--color-bg-base));
}

.topnav-brand-picker {
  position: relative;
  display: inline-flex;
}

.topnav-brand-picker-panel {
  position: absolute;
  top: calc(100% + 0.55rem);
  right: 0;
  z-index: 20;
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.55rem;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  border-radius: 999px;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--color-bg-surface) 94%, var(--color-bg-base)), color-mix(in srgb, var(--color-bg-surface) 82%, var(--color-bg-base)));
  box-shadow: 0 18px 34px color-mix(in srgb, var(--color-shadow-soft) 90%, transparent);
}

.topnav-brand-dot {
  --brand-dot-color: var(--color-primary);
  display: inline-flex;
  height: 1.15rem;
  width: 1.15rem;
  border-radius: 999px;
  border: none;
  padding: 0;
  background: var(--brand-dot-color);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, white 28%, transparent);
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease;
}

.topnav-brand-dot:hover {
  transform: translateY(-1px) scale(1.06);
}

.topnav-brand-dot[data-brand="green"] {
  --brand-dot-color: #16a34a;
}

.topnav-brand-dot[data-brand="cyan"] {
  --brand-dot-color: #0891b2;
}

.topnav-brand-dot[data-brand="blue"] {
  --brand-dot-color: #2563eb;
}

.topnav-brand-dot[data-brand="orange"] {
  --brand-dot-color: #e18a2a;
}

.topnav-brand-dot--active {
  box-shadow:
    0 0 0 3px color-mix(in srgb, var(--color-bg-base) 92%, transparent),
    0 0 0 5px color-mix(in srgb, var(--color-primary) 26%, transparent),
    inset 0 0 0 1px color-mix(in srgb, white 34%, transparent);
}

.topnav-actions :deep(.notification-trigger) {
  height: 2.5rem;
  width: 2.5rem;
  border-radius: 12px;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 74%, transparent);
  background: color-mix(in srgb, var(--color-bg-surface) 58%, var(--color-bg-base));
  color: var(--color-text-secondary);
  box-shadow: none;
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease,
    transform 0.2s ease;
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

.topnav-brand-dot:focus-visible {
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
