<template>
  <div class="contents">
    <button
      v-if="mobileOpen"
      type="button"
      class="backoffice-sidebar-backdrop fixed inset-0 z-40 md:hidden"
      aria-label="关闭导航"
      @click="emit('closeMobile')"
    />

    <aside
      class="backoffice-sidebar backoffice-sidebar--mobile backoffice-sidebar--expanded fixed inset-y-0 left-0 z-50 flex shrink-0 flex-col transition-all duration-300 md:hidden"
      :class="mobileOpen ? 'translate-x-0' : '-translate-x-full'"
    >
      <div
        class="backoffice-sidebar__header relative flex h-16 items-center px-5 overflow-hidden whitespace-nowrap"
      >
        <div class="flex items-center gap-3">
          <div
            class="backoffice-sidebar__logo-mark flex h-8.5 w-8.5 shrink-0 items-center justify-center rounded-xl shadow-sm"
          >
            <Box class="h-4 w-4" />
          </div>
          <span class="backoffice-sidebar__brand font-black text-lg tracking-tight uppercase">
            {{ brandKicker }}
          </span>
        </div>
        <button
          type="button"
          class="backoffice-sidebar__close ml-auto inline-flex h-9 w-9 items-center justify-center rounded-full p-1.5 shadow-sm transition-all"
          aria-label="关闭导航"
          @click="emit('closeMobile')"
        >
          <X class="h-4 w-4" />
        </button>
      </div>

      <div
        class="backoffice-sidebar__workspace px-6 py-5 overflow-hidden whitespace-nowrap transition-all duration-200"
      >
        <span class="backoffice-sidebar__workspace-label font-black uppercase tracking-widest">
          Workspace
        </span>
      </div>

      <nav class="backoffice-sidebar__nav flex-1 space-y-1.5 overflow-x-hidden px-4">
        <div
          v-for="item in backofficeItems"
          :key="item.name"
          class="w-full"
        >
          <button
            type="button"
            class="backoffice-sidebar__item w-full flex items-center justify-between py-2.5 rounded-xl text-sm transition-all overflow-hidden px-3"
            :class="backofficeItemButtonClass(item)"
            @click="navigate(item)"
          >
            <div class="flex items-center gap-3">
              <div
                class="backoffice-sidebar__item-icon shrink-0"
                :class="backofficeItemIconClass(item)"
              >
                <component
                  :is="item.icon"
                  class="backoffice-sidebar__icon-svg"
                />
              </div>
              <span class="transition-opacity duration-200 whitespace-nowrap">
                {{ item.title }}
              </span>
            </div>
            <ChevronDown
              v-if="item.children?.length"
              class="backoffice-sidebar__chevron h-3.5 w-3.5 transition-transform duration-200"
              :class="{ 'backoffice-sidebar__chevron--open': isBackofficeItemExpanded(item) }"
              @click.stop="toggleMenu(item.name)"
            />
          </button>

          <div
            v-if="item.children?.length && isBackofficeItemExpanded(item)"
            class="backoffice-sidebar__children mt-1.5 mb-3 pl-3 flex flex-col gap-1.5 animate-in slide-in-from-top-2 duration-200"
          >
            <button
              v-for="child in item.children"
              :key="child.name"
              type="button"
              class="backoffice-sidebar__child text-left py-2 px-3 rounded-lg transition-all relative group"
              :class="backofficeChildButtonClass(child)"
              @click="navigate(child)"
            >
              <div
                v-if="isItemActive(child)"
                class="backoffice-sidebar__child-indicator absolute top-1/2 -translate-y-1/2 h-4 rounded-full"
              />
              <span class="relative z-10">{{ child.title }}</span>
            </button>
          </div>
        </div>
      </nav>
    </aside>

    <aside
      class="backoffice-sidebar backoffice-sidebar--desktop sticky top-0 z-[60] hidden min-h-screen shrink-0 self-stretch flex-col transition-all duration-300 md:flex"
      :class="collapsed ? 'w-20' : 'backoffice-sidebar--expanded'"
    >
      <button
        type="button"
        class="backoffice-sidebar__collapse absolute -right-3.5 top-5 rounded-full p-1.5 shadow-sm z-10 transition-all cursor-pointer"
        :aria-label="collapsed ? '展开导航' : '折叠导航'"
        @click="emit('toggleCollapse')"
      >
        <ChevronRight
          v-if="collapsed"
          class="h-3.5 w-3.5"
        />
        <ChevronLeft
          v-else
          class="h-3.5 w-3.5"
        />
      </button>

      <div
        class="backoffice-sidebar__header h-16 flex items-center px-5 overflow-hidden whitespace-nowrap"
      >
        <div class="flex items-center gap-3">
          <div
            class="backoffice-sidebar__logo-mark w-8.5 h-8.5 shrink-0 rounded-xl flex items-center justify-center shadow-sm"
          >
            <Box class="h-4 w-4" />
          </div>
          <span
            class="backoffice-sidebar__brand font-black text-lg tracking-tight uppercase transition-opacity duration-200"
            :class="collapsed ? 'opacity-0' : 'opacity-100'"
          >
            {{ brandKicker }}
          </span>
        </div>
      </div>

      <div
        class="backoffice-sidebar__workspace px-6 py-5 overflow-hidden whitespace-nowrap transition-all duration-200"
        :class="collapsed ? 'opacity-0 h-0 p-0' : 'opacity-100 h-14'"
      >
        <span class="backoffice-sidebar__workspace-label font-black uppercase tracking-widest">
          Workspace
        </span>
      </div>

      <nav
        class="backoffice-sidebar__nav flex-1 space-y-1.5 overflow-x-hidden"
        :class="collapsed ? 'px-3 pt-4' : 'px-4'"
      >
        <div
          v-for="item in backofficeItems"
          :key="item.name"
          class="w-full"
        >
          <button
            type="button"
            class="backoffice-sidebar__item w-full flex items-center justify-between py-2.5 rounded-xl text-sm transition-all overflow-hidden"
            :class="[backofficeItemButtonClass(item), collapsed ? 'px-0 justify-center' : 'px-3']"
            :title="collapsed ? item.title : ''"
            @click="navigate(item)"
          >
            <div class="flex items-center gap-3">
              <div
                class="backoffice-sidebar__item-icon shrink-0"
                :class="backofficeItemIconClass(item)"
              >
                <component
                  :is="item.icon"
                  class="backoffice-sidebar__icon-svg"
                />
              </div>
              <span
                class="transition-opacity duration-200 whitespace-nowrap"
                :class="collapsed ? 'opacity-0 hidden' : 'opacity-100'"
              >
                {{ item.title }}
              </span>
            </div>
            <ChevronDown
              v-if="item.children?.length && !collapsed"
              class="backoffice-sidebar__chevron h-3.5 w-3.5 transition-transform duration-200"
              :class="{ 'backoffice-sidebar__chevron--open': isBackofficeItemExpanded(item) }"
              @click.stop="toggleMenu(item.name)"
            />
          </button>

          <div
            v-if="item.children?.length && isBackofficeItemExpanded(item) && !collapsed"
            class="backoffice-sidebar__children mt-1.5 mb-3 pl-3 flex flex-col gap-1.5 animate-in slide-in-from-top-2 duration-200"
          >
            <button
              v-for="child in item.children"
              :key="child.name"
              type="button"
              class="backoffice-sidebar__child text-left py-2 px-3 rounded-lg transition-all relative group"
              :class="backofficeChildButtonClass(child)"
              @click="navigate(child)"
            >
              <div
                v-if="isItemActive(child)"
                class="backoffice-sidebar__child-indicator absolute top-1/2 -translate-y-1/2 h-4 rounded-full"
              />
              <span class="relative z-10">{{ child.title }}</span>
            </button>
          </div>
        </div>
      </nav>
    </aside>
  </div>
</template>

<script setup lang="ts">
import type { Component } from 'vue'
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import {
  type WorkspaceShellModule,
  useWorkspaceShellNavigation,
} from '@/composables/useWorkspaceShellNavigation'
import {
  Box,
  ChevronLeft,
  ChevronDown,
  ChevronRight,
  Circle,
  GraduationCap,
  LayoutDashboard,
  Shield,
  Swords,
  Trophy,
  User,
  X,
} from 'lucide-vue-next'

import { useAuthStore } from '@/stores/auth'

type IconComp = Component
type NavQuery = Record<string, string>

type NavItem = {
  name: string
  path: string
  title: string
  icon: IconComp
  roles?: string[]
  query?: NavQuery
  children?: NavItem[]
  moduleKey?: string
  variant?: 'default' | 'backoffice-child'
}

type NavGroup = {
  key: string
  title: string
  shortTitle: string
  items: NavItem[]
}

const props = defineProps<{
  collapsed: boolean
  mobileOpen: boolean
}>()

const emit = defineEmits<{
  closeMobile: []
  toggleCollapse: []
}>()

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const expandedMenus = ref<Record<string, boolean>>({})
const shell = useWorkspaceShellNavigation(() => ({
  path: route.path,
  fullPath: route.fullPath,
  role: authStore.user?.role,
  routeName: String(route.name ?? ''),
}))
const brandKicker = computed(() => shell.value.brandKicker)
const currentBackofficeModuleKey = computed(() => shell.value.activeModuleKey)
const currentBackofficeSecondaryRouteName = computed(() => shell.value.activeSecondaryRouteName)
const activeBackofficeMenuName = computed(() =>
  currentBackofficeModuleKey.value ? `backoffice-${currentBackofficeModuleKey.value}` : null
)
const backofficeModuleIconMap: Record<string, IconComp> = {
  training: Swords,
  events: Trophy,
  account: User,
  overview: LayoutDashboard,
  operations: GraduationCap,
  resources: Swords,
  contestOps: Trophy,
  governance: Shield,
}

const defaultNavGroups = computed<NavGroup[]>(() => {
  const items: NavItem[] = shell.value.modules.map((module: WorkspaceShellModule) => ({
    name: `backoffice-${module.key}`,
    path: module.secondaryItems[0]?.path || '/',
    title: module.label,
    icon: backofficeModuleIconMap[module.key],
    moduleKey: module.key,
    children: module.secondaryItems.map((secondaryItem) => ({
      name: secondaryItem.routeName,
      path: secondaryItem.path,
      title: secondaryItem.label,
      icon: Circle,
      moduleKey: module.key,
      variant: 'backoffice-child',
    })),
  }))

  return items.length > 0 ? [{ key: 'backoffice', title: '后台', shortTitle: '台', items }] : []
})

const backofficeNavGroups = defaultNavGroups
const navGroups = backofficeNavGroups
const backofficeItems = computed(() => navGroups.value[0]?.items ?? [])

function queryMatches(query?: NavQuery): boolean {
  if (!query) return true
  return Object.entries(query).every(([key, value]) => String(route.query[key] ?? '') === value)
}

function isItemActive(item: NavItem): boolean {
  if (item.variant === 'backoffice-child') {
    return currentBackofficeSecondaryRouteName.value === item.name
  }
  if (item.moduleKey) {
    return currentBackofficeModuleKey.value === item.moduleKey
  }
  if (item.children?.some((child) => isItemActive(child))) return true
  if (!(route.path === item.path || route.path.startsWith(`${item.path}/`))) return false
  return queryMatches(item.query)
}

function itemClass(item: NavItem): string {
  return isItemActive(item) ? 'sidebar-item-active' : 'sidebar-item-idle'
}

function hasBackofficeChildren(item: NavItem): boolean {
  return (item.children?.length ?? 0) > 0
}

function isBackofficeParentOfActive(item: NavItem): boolean {
  return hasBackofficeChildren(item) && item.children!.some((child) => isItemActive(child))
}

function isBackofficeStandaloneActive(item: NavItem): boolean {
  return !hasBackofficeChildren(item) && isItemActive(item)
}

function isBackofficeParentHighlighted(item: NavItem): boolean {
  return (
    hasBackofficeChildren(item) &&
    (isBackofficeParentOfActive(item) || isBackofficeItemExpanded(item))
  )
}

function backofficeItemButtonClass(item: NavItem): string {
  if (isBackofficeStandaloneActive(item)) {
    return 'backoffice-sidebar__item--active'
  }

  if (isBackofficeParentHighlighted(item)) {
    return 'backoffice-sidebar__item--expanded'
  }

  return 'backoffice-sidebar__item--idle'
}

function backofficeItemIconClass(item: NavItem): string {
  return isBackofficeStandaloneActive(item) || isBackofficeParentHighlighted(item)
    ? 'backoffice-sidebar__item-icon--active'
    : 'backoffice-sidebar__item-icon--idle'
}

function backofficeChildButtonClass(item: NavItem): string {
  return isItemActive(item)
    ? 'backoffice-sidebar__child--active'
    : 'backoffice-sidebar__child--idle'
}

function childItemClass(item: NavItem): string {
  if (item.variant === 'backoffice-child') {
    return isItemActive(item)
      ? 'sidebar-child-active sidebar-child-active--backoffice'
      : 'sidebar-child-idle sidebar-child-idle--backoffice'
  }

  return isItemActive(item) ? 'sidebar-child-active' : 'sidebar-child-idle'
}

function isMenuExpanded(name: string): boolean {
  return expandedMenus.value[name] ?? activeBackofficeMenuName.value === name
}

function isBackofficeItemExpanded(item: NavItem): boolean {
  return (
    expandedMenus.value[item.name] ??
    (isBackofficeParentOfActive(item) || isMenuExpanded(item.name))
  )
}

function toggleMenu(name: string): void {
  expandedMenus.value[name] = !isMenuExpanded(name)
}

async function navigate(item: NavItem): Promise<void> {
  if (item.children?.length) {
    expandedMenus.value[item.name] = true
  }
  const targetQuery = item.query ?? {}
  const samePath = route.path === item.path
  const sameQuery =
    queryMatches(item.query) &&
    Object.keys(targetQuery).length ===
      Object.keys(route.query).filter((key) => typeof route.query[key] === 'string').length
  if (samePath && sameQuery) {
    emit('closeMobile')
    return
  }
  await router.push({ path: item.path, query: targetQuery })
  emit('closeMobile')
}
</script>

<style scoped>
.sidebar-shell {
  border-right: 1px solid color-mix(in srgb, var(--color-border-default) 78%, transparent);
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--color-bg-surface) 97%, var(--color-bg-base)),
      color-mix(in srgb, var(--color-bg-surface) 99%, var(--color-bg-base))
    ),
    radial-gradient(
      circle at top left,
      color-mix(in srgb, var(--color-primary) 7%, transparent),
      transparent 14rem
    );
}

.backoffice-sidebar {
  --backoffice-shell-surface: color-mix(in srgb, var(--color-bg-surface) 94%, var(--color-bg-base));
  --backoffice-shell-surface-subtle: color-mix(
    in srgb,
    var(--color-bg-elevated) 84%,
    var(--color-bg-surface)
  );
  --backoffice-shell-surface-strong: color-mix(
    in srgb,
    var(--color-bg-elevated) 94%,
    var(--color-bg-surface)
  );
  --backoffice-shell-line: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --backoffice-shell-line-strong: color-mix(in srgb, var(--color-border-default) 94%, transparent);
  --backoffice-shell-text: color-mix(in srgb, var(--color-text-primary) 96%, transparent);
  --backoffice-shell-muted: color-mix(in srgb, var(--color-text-secondary) 92%, transparent);
  --backoffice-shell-faint: color-mix(in srgb, var(--color-text-muted) 88%, transparent);
  border-right: 1px solid var(--backoffice-shell-line-strong);
  /* 允许按钮悬浮 */
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  white-space: nowrap;
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--backoffice-shell-surface) 99%, transparent),
      color-mix(in srgb, var(--backoffice-shell-surface-subtle) 96%, transparent)
    ),
    radial-gradient(
      circle at top left,
      color-mix(in srgb, var(--color-primary) 6%, transparent),
      transparent 22rem
    );
}

.backoffice-sidebar-backdrop {
  background: color-mix(in srgb, var(--color-bg-base) 55%, transparent);
}

.backoffice-sidebar--expanded {
  width: 17rem;
}

.backoffice-sidebar__header {
  border-bottom: 1px solid var(--backoffice-shell-line);
}

.backoffice-sidebar__logo-mark {
  border: 1px solid color-mix(in srgb, var(--color-primary) 40%, transparent);
  background: linear-gradient(
    135deg,
    color-mix(in srgb, var(--color-primary) 95%, var(--color-bg-base)),
    color-mix(in srgb, var(--color-primary) 85%, var(--color-bg-base))
  );
  color: var(--color-bg-base);
  box-shadow:
    0 4px 12px color-mix(in srgb, var(--color-primary) 24%, transparent),
    inset 0 1px 1px color-mix(in srgb, var(--color-bg-surface) 40%, transparent);
}

.backoffice-sidebar__brand {
  color: var(--backoffice-shell-text);
  font-weight: 900;
  letter-spacing: -0.01em;
  white-space: nowrap;
  overflow: hidden;
}

.backoffice-sidebar__brand-accent {
  color: var(--color-primary);
}

.backoffice-sidebar__close,
.backoffice-sidebar__collapse {
  border: 1px solid var(--backoffice-shell-line);
  background: var(--backoffice-shell-surface);
  color: var(--backoffice-shell-faint);
  box-shadow: 0 1px 2px color-mix(in srgb, var(--color-shadow-soft) 30%, transparent);
  z-index: 70; /* 提升层级，确保高于 TopNav (50) */
}

.backoffice-sidebar__close:hover,
.backoffice-sidebar__collapse:hover {
  border-color: color-mix(in srgb, var(--color-primary) 40%, var(--backoffice-shell-line));
  background: var(--backoffice-shell-surface-strong);
  color: var(--color-primary);
  transform: scale(1.05);
}

.backoffice-sidebar__workspace-label {
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.12em;
  color: var(--backoffice-shell-faint);
}

.backoffice-sidebar__icon-svg {
  height: 1.125rem;
  width: 1.125rem;
}

.backoffice-sidebar__nav {
  overflow-x: hidden;
}

.backoffice-sidebar__item {
  color: var(--backoffice-shell-muted);
  margin-inline: 0.25rem;
  width: calc(100% - 0.5rem);
  white-space: nowrap;
  overflow: hidden;
}

.backoffice-sidebar__item--idle {
  color: var(--backoffice-shell-muted);
  font-weight: 500;
}

.backoffice-sidebar__item--idle:hover {
  background: color-mix(in srgb, var(--color-primary) 6%, var(--backoffice-shell-surface-strong));
  color: var(--backoffice-shell-text);
}

.backoffice-sidebar__item--expanded {
  color: var(--backoffice-shell-text);
  background: color-mix(in srgb, var(--color-primary) 4%, var(--backoffice-shell-surface));
  font-weight: 700;
}

.backoffice-sidebar__item--active {
  background: var(--backoffice-shell-surface-strong);
  color: var(--color-primary);
  border: 1px solid color-mix(in srgb, var(--color-primary) 22%, transparent);
  font-weight: 700;
  box-shadow:
    0 4px 12px color-mix(in srgb, var(--color-shadow-soft) 40%, transparent),
    0 0 0 1px color-mix(in srgb, var(--color-bg-surface) 80%, transparent);
}

[data-theme='dark'] .backoffice-sidebar__item--active {
  box-shadow:
    0 4px 12px color-mix(in srgb, var(--color-shadow-strong) 24%, transparent),
    0 0 0 1px color-mix(in srgb, var(--color-bg-surface) 5%, transparent);
}

.backoffice-sidebar__item-icon--idle {
  color: var(--backoffice-shell-faint);
}

.backoffice-sidebar__item-icon--active {
  color: var(--color-primary);
}

.backoffice-sidebar__chevron {
  color: color-mix(in srgb, var(--backoffice-shell-faint) 94%, transparent);
}

.backoffice-sidebar__chevron--open {
  color: var(--color-primary);
  transform: rotate(180deg);
}

.backoffice-sidebar__children {
  margin-left: 1.45rem;
  border-left: 1px solid color-mix(in srgb, var(--backoffice-shell-line) 60%, transparent);
  padding-left: 0.5rem;
}

.backoffice-sidebar__child {
  font-size: 0.8125rem;
  line-height: 1.125rem;
  color: var(--backoffice-shell-muted);
}

.backoffice-sidebar__child--idle {
  color: var(--backoffice-shell-muted);
  font-weight: 500;
}

.backoffice-sidebar__child--idle:hover {
  background: color-mix(
    in srgb,
    var(--backoffice-shell-line) 24%,
    var(--backoffice-shell-surface-subtle)
  );
  color: var(--backoffice-shell-text);
}

.backoffice-sidebar__child--active {
  color: var(--color-primary);
  background: color-mix(in srgb, var(--color-primary) 8%, var(--backoffice-shell-surface-strong));
  font-weight: 700;
}

.backoffice-sidebar__child-indicator {
  left: -0.5rem;
  width: 3px;
  border-radius: 99px;
  background: var(--color-primary);
  box-shadow: 0 0 8px var(--color-primary);
}

.sidebar-shell-desktop {
  box-shadow: none;
  backdrop-filter: none;
}

.sidebar-shell-mobile {
  box-shadow: 0 18px 48px color-mix(in srgb, var(--color-shadow-strong) 8%, transparent);
}

.sidebar-brand-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.sidebar-brand-row--framed {
  padding-bottom: 1.25rem;
  border-bottom: 1px solid color-mix(in srgb, var(--color-border-default) 60%, transparent);
}

.sidebar-brand {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 0.75rem;
}

.sidebar-brand-button {
  border: 1px solid transparent;
  border-radius: 16px;
  background: transparent;
}

.sidebar-brand-button:hover {
  background: color-mix(in srgb, var(--color-bg-elevated) 40%, var(--color-bg-surface));
}

.sidebar-brand-mark {
  display: inline-flex;
  height: 2.2rem;
  width: 2.2rem;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  background: color-mix(in srgb, var(--color-primary) 10%, var(--color-bg-elevated));
  font-size: 0.8rem;
  font-weight: 800;
  border: 1px solid color-mix(in srgb, var(--color-primary) 20%, var(--color-border-default));
}

.sidebar-brand-kicker {
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.sidebar-group-title {
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--color-text-muted);
  display: flex;
  align-items: center;
  gap: 0.65rem;
}

.sidebar-group-title::after {
  content: '';
  flex: 1 1 auto;
  height: 1px;
  background: color-mix(in srgb, var(--color-border-default) 40%, transparent);
}

.sidebar-icon-button {
  display: inline-flex;
  height: 2.5rem;
  width: 2.5rem;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 76%, transparent);
  background: color-mix(in srgb, var(--color-bg-base) 48%, var(--color-bg-surface));
  color: var(--color-text-secondary);
  transition: all 0.2s ease;
}

.sidebar-icon-button:hover {
  color: var(--color-primary);
  border-color: color-mix(in srgb, var(--color-primary) 30%, var(--color-border-default));
  background: color-mix(in srgb, var(--color-primary) 6%, var(--color-bg-surface));
  transform: translateY(-1px);
}

.sidebar-item-button,
.sidebar-child-button {
  position: relative;
  border-radius: 12px;
  border: 1px solid transparent;
  min-height: 2.75rem;
}

.sidebar-item-icon-wrap {
  display: inline-flex;
  height: 1.75rem;
  width: 1.75rem;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  border: 1px solid transparent;
  background: transparent;
  transition: all 0.2s ease;
}

.sidebar-item-idle,
.sidebar-child-idle {
  color: var(--color-text-secondary);
  font-weight: 500;
}

.sidebar-item-idle:hover,
.sidebar-child-idle:hover {
  background: color-mix(in srgb, var(--color-bg-elevated) 60%, var(--color-bg-surface));
  color: var(--color-text-primary);
}

.sidebar-item-active,
.sidebar-child-active {
  background: color-mix(in srgb, var(--color-primary) 8%, var(--color-bg-surface));
  border-color: color-mix(in srgb, var(--color-primary) 20%, transparent);
  color: var(--color-primary);
  font-weight: 700;
  box-shadow:
    0 1px 2px color-mix(in srgb, var(--color-shadow-soft) 30%, transparent),
    0 0 0 1px color-mix(in srgb, var(--color-bg-surface) 80%, transparent);
}

[data-theme='dark'] .sidebar-item-active {
  box-shadow: 0 4px 12px color-mix(in srgb, var(--color-shadow-strong) 20%, transparent);
}

.sidebar-child-list {
  border-left: 1px solid color-mix(in srgb, var(--color-border-default) 60%, transparent);
}

:global([data-theme='light']) .sidebar-shell {
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--color-bg-surface) 97%, var(--color-bg-base)),
      color-mix(in srgb, var(--color-bg-elevated) 95%, var(--color-bg-base))
    ),
    radial-gradient(
      circle at top left,
      color-mix(in srgb, var(--color-primary) 6%, transparent),
      transparent 14rem
    );
}

:global([data-theme='light']) .backoffice-sidebar {
  --backoffice-shell-surface: var(--color-bg-surface);
  --backoffice-shell-surface-subtle: var(--color-bg-elevated);
  --backoffice-shell-surface-strong: var(--color-bg-surface);
  --backoffice-shell-line: color-mix(in srgb, var(--color-border-default) 92%, transparent);
  --backoffice-shell-line-strong: color-mix(in srgb, var(--color-border-default) 94%, transparent);
  --backoffice-shell-text: var(--color-text-primary);
  --backoffice-shell-muted: var(--color-text-secondary);
  --backoffice-shell-faint: var(--color-text-muted);
}

:global([data-theme='dark']) .backoffice-sidebar {
  --backoffice-shell-surface: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --backoffice-shell-surface-subtle: color-mix(
    in srgb,
    var(--color-bg-elevated) 84%,
    var(--color-bg-base)
  );
  --backoffice-shell-surface-strong: color-mix(
    in srgb,
    var(--color-bg-elevated) 92%,
    var(--color-bg-base)
  );
  --backoffice-shell-line: color-mix(in srgb, var(--color-border-default) 88%, transparent);
  --backoffice-shell-line-strong: color-mix(in srgb, var(--color-border-default) 94%, transparent);
  --backoffice-shell-text: color-mix(in srgb, var(--color-text-primary) 94%, transparent);
  --backoffice-shell-muted: color-mix(in srgb, var(--color-text-secondary) 90%, transparent);
  --backoffice-shell-faint: color-mix(in srgb, var(--color-text-muted) 90%, transparent);
  box-shadow:
    0 22px 56px color-mix(in srgb, var(--color-shadow-strong) 28%, transparent),
    0 0 0 1px color-mix(in srgb, var(--color-border-subtle) 48%, transparent);
}
</style>

<style scoped>
.backoffice-sidebar__nav {
  overflow-x: hidden;
}

.backoffice-sidebar__collapse {
  z-index: 100;
}
</style>
