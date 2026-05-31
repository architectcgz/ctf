<script setup lang="ts">
import { computed, watch } from 'vue'

import type { ContestDetailData } from '@/api/contracts'
import AppEmpty from '@/shared/ui/common/AppEmpty.vue'
import AppLoading from '@/shared/ui/common/AppLoading.vue'
import AdminSurfaceDrawer from '@/shared/ui/common/modal-templates/AdminSurfaceDrawer.vue'
import AppRouteLink from '@/shared/ui/navigation/AppRouteLink.vue'
import { useContestAnnouncementManagement } from '../model'

interface ContestAnnouncementFullPageRoute {
  name: string
  params: {
    id: string
  }
}

const props = defineProps<{
  open: boolean
  contest: ContestDetailData | null
  fullPageRoute: ContestAnnouncementFullPageRoute | null
}>()

const emit = defineEmits<{
  close: []
}>()
const management = useContestAnnouncementManagement(computed(() => props.contest))

function formatTime(value: string): string {
  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function handleClose(): void {
  emit('close')
}

function handleOpenChange(value: boolean): void {
  if (!value) {
    emit('close')
  }
}

async function handleSubmit(): Promise<void> {
  await management.publishAnnouncement()
}

async function handleDelete(announcementId: string): Promise<void> {
  await management.deleteAnnouncement(announcementId)
}

watch(
  () => [props.open, props.contest?.id] as const,
  ([open, contestId]) => {
    if (!open) {
      management.resetForm()
      return
    }
    if (!contestId) return
    void management.loadAnnouncements()
  }
)
</script>

<template>
  <AdminSurfaceDrawer
    :open="open"
    :title="contest ? `${contest.title} · 公告` : '发布通知'"
    subtitle="管理员可以为当前竞赛发布公告，也可以快速查看和删除未结束赛事的历史公告。"
    eyebrow="Contest Announcements"
    width="38rem"
    @close="handleClose"
    @update:open="handleOpenChange"
  >
    <div class="announcement-drawer">
      <div
        v-if="contest"
        class="announcement-drawer__header"
      >
        <div class="announcement-drawer__contest-meta">
          <span class="announcement-drawer__contest-label">赛事状态</span>
          <strong>{{ contest.status }}</strong>
        </div>
        <AppRouteLink
          v-if="fullPageRoute"
          id="contest-open-announcement-page"
          :to="fullPageRoute"
          class="ui-btn ui-btn--secondary ui-btn--sm"
          @click="emit('close')"
        >
          进入完整管理页
        </AppRouteLink>
      </div>

      <section
        v-if="contest && management.canManageAnnouncements.value"
        class="announcement-drawer__section"
      >
        <div class="announcement-drawer__section-head">
          <h3>发布公告</h3>
        </div>
        <form
          class="announcement-drawer__form"
          @submit.prevent="handleSubmit"
        >
          <label class="ui-field announcement-drawer__field">
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
              class="announcement-drawer__error"
            >{{ management.errors.title }}</span>
          </label>

          <label class="ui-field announcement-drawer__field">
            <span class="ui-field__label">内容</span>
            <span class="ui-control-wrap">
              <textarea
                v-model="management.form.content"
                rows="5"
                class="ui-control announcement-drawer__textarea"
                placeholder="输入面向参赛者的公告内容。"
              />
            </span>
            <span
              v-if="management.errors.content"
              class="announcement-drawer__error"
            >{{ management.errors.content }}</span>
          </label>

          <div class="announcement-drawer__actions">
            <button
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
        v-else-if="contest"
        class="announcement-drawer__section"
      >
        <div class="announcement-drawer__readonly">
          赛事已结束，公告区仅保留查看能力。
        </div>
      </section>

      <section class="announcement-drawer__section">
        <div class="announcement-drawer__section-head">
          <h3>历史公告</h3>
          <span>{{ management.announcements.value.length }} 条</span>
        </div>

        <AppLoading v-if="management.loading.value">
          正在加载公告...
        </AppLoading>

        <div
          v-else-if="management.loadError.value"
          class="announcement-drawer__load-error"
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
          class="announcement-drawer__list"
        >
          <article
            v-for="announcement in management.announcements.value"
            :key="announcement.id"
            class="announcement-drawer__item"
          >
            <div class="announcement-drawer__item-head">
              <div>
                <h4>{{ announcement.title }}</h4>
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
            <p class="announcement-drawer__content">
              {{ announcement.content || '暂无正文。' }}
            </p>
          </article>
        </div>
      </section>
    </div>
  </AdminSurfaceDrawer>
</template>

<style scoped>
.announcement-drawer {
  display: grid;
  gap: var(--space-5);
  overflow-y: auto;
  padding: var(--space-1) var(--space-6) var(--space-6);
}

.announcement-drawer__header,
.announcement-drawer__section-head,
.announcement-drawer__item-head,
.announcement-drawer__actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.announcement-drawer__contest-meta {
  display: grid;
  gap: var(--space-1);
}

.announcement-drawer__contest-label,
.announcement-drawer__item-head p,
.announcement-drawer__load-error,
.announcement-drawer__readonly,
.announcement-drawer__error {
  color: var(--color-text-muted);
  font-size: var(--font-size-0-875);
}

.announcement-drawer__section {
  display: grid;
  gap: var(--space-4);
}

.announcement-drawer__section-head h3,
.announcement-drawer__item-head h4 {
  margin: 0;
  color: var(--color-text-primary);
}

.announcement-drawer__form,
.announcement-drawer__field,
.announcement-drawer__list {
  display: grid;
  gap: var(--space-3);
}

.announcement-drawer__textarea {
  min-height: 9rem;
  resize: vertical;
}

.announcement-drawer__item {
  display: grid;
  gap: var(--space-2);
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-4);
  padding: var(--space-4);
  background: var(--color-bg-surface);
}

.announcement-drawer__item-head > div {
  display: grid;
  gap: var(--space-1);
}

.announcement-drawer__content {
  margin: 0;
  color: var(--color-text-secondary);
  line-height: 1.7;
  white-space: pre-wrap;
}
</style>
