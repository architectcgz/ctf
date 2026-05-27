<template>
  <div
    class="backoffice-sidebar__header flex h-16 items-center px-5 overflow-hidden whitespace-nowrap"
    :class="{ relative: mobile }"
  >
    <div class="flex items-center gap-3">
      <div
        class="backoffice-sidebar__logo-mark flex h-8.5 w-8.5 shrink-0 items-center justify-center rounded-xl shadow-sm"
      >
        <Box class="h-4 w-4" />
      </div>
      <span
        class="backoffice-sidebar__brand font-black text-lg tracking-tight uppercase transition-opacity duration-200"
        :class="showBrand ? 'opacity-100' : 'opacity-0'"
      >
        {{ brandKicker }}
      </span>
    </div>

    <button
      v-if="mobile"
      type="button"
      class="backoffice-sidebar__close ml-auto inline-flex h-9 w-9 items-center justify-center rounded-full p-1.5 shadow-sm transition-all"
      aria-label="关闭导航"
      @click="$emit('closeMobile')"
    >
      <X class="h-4 w-4" />
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Box, X } from 'lucide-vue-next'

const props = defineProps<{
  brandKicker: string
  collapsed: boolean
  mobile: boolean
}>()

defineEmits<{
  closeMobile: []
}>()

const showBrand = computed(() => props.mobile || !props.collapsed)
</script>
