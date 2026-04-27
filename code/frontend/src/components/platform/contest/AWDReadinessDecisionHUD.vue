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
    class="decision-hud progress-card metric-panel-card metric-panel-default-surface"
    :class="readinessDecision.key"
  >
    <div class="journal-note-label progress-card-label metric-panel-label">
      就绪态势
      <ShieldCheck
        v-if="readinessDecision.key === 'ready'"
        class="h-4 w-4"
      />
      <AlertTriangle
        v-else
        class="h-4 w-4"
      />
    </div>
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
        <h3 class="decision-title progress-card-value metric-panel-value">
          {{ readinessDecision.title }}
        </h3>
        <p class="decision-description progress-card-hint metric-panel-helper">
          {{ readinessDecision.description }}
        </p>
      </div>
    </div>
    <div
      v-if="blockingActionLabels.length > 0"
      class="decision-meta"
    >
      <div class="impact-tags">
        <span
          v-for="label in blockingActionLabels"
          :key="label"
          class="impact-tag"
        >{{ label }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.decision-hud {
  --metric-panel-padding: var(--space-2-5) var(--space-3);
  --metric-panel-value-size: var(--font-size-18);
  --metric-panel-value-line-height: 1.2;
  --metric-panel-value-spacing: 0;
  --metric-panel-helper-size: var(--font-size-12);
  --metric-panel-helper-line-height: 1.35;

  display: flex;
  min-width: var(--ui-selector-width-md);
  flex-direction: column;
  align-items: stretch;
  gap: var(--space-2);
}

.decision-hud.ready {
  --metric-panel-border: color-mix(in srgb, var(--color-success) 30%, var(--color-border-default));
}

.decision-hud.blocked {
  --metric-panel-border: color-mix(in srgb, var(--color-danger) 30%, var(--color-border-default));
}

.decision-main {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  min-width: 0;
}

.decision-icon {
  width: var(--ui-control-height-sm);
  height: var(--ui-control-height-sm);
  border-radius: var(--ui-control-radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.ready .decision-icon {
  background: color-mix(in srgb, var(--color-success) 14%, var(--color-bg-surface));
  color: var(--color-success);
}

.blocked .decision-icon {
  background: color-mix(in srgb, var(--color-danger) 14%, var(--color-bg-surface));
  color: var(--color-danger);
}

.decision-text {
  min-width: 0;
}

.decision-title {
  margin: 0;
  color: var(--color-text-primary);
}

.decision-description {
  margin: 0;
  color: var(--color-text-secondary);
}

.impact-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-1-5);
}

.impact-tag {
  font-size: var(--font-size-10);
  font-weight: 800;
  padding: var(--space-0-5) var(--space-2);
  border-radius: var(--ui-badge-radius-soft);
  background: var(--color-bg-elevated);
  color: var(--color-text-secondary);
}

.ready .impact-tag {
  background: color-mix(in srgb, var(--color-success) 14%, var(--color-bg-surface));
  color: var(--color-success);
}

.blocked .impact-tag {
  background: color-mix(in srgb, var(--color-danger) 14%, var(--color-bg-surface));
  color: var(--color-danger);
}
</style>
