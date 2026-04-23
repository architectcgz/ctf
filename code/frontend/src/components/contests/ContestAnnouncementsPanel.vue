<script setup lang="ts">
import type { ContestAnnouncement } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { formatTime } from '@/utils/format'

const props = withDefaults(
  defineProps<{
    announcements: ContestAnnouncement[]
    announcementsError: string
    emptyVariant?: 'inline' | 'empty-state'
  }>(),
  {
    emptyVariant: 'empty-state',
  }
)
</script>

<template>
  <div
    v-if="announcementsError"
    class="contest-alert contest-alert--warning"
  >
    {{ announcementsError }}
  </div>

  <div
    v-else-if="announcements.length === 0 && emptyVariant === 'inline'"
    class="contest-inline-note"
  >
    当前竞赛暂无新的公告通知。
  </div>

  <div
    v-else-if="announcements.length === 0"
    class="contest-empty-state"
  >
    <AppEmpty
      icon="Bell"
      title="暂无公告"
      description="当前竞赛暂无新的公告通知。"
    />
  </div>

  <div
    v-else
    class="announcement-list"
  >
    <article
      v-for="announcement in announcements"
      :key="announcement.id"
      class="announcement-item"
    >
      <div class="announcement-item__head">
        <h3 class="announcement-item__title">
          {{ announcement.title }}
        </h3>
        <time
          class="announcement-item__time"
          :datetime="announcement.created_at"
        >
          {{ formatTime(announcement.created_at) }}
        </time>
      </div>
      <p
        v-if="announcement.content"
        class="announcement-item__content"
      >
        {{ announcement.content }}
      </p>
    </article>
  </div>
</template>

<style scoped>
.contest-alert {
  margin-top: 0.8rem;
  border-inline-start: 2px solid transparent;
  padding: 0.6rem 0.75rem;
  font-size: var(--font-size-0-84);
  line-height: 1.6;
}

.contest-alert--warning {
  border-inline-start-color: color-mix(in srgb, var(--color-warning) 60%, transparent);
  background: color-mix(in srgb, var(--color-warning) 8%, transparent);
  color: color-mix(in srgb, var(--color-warning) 88%, var(--color-text-primary));
}

.contest-inline-note {
  border-inline-start: 2px solid color-mix(in srgb, var(--color-border-default) 84%, transparent);
  padding-inline-start: 0.85rem;
  font-size: var(--font-size-0-88);
  line-height: 1.7;
  color: var(--color-text-secondary);
}

.contest-empty-state {
  margin-top: 1rem;
}

.announcement-list {
  margin-top: 1rem;
  display: grid;
  gap: 0.9rem;
}

.announcement-item {
  border-bottom: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  padding-bottom: 0.9rem;
}

.announcement-item__head {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.45rem 1rem;
}

.announcement-item__title {
  font-size: var(--font-size-0-96);
  font-weight: 700;
  color: var(--color-text-primary);
}

.announcement-item__time {
  font-size: var(--font-size-0-76);
  color: var(--color-text-secondary);
}

.announcement-item__content {
  margin-top: 0.55rem;
  white-space: pre-wrap;
  font-size: var(--font-size-0-88);
  line-height: 1.75;
  color: var(--color-text-secondary);
}
</style>
