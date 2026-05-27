<template>
  <div>
    <button
      v-if="mobileOpen"
      type="button"
      class="backoffice-sidebar-backdrop fixed inset-0 z-40 md:hidden"
      aria-label="关闭导航"
      @click="$emit('closeMobile')"
    />

    <aside
      class="backoffice-sidebar backoffice-sidebar--mobile backoffice-sidebar--expanded fixed inset-y-0 left-0 z-50 flex shrink-0 flex-col transition-all duration-300 md:hidden"
      :class="mobileOpen ? 'translate-x-0' : '-translate-x-full'"
    >
      <SidebarPanelHeader
        :brand-kicker="brandKicker"
        :collapsed="false"
        mobile
        @close-mobile="$emit('closeMobile')"
      />
      <SidebarWorkspaceLabel :collapsed="false" mobile />
      <SidebarNavTree
        :items="items"
        :collapsed="false"
        mobile
        :item-button-class="itemButtonClass"
        :item-icon-class="itemIconClass"
        :child-button-class="childButtonClass"
        :is-item-expanded="isItemExpanded"
        :is-item-active="isItemActive"
        @navigate="$emit('navigate', $event)"
        @toggle-menu="$emit('toggleMenu', $event)"
      />
    </aside>
  </div>
</template>

<script setup lang="ts">
import SidebarNavTree from './SidebarNavTree.vue'
import SidebarPanelHeader from './SidebarPanelHeader.vue'
import SidebarWorkspaceLabel from './SidebarWorkspaceLabel.vue'
import type { NavItem } from './types'

defineProps<{
  mobileOpen: boolean
  brandKicker: string
  items: NavItem[]
  itemButtonClass: (item: NavItem) => string
  itemIconClass: (item: NavItem) => string
  childButtonClass: (item: NavItem) => string
  isItemExpanded: (item: NavItem) => boolean
  isItemActive: (item: NavItem) => boolean
}>()

defineEmits<{
  closeMobile: []
  navigate: [item: NavItem]
  toggleMenu: [name: string]
}>()
</script>
