<script setup lang="ts">
import { Bell, ChevronLeft, Save, ShieldCheck, Trophy } from 'lucide-vue-next'

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

function handleBack(): void {
  emit('back')
}

function handleOpenAnnouncements(): void {
  emit('openAnnouncements')
}

function handleSave(): void {
  emit('save')
}
</script>

<template>
  <div class="workspace-topbar">
    <header class="studio-topbar">
      <div class="studio-topbar-left">
        <button
          type="button"
          class="ui-btn ui-btn--ghost studio-back-btn"
          title="返回竞赛目录"
          @click="handleBack"
        >
          <ChevronLeft class="h-5 w-5" />
          返回竞赛目录
        </button>

        <div class="workspace-topbar__main studio-title-group">
          <div class="workspace-overline">Contest Editor</div>
          <h1 class="workspace-page-title">编辑竞赛</h1>
          <p
            class="workspace-page-copy studio-contest-title"
            :title="pageTitle"
          >
            {{ pageTitle }}
          </p>
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
      </div>

      <div class="studio-topbar-right">
        <button
          id="contest-open-announcements"
          type="button"
          class="ui-btn ui-btn--ghost studio-toolbar-btn"
          @click="handleOpenAnnouncements"
        >
          <Bell class="h-4 w-4" />
          <span>公告</span>
        </button>
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
  </div>
</template>

<style scoped>
.studio-topbar {
  height: 4.5rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 2rem;
  background: var(--color-bg-surface);
  border-bottom: 1px solid var(--workspace-line-soft);
  z-index: 10;
}

.studio-topbar-left {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.studio-topbar-right {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.studio-back-btn {
  width: 2.5rem;
  height: 2.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.75rem;
  color: var(--journal-muted);
  border: 1px solid var(--workspace-line-soft);
  transition: all 0.2s ease;
  cursor: pointer;
}

.studio-back-btn:hover {
  background: var(--color-bg-elevated);
  color: var(--color-text-primary);
  border-color: var(--color-border-default);
}

.studio-toolbar-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.65rem;
  height: 2.4rem;
  padding: 0 1rem;
  border-radius: 0.85rem;
  border: 1px solid var(--workspace-line-soft);
  background: var(--color-bg-surface);
  color: var(--color-text-secondary);
  font-size: 12px;
  font-weight: 800;
  transition: all 0.2s ease;
}

.studio-toolbar-btn:hover {
  background: var(--color-bg-elevated);
  color: var(--color-text-primary);
  border-color: var(--color-border-default);
}

.studio-save-btn {
  --ui-btn-height: 2.4rem;
  --ui-btn-padding: 0 1.25rem;
  --ui-btn-radius: 0.85rem;
  --ui-btn-font-size: var(--font-size-12);
  --ui-btn-font-weight: 800;
  --ui-btn-primary-hover-shadow: 0 10px 24px color-mix(in srgb, var(--color-primary) 30%, transparent);
  --ui-btn-hover-transform: translateY(-1px);
  box-shadow: 0 8px 20px color-mix(in srgb, var(--color-primary) 24%, transparent);
}

.studio-title-group {
  display: flex;
  align-items: baseline;
  gap: 1.25rem;
}

.studio-contest-title {
  font-size: 1rem;
  font-weight: 900;
  letter-spacing: -0.01em;
  color: var(--color-text-primary);
  margin: 0;
  max-width: 24rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.studio-contest-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.meta-tag {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.1rem 0.55rem;
  border-radius: 4px;
  font-size: 9px;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border: 1px solid transparent;
}

.meta-tag--awd {
  background: color-mix(in srgb, var(--color-primary) 8%, transparent);
  color: var(--color-primary);
  border-color: color-mix(in srgb, var(--color-primary) 20%, transparent);
}

.meta-tag--status {
  background: color-mix(in srgb, var(--journal-muted) 8%, transparent);
  color: var(--journal-muted);
  border-color: color-mix(in srgb, var(--journal-muted) 20%, transparent);
}
</style>
