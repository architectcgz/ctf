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
        <div class="studio-title-meta">
          <span class="workspace-overline">Contest Studio</span>
          <span class="studio-edit-label">编辑竞赛</span>
        </div>
        <h1
          class="studio-contest-heading"
          :title="pageTitle"
        >
          {{ pageTitle }}
        </h1>
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
  --studio-toolbar-button-height: calc(var(--ui-control-height-sm) + var(--space-1));

  background: var(--color-bg-surface);
  min-height: auto;
  padding: var(--space-4) var(--space-workspace-side-padding) 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-5);
}

.studio-title-block {
  display: flex;
  flex-direction: column;
  min-width: 0;
  gap: var(--space-1);
}

.studio-title-meta {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.studio-edit-label {
  display: inline-flex;
  align-items: center;
  border-left: 1px solid color-mix(in srgb, var(--color-border-default) 72%, transparent);
  padding-left: var(--space-2);
  font-size: var(--font-size-12);
  font-weight: 800;
  color: var(--color-text-muted);
}

.studio-contest-heading {
  margin: 0;
  max-width: var(--ui-selector-width-lg);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--color-text-primary);
  font-size: var(--font-size-20);
  font-weight: 800;
  line-height: 1.25;
}

.studio-contest-meta {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  margin-left: var(--space-5);
}

.meta-tag {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  padding: var(--space-1) var(--space-2-5);
  border-radius: var(--ui-badge-radius-soft);
  font-size: var(--font-size-11);
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
  --ui-btn-height: var(--studio-toolbar-button-height);
  font-size: var(--font-size-12);
  font-weight: 700;
}

.ops-divider {
  width: 1px;
  height: var(--space-5);
  background: var(--color-border-subtle);
  margin: 0 var(--space-2);
}

.studio-save-btn {
  --ui-btn-height: var(--studio-toolbar-button-height);
  --ui-btn-padding: 0 var(--space-5);
  --ui-btn-radius: var(--ui-control-radius-md);
  --ui-btn-font-size: var(--font-size-12);
  --ui-btn-font-weight: 800;
  box-shadow: 0 var(--space-1) var(--space-3)
    color-mix(in srgb, var(--color-primary) 20%, transparent);
}
</style>
