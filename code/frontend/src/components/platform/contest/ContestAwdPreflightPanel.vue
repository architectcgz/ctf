<script setup lang="ts">
import { computed } from 'vue'

import type { AWDReadinessData } from '@/api/contracts'

import AWDReadinessSummary from './AWDReadinessSummary.vue'
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
        <div class="workspace-overline">AWD Preflight Check</div>
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
          <span class="preflight-override-entry__label">强制启动赛事</span>
          <button
            id="contest-awd-preflight-force-start"
            type="button"
            class="ops-btn ops-btn--primary"
            @click="emit('open:override')"
          >
            强制放行
          </button>
        </div>
      </div>
    </header>

    <AWDReadinessSummary
      :readiness="readiness"
      :loading="loading"
      action-label="修正配置"
      @edit-config="handleNavigateChallenge"
    />
  </section>
</template>

<style scoped>
.studio-preflight {
  display: flex;
  flex-direction: column;
  gap: 2rem;
  padding: 1.5rem 2rem;
  background: var(--color-bg-base);
}

.studio-pane-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  border-bottom: 1px solid var(--color-border-subtle);
  padding-bottom: 1.5rem;
}

.pane-title {
  font-size: 1.5rem;
  font-weight: 900;
  color: var(--color-text-primary);
  margin: 0.25rem 0 0;
}

.pane-description {
  font-size: 14px;
  color: var(--color-text-secondary);
  margin: 0.5rem 0 0;
  max-width: 32rem;
  line-height: 1.6;
}

.header-side {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.preflight-override-entry {
  display: grid;
  gap: 0.4rem;
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
  color: white;
  border: none;
  box-shadow: 0 4px 12px color-mix(in srgb, var(--color-warning) 20%, transparent);
}

.ops-btn--primary:hover {
  background: color-mix(in srgb, var(--color-warning) 90%, black);
  transform: translateY(-1px);
}

@media (max-width: 1280px) {
  .studio-pane-header { flex-direction: column; align-items: flex-start; gap: 1.5rem; }
  .header-side { width: 100%; justify-content: space-between; }
}
</style>
