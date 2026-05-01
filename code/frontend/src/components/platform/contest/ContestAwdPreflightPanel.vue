<script setup lang="ts">
import type { AWDReadinessData } from '@/api/contracts'

import AWDReadinessChecklist from './AWDReadinessChecklist.vue'
import AWDReadinessDecisionHUD from './AWDReadinessDecisionHUD.vue'

const props = defineProps<{
  readiness: AWDReadinessData | null
  loading: boolean
}>()

const emit = defineEmits<{
  'navigate:challenge': [challengeId: string]
  'navigate:stage': [stage: 'awd-config']
}>()

function handleNavigateChallenge(challengeId: string) {
  emit('navigate:challenge', challengeId)
  emit('navigate:stage', 'awd-config')
}
</script>

<template>
  <section class="studio-preflight">
    <header class="studio-pane-header">
      <div class="header-main">
        <div class="workspace-overline">
          AWD Preflight Check
        </div>
        <h1 class="pane-title">
          赛前就绪检查
        </h1>
        <p class="pane-description">
          全自动审计所有 AWD 题目与服务的配置状态，确保比赛在裁判逻辑完整的前提下开启。
        </p>
      </div>

      <div class="header-side">
        <AWDReadinessDecisionHUD :readiness="readiness" />
      </div>
    </header>

    <AWDReadinessChecklist
      :readiness="readiness"
      action-label="修正配置"
      data-primary-action-class="ui-btn ui-btn--primary"
      @edit-config="handleNavigateChallenge"
    />
  </section>
</template>

<style scoped>
.studio-preflight {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  padding: var(--space-6) var(--space-8);
  background: transparent;
}

.studio-pane-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--space-5);
  border-bottom: 1px solid var(--color-border-subtle);
  padding-bottom: var(--space-5);
}

.header-main {
  flex: 1 1 auto;
  min-width: 0;
}

.pane-title {
  font-size: var(--font-size-1-5);
  font-weight: 900;
  color: var(--color-text-primary);
  margin: var(--space-1) 0 0;
}

.pane-description {
  font-size: var(--font-size-14);
  color: var(--color-text-secondary);
  margin: var(--space-2) 0 0;
  max-width: var(--ui-selector-width-lg);
  line-height: 1.6;
}

.header-side {
  display: flex;
  flex: 0 0 auto;
  align-items: flex-start;
  gap: var(--space-4);
}

@media (max-width: 1280px) {
  .studio-pane-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-5);
  }

  .header-side {
    width: 100%;
    justify-content: space-between;
  }
}
</style>
