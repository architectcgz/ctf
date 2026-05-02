<script setup lang="ts">
import AppEmpty from '@/components/common/AppEmpty.vue'
import { REVIEW_ARCHIVE_STATE_COPY } from './model/presentation'

defineProps<{
  loading: boolean
  error: string | null
  hasArchive: boolean
}>()

const emit = defineEmits<{
  reload: []
}>()
</script>

<template>
  <div
    v-if="loading"
    class="review-archive-loading"
  >
    <div class="review-archive-loading__hero" />
    <div class="review-archive-loading__grid">
      <div class="review-archive-loading__block" />
      <div class="review-archive-loading__block" />
    </div>
  </div>

  <AppEmpty
    v-else-if="error"
    :title="REVIEW_ARCHIVE_STATE_COPY.errorTitle"
    :description="error"
    icon="AlertTriangle"
  >
    <template #action>
      <button
        type="button"
        class="ui-btn ui-btn--primary"
        @click="emit('reload')"
      >
        {{ REVIEW_ARCHIVE_STATE_COPY.reload }}
      </button>
    </template>
  </AppEmpty>

  <AppEmpty
    v-else-if="!hasArchive"
    :title="REVIEW_ARCHIVE_STATE_COPY.emptyTitle"
    :description="REVIEW_ARCHIVE_STATE_COPY.emptyDescription"
    icon="FileChartColumnIncreasing"
  />

  <slot v-else />
</template>

<style scoped>
.review-archive-loading__hero,
.review-archive-loading__block {
  border-radius: 26px;
  background: linear-gradient(
    90deg,
    color-mix(in srgb, var(--journal-border) 80%, transparent),
    color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base))
  );
  animation: review-archive-pulse 1.35s ease-in-out infinite;
}

.review-archive-loading__hero {
  height: 220px;
}

.review-archive-loading__grid {
  display: grid;
  gap: var(--space-4);
  margin-top: var(--space-4);
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.review-archive-loading__block {
  height: 180px;
}

@keyframes review-archive-pulse {
  0%,
  100% {
    opacity: 0.58;
  }
  50% {
    opacity: 1;
  }
}

@media (max-width: 1023px) {
  .review-archive-loading__grid {
    grid-template-columns: 1fr;
  }
}
</style>
