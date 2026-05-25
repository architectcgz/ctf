<script setup lang="ts">
defineProps<{
  challengeOptions: Array<{
    key: string
    title: string
  }>
  activeChallengeKey: string
  targetKeyword: string
}>()

const emit = defineEmits<{
  'update:activeChallengeKey': [value: string]
  'update:targetKeyword': [value: string]
}>()
</script>

<template>
  <div class="ops-panel__toolbar">
    <div class="toolbar-field">
      <label for="awd-target-challenge">目标题目</label>
      <select
        id="awd-target-challenge"
        :value="activeChallengeKey"
        class="war-room-select"
        @change="emit('update:activeChallengeKey', String(($event.target as HTMLSelectElement).value))"
      >
        <option v-for="challenge in challengeOptions" :key="challenge.key" :value="challenge.key">
          {{ challenge.title }}
        </option>
      </select>
    </div>
    <div class="toolbar-field">
      <label for="awd-target-search">队伍筛选</label>
      <input
        id="awd-target-search"
        :value="targetKeyword"
        type="text"
        placeholder="按队伍名称筛选..."
        class="war-room-input"
        @input="emit('update:targetKeyword', String(($event.target as HTMLInputElement).value))"
      />
    </div>
  </div>
</template>

<style scoped>
.ops-panel__toolbar {
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--color-border-subtle);
  display: flex;
  gap: 1rem;
  background: var(--color-bg-surface);
}

.toolbar-field {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  flex: 1;
}

.toolbar-field label {
  font-size: 9px;
  font-weight: 900;
  color: var(--color-text-muted);
  letter-spacing: 0.1em;
}

.war-room-select,
.war-room-input {
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border-default);
  border-radius: 0.5rem;
  color: var(--color-text-primary);
  font-size: var(--font-size-12);
  font-weight: 700;
  padding: 0.5rem 0.75rem;
  outline: none;
  transition: all 0.2s ease;
}

.war-room-select:focus,
.war-room-input:focus {
  border-color: var(--color-primary);
}

@media (max-width: 720px) {
  .ops-panel__toolbar {
    flex-direction: column;
  }
}
</style>
