<script setup>
defineProps({
  title: {
    type: String,
    default: '',
  },
  subtitle: {
    type: String,
    default: '',
  },
})
</script>

<template>
  <section class="rounded-2xl border border-border bg-surface/90 p-5 shadow-[0_10px_30px_rgba(0,0,0,0.18)]">
    <header v-if="title || subtitle" class="mb-4 flex items-start justify-between gap-4">
      <div>
        <h2 v-if="title" class="text-lg font-semibold text-slate-100">{{ title }}</h2>
        <p v-if="subtitle" class="mt-1 text-sm text-slate-400">{{ subtitle }}</p>
      </div>
      <slot name="header" />
    </header>
    <slot />
  </section>
</template>
