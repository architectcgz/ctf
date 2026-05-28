<script setup lang="ts">
import { Bell, ChevronLeft } from 'lucide-vue-next'

defineProps<{
  contestTitle: string
  contestStatus: string
}>()

const emit = defineEmits<{
  back: []
}>()

function handleBack(): void {
  emit('back')
}
</script>

<template>
  <header class="contest-announcement-topbar">
    <div class="contest-announcement-topbar__left">
      <button
        type="button"
        class="contest-announcement-back"
        @click="handleBack"
      >
        <ChevronLeft class="h-5 w-5" />
      </button>
      <div class="contest-announcement-title-group">
        <div class="contest-announcement-overline">
          Contest Announcements
        </div>
        <h1>{{ contestTitle }}</h1>
      </div>
    </div>

    <div class="contest-announcement-status">
      <Bell class="h-4 w-4" />
      <span>{{ contestStatus }}</span>
    </div>
  </header>
</template>

<style scoped>
.contest-announcement-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.contest-announcement-topbar__left {
  display: flex;
  align-items: center;
  gap: var(--space-4);
}

.contest-announcement-back {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 0.75rem;
  border: 1px solid color-mix(in srgb, var(--journal-border) 84%, transparent);
  color: var(--color-text-secondary);
  background: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
}

.contest-announcement-title-group {
  display: grid;
  gap: var(--space-3);
}

.contest-announcement-overline,
.contest-announcement-status {
  color: var(--color-text-muted);
  font-size: var(--font-size-0-875);
}

.contest-announcement-title-group h1 {
  margin: 0;
}

.contest-announcement-status {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
}
</style>
