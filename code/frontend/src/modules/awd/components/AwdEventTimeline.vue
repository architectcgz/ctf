<script setup lang="ts">
import type { AwdTimelineItem } from '@/modules/awd/types'

defineProps<{
  items: AwdTimelineItem[]
  emptyText?: string
}>()
</script>

<template>
  <section class="awd-event-timeline">
    <div
      v-if="items.length === 0"
      class="awd-event-timeline__empty"
    >
      {{ emptyText || '暂无事件' }}
    </div>

    <div
      v-else
      class="awd-event-timeline__list"
    >
      <article
        v-for="item in items"
        :key="item.id"
        class="awd-event-timeline__item"
      >
        <time class="awd-event-timeline__time">{{ item.time }}</time>
        <div class="awd-event-timeline__body">
          <strong>{{ item.title }}</strong>
          <p>{{ item.description }}</p>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.awd-event-timeline__list {
  display: grid;
  gap: 0.85rem;
}

.awd-event-timeline__item {
  display: grid;
  grid-template-columns: 68px minmax(0, 1fr);
  gap: 0.85rem;
  padding-bottom: 0.85rem;
  border-bottom: 1px solid color-mix(in srgb, var(--color-border-default) 72%, transparent);
}

.awd-event-timeline__item:last-child {
  padding-bottom: 0;
  border-bottom: 0;
}

.awd-event-timeline__time {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-12);
  color: var(--color-text-secondary);
}

.awd-event-timeline__body strong {
  display: block;
  font-size: var(--font-size-14);
}

.awd-event-timeline__body p {
  margin: 0.3rem 0 0;
  font-size: var(--font-size-13);
  color: var(--color-text-secondary);
  line-height: 1.5;
}

.awd-event-timeline__empty {
  color: var(--color-text-secondary);
  font-size: var(--font-size-13);
}
</style>
