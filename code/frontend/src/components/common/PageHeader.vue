<script setup lang="ts">
defineProps<{
  title: string
  description?: string
  eyebrow?: string
}>()
</script>

<template>
  <header class="flex flex-col gap-4 rounded-[28px] border border-border bg-surface/82 px-5 py-5 shadow-[0_20px_50px_var(--color-shadow-soft)] lg:flex-row lg:items-end lg:justify-between lg:px-6">
    <div class="space-y-3">
      <p
        v-if="eyebrow"
        class="inline-flex rounded-full border border-primary/30 bg-primary/10 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.28em] text-primary"
      >
        {{ eyebrow }}
      </p>
      <div class="space-y-1">
        <h1 class="text-3xl font-semibold tracking-tight text-text-primary">{{ title }}</h1>
        <p v-if="description" class="max-w-3xl text-sm leading-6 text-text-secondary">
          {{ description }}
        </p>
      </div>
    </div>
    <div class="flex shrink-0 items-center gap-3">
      <slot />
    </div>
  </header>
</template>
