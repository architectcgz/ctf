<template>
  <section :class="['section-card', rootClass]">
    <header
      v-if="title || subtitle || $slots.header"
      class="section-card__header flex items-start justify-between gap-4"
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
    <div class="section-card__body">
      <slot />
    </div>
  </section>
</template>

<style scoped>
.section-card {
  --section-card-padding-block-start: 0.95rem;
  --section-card-padding-inline: 0.1rem;
  --section-card-padding-block-end: 0.35rem;
  --section-card-border: none;
  --section-card-border-top-width: 1px;
  --section-card-border-top-style: solid;
  --section-card-border-top-color: color-mix(
    in srgb,
    var(--color-primary) 22%,
    var(--color-border-default)
  );
  --section-card-border-radius: 0;
  --section-card-background: transparent;
  --section-card-box-shadow: none;
  --section-card-header-margin-bottom: var(--space-4);
  --section-card-header-padding-bottom: var(--space-3);
  --section-card-header-border-bottom: 1px solid
    color-mix(in srgb, var(--color-primary) 14%, var(--color-border-subtle));
  --section-card-body-gap: var(--space-4);
  --section-card-body-padding-left: 0.12rem;

  padding: var(--section-card-padding-block-start) var(--section-card-padding-inline)
    var(--section-card-padding-block-end);
  border: var(--section-card-border);
  border-top: var(--section-card-border-top-width) var(--section-card-border-top-style)
    var(--section-card-border-top-color);
  border-radius: var(--section-card-border-radius);
  background: var(--section-card-background);
  box-shadow: var(--section-card-box-shadow);
}

.section-card__header {
  margin-bottom: var(--section-card-header-margin-bottom);
  padding-bottom: var(--section-card-header-padding-bottom);
  border-bottom: var(--section-card-header-border-bottom);
}

.section-card__body {
  display: grid;
  gap: var(--section-card-body-gap);
  padding-left: var(--section-card-body-padding-left);
}

.section-card--teacher-flat {
  --section-card-padding-block-start: var(--space-4);
  --section-card-padding-inline: var(--space-1);
  --section-card-padding-block-end: var(--space-3);
  --section-card-border-top-color: color-mix(in srgb, var(--teacher-divider) 88%, transparent);
  --section-card-header-border-bottom: 1px dashed
    color-mix(in srgb, var(--teacher-divider) 86%, transparent);
  --section-card-body-padding-left: 0;
}

.section-card--teacher-surface {
  --section-card-border: 1px solid var(--teacher-card-border);
  --section-card-border-top-width: 0;
  --section-card-background: var(--journal-surface-subtle);
  --section-card-box-shadow: 0 10px 24px var(--color-shadow-soft);
  --section-card-header-border-bottom: 1px dashed var(--teacher-divider);
}

@media (max-width: 767px) {
  .section-card {
    --section-card-padding-block-start: 0.82rem;
    --section-card-padding-inline: 0;
    --section-card-padding-block-end: 0.25rem;
  }
}
</style>

<script setup lang="ts">
import { computed } from 'vue'

type SectionCardVariant = 'default' | 'teacher-flat' | 'teacher-surface'

const props = withDefaults(
  defineProps<{
    title?: string
    subtitle?: string
    variant?: SectionCardVariant
  }>(),
  {
    title: '',
    subtitle: '',
    variant: 'default',
  }
)

const rootClass = computed(() => `section-card--${props.variant}`)
</script>
