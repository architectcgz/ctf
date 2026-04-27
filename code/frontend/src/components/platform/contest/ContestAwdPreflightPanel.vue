<script setup lang="ts">
import { computed } from 'vue'

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

        <!-- Override Entry - Compact -->
        <div
          v-if="canForceStart"
          class="preflight-override-entry"
        >
          <header class="list-heading contest-awd-preflight-panel__override-head">
            <div>
              <div class="journal-note-label">
                Override Entry
              </div>
              <h3 class="list-heading__title">
                强制启动赛事
              </h3>
            </div>
          </header>
          <button
            id="contest-awd-preflight-force-start"
            type="button"
            class="ui-btn ui-btn--primary"
            @click="emit('open:override')"
          >
            强制放行
          </button>
        </div>
      </div>
    </header>

    <AWDReadinessChecklist
      :readiness="readiness"
      action-label="修正配置"
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
  background: var(--color-bg-base);
}

.studio-pane-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  border-bottom: 1px solid var(--color-border-subtle);
  padding-bottom: var(--space-5);
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
  align-items: center;
  gap: var(--space-4);
}

.preflight-override-entry {
  display: grid;
  gap: var(--space-1-5);
  justify-items: end;
}

.preflight-override-entry__label {
  font-size: var(--font-size-11);
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.ops-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 2.25rem;
  padding: 0 1.25rem;
  border-radius: 0.75rem;
  font-size: 12px;
  font-weight: 800;
  cursor: pointer;
  transition: all 0.2s ease;
}

.ops-btn--primary {
  background: var(--color-warning);
  color: var(--color-bg-base);
  border: none;
  box-shadow: 0 4px 12px color-mix(in srgb, var(--color-warning) 20%, transparent);
}

.ops-btn--primary:hover {
  background: color-mix(in srgb, var(--color-warning) 90%, var(--color-bg-base));
  transform: translateY(-1px);
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
