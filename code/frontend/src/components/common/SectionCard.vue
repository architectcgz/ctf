<script setup lang="ts">
defineProps<{
  title?: string
  subtitle?: string
}>()
</script>

<template>
  <section class="relative overflow-hidden rounded-[26px] border border-border bg-surface/88 p-5 shadow-[0_18px_40px_var(--color-shadow-soft)]">
    <div class="pointer-events-none absolute inset-x-0 top-0 h-px bg-[linear-gradient(90deg,transparent,rgba(8,145,178,0.45),transparent)]" />
    <header v-if="title || subtitle" class="mb-4 flex items-start justify-between gap-4 border-b border-border-subtle pb-4">
      <div>
        <h2 v-if="title" class="text-lg font-semibold text-text-primary">{{ title }}</h2>
        <p v-if="subtitle" class="mt-1 text-sm text-text-secondary">{{ subtitle }}</p>
      </div>
      <slot name="header" />
    </header>
    <slot />
  </section>
</template>
