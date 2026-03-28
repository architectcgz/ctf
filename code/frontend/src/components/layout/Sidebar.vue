<template>
  <div class="contents">
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
      <div class="sidebar-brand-row px-2">
        <div class="sidebar-brand">
          <div class="sidebar-brand-mark tech-accent">
            CTF
          </div>
          <div class="min-w-0">
            <div class="sidebar-brand-kicker tech-accent">Academic Ops</div>
            <div class="truncate text-sm font-semibold text-text-primary">攻防实训平台</div>
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

      <div class="mt-6 min-h-0 flex-1 overflow-y-auto">
        <nav class="space-y-7">
          <section v-for="group in navGroups" :key="group.key" class="space-y-2.5">
            <div class="sidebar-group-title px-2">
              {{ group.title }}
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
                      <ChevronDown class="h-4 w-4 transition-transform" :class="isMenuExpanded(item.name) ? 'rotate-180' : ''" />
                    </button>
                  </div>

                  <div v-if="isMenuExpanded(item.name)" class="sidebar-child-list ml-5 space-y-1.5 pl-3">
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
      class="sidebar-shell sidebar-shell-desktop sticky top-0 hidden h-screen shrink-0 px-3 py-4 md:flex md:flex-col"
      :class="collapsed ? 'w-[var(--shell-sidebar-collapsed)]' : 'w-[var(--shell-sidebar-expanded)]'"
    >
      <div class="sidebar-brand-row px-1">
        <button
          type="button"
          class="sidebar-brand-button flex min-w-0 items-center gap-3 px-2.5 py-2 text-left transition"
          :class="collapsed ? 'w-12 justify-center px-0 border-transparent bg-transparent shadow-none' : 'w-full'"
          :title="collapsed ? 'CTF 靶场平台' : undefined"
          @click="emit('toggleCollapse')"
        >
          <div
            class="sidebar-brand-mark tech-accent"
            :style="collapsed ? { background: 'transparent', boxShadow: 'none' } : {}"
          >
            CTF
          </div>
          <div v-if="!collapsed" class="min-w-0">
            <div class="sidebar-brand-kicker tech-accent">Academic Ops</div>
            <div class="truncate text-sm font-semibold text-text-primary">攻防实训平台</div>
            <div class="truncate text-xs text-text-muted">{{ roleBadge }}</div>
          </div>
        </button>
      </div>

      <div class="mt-6 min-h-0 flex-1 overflow-y-auto">
        <nav class="space-y-7">
          <section v-for="group in navGroups" :key="group.key" class="space-y-2.5">
            <div
              v-if="!collapsed"
              class="sidebar-group-title px-2"
            >
              {{ group.title }}
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
                      <ChevronDown class="h-4 w-4 transition-transform" :class="isMenuExpanded(item.name) ? 'rotate-180' : ''" />
                    </button>
                  </div>

                  <div v-if="isMenuExpanded(item.name)" class="sidebar-child-list ml-5 space-y-1.5 pl-3">
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
                    collapsed ? 'justify-center px-0 py-3' : 'gap-3 px-3 py-2.5',
                  ]"
                  :title="collapsed ? item.title : undefined"
                  @click="navigate(item)"
                >
                  <span class="sidebar-item-icon-wrap" :class="collapsed ? 'sidebar-item-icon-wrap-collapsed' : ''">
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
const roleBadge = computed(() => {
  const role = authStore.user?.role
  if (role === 'admin') return 'Administrator Panel'
  if (role === 'teacher') return 'Instructor Workspace'
  return 'Student Console'
})

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
    ? 'sidebar-item-active'
    : 'sidebar-item-idle'
}

function childItemClass(item: NavItem): string {
  return isItemActive(item)
    ? 'sidebar-child-active'
    : 'sidebar-child-idle'
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

<style scoped>
.sidebar-shell {
  border-right: 1px solid color-mix(in srgb, var(--color-border-default) 78%, transparent);
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--color-bg-surface) 96%, var(--color-bg-base)), color-mix(in srgb, var(--color-bg-surface) 98%, var(--color-bg-base))),
    radial-gradient(circle at top left, color-mix(in srgb, var(--color-primary) 12%, transparent), transparent 14rem);
}

.sidebar-shell-desktop {
  box-shadow: inset -1px 0 0 color-mix(in srgb, var(--color-border-subtle) 88%, transparent);
}

.sidebar-shell-mobile {
  box-shadow: 0 18px 48px rgba(15, 23, 42, 0.18);
}

.sidebar-brand-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.sidebar-brand {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 0.75rem;
}

.sidebar-brand-button {
  border: 1px solid color-mix(in srgb, var(--color-border-default) 74%, transparent);
  border-radius: 18px;
  background: color-mix(in srgb, var(--color-bg-surface) 72%, var(--color-bg-base));
  box-shadow: 0 10px 28px rgba(15, 23, 42, 0.06);
}

.sidebar-brand-button:hover {
  border-color: color-mix(in srgb, var(--color-primary) 34%, var(--color-border-default));
}

.sidebar-brand-mark {
  display: inline-flex;
  height: 2.6rem;
  width: 2.6rem;
  align-items: center;
  justify-content: center;
  border-radius: 14px;
  background: color-mix(in srgb, var(--color-primary) 12%, var(--color-bg-surface));
  font-size: 0.85rem;
  font-weight: 700;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.sidebar-brand-kicker,
.sidebar-group-title,
.sidebar-footer-title,
.tech-accent {
  font-family: "JetBrains Mono", "Fira Code", "SFMono-Regular", monospace;
}

.sidebar-brand-kicker,
.sidebar-group-title,
.sidebar-footer-title {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.sidebar-icon-button {
  display: inline-flex;
  height: 2.5rem;
  width: 2.5rem;
  align-items: center;
  justify-content: center;
  border-radius: 14px;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 78%, transparent);
  background: color-mix(in srgb, var(--color-bg-base) 66%, var(--color-bg-surface));
  color: var(--color-text-secondary);
  transition: all 0.2s ease;
}

.sidebar-icon-button:hover {
  color: var(--color-text-primary);
  border-color: color-mix(in srgb, var(--color-primary) 34%, var(--color-border-default));
  box-shadow: 0 0 18px color-mix(in srgb, var(--color-primary) 14%, transparent);
}

.sidebar-item-button,
.sidebar-child-button {
  border-radius: 16px;
  border: 1px solid transparent;
}

.sidebar-item-icon-wrap {
  display: inline-flex;
  height: 2rem;
  width: 2rem;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 68%, transparent);
  background: color-mix(in srgb, var(--color-bg-base) 72%, var(--color-bg-surface));
  transition: all 0.2s ease;
}

.sidebar-item-icon-wrap-child {
  height: 1.8rem;
  width: 1.8rem;
  border-radius: 10px;
}

.sidebar-item-icon-wrap-collapsed {
  margin: 0 auto;
}

.sidebar-item-icon {
  color: currentColor;
}

.sidebar-item-idle,
.sidebar-child-idle {
  color: var(--color-text-secondary);
}

.sidebar-item-idle:hover,
.sidebar-child-idle:hover {
  border-color: color-mix(in srgb, var(--color-border-default) 88%, transparent);
  background: color-mix(in srgb, var(--color-bg-base) 62%, var(--color-bg-surface));
  color: var(--color-text-primary);
  box-shadow: 0 0 24px color-mix(in srgb, var(--color-primary) 10%, transparent);
}

.sidebar-item-idle:hover .sidebar-item-icon-wrap,
.sidebar-child-idle:hover .sidebar-item-icon-wrap {
  border-color: transparent;
  box-shadow: none;
}

.sidebar-item-active,
.sidebar-child-active {
  border-color: color-mix(in srgb, var(--color-primary) 26%, var(--color-border-default));
  background:
    linear-gradient(90deg, color-mix(in srgb, var(--color-primary) 12%, transparent), transparent 70%),
    color-mix(in srgb, var(--color-bg-base) 58%, var(--color-bg-surface));
  color: var(--color-text-primary);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.04);
}

.sidebar-item-active .sidebar-item-icon-wrap,
.sidebar-child-active .sidebar-item-icon-wrap {
  border-color: transparent;
  background: color-mix(in srgb, var(--color-primary) 12%, var(--color-bg-surface));
  color: var(--color-primary);
  box-shadow: none;
}

.sidebar-child-list {
  border-left: 1px solid color-mix(in srgb, var(--color-border-subtle) 88%, transparent);
}

:global([data-theme="light"]) .sidebar-shell {
  background:
    linear-gradient(180deg, #ffffff, #f8fafc),
    radial-gradient(circle at top left, rgba(99, 102, 241, 0.08), transparent 14rem);
}

:global([data-theme="light"]) .sidebar-brand-button,
:global([data-theme="light"]) .sidebar-footer-card,
:global([data-theme="light"]) .sidebar-icon-button,
:global([data-theme="light"]) .sidebar-item-icon-wrap {
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.04);
}
</style>
