<script setup lang="ts">
defineProps<{
  title: string
  description?: string
  eyebrow?: string
}>()
</script>

<template>
  <header class="page-header flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
    <div class="space-y-2">
      <p
        v-if="eyebrow"
        class="page-header__eyebrow inline-flex border-l-2 border-primary pl-3 text-[11px] font-semibold uppercase tracking-[0.26em] text-primary"
      >
        {{ eyebrow }}
      </p>
      <div class="space-y-2">
        <h1 class="workspace-page-title">{{ title }}</h1>
        <p v-if="description" class="workspace-page-copy">
          {{ description }}
        </p>
      </div>
    </div>
    <div class="flex shrink-0 items-center gap-3">
      <slot />
    </div>
  </header>
</template>

<style scoped>
.page-header {
  padding: 1rem 1.25rem;
  border: 1px solid color-mix(in srgb, var(--color-primary) 16%, var(--color-border-default));
  border-radius: 18px;
  background: linear-gradient(
    145deg,
    color-mix(in srgb, var(--color-primary) 8%, var(--color-bg-surface)),
    color-mix(in srgb, var(--color-bg-surface) 94%, var(--color-bg-base))
  );
  box-shadow: 0 12px 28px color-mix(in srgb, var(--color-shadow-soft) 36%, transparent);
}

.page-header__eyebrow {
  border-left-color: color-mix(in srgb, var(--color-primary) 90%, var(--color-bg-surface));
}

:global([data-theme='light']) .page-header {
  border-color: color-mix(in srgb, var(--color-primary) 10%, var(--color-border-default));
  background: linear-gradient(
    145deg,
    color-mix(in srgb, var(--color-primary) 4%, var(--color-bg-surface)),
    color-mix(in srgb, var(--color-bg-elevated) 94%, var(--color-bg-surface))
  );
  box-shadow: 0 10px 24px color-mix(in srgb, var(--color-shadow-soft) 72%, transparent);
}

@media (max-width: 767px) {
  .page-header {
    padding: 0.875rem 1rem;
    border-radius: 16px;
  }
}
</style>
