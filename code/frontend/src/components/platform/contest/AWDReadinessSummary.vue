<script setup lang="ts">
import { computed } from 'vue'

import type { AWDReadinessData } from '@/api/contracts'
import AWDReadinessChecklist from '@/components/platform/contest/AWDReadinessChecklist.vue'

const props = withDefaults(
  defineProps<{
    readiness: AWDReadinessData | null
    loading: boolean
    actionLabel?: string
    hideActions?: boolean
  }>(),
  {
    actionLabel: '编辑配置',
    hideActions: false,
  }
)

const emit = defineEmits<{
  editConfig: [challengeId: string]
}>()

const readinessDecision = computed(() => {
  const readiness = props.readiness
  const hasGlobalBlockers = (readiness?.global_blocking_reasons?.length ?? 0) > 0
  const hasChallengeBlockers = (readiness?.blocking_count ?? 0) > 0 || (readiness?.items?.length ?? 0) > 0

  if (readiness?.ready) {
    return {
      label: '可开赛',
      helper: '题目侧 checker 与系统级门禁均已满足开赛要求。',
      tone: 'ready',
    }
  }

  if (!hasGlobalBlockers && (readiness?.total_challenges ?? 0) > 0 && hasChallengeBlockers) {
    return {
      label: '待修复',
      helper: '题目侧仍有阻塞项，修复并重新校验后才能开赛。',
      tone: 'blocked',
    }
  }

  return {
    label: '不可开赛',
    helper: hasGlobalBlockers
      ? '系统级阻塞仍会拦截开赛关键动作。'
      : '当前仍有未完成的赛前条件，暂不建议开赛。',
    tone: 'blocked',
  }
})

</script>

<template>
  <div class="studio-readiness-flow">
    <header class="list-heading readiness-decision__head">
      <div>
        <div class="workspace-overline">
          AWD Readiness
        </div>
        <h2 class="list-heading__title">
          开赛就绪摘要
        </h2>
      </div>
    </header>

    <section
      v-if="readiness"
      class="workspace-directory-section readiness-decision-card"
      :class="`readiness-decision-card--${readinessDecision.tone}`"
    >
      <div class="journal-note-label readiness-decision-card__label">
        就绪决策
      </div>
      <div class="readiness-decision-card__value">
        {{ readinessDecision.label }}
      </div>
      <p class="readiness-decision-card__helper">
        {{ readinessDecision.helper }}
      </p>
    </section>

    <AWDReadinessChecklist
      :readiness="readiness"
      :action-label="actionLabel"
      :hide-actions="hideActions"
      @edit-config="emit('editConfig', $event)"
    />
  </div>
</template>

<style scoped>
.studio-readiness-flow { display: flex; flex-direction: column; gap: 2rem; }

.readiness-decision-card {
  display: grid;
  gap: var(--space-2);
  padding: var(--space-6);
}

.readiness-decision-card--ready {
  border-color: color-mix(in srgb, var(--color-success) 22%, var(--journal-border));
}

.readiness-decision-card--blocked {
  border-color: color-mix(in srgb, var(--color-danger) 24%, var(--journal-border));
}

.readiness-decision-card__label {
  color: var(--journal-muted);
}

.readiness-decision-card__value {
  font-size: var(--font-size-1-45);
  font-weight: 900;
  color: var(--journal-ink);
}

.readiness-decision-card__helper {
  margin: 0;
  color: var(--journal-muted);
  line-height: 1.7;
}
</style>
