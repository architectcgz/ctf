<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getContest } from '@/api/admin'
import type { ContestDetailData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import ContestAnnouncementsTopbarPanel from '@/components/platform/contest/ContestAnnouncementsTopbarPanel.vue'
import ContestAnnouncementsWorkspacePanel from '@/components/platform/contest/ContestAnnouncementsWorkspacePanel.vue'
import { useContestAnnouncementManagement } from '@/composables/useContestAnnouncementManagement'
import { useToast } from '@/composables/useToast'
import { ApiError } from '@/api/request'

const route = useRoute()
const router = useRouter()
const toast = useToast()

const contestId = computed(() => String(route.params.id ?? ''))
const contest = ref<ContestDetailData | null>(null)
const loading = ref(true)
const loadError = ref('')

const management = useContestAnnouncementManagement(computed(() => contest.value))

function humanizeRequestError(error: unknown, fallback: string): string {
  if (error instanceof ApiError && error.message.trim()) {
    return error.message
  }
  if (error instanceof Error && error.message.trim()) {
    return error.message
  }
  return fallback
}

function formatTime(value: string): string {
  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function goBackToStudio(): void {
  void router.push({ name: 'ContestEdit', params: { id: contestId.value } })
}

async function loadPage(): Promise<void> {
  if (!contestId.value) {
    loadError.value = '缺少竞赛编号。'
    loading.value = false
    return
  }

  loading.value = true
  loadError.value = ''
  try {
    contest.value = await getContest(contestId.value)
    await management.loadAnnouncements()
  } catch (error) {
    loadError.value = humanizeRequestError(error, '竞赛公告加载失败')
    toast.error(loadError.value)
  } finally {
    loading.value = false
  }
}

async function handleSubmit(): Promise<void> {
  await management.publishAnnouncement()
}

async function handleDelete(announcementId: string): Promise<void> {
  await management.deleteAnnouncement(announcementId)
}

onMounted(() => {
  void loadPage()
})
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
