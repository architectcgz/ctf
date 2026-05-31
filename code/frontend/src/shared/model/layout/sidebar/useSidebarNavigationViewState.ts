import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Circle,
  GraduationCap,
  LayoutDashboard,
  Shield,
  Swords,
  Trophy,
  User,
} from 'lucide-vue-next'

import {
  type WorkspaceShellModule,
  useWorkspaceShellNavigation,
} from '@/shared/model/layout/useWorkspaceShellNavigation'
import { useAuthStore } from '@/stores/auth'

import type { IconComp, NavGroup, NavItem, NavQuery } from './types'

interface UseSidebarNavigationViewStateOptions {
  closeMobile: () => void
}

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

export function useSidebarNavigationViewState({
  closeMobile,
}: UseSidebarNavigationViewStateOptions) {
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

  function hasBackofficeChildren(item: NavItem): boolean {
    return (item.children?.length ?? 0) > 0
  }

  function isBackofficeParentOfActive(item: NavItem): boolean {
    return hasBackofficeChildren(item) && item.children!.some((child) => isItemActive(child))
  }

  function isBackofficeStandaloneActive(item: NavItem): boolean {
    return !hasBackofficeChildren(item) && isItemActive(item)
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
      closeMobile()
      return
    }

    await router.push({ path: item.path, query: targetQuery })
    closeMobile()
  }

  return {
    brandKicker,
    backofficeItems,
    backofficeItemButtonClass,
    backofficeItemIconClass,
    backofficeChildButtonClass,
    isBackofficeItemExpanded,
    isItemActive,
    toggleMenu,
    navigate,
  }
}
