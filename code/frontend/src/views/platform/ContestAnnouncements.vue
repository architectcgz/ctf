<script setup lang="ts">
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import ContestAnnouncementsTopbarPanel from '@/components/platform/contest/ContestAnnouncementsTopbarPanel.vue'
import ContestAnnouncementsWorkspacePanel from '@/components/platform/contest/ContestAnnouncementsWorkspacePanel.vue'
import { useContestAnnouncementsPage } from '@/features/platform-contests'

const {
  contest,
  loading,
  loadError,
  management,
  formatTime,
  goBackToStudio,
  loadPage,
  handleSubmit,
  handleDelete,
} = useContestAnnouncementsPage()
</script>

<template>
  <div class="contest-announcement-shell">
    <div
      v-if="loading"
      class="contest-announcement-loading"
    >
      <AppLoading>正在同步竞赛公告...</AppLoading>
    </div>

    <main
      v-else
      class="contest-announcement-content"
    >
      <ContestAnnouncementsTopbarPanel
        v-if="contest"
        :contest-title="contest.title"
        :contest-status="contest.status"
        @back="goBackToStudio"
      />

      <AppEmpty
        v-if="loadError"
        title="竞赛公告加载失败"
        :description="loadError"
        icon="AlertTriangle"
      >
        <template #action>
          <button
            type="button"
            class="ui-btn ui-btn--ghost"
            @click="goBackToStudio"
          >
            返回竞赛工作台
          </button>
        </template>
      </AppEmpty>

      <ContestAnnouncementsWorkspacePanel
        v-else-if="contest"
        :can-manage-announcements="management.canManageAnnouncements.value"
        :title="management.form.title"
        :content="management.form.content"
        :title-error="management.errors.title"
        :content-error="management.errors.content"
        :publishing="management.publishing.value"
        :announcements="management.announcements.value"
        :loading="management.loading.value"
        :load-error="management.loadError.value"
        :deleting-announcement-id="management.deletingAnnouncementId.value"
        :format-time="formatTime"
        @submit="void handleSubmit()"
        @delete="void handleDelete($event)"
        @update:title="management.form.title = $event"
        @update:content="management.form.content = $event"
      />
    </main>
  </div>
</template>

<style scoped>
.contest-announcement-shell {
  display: flex;
  flex-direction: column;
  min-height: calc(100vh - 64px);
  background: var(--color-bg-base);
}

.contest-announcement-loading {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.contest-announcement-content {
  display: grid;
  gap: var(--space-6);
  padding: var(--space-6);
}
</style>
