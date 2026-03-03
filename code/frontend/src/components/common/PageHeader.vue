<script setup>
defineProps({
  title: {
    type: String,
    required: true,
  },
  description: {
    type: String,
    default: '',
  },
  eyebrow: {
    type: String,
    default: '',
  },
})
</script>

<template>
  <header class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
    <div class="space-y-2">
      <p v-if="eyebrow" class="text-xs font-semibold uppercase tracking-[0.28em] text-cyan-400/80">
        {{ eyebrow }}
      </p>
      <div class="space-y-1">
        <h1 class="text-3xl font-semibold tracking-tight text-slate-50">{{ title }}</h1>
        <p v-if="description" class="max-w-3xl text-sm leading-6 text-slate-400">
          {{ description }}
        </p>
      </div>
    </div>
    <slot />
  </header>
</template>
