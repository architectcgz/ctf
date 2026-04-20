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
  <section class="studio-preflight">
    <header class="studio-pane-header">
      <div class="header-main">
        <h1 class="pane-title">赛前就绪检查</h1>
        <p class="pane-description">全自动审计所有 AWD 题目与服务的配置状态，确保比赛在裁判逻辑完整的前提下开启。</p>
      </div>

      <!-- Override Entry with a more balanced style -->
      <div v-if="canForceStart" class="studio-override-card">
        <div class="override-content">
          <div class="override-overline">Operational Bypass</div>
          <h3 class="override-title">强制启动赛事</h3>
          <p class="override-hint">针对紧急演练或特定场景，可跳过就绪校验。请保留操作备注。</p>
        </div>
        <button
          id="contest-awd-preflight-force-start"
          type="button"
          class="ops-btn ops-btn--primary"
          @click="emit('open:override')"
        >
          强制放行
        </button>
      </div>
    </header>

    <div class="studio-preflight-body">
      <AWDReadinessSummary
        :readiness="readiness"
        :loading="loading"
        action-label="修正配置"
        @edit-config="handleNavigateChallenge"
      />
    </div>
  </section>
</template>

<style scoped>
.studio-preflight {
  display: flex;
  flex-direction: column;
  gap: 2rem;
  padding: 1.5rem 2rem;
  background: #fdfdfd;
}

.studio-pane-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.pane-title {
  font-size: 1.25rem;
  font-weight: 900;
  color: #0f172a;
  margin: 0;
}

.pane-description {
  font-size: 13px;
  color: #64748b;
  margin: 0.5rem 0 0;
  max-width: 32rem;
  line-height: 1.6;
}

/* Override Card Styles */
.studio-override-card {
  background: #fff7ed;
  border: 1px solid #ffedd5;
  border-radius: 1rem;
  padding: 1.25rem 1.5rem;
  display: flex;
  align-items: center;
  gap: 2rem;
  max-width: 40rem;
}

.override-overline { font-size: 9px; font-weight: 800; text-transform: uppercase; color: #c2410c; letter-spacing: 0.1em; }
.override-title { font-size: 14px; font-weight: 900; color: #9a3412; margin: 0.15rem 0; }
.override-hint { font-size: 11px; color: #9a3412; opacity: 0.8; margin: 0; }

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
  background: #f97316;
  color: white;
  border: none;
  box-shadow: 0 4px 12px rgba(249, 115, 22, 0.2);
}

.ops-btn--primary:hover {
  background: #ea580c;
  transform: translateY(-1px);
}

.studio-preflight-body {
  background: white;
  border-radius: 1.25rem;
  border: 1px solid #e2e8f0;
  overflow: hidden;
}

@media (max-width: 1280px) {
  .studio-pane-header { flex-direction: column; gap: 1.5rem; }
  .studio-override-card { max-width: 100%; width: 100%; }
}
</style>
