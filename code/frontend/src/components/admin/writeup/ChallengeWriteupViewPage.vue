<script setup lang="ts">
import { Edit3, RefreshCw } from 'lucide-vue-next'

import ChallengeDescriptionPanel from '@/components/admin/challenge/ChallengeDescriptionPanel.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import { useChallengeWriteupEditorPage } from '@/composables/useChallengeWriteupEditorPage'

const props = defineProps<{
  challengeId: string
}>()

const emit = defineEmits<{
  back: []
  edit: []
}>()

const {
  loading,
  challenge,
  writeup,
  loadPage,
} = useChallengeWriteupEditorPage(props.challengeId)
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero flex min-h-full flex-1 flex-col rounded-[24px] border px-6 py-6 md:px-8"
  >
    <header class="workspace-topbar">
      <div class="topbar-leading">
        <span class="workspace-overline">Challenge Workspace</span>
        <span class="class-chip">查看题解</span>
      </div>
      <div class="writeup-top-actions">
        <button class="admin-btn admin-btn-ghost" type="button" @click="emit('back')">返回题目</button>
        <button
          v-if="writeup"
          class="admin-btn admin-btn-primary"
          type="button"
          @click="emit('edit')"
        >
          <Edit3 class="h-4 w-4" />
          编辑题解
        </button>
        <button class="admin-btn admin-btn-ghost" type="button" @click="void loadPage()">
          <RefreshCw class="h-4 w-4" />
          刷新
        </button>
      </div>
    </header>

    <header v-if="writeup" class="writeup-reading-card__hero writeup-reading-card__hero--page">
      <div class="writeup-reading-card__intro">
        <div class="journal-note-label">Admin Writeup</div>
        <h1 class="workspace-tab-heading__title">{{ writeup.title }}</h1>
        <p class="workspace-tab-copy">
          当前保存版本会按这里的正文与公开范围对外展示，适合用于复核发布前的阅读效果。
        </p>
      </div>
    </header>

    <div v-if="writeup" class="journal-divider" />

    <AppLoading v-if="loading" class="writeup-loading">正在加载题解数据...</AppLoading>

    <main v-else class="content-pane writeup-workspace">
      <dl v-if="writeup" class="writeup-snapshot-grid">
        <div class="writeup-snapshot-item">
          <dt>可见性</dt>
          <dd>{{ writeup.visibility }}</dd>
        </div>
        <div class="writeup-snapshot-item">
          <dt>推荐状态</dt>
          <dd>{{ writeup.is_recommended ? '推荐题解' : '未推荐' }}</dd>
        </div>
        <div class="writeup-snapshot-item">
          <dt>创建时间</dt>
          <dd>{{ writeup.created_at }}</dd>
        </div>
        <div class="writeup-snapshot-item">
          <dt>更新时间</dt>
          <dd>{{ writeup.updated_at }}</dd>
        </div>
      </dl>

      <div v-if="writeup" class="writeup-view-body">
        <ChallengeDescriptionPanel
          :content="writeup.content"
          label="题解正文"
          test-id="admin-writeup-view-content"
        />
      </div>

      <section v-else class="writeup-view-section">
        <AppEmpty
          title="当前还没有管理员题解"
          description="可以先进入编辑页创建题解，再回到查看页确认最终展示效果。"
          icon="BookOpen"
        />
      </section>
    </main>
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-topbar-padding-bottom: var(--space-3);
  --journal-shell-hero-radial-strength: 7%;
  --journal-shell-hero-radial-size: 22rem;
  --journal-shell-hero-end: var(--journal-surface);
  --journal-shell-hero-shadow: 0 22px 50px var(--color-shadow-soft);
}

.writeup-top-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
}

.writeup-loading {
  padding-block: var(--space-7);
}

.writeup-view-section {
  display: grid;
  gap: var(--space-4);
}

.writeup-workspace {
  display: grid;
  gap: var(--space-4);
}

.writeup-reading-card__hero {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  gap: var(--space-4);
}

.writeup-reading-card__hero--page {
  padding-bottom: var(--space-4);
}

.writeup-reading-card__intro {
  display: grid;
  gap: var(--space-2);
  max-width: min(42rem, 100%);
}

.writeup-editor-head {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  gap: var(--space-3);
}

.writeup-snapshot-grid {
  display: grid;
  gap: var(--space-3);
  grid-template-columns: repeat(auto-fit, minmax(11rem, 1fr));
  margin: 0;
}

.writeup-snapshot-item {
  display: grid;
  gap: var(--space-1);
}

.writeup-snapshot-item dt {
  font-size: var(--font-size-0-76);
  font-weight: 700;
  color: var(--journal-muted);
}

.writeup-snapshot-item dd {
  margin: 0;
  color: var(--journal-ink);
}

.writeup-view-body {
  display: grid;
  gap: var(--space-2);
}

.admin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: 2.4rem;
  border-radius: 0.75rem;
  border: 1px solid transparent;
  padding: var(--space-2) var(--space-4);
  font-size: var(--font-size-0-875);
  font-weight: 600;
  transition:
    border-color 150ms ease,
    background 150ms ease,
    color 150ms ease;
}

.admin-btn-ghost {
  border-color: var(--journal-border);
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  color: var(--journal-ink);
}

.admin-btn-primary {
  border-color: color-mix(in srgb, var(--journal-accent) 18%, transparent);
  background: var(--journal-accent);
  color: #fff;
}

@media (max-width: 960px) {
  .writeup-reading-card {
    gap: var(--space-4);
    padding: var(--space-4);
  }
}
</style>
