<template>
  <button
    type="button"
    class="topnav-icon-button"
    aria-label="切换主题色"
    aria-controls="topnav-brand-picker-panel"
    :aria-expanded="open ? 'true' : 'false'"
    :title="`当前主题色：${currentBrandLabel}`"
    @click="$emit('toggle')"
  >
    <Palette class="h-4 w-4" />
  </button>

  <div
    v-if="open"
    id="topnav-brand-picker-panel"
    class="topnav-brand-picker-panel"
    role="menu"
    aria-label="主题色选择"
  >
    <button
      v-for="option in availableBrands"
      :key="option.value"
      type="button"
      class="topnav-brand-dot"
      :class="{ 'topnav-brand-dot--active': option.value === brand }"
      role="menuitemradio"
      :aria-checked="option.value === brand"
      :aria-label="`切换到${option.label}主题`"
      :data-brand="option.value"
      :title="option.label"
      @click="$emit('select', option.value)"
    >
      <span class="sr-only">{{ option.label }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { Palette } from 'lucide-vue-next'

import type { ThemeBrand } from '@/composables/useTheme'

interface ThemeBrandOption {
  value: ThemeBrand
  label: string
}

defineProps<{
  open: boolean
  brand: ThemeBrand
  currentBrandLabel: string
  availableBrands: ThemeBrandOption[]
}>()

defineEmits<{
  toggle: []
  select: [brand: ThemeBrand]
}>()
</script>
