<script setup lang="ts">
import { computed } from 'vue'

import type { AWDReadinessData } from '@/api/contracts'

import AWDReadinessSummary from './AWDReadinessSummary.vue'

const props = defineProps<{
  readiness: AWDReadinessData | null
  loading: boolean
}>()

const emit = defineEmits<{
  'navigate:challenge': [challengeId: string]
  'navigate:stage': [stage: 'awd-config']
  'open:override': []
}>()

const canForceStart = computed(
  () => Boolean(props.readiness) && !props.readiness?.ready && (props.readiness?.global_blocking_reasons?.length ?? 0) === 0
)

function handleNavigateChallenge(challengeId: string) {
  emit('navigate:challenge', challengeId)
  emit('navigate:stage', 'awd-config')
}
</script>

<template>
  <section class="contest-awd-preflight-panel">
    <header class="contest-awd-preflight-panel__header">
      <div class="workspace-tab-heading__main">
        <div class="workspace-overline">Preflight</div>
        <h2 class="workspace-page-title">赛前检查</h2>
        <p class="workspace-page-copy">
          开赛前先看这里，集中确认还有哪些题目需要回到 AWD 配置继续补齐。
        </p>
      </div>

      <section
        v-if="canForceStart"
        class="workspace-directory-section contest-awd-preflight-panel__override"
      >
        <div>
          <div class="journal-note-label">Override Entry</div>
          <h3 class="list-heading__title">强制开赛</h3>
          <p class="contest-awd-preflight-panel__override-copy">
            如果这是演练或临时放行场景，可以直接打开强制开赛弹层，保留原因说明。
          </p>
        </div>
        <button
          id="contest-awd-preflight-force-start"
          type="button"
          class="ui-btn ui-btn--primary"
          @click="emit('open:override')"
        >
          强制开赛
        </button>
      </section>
    </header>

    <AWDReadinessSummary
      :readiness="readiness"
      :loading="loading"
      action-label="返回 AWD 配置"
      @edit-config="handleNavigateChallenge"
    />
  </section>
</template>

<style scoped>
.contest-awd-preflight-panel {
  display: grid;
  gap: var(--space-5);
}

.contest-awd-preflight-panel__header {
  display: grid;
  gap: var(--space-4);
}

.contest-awd-preflight-panel__override {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.25rem 1.35rem;
}

.contest-awd-preflight-panel__override-copy {
  margin: 0.45rem 0 0;
  max-width: 44rem;
  color: var(--journal-ink);
  line-height: 1.7;
}

.list-heading__title {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-1-20);
  font-weight: 700;
  color: var(--journal-ink);
}

@media (max-width: 767px) {
  .contest-awd-preflight-panel__override {
    align-items: flex-start;
  }
}
</style>
