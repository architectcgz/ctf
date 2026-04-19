<template>
  <div class="contents">
    <template v-if="isBackofficeRoute">
      <button
        v-if="mobileOpen"
        type="button"
        class="fixed inset-0 z-40 bg-black/55 md:hidden"
        aria-label="关闭导航"
        @click="emit('closeMobile')"
      />

      <aside
        class="backoffice-sidebar backoffice-sidebar--mobile backoffice-sidebar--expanded fixed inset-y-0 left-0 z-50 flex shrink-0 flex-col transition-all duration-300 md:hidden"
        :class="mobileOpen ? 'translate-x-0' : '-translate-x-full'"
      >
        <div class="backoffice-sidebar__header relative flex h-16 items-center px-5 overflow-hidden whitespace-nowrap">
          <div class="flex items-center gap-3">
            <div class="backoffice-sidebar__logo-mark flex h-8.5 w-8.5 shrink-0 items-center justify-center rounded-xl shadow-sm">
              <Box class="h-4 w-4" />
            </div>
            <span class="backoffice-sidebar__brand font-black text-lg tracking-tight uppercase">
              Challenge<span class="backoffice-sidebar__brand-accent">Ops</span>
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

        <div class="backoffice-sidebar__workspace px-6 py-5 overflow-hidden whitespace-nowrap transition-all duration-200">
          <span class="backoffice-sidebar__workspace-label font-black uppercase tracking-widest">
            Workspace
          </span>
        </div>

        <nav class="backoffice-sidebar__nav flex-1 space-y-1.5 overflow-x-hidden px-4">
          <div v-for="item in backofficeItems" :key="item.name" class="w-full">
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
                  <component :is="item.icon" class="backoffice-sidebar__icon-svg" />
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
        class="backoffice-sidebar backoffice-sidebar--desktop sticky top-0 z-[60] hidden h-screen shrink-0 flex-col transition-all duration-300 md:flex"
        :class="collapsed ? 'w-20' : 'backoffice-sidebar--expanded'"
      >
        <button
          type="button"
          class="backoffice-sidebar__collapse absolute -right-3.5 top-5 rounded-full p-1.5 shadow-sm z-10 transition-all cursor-pointer"
          :aria-label="collapsed ? '展开导航' : '折叠导航'"
          @click="emit('toggleCollapse')"
        >
          <ChevronRight v-if="collapsed" class="h-3.5 w-3.5" />
          <ChevronLeft v-else class="h-3.5 w-3.5" />
        </button>

        <div
          class="backoffice-sidebar__header h-16 flex items-center px-5 overflow-hidden whitespace-nowrap"
        >
          <div class="flex items-center gap-3">
            <div class="backoffice-sidebar__logo-mark w-8.5 h-8.5 shrink-0 rounded-xl flex items-center justify-center shadow-sm">
              <Box class="h-4 w-4" />
            </div>
            <span
              class="backoffice-sidebar__brand font-black text-lg tracking-tight uppercase transition-opacity duration-200"
              :class="collapsed ? 'opacity-0' : 'opacity-100'"
            >
              Challenge<span class="backoffice-sidebar__brand-accent">Ops</span>
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
          <div v-for="item in backofficeItems" :key="item.name" class="w-full">
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
                  <component :is="item.icon" class="backoffice-sidebar__icon-svg" />
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
    </template>

    <template v-else>
      <button
        v-if="mobileOpen"
        type="button"
        class="fixed inset-0 z-40 bg-black/55 md:hidden"
        aria-label="关闭导航"
        @click="emit('closeMobile')"
      />

      <aside
        class="sidebar-shell sidebar-shell-mobile fixed inset-y-0 left-0 z-50 flex w-[var(--shell-sidebar-expanded)] flex-col px-3 py-4 transition-transform duration-200 md:hidden"
        :class="mobileOpen ? 'translate-x-0' : '-translate-x-full'"
      >
        <div class="sidebar-brand-row sidebar-brand-row--framed px-2">
          <div class="sidebar-brand">
            <div class="sidebar-brand-mark tech-accent">
              {{ brandMark }}
            </div>
            <div class="min-w-0">
              <div class="sidebar-brand-kicker tech-accent">{{ brandKicker }}</div>
              <div class="truncate text-sm font-semibold text-text-primary">{{ brandTitle }}</div>
              <div class="truncate text-xs text-text-muted">{{ roleBadge }}</div>
            </div>
          </div>
          <button
            type="button"
            class="sidebar-icon-button"
            aria-label="关闭导航"
            @click="emit('closeMobile')"
          >
            <X class="h-4 w-4" />
          </button>
        </div>

        <div class="sidebar-nav-scroll mt-6 flex min-h-0 flex-1 overflow-y-auto">
          <nav class="sidebar-nav-list flex min-h-full flex-col space-y-7">
            <section v-for="group in navGroups" :key="group.key" class="sidebar-group space-y-2.5">
              <div class="sidebar-group-title px-2">
                <span>{{ group.title }}</span>
              </div>
              <div class="space-y-1.5">
                <template v-for="item in group.items" :key="item.name">
                  <div v-if="item.children?.length" class="space-y-1.5">
                    <div class="flex items-center gap-2">
                      <button
                        type="button"
                        class="sidebar-item-button flex min-w-0 flex-1 items-center gap-3 px-3 py-2.5 text-left text-sm transition"
                        :class="itemClass(item)"
                        @click="navigate(item)"
                      >
                        <span class="sidebar-item-icon-wrap">
                          <component :is="item.icon" class="sidebar-item-icon h-4 w-4 shrink-0" />
                        </span>
                        <span class="truncate">{{ item.title }}</span>
                      </button>
                      <button
                        type="button"
                        class="sidebar-icon-button"
                        :aria-label="`${item.title}子菜单`"
                        @click="toggleMenu(item.name)"
                      >
                        <ChevronDown
                          class="h-4 w-4 transition-transform"
                          :class="isMenuExpanded(item.name) ? 'rotate-180' : ''"
                        />
                      </button>
                    </div>

                    <div
                      v-if="isMenuExpanded(item.name)"
                      class="sidebar-child-list ml-5 space-y-1.5 pl-3"
                    >
                      <button
                        v-for="child in item.children"
                        :key="child.name"
                        type="button"
                        class="sidebar-child-button flex w-full items-center gap-3 px-3 py-2.5 text-left text-sm transition"
                        :class="childItemClass(child)"
                        @click="navigate(child)"
                      >
                        <span class="sidebar-item-icon-wrap sidebar-item-icon-wrap-child">
                          <component :is="child.icon" class="sidebar-item-icon h-4 w-4 shrink-0" />
                        </span>
                        <span class="truncate">{{ child.title }}</span>
                      </button>
                    </div>
                  </div>

                  <button
                    v-else
                    type="button"
                    class="sidebar-item-button flex w-full items-center gap-3 px-3 py-2.5 text-left text-sm transition"
                    :class="itemClass(item)"
                    @click="navigate(item)"
                  >
                    <span class="sidebar-item-icon-wrap">
                      <component :is="item.icon" class="sidebar-item-icon h-4 w-4 shrink-0" />
                    </span>
                    <span class="truncate">{{ item.title }}</span>
                  </button>
                </template>
              </div>
            </section>
          </nav>
        </div>
      </aside>

      <aside
        class="sidebar-shell sidebar-shell-desktop sticky top-0 hidden min-h-screen shrink-0 self-stretch px-3 py-4 md:flex md:flex-col"
        :class="collapsed ? 'w-[var(--shell-sidebar-collapsed)]' : 'w-[var(--shell-sidebar-expanded)]'"
      >
        <div class="sidebar-brand-row sidebar-brand-row--framed px-1">
          <button
            type="button"
            class="sidebar-brand-button flex min-w-0 items-center gap-3 px-2.5 py-2 text-left transition"
            :class="
              collapsed
                ? 'w-12 justify-center px-0 border-transparent bg-transparent shadow-none'
                : 'w-full'
            "
            :title="collapsed ? brandTooltip : undefined"
            @click="emit('toggleCollapse')"
          >
            <div
              class="sidebar-brand-mark tech-accent"
              :style="collapsed ? { background: 'transparent', boxShadow: 'none' } : {}"
            >
              {{ brandMark }}
            </div>
            <div v-if="!collapsed" class="min-w-0">
              <div class="sidebar-brand-kicker tech-accent">{{ brandKicker }}</div>
              <div class="truncate text-sm font-semibold text-text-primary">{{ brandTitle }}</div>
              <div class="truncate text-xs text-text-muted">{{ roleBadge }}</div>
            </div>
          </button>
        </div>

        <div class="sidebar-nav-scroll mt-6 flex min-h-0 flex-1 overflow-y-auto">
          <nav class="sidebar-nav-list flex min-h-full flex-col space-y-7">
            <section v-for="group in navGroups" :key="group.key" class="sidebar-group space-y-2.5">
              <div v-if="!collapsed" class="sidebar-group-title px-2">
                <span>{{ group.title }}</span>
              </div>
              <div
                v-else
                class="sidebar-group-title sidebar-group-title--collapsed"
                :title="group.title"
              >
                {{ group.shortTitle }}
              </div>

              <div class="space-y-1.5">
                <template v-for="item in group.items" :key="item.name">
                  <div v-if="item.children?.length && !collapsed" class="space-y-1.5">
                    <div class="flex items-center gap-2">
                      <button
                        type="button"
                        class="sidebar-item-button group flex min-w-0 flex-1 items-center gap-3 px-3 py-2.5 text-left transition"
                        :class="itemClass(item)"
                        @click="navigate(item)"
                      >
                        <span class="sidebar-item-icon-wrap">
                          <component :is="item.icon" class="sidebar-item-icon h-4 w-4 shrink-0" />
                        </span>
                        <span class="truncate text-sm">{{ item.title }}</span>
                      </button>
                      <button
                        type="button"
                        class="sidebar-icon-button"
                        :aria-label="`${item.title}子菜单`"
                        @click="toggleMenu(item.name)"
                      >
                        <ChevronDown
                          class="h-4 w-4 transition-transform"
                          :class="isMenuExpanded(item.name) ? 'rotate-180' : ''"
                        />
                      </button>
                    </div>

                    <div
                      v-if="isMenuExpanded(item.name)"
                      class="sidebar-child-list ml-5 space-y-1.5 pl-3"
                    >
                      <button
                        v-for="child in item.children"
                        :key="child.name"
                        type="button"
                        class="sidebar-child-button flex w-full items-center gap-3 px-3 py-2.5 text-left text-sm transition"
                        :class="childItemClass(child)"
                        @click="navigate(child)"
                      >
                        <span class="sidebar-item-icon-wrap sidebar-item-icon-wrap-child">
                          <component :is="child.icon" class="sidebar-item-icon h-4 w-4 shrink-0" />
                        </span>
                        <span class="truncate">{{ child.title }}</span>
                      </button>
                    </div>
                  </div>

                  <button
                    v-else
                    type="button"
                    class="sidebar-item-button group flex w-full items-center text-left transition"
                    :class="[
                      itemClass(item),
                      collapsed
                        ? 'sidebar-item-button--collapsed justify-center px-0 py-3'
                        : 'gap-3 px-3 py-2.5',
                    ]"
                    :title="collapsed ? item.title : undefined"
                    @click="navigate(item)"
                  >
                    <span
                      class="sidebar-item-icon-wrap"
                      :class="collapsed ? 'sidebar-item-icon-wrap-collapsed' : ''"
                    >
                      <component :is="item.icon" class="sidebar-item-icon h-4 w-4 shrink-0" />
                    </span>
                    <span v-if="!collapsed" class="truncate text-sm">{{ item.title }}</span>
                  </button>
                </template>
              </div>
            </section>
          </nav>
        </div>
      </aside>
    </template>
  </div>
</template>

<script setup lang="ts">
import type { Component } from 'vue'
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import {
  type BackofficeModuleKey,
  getBackofficeActiveSecondaryRouteName,
  getBackofficeModuleByPath,
  getVisibleBackofficeSecondaryItems,
  getVisibleBackofficeModules,
} from '@/config/backofficeNavigation'
import {
  BarChart3,
  Bell,
  Box,
  ChevronLeft,
  ChevronDown,
  ChevronRight,
  Circle,
  ClipboardList,
  FileDown,
  GraduationCap,
  LayoutDashboard,
  Layers,
  Radar,
  ScanEye,
  Server,
  Settings,
  Shield,
  Swords,
  Trophy,
  User,
  Users,
  X,
} from 'lucide-vue-next'

import { routes } from '@/router'
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
  moduleKey?: BackofficeModuleKey
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

const iconRegistry: Record<string, Component> = {
  BarChart3,
  Bell,
  ClipboardList,
  FileDown,
  GraduationCap,
  LayoutDashboard,
  Layers,
  Radar,
  ScanEye,
  Server,
  Settings,
  Shield,
  Swords,
  Trophy,
  User,
  Users,
}
const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const expandedMenus = ref<Record<string, boolean>>({})
const isBackofficeRoute = computed(
  () => route.path.startsWith('/academy/') || route.path.startsWith('/platform/')
)
const brandMark = computed(() => (isBackofficeRoute.value ? 'OPS' : 'CTF'))
const brandKicker = computed(() => (isBackofficeRoute.value ? 'ChallengeOps' : 'Student Space'))
const brandTitle = computed(() => (isBackofficeRoute.value ? '后台工作台' : '攻防实训平台'))
const brandTooltip = computed(() => (isBackofficeRoute.value ? 'ChallengeOps 后台' : 'CTF 靶场平台'))
const roleBadge = computed(() => {
  const role = authStore.user?.role
  if (isBackofficeRoute.value) {
    if (role === 'admin') return 'Administrator Panel'
    if (role === 'teacher') return 'Instructor Access'
    return 'Platform Console'
  }
  if (role === 'admin') return 'Administrator Panel'
  if (role === 'teacher') return 'Instructor Workspace'
  return 'Student Console'
})
const currentBackofficeModuleKey = computed(() => getBackofficeModuleByPath(route.path)?.key ?? null)
const currentBackofficeSecondaryRouteName = computed(() => getBackofficeActiveSecondaryRouteName(route.path))
const activeBackofficeMenuName = computed(() =>
  currentBackofficeModuleKey.value ? `backoffice-${currentBackofficeModuleKey.value}` : null
)
const backofficeModuleIconMap: Record<BackofficeModuleKey, IconComp> = {
  overview: LayoutDashboard,
  operations: GraduationCap,
  resources: Swords,
  contestOps: Trophy,
  governance: Shield,
}

function resolveIcon(name?: string): IconComp {
  if (!name) return Circle
  return iconRegistry[name] || Circle
}

const defaultNavGroups = computed<NavGroup[]>(() => {
  const root = routes.find((r) => r.path === '/')
  const children = (root?.children || []).filter(
    (r) => !!r.meta?.icon && !!r.meta?.title && !r.path.includes(':')
  )

  const role = authStore.user?.role
  const visible = children.filter((r) => {
    const required = r.meta?.roles
    if (!required || required.length === 0) return true
    if (!role) return false
    return required.includes(role)
  })
  const sidebarVisible = visible.filter(
    (r) => !(role === 'teacher' && r.name === 'TeacherDashboard')
  )

  const items: NavItem[] = sidebarVisible.map((r) => ({
    name: String(r.name || r.path),
    path: r.path.startsWith('/') ? r.path : `/${r.path}`,
    title: String(r.meta?.title || r.name || r.path),
    icon: resolveIcon(String(r.meta?.icon || '')),
    roles: r.meta?.roles as string[] | undefined,
  }))

  const mainItems = items.filter(
    (i) =>
      !i.path.startsWith('/academy/') &&
      !i.path.startsWith('/platform/')
  )
  const teacherItems = items.filter((i) => i.path.startsWith('/academy/'))
  const adminItems = items.filter((i) => i.path.startsWith('/platform/'))

  const groups: NavGroup[] = [{ key: 'main', title: '导航', shortTitle: '导', items: mainItems }]
  if (teacherItems.length > 0)
    groups.push({ key: 'teacher', title: '教学', shortTitle: '教', items: teacherItems })
  if (adminItems.length > 0)
    groups.push({ key: 'admin', title: '管理', shortTitle: '管', items: adminItems })
  return groups
})

const backofficeNavGroups = computed<NavGroup[]>(() => {
  const modules = getVisibleBackofficeModules(authStore.user?.role)
  const activeSecondaryItems = getVisibleBackofficeSecondaryItems(
    route.path,
    authStore.user?.role ?? null
  )
  const items: NavItem[] = modules.map((module) => ({
    name: `backoffice-${module.key}`,
    path: module.secondaryItems[0]?.path || '/',
    title: module.label,
    icon: backofficeModuleIconMap[module.key],
    moduleKey: module.key,
    children: module.secondaryItems.map((secondaryItem) => ({
      name: secondaryItem.routeName,
      path: secondaryItem.path,
      title:
        activeSecondaryItems.find((item) => item.routeName === secondaryItem.routeName)?.label ??
        secondaryItem.label,
      icon: Circle,
      moduleKey: module.key,
      variant: 'backoffice-child',
    })),
  }))

  return items.length > 0
    ? [{ key: 'backoffice', title: '后台', shortTitle: '台', items }]
    : []
})

const navGroups = computed<NavGroup[]>(() =>
  isBackofficeRoute.value ? backofficeNavGroups.value : defaultNavGroups.value
)
const backofficeItems = computed(() => backofficeNavGroups.value[0]?.items ?? [])

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
  return hasBackofficeChildren(item) && (isBackofficeParentOfActive(item) || isBackofficeItemExpanded(item))
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
  return expandedMenus.value[item.name] ?? (isBackofficeParentOfActive(item) || isMenuExpanded(item.name))
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
  --backoffice-shell-surface-subtle: color-mix(in srgb, var(--color-bg-elevated) 84%, var(--color-bg-surface));
  --backoffice-shell-surface-strong: color-mix(in srgb, var(--color-bg-elevated) 94%, var(--color-bg-surface));
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
    color-mix(in srgb, var(--color-primary) 95%, black),
    color-mix(in srgb, var(--color-primary) 85%, black)
  );
  color: white;
  box-shadow: 
    0 4px 12px color-mix(in srgb, var(--color-primary) 24%, transparent),
    inset 0 1px 1px rgba(255, 255, 255, 0.4);
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
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
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
    0 4px 12px rgba(0, 0, 0, 0.06),
    0 0 0 1px rgba(255, 255, 255, 0.8);
}

[data-theme='dark'] .backoffice-sidebar__item--active {
  box-shadow: 
    0 4px 12px rgba(0, 0, 0, 0.24),
    0 0 0 1px rgba(255, 255, 255, 0.05);
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
  background: color-mix(in srgb, var(--backoffice-shell-line) 24%, var(--backoffice-shell-surface-subtle));
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
  box-shadow: 0 18px 48px rgba(0, 0, 0, 0.08);
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
    0 1px 2px rgba(0, 0, 0, 0.05),
    0 0 0 1px rgba(255, 255, 255, 0.8);
}

[data-theme='dark'] .sidebar-item-active {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
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
  --backoffice-shell-surface: white;
  --backoffice-shell-surface-subtle: #f8fafc;
  --backoffice-shell-surface-strong: white;
  --backoffice-shell-line: color-mix(in srgb, #e2e8f0 92%, transparent);
  --backoffice-shell-line-strong: color-mix(in srgb, #d9e1ec 94%, transparent);
  --backoffice-shell-text: #0f172a;
  --backoffice-shell-muted: #64748b;
  --backoffice-shell-faint: #94a3b8;
}

:global([data-theme='dark']) .backoffice-sidebar {
  --backoffice-shell-surface: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --backoffice-shell-surface-subtle: color-mix(in srgb, var(--color-bg-elevated) 84%, var(--color-bg-surface));
  --backoffice-shell-surface-strong: color-mix(in srgb, var(--color-bg-elevated) 92%, var(--color-bg-surface));
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
  z-index: 100 !important; /* 确保不被顶部导航遮挡 */
}
</style>
