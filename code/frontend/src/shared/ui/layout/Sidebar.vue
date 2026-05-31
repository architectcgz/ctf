<template>
  <div class="sidebar-host contents">
    <SidebarMobilePanel
      :mobile-open="mobileOpen"
      :brand-kicker="brandKicker"
      :items="backofficeItems"
      :item-button-class="backofficeItemButtonClass"
      :item-icon-class="backofficeItemIconClass"
      :child-button-class="backofficeChildButtonClass"
      :is-item-expanded="isBackofficeItemExpanded"
      :is-item-active="isItemActive"
      @close-mobile="emit('closeMobile')"
      @navigate="navigate"
      @toggle-menu="toggleMenu"
    />

    <SidebarDesktopPanel
      :collapsed="collapsed"
      :brand-kicker="brandKicker"
      :items="backofficeItems"
      :item-button-class="backofficeItemButtonClass"
      :item-icon-class="backofficeItemIconClass"
      :child-button-class="backofficeChildButtonClass"
      :is-item-expanded="isBackofficeItemExpanded"
      :is-item-active="isItemActive"
      @toggle-collapse="emit('toggleCollapse')"
      @navigate="navigate"
      @toggle-menu="toggleMenu"
    />
  </div>
</template>

<script setup lang="ts">
import SidebarDesktopPanel from './sidebar/SidebarDesktopPanel.vue'
import SidebarMobilePanel from './sidebar/SidebarMobilePanel.vue'
import { useSidebarNavigationViewState } from '@/shared/model/layout/sidebar/useSidebarNavigationViewState'
import './sidebar/sidebarShell.css'

defineProps<{
  collapsed: boolean
  mobileOpen: boolean
}>()

const emit = defineEmits<{
  closeMobile: []
  toggleCollapse: []
}>()

const {
  brandKicker,
  backofficeItems,
  backofficeItemButtonClass,
  backofficeItemIconClass,
  backofficeChildButtonClass,
  isBackofficeItemExpanded,
  isItemActive,
  toggleMenu,
  navigate,
} = useSidebarNavigationViewState({
  closeMobile: () => emit('closeMobile'),
})
</script>
