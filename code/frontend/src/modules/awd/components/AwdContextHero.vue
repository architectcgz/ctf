<script setup lang="ts">
import AwdMetricStrip from '@/modules/awd/components/AwdMetricStrip.vue'
import type { AwdHeroMetric } from '@/modules/awd/types'

withDefaults(
  defineProps<{
    eyebrow?: string
    contestTitle: string
    pageTitle: string
    pageDescription?: string
    metrics?: AwdHeroMetric[]
  }>(),
  {
    eyebrow: 'AWD Workspace',
    pageDescription: '',
    metrics: () => [],
  }
)
</script>

<template>
  <header class="awd-context-hero">
    <div class="awd-context-hero__copy">
      <div class="workspace-overline awd-context-hero__eyebrow">
        {{ eyebrow }}
      </div>
      <p class="awd-context-hero__contest">
        {{ contestTitle }}
      </p>
      <h1 class="workspace-page-title awd-context-hero__title">
        {{ pageTitle }}
      </h1>
      <p
        v-if="pageDescription"
        class="workspace-page-copy awd-context-hero__description"
      >
        {{ pageDescription }}
      </p>
    </div>

    <AwdMetricStrip :metrics="metrics" />
  </header>
</template>

<style scoped>
.awd-context-hero {
  display: grid;
  gap: 1rem;
}

.awd-context-hero__copy {
  display: grid;
  gap: 0.5rem;
}

.awd-context-hero__contest {
  margin: 0;
  font-size: var(--font-size-13);
  font-weight: 600;
  color: var(--color-text-secondary);
}

.awd-context-hero__title {
  margin: 0;
}

.awd-context-hero__description {
  margin: 0;
}
</style>
