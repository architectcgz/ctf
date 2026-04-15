<script setup lang="ts">
import type { ReviewArchiveObservationItemData } from '@/api/contracts'

defineProps<{
  items: ReviewArchiveObservationItemData[]
}>()

function levelClass(level: string): string {
  if (level === 'good') return 'observation observation--good'
  if (level === 'attention') return 'observation observation--attention'
  return 'observation observation--neutral'
}
</script>

<template>
  <section class="observation-strip teacher-surface-section">
    <header class="observation-strip__header">
      <div>
        <div class="observation-strip__eyebrow">Teaching Signals</div>
        <h2 class="observation-strip__title">教学观察摘要</h2>
      </div>
      <p class="observation-strip__hint">
        这些结论全部来自当前归档中的训练事件与评阅记录，没有附加 AI 黑盒判断。
      </p>
    </header>

    <div class="observation-strip__grid">
      <article
        v-for="item in items"
        :key="item.key"
        :class="[levelClass(item.level), 'teacher-surface-metric']"
      >
        <div class="observation__head">
          <span class="observation__label">{{ item.label }}</span>
          <span class="observation__level">{{ item.level }}</span>
        </div>
        <p class="observation__summary">{{ item.summary }}</p>
        <p v-if="item.evidence" class="observation__evidence">{{ item.evidence }}</p>
      </article>
    </div>
  </section>
</template>

<style scoped>
.observation-strip {
  padding: var(--space-4-5) var(--space-4-5) var(--space-5);
}

.observation-strip__header {
  display: flex;
  gap: var(--space-4);
  justify-content: space-between;
  align-items: flex-end;
  padding-bottom: var(--space-3);
  border-bottom: 1px dashed color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.observation-strip__eyebrow {
  font-size: var(--font-size-0-72);
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--journal-accent-strong);
  font-family: var(--font-family-sans);
}

.observation-strip__title {
  margin-top: var(--space-2);
  font-size: var(--font-size-1-35);
  color: var(--journal-ink);
}

.observation-strip__hint {
  max-width: 32rem;
  font-size: var(--font-size-0-93);
  line-height: 1.7;
  color: var(--color-text-secondary);
}

.observation-strip__grid {
  display: grid;
  gap: var(--space-4);
  margin-top: var(--space-4);
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.observation {
  padding: var(--space-4) var(--space-4);
  border-radius: 20px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 76%, transparent);
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base)),
    color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base))
  );
}

.observation--good {
  border-color: color-mix(in srgb, var(--color-success) 24%, var(--journal-border));
}

.observation--attention {
  border-color: color-mix(in srgb, var(--color-warning) 24%, var(--journal-border));
}

.observation__head {
  display: flex;
  gap: var(--space-3);
  justify-content: space-between;
  align-items: center;
}

.observation__label {
  font-weight: 700;
  color: var(--journal-ink);
}

.observation__level {
  font-size: var(--font-size-0-72);
  text-transform: uppercase;
  letter-spacing: 0.16em;
  color: var(--journal-muted);
  font-family: var(--font-family-mono);
}

.observation__summary {
  margin-top: var(--space-3);
  line-height: 1.75;
  color: color-mix(in srgb, var(--journal-muted) 80%, var(--journal-ink));
}

.observation__evidence {
  margin-top: var(--space-2-5);
  color: var(--journal-muted);
  font-size: var(--font-size-0-93);
  line-height: 1.7;
}

@media (max-width: 767px) {
  .observation-strip__header,
  .observation-strip__grid {
    grid-template-columns: 1fr;
    flex-direction: column;
    align-items: flex-start;
  }

  .observation-strip__grid {
    display: grid;
  }
}
</style>
