<script setup lang="ts">
import { Bell, Save, ShieldCheck, Trophy } from 'lucide-vue-next'

defineProps<{
  pageTitle: string
  contestMode: string
  contestStatus: string
  contestModeLabel: string
  contestStatusLabel: string
  activeStage: string
  saving: boolean
}>()

const emit = defineEmits<{
  back: []
  openAnnouncements: []
  save: []
}>()

function handleOpenAnnouncements(): void {
  emit('openAnnouncements')
}

function handleSave(): void {
  emit('save')
}
</script>

<template>
  <header class="workspace-topbar studio-topbar-wrapper">
    <div class="topbar-leading">
      <div class="studio-title-block">
        <div class="workspace-overline">Contest Studio</div>
        <div class="studio-title-row">
          <h1 class="workspace-page-title">编辑竞赛</h1>
          <span class="title-separator">/</span>
          <span class="studio-contest-name">{{ pageTitle }}</span>
        </div>
      </div>

      <div class="studio-contest-meta">
        <span
          class="meta-tag"
          :class="`meta-tag--${contestMode}`"
        >
          <Trophy class="h-3 w-3" /> {{ contestModeLabel }}
        </span>
        <span class="meta-tag meta-tag--status">
          <ShieldCheck class="h-3 w-3" /> {{ contestStatusLabel }}
        </span>
      </div>
    </div>

    <div class="top-note">
      <button
        id="contest-open-announcements"
        type="button"
        class="ui-btn ui-btn--secondary studio-toolbar-btn"
        @click="handleOpenAnnouncements"
      >
        <Bell class="h-3.5 w-3.5" />
        <span>公告管理</span>
      </button>
      
      <div class="ops-divider" />

      <button
        v-if="activeStage === 'basics'"
        type="button"
        class="ui-btn ui-btn--primary studio-save-btn"
        :disabled="saving"
        @click="handleSave"
      >
        <Save class="h-4 w-4" />
        <span>{{ saving ? '正在保存...' : '保存变更' }}</span>
      </button>
    </div>
  </header>
</template>

<style scoped>
.studio-topbar-wrapper {
  background: var(--color-bg-surface);
  min-height: 5.5rem;
  padding: var(--space-workspace-topbar-padding-top) var(--space-workspace-side-padding) var(--space-2);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.studio-title-block {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.studio-title-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.title-separator {
  color: var(--color-text-muted);
  font-weight: 300;
  font-size: var(--font-size-18);
  opacity: 0.5;
}

.studio-contest-name {
  font-size: var(--font-size-15);
  font-weight: 700;
  color: var(--color-text-secondary);
  max-width: 18rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.studio-contest-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-left: 1.5rem;
}

.meta-tag {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.2rem 0.65rem;
  border-radius: 0.5rem;
  font-size: 10px;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.02em;
}

.meta-tag--awd {
  background: color-mix(in srgb, var(--color-primary) 10%, transparent);
  color: var(--color-primary);
  border: 1px solid color-mix(in srgb, var(--color-primary) 15%, transparent);
}

.meta-tag--status {
  background: var(--color-bg-elevated);
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border-subtle);
}

.studio-toolbar-btn {
  --ui-btn-height: 2.25rem;
  font-size: var(--font-size-12);
  font-weight: 700;
}

.ops-divider {
  width: 1px;
  height: 1.25rem;
  background: var(--color-border-subtle);
  margin: 0 0.5rem;
}

.studio-save-btn {
  --ui-btn-height: 2.25rem;
  --ui-btn-padding: 0 1.25rem;
  --ui-btn-radius: 0.75rem;
  --ui-btn-font-size: var(--font-size-12);
  --ui-btn-font-weight: 800;
  box-shadow: 0 4px 12px color-mix(in srgb, var(--color-primary) 20%, transparent);
}
</style>
