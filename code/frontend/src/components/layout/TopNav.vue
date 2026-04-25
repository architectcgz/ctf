<template>
  <header
    class="topnav-shell sticky top-0 z-50"
    :class="{ 'topnav-shell--admin': isBackofficeRoute }"
  >
    <div
      class="topnav-inner topnav-inner-shell mx-auto flex h-16 w-full items-center justify-between gap-4 px-4 md:px-6 xl:px-8"
    >
      <div class="topnav-main flex min-w-0 items-center gap-3 md:gap-4">
        <button
          v-if="!isBackofficeRoute || isMobile"
          type="button"
          class="topnav-icon-button"
          :aria-label="sidebarCollapsed ? '展开导航' : '折叠导航'"
          @click="isMobile ? $emit('toggleSidebar') : $emit('toggleCollapse')"
        >
          <Menu class="h-4 w-4 md:hidden" />
          <PanelLeftClose
            v-if="!sidebarCollapsed"
            class="hidden h-4 w-4 md:block"
          />
          <PanelLeftOpen
            v-else
            class="hidden h-4 w-4 md:block"
          />
        </button>

        <div
          v-if="isBackofficeRoute"
          class="flex min-w-0 items-center text-sm font-bold text-slate-500"
        >
          <span class="text-slate-400 whitespace-nowrap">Workspace</span>
          <span class="mx-2 text-slate-300">/</span>
          <span class="whitespace-nowrap">{{ backofficeBreadcrumb.moduleLabel }}</span>
          <span class="mx-2 text-slate-300">/</span>
          <span class="truncate text-slate-900 font-black">
            {{ backofficeBreadcrumb.secondaryLabel }}
          </span>
        </div>

        <div
          v-else
          class="topnav-title-block min-w-0"
        >
          <div class="topnav-page-title truncate font-semibold text-text-primary">
            {{ pageTitle }}
          </div>
        </div>
      </div>

      <div class="topnav-actions flex shrink-0 items-center gap-3">
        <div
          class="topnav-tool-cluster"
          :class="{ 'topnav-tool-cluster--admin': isBackofficeRoute }"
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

          <div
            ref="brandPickerRef"
            class="topnav-brand-picker"
          >
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

        <div
          class="topnav-user-card flex items-center gap-3 px-2.5 py-1.5 sm:px-3"
          :class="{ 'topnav-user-card--admin': isBackofficeRoute }"
        >
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
import {
  getBackofficeModuleByPath,
  getVisibleBackofficeSecondaryItems,
} from '@/config/backofficeNavigation'
import { useAuth } from '@/composables/useAuth'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/composables/useTheme'
import type { WebSocketStatus } from '@/composables/useWebSocket'
import { isBackofficeRoute as checkBackofficeRoute } from '@/utils/backofficeRouteMeta'
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
const isBackofficeRoute = computed(() => checkBackofficeRoute(route.path))

const isMobile = ref(window.innerWidth < 768)
const brandPickerRef = ref<HTMLElement | null>(null)
const brandPickerOpen = ref(false)

function onResize() {
  isMobile.value = window.innerWidth < 768
}
const { logout } = useAuth()
const { availableBrands, brand, setBrand, theme, toggleTheme } = useTheme()

const pageTitle = computed(() => resolveRouteTitle(route))
const backofficeBreadcrumb = computed(() => {
  const module = getBackofficeModuleByPath(route.path)
  const secondaryItems = getVisibleBackofficeSecondaryItems(route.path, authStore.user?.role ?? null)
  const activeSecondaryItem = secondaryItems.find((item) => item.active) ?? null

  return {
    moduleLabel: module?.label ?? '后台',
    secondaryLabel: activeSecondaryItem?.label ?? pageTitle.value ?? '工作区',
  }
})
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
  --topnav-surface: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --topnav-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 84%, var(--color-bg-base));
  --topnav-surface-elevated: color-mix(in srgb, var(--color-bg-elevated) 86%, var(--color-bg-surface));
  --topnav-line: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --topnav-line-strong: color-mix(in srgb, var(--color-border-default) 88%, transparent);
  --topnav-text: color-mix(in srgb, var(--color-text-primary) 94%, transparent);
  --topnav-muted: color-mix(in srgb, var(--color-text-secondary) 92%, transparent);
  --topnav-faint: color-mix(in srgb, var(--color-text-muted) 92%, transparent);
  --topnav-soft-accent: color-mix(in srgb, var(--color-primary) 10%, transparent);
  --topnav-danger-line: color-mix(in srgb, var(--color-danger) 30%, var(--topnav-line));
  --topnav-danger-soft: color-mix(in srgb, var(--color-danger) 12%, var(--topnav-surface));
  border-bottom: 1px solid color-mix(in srgb, var(--color-border-default) 76%, transparent);
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--color-bg-base) 97%, var(--color-bg-surface)),
      color-mix(in srgb, var(--color-bg-base) 99%, var(--color-bg-surface))
    ),
    radial-gradient(
      circle at top left,
      color-mix(in srgb, var(--color-primary) 8%, transparent),
      transparent 18rem
    );
  backdrop-filter: blur(14px);
}

.topnav-shell--admin {
  border-bottom-color: var(--topnav-line-strong);
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--topnav-surface) 98%, var(--color-bg-base)),
      color-mix(in srgb, var(--topnav-surface-subtle) 96%, var(--color-bg-base))
    ),
    radial-gradient(
      circle at top left,
      color-mix(in srgb, var(--color-primary) 8%, transparent),
      transparent 20rem
    );
  backdrop-filter: none;
}

.topnav-inner {
  position: relative;
}

.topnav-inner-shell {
  max-width: 100rem;
}

.topnav-main,
.topnav-actions {
  min-height: 2.75rem;
}

.topnav-tool-cluster {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.2rem;
  border: 1px solid var(--topnav-line);
  border-radius: 12px;
  background: var(--topnav-surface-subtle);
}

.topnav-tool-cluster--admin {
  border-color: var(--topnav-line);
  border-radius: 999px;
  background: var(--topnav-surface);
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
  border: 1px solid var(--topnav-line-strong);
  border-radius: 999px;
  background: var(--topnav-surface-elevated);
  box-shadow: 0 10px 25px color-mix(in srgb, var(--color-shadow-soft) 68%, transparent);
}

.topnav-brand-dot {
  --brand-dot-color: var(--color-primary);
  display: inline-flex;
  height: 1rem;
  width: 1rem;
  border-radius: 999px;
  border: none;
  padding: 0;
  background: var(--brand-dot-color);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.2);
  transition: all 0.2s ease;
}

.topnav-brand-dot:hover {
  transform: translateY(-1px) scale(1.1);
}

.topnav-brand-dot[data-brand='green'] {
  --brand-dot-color: var(--color-brand-swatch-green);
}

.topnav-brand-dot[data-brand='cyan'] {
  --brand-dot-color: var(--color-brand-swatch-cyan);
}

.topnav-brand-dot[data-brand='blue'] {
  --brand-dot-color: var(--color-brand-swatch-blue);
}

.topnav-brand-dot[data-brand='orange'] {
  --brand-dot-color: var(--color-brand-swatch-orange);
}

.topnav-brand-dot--active {
  box-shadow:
    0 0 0 2px var(--topnav-surface-elevated),
    0 0 0 4px color-mix(in srgb, var(--color-primary) 22%, transparent);
}

.topnav-actions :deep(.notification-trigger) {
  height: 2.25rem;
  width: 2.25rem;
  border-radius: 10px;
  border: 1px solid var(--topnav-line);
  background: var(--topnav-surface);
  color: var(--topnav-muted);
  box-shadow: none;
  transition: all 0.2s ease;
}

.topnav-actions :deep(.notification-trigger:hover) {
  border-color: var(--topnav-line-strong);
  background: var(--topnav-surface-subtle);
  color: var(--topnav-text);
}

.topnav-icon-button {
  display: inline-flex;
  height: 2.25rem;
  width: 2.25rem;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  border: 1px solid var(--topnav-line);
  background: var(--topnav-surface);
  color: var(--topnav-muted);
  transition: all 0.2s ease;
}

.topnav-icon-button:hover {
  color: var(--topnav-text);
  border-color: var(--topnav-line-strong);
  background: var(--topnav-surface-subtle);
}

.topnav-icon-button--quiet {
  height: 2rem;
  width: 2rem;
}

.text-slate-500 {
  color: var(--topnav-muted);
}

.text-slate-400 {
  color: var(--topnav-faint);
}

.text-slate-300 {
  color: color-mix(in srgb, var(--topnav-line-strong) 92%, transparent);
}

.text-slate-900 {
  color: var(--topnav-text);
}

.font-bold {
  font-weight: 700;
}

.font-black {
  font-weight: 900;
}

.topnav-title-block {
  min-width: 0;
}

.topnav-page-title {
  font-size: 0.875rem;
  line-height: 1.25rem;
  letter-spacing: -0.01em;
}

.topnav-user-card {
  border: 1px solid var(--topnav-line);
  border-radius: 12px;
  background: var(--topnav-surface);
  min-height: 2.5rem;
  padding: 0 0.75rem;
}

.topnav-user-card--admin {
  border-radius: 999px;
  background: var(--topnav-surface);
  border-color: var(--topnav-line);
}

.topnav-user-mark {
  display: inline-flex;
  height: 1.75rem;
  width: 1.75rem;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  border: 1px solid var(--topnav-line-strong);
  background: var(--topnav-surface-subtle);
  font-size: 0.7rem;
  font-weight: 800;
  color: var(--topnav-muted);
}

.topnav-shell--admin .topnav-user-mark {
  border-radius: 999px;
}

.topnav-shell--admin .topnav-icon-button,
.topnav-shell--admin .topnav-actions :deep(.notification-trigger) {
  border-radius: 999px;
  border-color: var(--topnav-line);
  background: var(--topnav-surface);
}

.topnav-shell--admin .topnav-icon-button:hover,
.topnav-shell--admin .topnav-actions :deep(.notification-trigger:hover) {
  border-color: var(--topnav-line-strong);
  background: var(--topnav-surface-subtle);
}

.topnav-user-identity {
  min-width: 0;
}

.topnav-user-role {
  font-size: 9px;
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--topnav-faint);
}

.topnav-logout {
  border-color: var(--topnav-danger-line);
  color: color-mix(in srgb, var(--color-danger) 90%, var(--topnav-text));
}

.topnav-logout:hover {
  border-color: color-mix(in srgb, var(--color-danger) 42%, var(--topnav-line));
  background: var(--topnav-danger-soft);
  color: color-mix(in srgb, var(--color-danger) 96%, var(--topnav-text));
}

:global([data-theme='light']) .topnav-shell {
  --topnav-surface: color-mix(in srgb, white 96%, var(--color-bg-surface));
  --topnav-surface-subtle: color-mix(in srgb, #f8fafc 92%, var(--color-bg-surface));
  --topnav-surface-elevated: white;
  --topnav-line: color-mix(in srgb, #e2e8f0 90%, transparent);
  --topnav-line-strong: color-mix(in srgb, #d9e1ec 94%, transparent);
  --topnav-text: #0f172a;
  --topnav-muted: #64748b;
  --topnav-faint: #94a3b8;
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 97%, var(--color-bg-base)),
      color-mix(
        in srgb,
        var(--journal-surface-subtle, var(--color-bg-elevated)) 95%,
        var(--color-bg-base)
      )
    ),
    radial-gradient(
      circle at top left,
      color-mix(in srgb, var(--color-primary-hover) 6%, transparent),
      transparent 18rem
    );
}

:global([data-theme='dark']) .topnav-shell--admin {
  --topnav-surface: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --topnav-surface-subtle: color-mix(in srgb, var(--color-bg-elevated) 84%, var(--color-bg-base));
  --topnav-surface-elevated: color-mix(in srgb, var(--color-bg-elevated) 92%, var(--color-bg-base));
  --topnav-line: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  --topnav-line-strong: color-mix(in srgb, var(--color-border-default) 92%, transparent);
  --topnav-text: color-mix(in srgb, var(--color-text-primary) 94%, transparent);
  --topnav-muted: color-mix(in srgb, var(--color-text-secondary) 90%, transparent);
  --topnav-faint: color-mix(in srgb, var(--color-text-muted) 90%, transparent);
}

:global([data-theme='dark']) .topnav-tool-cluster--admin,
:global([data-theme='dark']) .topnav-user-card--admin,
:global([data-theme='dark']) .topnav-brand-picker-panel {
  box-shadow:
    0 18px 38px color-mix(in srgb, var(--color-shadow-strong) 34%, transparent),
    0 0 0 1px color-mix(in srgb, var(--color-border-subtle) 56%, transparent);
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

@media (min-width: 768px) {
  .topnav-page-title {
    font-size: 0.9375rem;
    line-height: 1.375rem;
  }
}
</style>
