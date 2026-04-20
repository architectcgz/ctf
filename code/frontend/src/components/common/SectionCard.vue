<script setup lang="ts">
defineProps<{
  title?: string
  subtitle?: string
}>()
</script>

<template>
  <section class="section-card">
    <header
      v-if="title || subtitle || $slots.header"
      class="section-card__header mb-4 flex items-start justify-between gap-4 border-b border-border-subtle pb-3"
    >
      <div class="min-w-0">
        <h2
          v-if="title"
          class="text-base font-semibold text-text-primary"
        >
          {{ title }}
        </h2>
        <p
          v-if="subtitle"
          class="mt-1 text-sm leading-6 text-text-secondary"
        >
          {{ subtitle }}
        </p>
      </div>
      <div class="shrink-0">
        <slot name="header" />
      </div>
    </header>
    <div class="section-card__body space-y-4">
      <slot />
    </div>
  </section>
</template>

<style scoped>
.section-card {
  padding: 0.95rem 0.1rem 0.35rem;
  border-top: 1px solid color-mix(in srgb, var(--color-primary) 22%, var(--color-border-default));
  background: transparent;
}

.section-card__header {
  border-bottom-color: color-mix(in srgb, var(--color-primary) 14%, var(--color-border-subtle));
}

.section-card__body {
  padding-left: 0.12rem;
}

@media (max-width: 767px) {
  .section-card {
    padding: 0.82rem 0 0.25rem;
  }
}
</style>
