<template>
  <aside
    class="backoffice-sidebar backoffice-sidebar--desktop sticky top-0 z-[60] hidden min-h-screen shrink-0 self-stretch flex-col transition-all duration-300 md:flex"
    :class="collapsed ? 'w-20' : 'backoffice-sidebar--expanded'"
  >
    <button
      type="button"
      class="backoffice-sidebar__collapse absolute -right-3.5 top-5 rounded-full p-1.5 shadow-sm z-10 transition-all cursor-pointer"
      :aria-label="collapsed ? '展开导航' : '折叠导航'"
      @click="$emit('toggleCollapse')"
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

    <SidebarPanelHeader
      :brand-kicker="brandKicker"
      :collapsed="collapsed"
      :mobile="false"
    />
    <SidebarWorkspaceLabel :collapsed="collapsed" :mobile="false" />
    <SidebarNavTree
      :items="items"
      :collapsed="collapsed"
      :mobile="false"
      :item-button-class="itemButtonClass"
      :item-icon-class="itemIconClass"
      :child-button-class="childButtonClass"
      :is-item-expanded="isItemExpanded"
      :is-item-active="isItemActive"
      @navigate="$emit('navigate', $event)"
      @toggle-menu="$emit('toggleMenu', $event)"
    />
  </aside>
</template>

<script setup lang="ts">
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'

import SidebarNavTree from './SidebarNavTree.vue'
import SidebarPanelHeader from './SidebarPanelHeader.vue'
import SidebarWorkspaceLabel from './SidebarWorkspaceLabel.vue'
import type { NavItem } from '@/shared/model/layout/sidebar/types'

defineProps<{
  collapsed: boolean
  brandKicker: string
  items: NavItem[]
  itemButtonClass: (item: NavItem) => string
  itemIconClass: (item: NavItem) => string
  childButtonClass: (item: NavItem) => string
  isItemExpanded: (item: NavItem) => boolean
  isItemActive: (item: NavItem) => boolean
}>()

defineEmits<{
  toggleCollapse: []
  navigate: [item: NavItem]
  toggleMenu: [name: string]
}>()
</script>
