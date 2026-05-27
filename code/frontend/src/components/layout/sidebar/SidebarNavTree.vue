<template>
  <nav
    class="backoffice-sidebar__nav flex-1 space-y-1.5 overflow-x-hidden"
    :class="navClass"
  >
    <div
      v-for="item in items"
      :key="item.name"
      class="w-full"
    >
      <button
        type="button"
        class="backoffice-sidebar__item w-full flex items-center justify-between py-2.5 rounded-xl text-sm transition-all overflow-hidden"
        :class="itemButtonClasses(item)"
        :title="collapsed && !mobile ? item.title : ''"
        @click="$emit('navigate', item)"
      >
        <div class="flex items-center gap-3">
          <div
            class="backoffice-sidebar__item-icon shrink-0"
            :class="itemIconClass(item)"
          >
            <component
              :is="item.icon"
              class="backoffice-sidebar__icon-svg"
            />
          </div>
          <span
            class="transition-opacity duration-200 whitespace-nowrap"
            :class="itemLabelClass"
          >
            {{ item.title }}
          </span>
        </div>
        <ChevronDown
          v-if="showChevron(item)"
          class="backoffice-sidebar__chevron h-3.5 w-3.5 transition-transform duration-200"
          :class="{ 'backoffice-sidebar__chevron--open': isItemExpanded(item) }"
          @click.stop="$emit('toggleMenu', item.name)"
        />
      </button>

      <div
        v-if="showChildren(item)"
        class="backoffice-sidebar__children mt-1.5 mb-3 pl-3 flex flex-col gap-1.5 animate-in slide-in-from-top-2 duration-200"
      >
        <button
          v-for="child in item.children"
          :key="child.name"
          type="button"
          class="backoffice-sidebar__child text-left py-2 px-3 rounded-lg transition-all relative group"
          :class="childButtonClass(child)"
          @click="$emit('navigate', child)"
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
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ChevronDown } from 'lucide-vue-next'

import type { NavItem } from './types'

const props = defineProps<{
  items: NavItem[]
  collapsed: boolean
  mobile: boolean
  itemButtonClass: (item: NavItem) => string
  itemIconClass: (item: NavItem) => string
  childButtonClass: (item: NavItem) => string
  isItemExpanded: (item: NavItem) => boolean
  isItemActive: (item: NavItem) => boolean
}>()

defineEmits<{
  navigate: [item: NavItem]
  toggleMenu: [name: string]
}>()

const navClass = computed(() => {
  if (props.mobile) {
    return 'px-4'
  }
  return props.collapsed ? 'px-3 pt-4' : 'px-4'
})

const itemLabelClass = computed(() =>
  props.mobile || !props.collapsed ? 'opacity-100' : 'opacity-0 hidden'
)

function itemButtonClasses(item: NavItem): Array<string> {
  return [
    props.itemButtonClass(item),
    props.mobile ? 'px-3' : props.collapsed ? 'px-0 justify-center' : 'px-3',
  ]
}

function showChevron(item: NavItem): boolean {
  return Boolean(item.children?.length) && (props.mobile || !props.collapsed)
}

function showChildren(item: NavItem): boolean {
  return Boolean(item.children?.length) && props.isItemExpanded(item) && (props.mobile || !props.collapsed)
}
</script>
