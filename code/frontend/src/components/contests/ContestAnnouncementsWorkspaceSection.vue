<script setup lang="ts">
import type { ContestAnnouncement } from '@/api/contracts'

import ContestAnnouncementsPanel from './ContestAnnouncementsPanel.vue'

defineProps<{
  panelId: string
  tabId: string
  announcements: ContestAnnouncement[]
  announcementsError: string
}>()
</script>

<template>
  <section
    :id="panelId"
    class="workspace-panel"
    role="tabpanel"
    :aria-labelledby="tabId"
  >
    <section class="contest-section">
      <div class="contest-section__head workspace-tab-heading">
        <div class="workspace-tab-heading__main">
          <div class="workspace-overline">
            Announcements
          </div>
          <h2 class="contest-section__title workspace-tab-heading__title">
            公告
          </h2>
        </div>
        <div class="contest-section__hint">
          {{ announcements.length }} 条
        </div>
      </div>

      <ContestAnnouncementsPanel
        :announcements="announcements"
        :announcements-error="announcementsError"
      />
    </section>
  </section>
</template>

<style scoped>
.contest-section__head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-3);
}

.contest-section__title:not(.workspace-tab-heading__title) {
  margin-top: var(--space-1-5);
  font-size: var(--font-size-1-10);
  font-weight: 700;
  color: var(--color-text-primary);
}

.contest-section__hint {
  font-size: var(--font-size-0-82);
  color: var(--color-text-secondary);
}
</style>
