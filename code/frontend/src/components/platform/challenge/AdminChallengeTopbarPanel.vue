<script setup lang="ts">
defineProps<{
  workspaceLabel: string
  hasChallengeId: boolean
}>()

const emit = defineEmits<{
  openTopology: []
  openChallengeList: []
}>()

function handleOpenTopology(): void {
  emit('openTopology')
}

function handleOpenChallengeList(): void {
  emit('openChallengeList')
}
</script>

<template>
  <header class="workspace-topbar">
    <div class="topbar-leading">
      <span class="workspace-overline">Challenge Profile</span>
      <span class="class-chip">{{ workspaceLabel }}</span>
    </div>
    <div class="topbar-actions">
      <button
        v-if="hasChallengeId"
        class="ui-btn ui-btn--primary"
        type="button"
        @click="handleOpenTopology"
      >
        拓扑编排
      </button>
      <button
        class="ui-btn ui-btn--ghost"
        type="button"
        @click="handleOpenChallengeList"
      >
        返回 Jeopardy题库
      </button>
    </div>
  </header>
</template>

<style scoped>
.topbar-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.topbar-actions > .ui-btn {
  --ui-btn-height: 2.45rem;
  --ui-btn-radius: 0.75rem;
  --ui-btn-padding: var(--space-2) var(--space-4);
  --ui-btn-font-size: var(--font-size-0-875);
  --ui-btn-font-weight: 600;
  --ui-btn-primary-border: color-mix(in srgb, var(--journal-accent) 18%, transparent);
  --ui-btn-primary-background: var(--journal-accent);
  --ui-btn-primary-hover-background: color-mix(in srgb, var(--journal-accent) 88%, var(--color-bg-base));
  --ui-btn-ghost-color: var(--journal-ink);
  --ui-btn-ghost-hover-color: var(--journal-accent);
  --ui-btn-ghost-hover-background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
  --ui-btn-focus-ring: color-mix(in srgb, var(--journal-accent) 18%, transparent);
}
</style>
