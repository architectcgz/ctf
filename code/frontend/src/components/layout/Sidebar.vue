<template>
  <div class="contents">
    <button
      v-if="mobileOpen"
      type="button"
      class="fixed inset-0 z-40 bg-black/55 backdrop-blur-sm md:hidden"
      aria-label="关闭导航"
      @click="emit('closeMobile')"
    />

    <aside
      class="fixed inset-y-0 left-0 z-50 flex w-[var(--shell-sidebar-expanded)] flex-col border-r border-border bg-surface/92 px-3 py-4 shadow-[0_24px_60px_var(--color-shadow-strong)] transition-transform duration-200 md:hidden"
      :class="mobileOpen ? 'translate-x-0' : '-translate-x-full'"
    >
      <div class="flex items-center justify-between gap-3 px-2">
        <div class="flex items-center gap-3">
          <div class="flex h-11 w-11 items-center justify-center rounded-2xl border border-primary/35 bg-primary/10 text-sm font-semibold text-primary">
            CTF
          </div>
          <div>
            <div class="text-sm font-semibold text-text-primary">攻防实训平台</div>
          </div>
        </div>
        <button
          type="button"
          class="inline-flex h-9 w-9 items-center justify-center rounded-xl border border-border bg-elevated text-text-secondary transition hover:border-primary/45 hover:text-text-primary"
          aria-label="关闭导航"
          @click="emit('closeMobile')"
        >
          <X class="h-4 w-4" />
        </button>
      </div>

      <div class="mt-5 min-h-0 flex-1 overflow-y-auto">
        <nav class="space-y-6">
          <section v-for="group in navGroups" :key="group.key" class="space-y-2">
            <div class="px-2 text-[11px] font-semibold uppercase tracking-[0.24em] text-text-muted">
              {{ group.title }}
            </div>
            <div class="space-y-1">
              <template v-for="item in group.items" :key="item.name">
                <div v-if="item.children?.length" class="space-y-1">
                  <div class="flex items-center gap-2">
                    <button
                      type="button"
                      class="flex min-w-0 flex-1 items-center gap-3 rounded-2xl border px-3 py-2.5 text-left text-sm transition"
                      :class="itemClass(item)"
                      @click="navigate(item)"
                    >
                      <component :is="item.icon" class="h-4 w-4 shrink-0" />
                      <span class="truncate">{{ item.title }}</span>
                    </button>
                    <button
                      type="button"
                      class="inline-flex h-10 w-10 items-center justify-center rounded-2xl border border-border bg-[var(--color-bg-base)] text-text-secondary transition hover:border-primary/40 hover:text-text-primary"
                      :aria-label="`${item.title}子菜单`"
                      @click="toggleMenu(item.name)"
                    >
                      <ChevronDown class="h-4 w-4 transition-transform" :class="isMenuExpanded(item.name) ? 'rotate-180' : ''" />
                    </button>
                  </div>

                  <div v-if="isMenuExpanded(item.name)" class="ml-5 space-y-1 border-l border-border-subtle pl-3">
                    <button
                      v-for="child in item.children"
                      :key="child.name"
                      type="button"
                      class="flex w-full items-center gap-3 rounded-2xl border px-3 py-2.5 text-left text-sm transition"
                      :class="childItemClass(child)"
                      @click="navigate(child)"
                    >
                      <component :is="child.icon" class="h-4 w-4 shrink-0" />
                      <span class="truncate">{{ child.title }}</span>
                    </button>
                  </div>
                </div>

                <button
                  v-else
                  type="button"
                  class="flex w-full items-center gap-3 rounded-2xl border px-3 py-2.5 text-left text-sm transition"
                  :class="itemClass(item)"
                  @click="navigate(item)"
                >
                  <component :is="item.icon" class="h-4 w-4 shrink-0" />
                  <span class="truncate">{{ item.title }}</span>
                </button>
              </template>
            </div>
          </section>
        </nav>
      </div>
    </aside>

    <aside
      class="sticky top-0 hidden h-screen shrink-0 border-r border-border bg-surface/78 px-3 py-4 shadow-[inset_-1px_0_0_rgba(255,255,255,0.02)] backdrop-blur md:flex md:flex-col"
      :class="collapsed ? 'w-[var(--shell-sidebar-collapsed)]' : 'w-[var(--shell-sidebar-expanded)]'"
    >
      <div class="flex items-center justify-between gap-2 px-1">
        <button
          type="button"
          class="flex min-w-0 items-center gap-3 rounded-2xl border border-border bg-elevated/85 px-2.5 py-2 text-left transition hover:border-primary/40"
          :class="collapsed ? 'w-12 justify-center px-0' : 'w-full'"
          :title="collapsed ? 'CTF 靶场平台' : undefined"
          @click="emit('toggleCollapse')"
        >
          <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-primary/12 text-sm font-semibold text-primary">
            CTF
          </div>
          <div v-if="!collapsed" class="min-w-0">
            <div class="truncate text-sm font-semibold text-text-primary">攻防实训平台</div>
          </div>
        </button>
      </div>

      <div class="mt-5 min-h-0 flex-1 overflow-y-auto">
        <nav class="space-y-6">
          <section v-for="group in navGroups" :key="group.key" class="space-y-2">
            <div
              class="px-2 text-[11px] font-semibold uppercase tracking-[0.22em] text-text-muted"
              :class="collapsed ? 'text-center' : ''"
            >
              <span v-if="!collapsed">{{ group.title }}</span>
              <span v-else>{{ group.shortTitle }}</span>
            </div>

            <div class="space-y-1">
              <template v-for="item in group.items" :key="item.name">
                <div v-if="item.children?.length && !collapsed" class="space-y-1">
                  <div class="flex items-center gap-2">
                    <button
                      type="button"
                      class="group flex min-w-0 flex-1 items-center gap-3 rounded-2xl border px-3 py-2.5 text-left transition"
                      :class="itemClass(item)"
                      @click="navigate(item)"
                    >
                      <component :is="item.icon" class="h-4 w-4 shrink-0" />
                      <span class="truncate text-sm">{{ item.title }}</span>
                    </button>
                    <button
                      type="button"
                      class="inline-flex h-10 w-10 items-center justify-center rounded-2xl border border-border bg-[var(--color-bg-base)] text-text-secondary transition hover:border-primary/40 hover:text-text-primary"
                      :aria-label="`${item.title}子菜单`"
                      @click="toggleMenu(item.name)"
                    >
                      <ChevronDown class="h-4 w-4 transition-transform" :class="isMenuExpanded(item.name) ? 'rotate-180' : ''" />
                    </button>
                  </div>

                  <div v-if="isMenuExpanded(item.name)" class="ml-5 space-y-1 border-l border-border-subtle pl-3">
                    <button
                      v-for="child in item.children"
                      :key="child.name"
                      type="button"
                      class="flex w-full items-center gap-3 rounded-2xl border px-3 py-2.5 text-left text-sm transition"
                      :class="childItemClass(child)"
                      @click="navigate(child)"
                    >
                      <component :is="child.icon" class="h-4 w-4 shrink-0" />
                      <span class="truncate">{{ child.title }}</span>
                    </button>
                  </div>
                </div>

                <button
                  v-else
                  type="button"
                  class="group flex w-full items-center rounded-2xl border text-left transition"
                  :class="[
                    itemClass(item),
                    collapsed ? 'justify-center px-0 py-3' : 'gap-3 px-3 py-2.5',
                  ]"
                  :title="collapsed ? item.title : undefined"
                  @click="navigate(item)"
                >
                  <component :is="item.icon" class="h-4 w-4 shrink-0" />
                  <span v-if="!collapsed" class="truncate text-sm">{{ item.title }}</span>
                </button>
              </template>
            </div>
          </section>
        </nav>
      </div>

    </aside>
  </div>
</template>

<script setup lang="ts">
import type { Component } from 'vue'
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import {
  BarChart3,
  Bell,
  ChevronDown,
  Circle,
  ClipboardList,
  FileChartColumnIncreasing,
  FileDown,
  GraduationCap,
  LayoutDashboard,
  LayoutList,
  Layers,
  Lightbulb,
  Radar,
  ScanEye,
  Server,
  Settings,
  Shield,
  Swords,
  Timer,
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

function resolveIcon(name?: string): IconComp {
  if (!name) return Circle
  return iconRegistry[name] || Circle
}

function buildDashboardChildren(): NavItem[] {
  return [
    {
      name: 'DashboardRecommendation',
      path: '/dashboard',
      title: '训练建议',
      icon: Lightbulb,
      query: { panel: 'recommendation' },
    },
    {
      name: 'DashboardCategory',
      path: '/dashboard',
      title: '分类进度',
      icon: LayoutList,
      query: { panel: 'category' },
    },
    {
      name: 'DashboardTimeline',
      path: '/dashboard',
      title: '近期动态',
      icon: Timer,
      query: { panel: 'timeline' },
    },
    {
      name: 'DashboardDifficulty',
      path: '/dashboard',
      title: '难度分布',
      icon: FileChartColumnIncreasing,
      query: { panel: 'difficulty' },
    },
  ]
}

const navGroups = computed<NavGroup[]>(() => {
  const root = routes.find((r) => r.path === '/')
  const children = (root?.children || []).filter((r) => !!r.meta?.icon && !!r.meta?.title)

  const role = authStore.user?.role
  const visible = children.filter((r) => {
    const required = r.meta?.roles
    if (!required || required.length === 0) return true
    if (!role) return false
    return required.includes(role)
  })
  const sidebarVisible = visible.filter((r) => !(role === 'teacher' && r.name === 'TeacherDashboard'))

  const items: NavItem[] = sidebarVisible.map((r) => ({
    name: String(r.name || r.path),
    path: r.path.startsWith('/') ? r.path : `/${r.path}`,
    title: String(r.meta?.title || r.name || r.path),
    icon: resolveIcon(String(r.meta?.icon || '')),
    roles: r.meta?.roles as string[] | undefined,
  }))

  if (role === 'student') {
    const dashboardItem = items.find((item) => item.path === '/dashboard')
    if (dashboardItem) {
      dashboardItem.children = buildDashboardChildren()
    }
  }

  const mainItems = items.filter((i) => !i.path.startsWith('/admin/') && !i.path.startsWith('/teacher/'))
  const teacherItems = items.filter((i) => i.path.startsWith('/teacher/'))
  const adminItems = items.filter((i) => i.path.startsWith('/admin/'))

  const groups: NavGroup[] = [{ key: 'main', title: '导航', shortTitle: '导', items: mainItems }]
  if (teacherItems.length > 0) groups.push({ key: 'teacher', title: '教学', shortTitle: '教', items: teacherItems })
  if (adminItems.length > 0) groups.push({ key: 'admin', title: '管理', shortTitle: '管', items: adminItems })
  return groups
})

watch(
  () => route.fullPath,
  () => {
    if (route.path === '/dashboard') {
      expandedMenus.value.Dashboard = true
    }
  },
  { immediate: true },
)

function queryMatches(query?: NavQuery): boolean {
  if (!query) return true
  return Object.entries(query).every(([key, value]) => String(route.query[key] ?? '') === value)
}

function isItemActive(item: NavItem): boolean {
  if (item.children?.some((child) => isItemActive(child))) return true
  if (!(route.path === item.path || route.path.startsWith(`${item.path}/`))) return false
  return queryMatches(item.query)
}

function itemClass(item: NavItem): string {
  return isItemActive(item)
    ? 'border-primary/45 bg-primary/12 text-text-primary shadow-[inset_0_1px_0_rgba(255,255,255,0.04)]'
    : 'border-transparent bg-transparent text-text-secondary hover:border-border hover:bg-elevated/85 hover:text-text-primary'
}

function childItemClass(item: NavItem): string {
  return isItemActive(item)
    ? 'border-primary/45 bg-primary/10 text-text-primary'
    : 'border-transparent bg-transparent text-text-secondary hover:border-border hover:bg-elevated/85 hover:text-text-primary'
}

function isMenuExpanded(name: string): boolean {
  return expandedMenus.value[name] ?? false
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
  const sameQuery = queryMatches(item.query) && Object.keys(targetQuery).length === Object.keys(route.query).filter((key) => typeof route.query[key] === 'string').length
  if (samePath && sameQuery) {
    emit('closeMobile')
    return
  }
  await router.push({ path: item.path, query: targetQuery })
  emit('closeMobile')
}
</script>
