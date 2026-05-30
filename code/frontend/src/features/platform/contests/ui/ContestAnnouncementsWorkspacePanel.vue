<script setup lang="ts">
import type { ContestAnnouncement } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'

defineProps<{
  canManageAnnouncements: boolean
  title: string
  content: string
  titleError?: string
  contentError?: string
  publishing: boolean
  announcements: ContestAnnouncement[]
  loading: boolean
  loadError: string
  deletingAnnouncementId: string | null
  formatTime: (value: string) => string
}>()

const emit = defineEmits<{
  submit: []
  delete: [announcementId: string]
  'update:title': [value: string]
  'update:content': [value: string]
}>()

function handleSubmit(): void {
  emit('submit')
}

function handleDelete(announcementId: string): void {
  emit('delete', announcementId)
}

function handleTitleInput(event: Event): void {
  emit('update:title', (event.target as HTMLInputElement).value)
}

function handleContentInput(event: Event): void {
  emit('update:content', (event.target as HTMLTextAreaElement).value)
}
</script>

<template>
  <section
    v-if="canManageAnnouncements"
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
            :value="title"
            type="text"
            class="ui-control"
            placeholder="例如：开赛通知"
            @input="handleTitleInput"
          >
        </span>
        <span
          v-if="titleError"
          class="contest-announcement-error"
        >{{ titleError }}</span>
      </label>

      <label class="ui-field contest-announcement-field">
        <span class="ui-field__label">内容</span>
        <span class="ui-control-wrap">
          <textarea
            :value="content"
            rows="6"
            class="ui-control contest-announcement-textarea"
            placeholder="输入面向参赛者展示的公告内容。"
            @input="handleContentInput"
          />
        </span>
        <span
          v-if="contentError"
          class="contest-announcement-error"
        >{{ contentError }}</span>
      </label>

      <div class="contest-announcement-actions">
        <button
          id="contest-announcement-submit"
          type="submit"
          class="ui-btn ui-btn--primary"
          :disabled="publishing"
        >
          {{ publishing ? '发布中...' : '发布公告' }}
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
      <span>{{ announcements.length }} 条</span>
    </header>

    <AppLoading v-if="loading">
      正在读取公告列表...
    </AppLoading>

    <div
      v-else-if="loadError"
      class="contest-announcement-inline-error"
    >
      {{ loadError }}
    </div>

    <AppEmpty
      v-else-if="announcements.length === 0"
      icon="Bell"
      title="暂无公告"
      description="当前竞赛还没有发布公告。"
    />

    <div
      v-else
      class="contest-announcement-list"
    >
      <article
        v-for="announcement in announcements"
        :key="announcement.id"
        class="contest-announcement-item"
      >
        <div class="contest-announcement-item__head">
          <div>
            <h3>{{ announcement.title }}</h3>
            <p>{{ formatTime(announcement.created_at) }}</p>
          </div>
          <button
            v-if="canManageAnnouncements"
            :id="`contest-announcement-delete-${announcement.id}`"
            type="button"
            class="ui-btn ui-btn--ghost ui-btn--sm"
            :disabled="deletingAnnouncementId === announcement.id"
            @click="handleDelete(announcement.id)"
          >
            {{ deletingAnnouncementId === announcement.id ? '删除中...' : '删除' }}
          </button>
        </div>
        <p class="contest-announcement-item__content">
          {{ announcement.content || '暂无正文。' }}
        </p>
      </article>
    </div>
  </section>
</template>

<style scoped>
.contest-announcement-panel__head,
.contest-announcement-item__head,
.contest-announcement-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.contest-announcement-form,
.contest-announcement-field,
.contest-announcement-list {
  display: grid;
  gap: var(--space-3);
}

.contest-announcement-panel__overline,
.contest-announcement-item__head p,
.contest-announcement-inline-error,
.contest-announcement-error {
  color: var(--color-text-muted);
  font-size: var(--font-size-0-875);
}

.contest-announcement-panel h2,
.contest-announcement-item h3,
.contest-announcement-item__content {
  margin: 0;
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
