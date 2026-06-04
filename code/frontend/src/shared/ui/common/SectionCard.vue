<template>
  <section :class="['section-card', rootClass]">
    <header
      v-if="title || subtitle || $slots.header"
      class="section-card__header flex items-start justify-between gap-4"
    >
      <div class="min-w-0">
        <h2
          v-if="title"
          class="section-card__title"
        >
          {{ title }}
        </h2>
        <p
          v-if="subtitle"
          class="section-card__subtitle"
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
  padding:
    var(--section-card-padding-block-start, 0.95rem)
    var(--section-card-padding-inline, 0.1rem)
    var(--section-card-padding-block-end, 0.35rem);
  border: var(--section-card-border, none);
  border-top:
    var(--section-card-border-top-width, 1px)
    var(--section-card-border-top-style, solid)
    var(
      --section-card-border-top-color,
      color-mix(in srgb, var(--color-primary) 22%, var(--color-border-default))
    );
  border-radius: var(--section-card-border-radius, 0);
  background: var(--section-card-background, transparent);
  box-shadow: var(--section-card-box-shadow, none);
}

.section-card__header {
  align-items: var(--section-card-header-align-items, flex-start);
  margin-bottom: var(--section-card-header-margin-bottom, var(--space-4));
  padding: var(--section-card-header-padding, 0 0 var(--space-3) 0);
  border: var(--section-card-header-border, 0);
  border-bottom:
    var(
      --section-card-header-border-bottom,
      1px solid color-mix(in srgb, var(--color-primary) 14%, var(--color-border-subtle))
    );
  border-left: var(--section-card-header-border-left, 0);
  border-radius: var(--section-card-header-border-radius, 0);
  background: var(--section-card-header-background, transparent);
}

.section-card__title {
  font-size: var(--section-card-title-font-size, var(--font-size-1-00));
  line-height: var(--section-card-title-line-height, 1.5);
  font-weight: var(--section-card-title-font-weight, 600);
  color: var(--section-card-title-color, var(--color-text-primary));
}

.section-card__subtitle {
  margin-top: var(--section-card-subtitle-margin-top, var(--space-1));
  font-size: var(--section-card-subtitle-font-size, var(--font-size-0-88));
  line-height: var(--section-card-subtitle-line-height, 1.5);
  color: var(--section-card-subtitle-color, var(--color-text-secondary));
}

.section-card__body {
  display: grid;
  gap: var(--section-card-body-gap, var(--space-4));
  padding-left: var(--section-card-body-padding-left, 0.12rem);
}

.section-card__body > .rounded-2xl,
.section-card__body > .rounded-xl {
  border-color: var(--section-card-direct-surface-border-color);
  background: var(--section-card-direct-surface-background);
  box-shadow: var(--section-card-direct-surface-box-shadow);
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
    padding:
      var(--section-card-padding-block-start-mobile, 0.82rem)
      var(--section-card-padding-inline-mobile, 0)
      var(--section-card-padding-block-end-mobile, 0.25rem);
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
