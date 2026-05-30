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
    <div class="header-actions topbar-actions">
      <button
        v-if="hasChallengeId"
        class="header-btn header-btn--primary"
        type="button"
        @click="handleOpenTopology"
      >
        拓扑编排
      </button>
      <button
        class="header-btn header-btn--ghost"
        type="button"
        @click="handleOpenChallengeList"
      >
        返回 Jeopardy题库
      </button>
    </div>
  </header>
</template>

<style scoped>
</style>
