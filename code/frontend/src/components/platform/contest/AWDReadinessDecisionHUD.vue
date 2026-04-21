<script setup lang="ts">
import { computed } from 'vue'
import { ShieldCheck, AlertTriangle } from 'lucide-vue-next'
import type { AWDReadinessData } from '@/api/contracts'

const props = defineProps<{
  readiness: AWDReadinessData | null
}>()

const readinessDecision = computed(() => {
  if (!props.readiness) return { key: 'pending', title: '正在审计...', description: '请稍候，系统正在扫描题目状态。' }
  if (props.readiness.ready) return { key: 'ready', title: '环境已就绪', description: '所有服务均通过验证。' }
  return { key: 'blocked', title: '存在阻塞风险', description: '部分题目校验失败。' }
})

const blockingActionLabels = computed(() => {
  if (!props.readiness) return []
  const labels: string[] = []
  const actions = props.readiness.blocking_actions || []
  if (actions.includes('start_contest')) labels.push('开启比赛')
  if (actions.includes('create_round')) labels.push('创建轮次')
  if (actions.includes('run_current_round_check')) labels.push('即时巡检')
  return labels
})
</script>

<template>
  <div
    v-if="readiness"
    class="decision-hud"
    :class="readinessDecision.key"
  >
    <div class="decision-main">
      <div class="decision-icon">
        <ShieldCheck
          v-if="readinessDecision.key === 'ready'"
          class="h-5 w-5"
        />
        <AlertTriangle
          v-else
          class="h-5 w-5"
        />
      </div>
      <div class="decision-text">
        <h3 class="decision-title">
          {{ readinessDecision.title }}
        </h3>
        <p class="decision-description">
          {{ readinessDecision.description }}
        </p>
      </div>
    </div>
    <div class="decision-meta">
      <div class="impact-tags">
        <span
          v-for="label in blockingActionLabels"
          :key="label"
          class="impact-tag"
        >{{ label }}</span>
        <span
          v-if="blockingActionLabels.length === 0"
          class="impact-tag neutral"
        >无阻塞</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.decision-hud {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  padding: 0.75rem 1.25rem;
  border-radius: 0.85rem;
  border: 1px solid var(--color-border-default);
  background: var(--color-bg-surface);
  min-width: 24rem;
}

.decision-hud.ready { border-color: color-mix(in srgb, var(--color-success) 30%, var(--color-border-default)); }
.decision-hud.blocked { border-color: color-mix(in srgb, var(--color-danger) 30%, var(--color-border-default)); }

.decision-main { display: flex; align-items: center; gap: 0.75rem; flex: 1; }
.decision-icon {
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 0.65rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.ready .decision-icon { background: var(--color-success-soft); color: var(--color-success); }
.blocked .decision-icon { background: var(--color-danger-soft); color: var(--color-danger); }

.decision-title { font-size: var(--font-size-13); font-weight: 900; margin: 0; color: var(--color-text-primary); white-space: nowrap; }
.decision-description { font-size: var(--font-size-12); color: var(--color-text-secondary); font-weight: 500; margin: 0; white-space: nowrap; }

.impact-tags { display: flex; gap: 0.35rem; }
.impact-tag { font-size: 10px; font-weight: 800; padding: 0.15rem 0.5rem; border-radius: 4px; background: var(--color-bg-elevated); color: var(--color-text-secondary); }
.impact-tag.neutral { opacity: 0.5; }
.ready .impact-tag { background: var(--color-success-soft); color: var(--color-success); }
.blocked .impact-tag { background: var(--color-danger-soft); color: var(--color-danger); }
</style>