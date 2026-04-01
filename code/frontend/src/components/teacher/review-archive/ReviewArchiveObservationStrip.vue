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
  <section class="observation-strip">
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
        :class="levelClass(item.level)"
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
  padding: 0.3rem 0;
}

.observation-strip__header {
  display: flex;
  gap: 1rem;
  justify-content: space-between;
  align-items: flex-end;
  padding-bottom: 0.85rem;
  border-bottom: 1px solid color-mix(in srgb, #1e40af 14%, var(--color-border-subtle));
}

.observation-strip__eyebrow {
  font-size: 0.72rem;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: #1d4ed8;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
}

.observation-strip__title {
  margin-top: 0.55rem;
  font-size: 1.35rem;
  color: #0f172a;
}

.observation-strip__hint {
  max-width: 32rem;
  font-size: 0.93rem;
  line-height: 1.7;
  color: var(--color-text-secondary);
}

.observation-strip__grid {
  display: grid;
  gap: 1rem;
  margin-top: 1rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.observation {
  padding: 1rem 1.05rem;
  border: 1px solid color-mix(in srgb, #94a3b8 55%, white);
  border-radius: 20px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.94), rgba(248, 250, 252, 0.9));
}

.observation--good {
  border-color: rgba(16, 185, 129, 0.24);
}

.observation--attention {
  border-color: rgba(245, 158, 11, 0.26);
}

.observation__head {
  display: flex;
  gap: 0.8rem;
  justify-content: space-between;
  align-items: center;
}

.observation__label {
  font-weight: 700;
  color: #0f172a;
}

.observation__level {
  font-size: 0.72rem;
  text-transform: uppercase;
  letter-spacing: 0.16em;
  color: #475569;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
}

.observation__summary {
  margin-top: 0.8rem;
  line-height: 1.75;
  color: #1e293b;
}

.observation__evidence {
  margin-top: 0.65rem;
  color: #475569;
  font-size: 0.93rem;
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
