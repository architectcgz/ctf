<template>
  <Transition name="notification-list-fade" mode="out-in">
    <section v-if="items.length === 0" key="empty" class="notification-empty">
      <div class="notification-empty__icon">
        <Bell class="h-8 w-8" />
      </div>
      <p class="notification-empty__title">
        {{ emptyState.title }}
      </p>
      <p class="notification-empty__copy">
        {{ emptyState.copy }}
      </p>
    </section>

    <section v-else key="list" class="notification-list" aria-label="通知列表">
      <button
        v-for="item in items"
        :key="item.id"
        type="button"
        class="notice-card"
        :class="{ 'is-unread': item.unread, 'is-read': !item.unread }"
        @click="selectNotification(item.id)"
      >
        <span class="notice-icon" :style="{ color: typeMeta(item.type).accentColor }">
          <component :is="typeMeta(item.type).icon" class="notice-icon__glyph" />
        </span>

        <span class="notice-body">
          <span class="notice-category" :style="{ color: typeMeta(item.type).accentColor }">
            {{ typeMeta(item.type).label }}
          </span>

          <span class="notice-title-row">
            <span class="notice-title">{{ item.title }}</span>
            <time>{{ formatDate(item.created_at) }}</time>
          </span>

          <span v-if="item.content" class="notice-copy">
            {{ item.content }}
          </span>
        </span>

        <span v-if="item.unread" class="unread-dot" aria-label="未读" />
      </button>
    </section>
  </Transition>
</template>

<script setup lang="ts">
import { Bell } from 'lucide-vue-next'

import type { NotificationTypeMeta } from '@/entities/notification'
import type { StoredNotificationItem } from '@/stores/notification'
import { formatDate } from '@/utils/format'

defineProps<{
  items: StoredNotificationItem[]
  emptyState: {
    title: string
    copy: string
  }
  typeMeta: (type: string) => NotificationTypeMeta
}>()

const emit = defineEmits<{
  select: [id: string]
}>()

function selectNotification(id: string): void {
  emit('select', id)
}
</script>

<style scoped>
.notification-list-fade-enter-active,
.notification-list-fade-leave-active {
  transition:
    opacity 0.25s ease,
    transform 0.25s ease;
}

.notification-list-fade-enter-from {
  opacity: 0;
  transform: translateY(var(--space-2));
}

.notification-list-fade-leave-to {
  opacity: 0;
  transform: translateY(calc(var(--space-2) * -1));
}

.notification-empty {
  display: flex;
  min-height: 18rem;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-10) var(--space-6);
  text-align: center;
}

.notification-empty__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 4rem;
  height: 4rem;
  border-radius: 999px;
  background: var(--notification-empty-icon-bg);
  color: var(--notification-panel-muted);
}

.notification-empty__title {
  margin-top: var(--space-4);
  font-size: var(--font-size-14);
  font-weight: 700;
  color: var(--notification-summary);
}

.notification-empty__copy {
  margin-top: var(--space-1);
  font-size: var(--font-size-12);
  color: var(--notification-panel-muted);
}

.notification-list {
  display: grid;
  gap: var(--space-2-5);
  margin-top: var(--space-4);
}

.notice-card {
  position: relative;
  min-height: 5.875rem;
  display: grid;
  grid-template-columns: 2.75rem minmax(0, 1fr);
  gap: var(--space-2-5);
  width: 100%;
  padding: var(--space-4) 2.25rem var(--space-4) var(--space-4);
  border: 1px solid var(--notification-card-border);
  border-radius: var(--ui-control-radius-lg);
  background: var(--notification-card-bg);
  box-shadow: var(--notification-card-shadow);
  text-align: left;
  cursor: pointer;
  transition:
    border-color var(--ui-motion-fast),
    transform var(--ui-motion-fast),
    box-shadow var(--ui-motion-fast);
}

.notice-card:hover,
.notice-card:focus-visible {
  border-color: var(--notification-card-border-hover);
  transform: translateY(-0.0625rem);
  box-shadow: var(--notification-card-shadow-hover);
}

.notice-card.is-unread {
  border-color: var(--notification-card-border-hover);
}

.notice-card.is-read {
  opacity: 0.95;
}

.notice-card:focus-visible {
  outline: var(--ui-focus-ring-width) solid
    color-mix(in srgb, var(--color-primary) 46%, var(--color-border-default));
  outline-offset: var(--space-1);
}

.notice-icon {
  width: 2.125rem;
  height: 2.125rem;
  margin-top: var(--space-0-5);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: radial-gradient(
    circle at 48% 43%,
    color-mix(in srgb, var(--color-success) 20%, transparent),
    color-mix(in srgb, var(--color-success) 14%, transparent) 52%,
    color-mix(in srgb, var(--color-success) 9%, transparent)
  );
  box-shadow:
    0 0 1.75rem color-mix(in srgb, var(--color-success) 9%, transparent),
    inset 0 0 0 1px color-mix(in srgb, var(--color-success) 4%, transparent);
}

.notice-icon__glyph {
  width: var(--space-4);
  height: var(--space-4);
}

.notice-body {
  min-width: 0;
  display: block;
}

.notice-category {
  display: block;
  margin-bottom: var(--space-1);
  font-size: var(--font-size-12);
  line-height: 1;
  font-weight: 600;
}

.notice-title-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  min-width: 0;
}

.notice-title {
  min-width: 0;
  flex: 0 1 auto;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--notification-card-title);
  font-size: var(--font-size-15);
  line-height: 1.18;
  font-weight: 700;
  letter-spacing: 0.0125rem;
}

.notice-title-row time {
  min-width: 0;
  color: var(--notification-card-time);
  font-size: var(--font-size-12);
  line-height: 1.1;
  font-weight: 400;
  white-space: nowrap;
}

.notice-copy {
  display: -webkit-box;
  width: 100%;
  max-width: 20.625rem;
  margin: var(--space-2) 0 0;
  color: var(--notification-card-copy);
  font-size: var(--font-size-13);
  line-height: 1.45;
  font-weight: 400;
  letter-spacing: 0.00625rem;
  overflow: hidden;
  text-overflow: ellipsis;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.unread-dot {
  position: absolute;
  top: var(--space-5);
  right: var(--space-4);
  width: var(--space-2);
  height: var(--space-2);
  border-radius: 50%;
  background: var(--notification-signal);
  box-shadow:
    0 0 1.125rem color-mix(in srgb, var(--notification-signal) 72%, transparent),
    0 0 0.125rem color-mix(in srgb, var(--color-bg-surface) 40%, transparent) inset;
}

@media (max-width: 768px) {
  .notice-card {
    grid-template-columns: 2.5rem minmax(0, 1fr);
    gap: var(--space-2-5);
    min-height: 5.75rem;
    padding: var(--space-3-5) 2rem var(--space-3-5) var(--space-3-5);
  }

  .notice-icon {
    width: var(--space-8);
    height: var(--space-8);
  }

  .notice-title {
    font-size: var(--font-size-14);
  }

  .notice-title-row time {
    font-size: var(--font-size-13);
  }

  .notice-copy {
    font-size: var(--font-size-13);
  }
}
</style>
