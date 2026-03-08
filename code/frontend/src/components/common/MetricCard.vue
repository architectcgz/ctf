<script setup lang="ts">
withDefaults(
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

const accentClassMap = {
  primary: 'bg-[var(--color-primary-soft)] text-[var(--color-primary)]',
  success: 'bg-[color:rgba(63,185,80,0.14)] text-[var(--color-success)]',
  warning: 'bg-[color:rgba(210,153,34,0.16)] text-[var(--color-warning)]',
  danger: 'bg-[color:rgba(248,81,73,0.16)] text-[var(--color-danger)]',
}
</script>

<template>
  <section class="overflow-hidden rounded-[24px] border border-border bg-surface/88 p-5 shadow-[0_18px_40px_var(--color-shadow-soft)]">
    <div class="flex items-start justify-between gap-3">
      <div class="min-w-0">
        <div class="text-[11px] font-semibold uppercase tracking-[0.18em] text-text-muted">{{ label }}</div>
        <div class="mt-3 truncate text-3xl font-semibold tracking-tight text-text-primary">{{ value }}</div>
      </div>
      <div class="h-12 w-1.5 rounded-full" :class="accentClassMap[accent]" />
    </div>
    <div v-if="hint || $slots.default" class="mt-4 border-t border-border-subtle pt-3 text-sm text-text-secondary">
      <slot>
        {{ hint }}
      </slot>
    </div>
  </section>
</template>
