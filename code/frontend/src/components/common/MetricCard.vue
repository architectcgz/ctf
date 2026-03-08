<script setup lang="ts">
import { computed } from 'vue'

import AppCard from '@/components/common/AppCard.vue'

const props = withDefaults(
  defineProps<{
    label: string
    value: string | number
    hint?: string
    accent?: 'primary' | 'success' | 'warning' | 'danger'
  }>(),
  {
    hint: '',
    accent: 'primary',
  },
)

const displayValue = computed(() => String(props.value))
</script>

<template>
  <AppCard variant="metric" :accent="props.accent" :eyebrow="props.label" :title="displayValue">
    <div v-if="props.hint || $slots.default" class="text-sm leading-6 text-text-secondary">
      <slot>
        {{ props.hint }}
      </slot>
    </div>
  </AppCard>
</template>
