<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Bell, ChevronLeft } from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'

import { getContest } from '@/api/admin'
import type { ContestDetailData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
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
      <header
        v-if="contest"
        class="contest-announcement-topbar"
      >
        <div class="contest-announcement-topbar__left">
          <button
            type="button"
            class="contest-announcement-back"
            @click="goBackToStudio"
          >
            <ChevronLeft class="h-5 w-5" />
          </button>
          <div class="contest-announcement-title-group">
            <div class="contest-announcement-overline">
              Contest Announcements
            </div>
            <h1>{{ contest.title }}</h1>
          </div>
        </div>

        <div class="contest-announcement-status">
          <Bell class="h-4 w-4" />
          <span>{{ contest.status }}</span>
        </div>
      </header>

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

      <template v-else-if="contest">
        <section
          v-if="management.canManageAnnouncements.value"
          class="contest-announcement-panel"
        >
          <header class="contest-announcement-panel__head">
            <div>
              <div class="contest-announcement-panel__overline">
                Publish
              </div>
              <h2>发布公告</h2>
            </div>
          </header>

          <form
            class="contest-announcement-form"
            @submit.prevent="handleSubmit"
          >
            <label class="ui-field contest-announcement-field">
              <span class="ui-field__label">标题</span>
              <span class="ui-control-wrap">
                <input
                  v-model="management.form.title"
                  type="text"
                  class="ui-control"
                  placeholder="例如：开赛通知"
                >
              </span>
              <span
                v-if="management.errors.title"
                class="contest-announcement-error"
              >{{ management.errors.title }}</span>
            </label>

            <label class="ui-field contest-announcement-field">
              <span class="ui-field__label">内容</span>
              <span class="ui-control-wrap">
                <textarea
                  v-model="management.form.content"
                  rows="6"
                  class="ui-control contest-announcement-textarea"
                  placeholder="输入面向参赛者展示的公告内容。"
                />
              </span>
              <span
                v-if="management.errors.content"
                class="contest-announcement-error"
              >{{ management.errors.content }}</span>
            </label>

            <div class="contest-announcement-actions">
              <button
                id="contest-announcement-submit"
                type="submit"
                class="ui-btn ui-btn--primary"
                :disabled="management.publishing.value"
              >
                {{ management.publishing.value ? '发布中...' : '发布公告' }}
              </button>
            </div>
          </form>
        </section>

        <section
          v-else
          class="contest-announcement-panel contest-announcement-panel--readonly"
        >
          <div class="contest-announcement-panel__overline">
            Read Only
          </div>
          <h2>赛事已结束，公告区仅保留查看能力。</h2>
        </section>

        <section class="contest-announcement-panel">
          <header class="contest-announcement-panel__head">
            <div>
              <div class="contest-announcement-panel__overline">
                History
              </div>
              <h2>历史公告</h2>
            </div>
            <span>{{ management.announcements.value.length }} 条</span>
          </header>

          <AppLoading v-if="management.loading.value">
            正在读取公告列表...
          </AppLoading>

          <div
            v-else-if="management.loadError.value"
            class="contest-announcement-inline-error"
          >
            {{ management.loadError.value }}
          </div>

          <AppEmpty
            v-else-if="management.announcements.value.length === 0"
            icon="Bell"
            title="暂无公告"
            description="当前竞赛还没有发布公告。"
          />

          <div
            v-else
            class="contest-announcement-list"
          >
            <article
              v-for="announcement in management.announcements.value"
              :key="announcement.id"
              class="contest-announcement-item"
            >
              <div class="contest-announcement-item__head">
                <div>
                  <h3>{{ announcement.title }}</h3>
                  <p>{{ formatTime(announcement.created_at) }}</p>
                </div>
                <button
                  v-if="management.canManageAnnouncements.value"
                  :id="`contest-announcement-delete-${announcement.id}`"
                  type="button"
                  class="ui-btn ui-btn--ghost ui-btn--sm"
                  :disabled="management.deletingAnnouncementId.value === announcement.id"
                  @click="handleDelete(announcement.id)"
                >
                  {{
                    management.deletingAnnouncementId.value === announcement.id ? '删除中...' : '删除'
                  }}
                </button>
              </div>
              <p class="contest-announcement-item__content">
                {{ announcement.content || '暂无正文。' }}
              </p>
            </article>
          </div>
        </section>
      </template>
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

.contest-announcement-topbar,
.contest-announcement-panel__head,
.contest-announcement-item__head,
.contest-announcement-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.contest-announcement-topbar {
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

.contest-announcement-title-group,
.contest-announcement-form,
.contest-announcement-field,
.contest-announcement-list {
  display: grid;
  gap: var(--space-3);
}

.contest-announcement-overline,
.contest-announcement-panel__overline,
.contest-announcement-item__head p,
.contest-announcement-inline-error,
.contest-announcement-status,
.contest-announcement-error {
  color: var(--color-text-muted);
  font-size: var(--font-size-0-875);
}

.contest-announcement-title-group h1,
.contest-announcement-panel h2,
.contest-announcement-item h3,
.contest-announcement-item__content {
  margin: 0;
}

.contest-announcement-status {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
}

.contest-announcement-panel {
  display: grid;
  gap: var(--space-4);
  border-radius: 1.25rem;
  border: 1px solid color-mix(in srgb, var(--journal-border) 84%, transparent);
  background: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  padding: var(--space-5);
}

.contest-announcement-panel--readonly,
.contest-announcement-inline-error {
  white-space: pre-wrap;
}

.contest-announcement-textarea {
  min-height: 9rem;
  resize: vertical;
}

.contest-announcement-item {
  display: grid;
  gap: var(--space-3);
  border-radius: 1rem;
  border: 1px solid color-mix(in srgb, var(--journal-border) 84%, transparent);
  background: color-mix(in srgb, var(--color-bg-surface) 84%, var(--color-bg-base));
  padding: var(--space-4);
}

.contest-announcement-item__head {
  align-items: flex-start;
}

.contest-announcement-item__head > div {
  display: grid;
  gap: var(--space-1);
}

.contest-announcement-item__content {
  white-space: pre-wrap;
  line-height: 1.65;
}
</style>
