<template>
  <aside class="h-screen w-72 shrink-0 border-r border-border bg-surface/60 px-4 py-5">
    <div class="flex items-center justify-between">
      <div class="text-sm font-semibold">CTF 靶场平台</div>
      <div class="text-xs text-text-muted">{{ roleLabel }}</div>
    </div>

    <nav class="mt-6 space-y-6">
      <div v-for="group in navGroups" :key="group.key">
        <div class="px-2 text-xs font-semibold text-text-muted">{{ group.title }}</div>
        <div class="mt-2 space-y-1">
          <button
            v-for="item in group.items"
            :key="item.name"
            type="button"
            class="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-left text-sm transition hover:bg-elevated"
            :class="isActive(item.path) ? 'bg-elevated' : ''"
            @click="navigate(item.path)"
          >
            <component :is="item.icon" class="h-4 w-4 text-text-secondary" />
            <span class="truncate">{{ item.title }}</span>
          </button>
        </div>
      </div>
    </nav>
  </aside>
</template>

<script setup lang="ts">
import type { Component } from 'vue'
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import * as LucideIcons from 'lucide-vue-next'
import { Circle } from 'lucide-vue-next'

import { routes } from '@/router'
import { useAuthStore } from '@/stores/auth'

type IconComp = Component
const iconRegistry = LucideIcons as unknown as Record<string, Component>

type NavItem = {
  name: string
  path: string
  title: string
  icon: IconComp
  roles?: string[]
}

type NavGroup = {
  key: string
  title: string
  items: NavItem[]
}

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const roleLabel = computed(() => authStore.user?.role || '-')

function resolveIcon(name?: string): IconComp {
  if (!name) return Circle
  const icon = iconRegistry[name]
  return icon || Circle
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

  const items: NavItem[] = visible.map((r) => ({
    name: String(r.name || r.path),
    path: r.path.startsWith('/') ? r.path : `/${r.path}`,
    title: String(r.meta?.title || r.name || r.path),
    icon: resolveIcon(String(r.meta?.icon || '')),
    roles: r.meta?.roles as string[] | undefined,
  }))

  const mainItems = items.filter((i) => !i.path.startsWith('/admin/') && !i.path.startsWith('/teacher/'))
  const teacherItems = items.filter((i) => i.path.startsWith('/teacher/'))
  const adminItems = items.filter((i) => i.path.startsWith('/admin/'))

  const groups: NavGroup[] = [{ key: 'main', title: '导航', items: mainItems }]
  if (teacherItems.length > 0) groups.push({ key: 'teacher', title: '教师', items: teacherItems })
  if (adminItems.length > 0) groups.push({ key: 'admin', title: '管理员', items: adminItems })
  return groups
})

function isActive(path: string): boolean {
  return route.path === path
}

async function navigate(path: string): Promise<void> {
  if (route.path === path) return
  await router.push(path)
}
</script>
