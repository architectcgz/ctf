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
        :back-route="backToStudioRoute"
      />

      <AppEmpty
        v-if="loadError"
        title="竞赛公告加载失败"
        :description="loadError"
        icon="AlertTriangle"
      >
        <template #action>
          <AppRouteLink
            id="contest-announcements-error-back"
            :to="backToStudioRoute"
            class="ui-btn ui-btn--ghost"
          >
            返回竞赛工作台
          </AppRouteLink>
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

<script setup lang="ts">
import { toRef } from 'vue'

import AppEmpty from '@/shared/ui/common/AppEmpty.vue'
import AppLoading from '@/shared/ui/common/AppLoading.vue'
import AppRouteLink from '@/shared/ui/navigation/AppRouteLink.vue'
import {
  ContestAnnouncementsTopbarPanel,
  ContestAnnouncementsWorkspacePanel,
  useContestAnnouncementsPage,
} from '@/features/platform/contests'

const props = defineProps<{
  contestId: string
}>()

const {
  contest,
  loading,
  loadError,
  backToStudioRoute,
  management,
  formatTime,
  loadPage,
  handleSubmit,
  handleDelete,
} = useContestAnnouncementsPage(toRef(props, 'contestId'))
</script>

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
