<script setup lang="ts">
interface RoundSummary {
  id: string
  round_number: number
}

defineProps<{
  rounds: RoundSummary[]
  selectedRoundNumber?: number
}>()

const emit = defineEmits<{
  setRound: [roundNumber?: number]
}>()
</script>

<template>
  <section class="workspace-directory-section teacher-directory-section awd-review-round-section">
    <header class="list-heading">
      <div>
        <div class="journal-note-label">
          Review Scope
        </div>
        <h3 class="list-heading__title">
          轮次切换
        </h3>
      </div>
      <div class="teacher-directory-meta">
        默认展示整场总览；可切到单轮查看本轮服务、攻击和流量证据。
      </div>
    </header>

    <div class="awd-review-round-list custom-scrollbar">
      <button
        type="button"
        class="teacher-directory-chip awd-review-round-chip"
        :class="{ 'awd-review-round-chip--active': !selectedRoundNumber }"
        @click="emit('setRound', undefined)"
      >
        整场总览
      </button>
      <button
        v-for="round in rounds"
        :key="round.id"
        type="button"
        class="teacher-directory-chip awd-review-round-chip"
        :class="{ 'awd-review-round-chip--active': selectedRoundNumber === round.round_number }"
        @click="emit('setRound', round.round_number)"
      >
        R{{ round.round_number }}
      </button>
    </div>
  </section>
</template>

<style scoped>
.awd-review-round-list {
  display: flex;
  flex-wrap: nowrap;
  gap: var(--space-3);
  overflow-x: auto;
}

.awd-review-round-chip {
  border: 1px solid var(--awd-review-line);
  background: color-mix(in srgb, var(--awd-review-surface-subtle) 92%, transparent);
  color: var(--awd-review-muted);
}

.awd-review-round-chip--active {
  border-color: color-mix(in srgb, var(--awd-review-primary) 28%, transparent);
  background: color-mix(in srgb, var(--awd-review-primary) 12%, transparent);
  color: var(--awd-review-primary-strong);
}
</style>
