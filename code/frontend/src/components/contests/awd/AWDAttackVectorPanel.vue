<script setup lang="ts">
import { Sword } from 'lucide-vue-next'

import type {
  ContestAWDWorkspaceTargetServiceData,
  ContestAWDWorkspaceTargetTeamData,
  ID,
} from '@/api/contracts'
import AWDAttackResultFooter from './AWDAttackResultFooter.vue'
import AWDAttackTargetGrid from './AWDAttackTargetGrid.vue'
import AWDAttackToolbar from './AWDAttackToolbar.vue'

type AttackTargetItem = ContestAWDWorkspaceTargetTeamData & {
  active_service?: ContestAWDWorkspaceTargetServiceData
}

defineProps<{
  challengeOptions: Array<{
    key: string
    title: string
  }>
  activeChallengeKey: string
  targetKeyword: string
  hasActiveChallenge: boolean
  targets: AttackTargetItem[]
  activeChallengeRuntimeKey: string
  openingTargetKey: string
  submittingKey: string
  flagInputs: Record<string, string>
  showResult: boolean
  resultSuccess: boolean
  resultMessage: string
  formatServiceRef: (serviceId?: string) => string
}>()

const emit = defineEmits<{
  'update:activeChallengeKey': [value: string]
  'update:targetKeyword': [value: string]
  openTarget: [teamId: ID]
  updateFlag: [payload: { stateKey: string; value: string }]
  submit: [teamId: ID]
}>()
</script>

<template>
  <section class="ops-panel">
    <header class="ops-panel__header">
      <Sword class="ops-panel__icon ops-panel__icon--danger h-4 w-4" />
      <h3 class="ops-panel__title">攻击向量</h3>
    </header>

    <AWDAttackToolbar
      :challenge-options="challengeOptions"
      :active-challenge-key="activeChallengeKey"
      :target-keyword="targetKeyword"
      @update:active-challenge-key="emit('update:activeChallengeKey', $event)"
      @update:target-keyword="emit('update:targetKeyword', $event)"
    />

    <div class="ops-panel__content custom-scrollbar">
      <div v-if="challengeOptions.length === 0" class="panel-note">
        当前竞赛暂无可部署服务。
      </div>
      <div v-else-if="!hasActiveChallenge" class="panel-note">请选择目标题目后开始攻击。</div>
      <div v-else-if="targets.length === 0" class="panel-note">
        当前题目下没有匹配的目标队伍。
      </div>
      <AWDAttackTargetGrid
        v-else
        :targets="targets"
        :active-challenge-runtime-key="activeChallengeRuntimeKey"
        :opening-target-key="openingTargetKey"
        :submitting-key="submittingKey"
        :flag-inputs="flagInputs"
        :format-service-ref="formatServiceRef"
        @open-target="emit('openTarget', $event)"
        @update-flag="emit('updateFlag', $event)"
        @submit="emit('submit', $event)"
      />
    </div>

    <AWDAttackResultFooter
      v-if="showResult"
      :success="resultSuccess"
      :message="resultMessage"
    />
  </section>
</template>

<style scoped>
.ops-panel {
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-default);
  border-radius: 1rem;
  display: flex;
  flex-direction: column;
  min-height: 0;
  box-shadow: var(--color-shadow-soft);
  overflow: hidden;
}

.ops-panel__header {
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--color-border-subtle);
  display: flex;
  align-items: center;
  gap: 0.75rem;
  background: var(--color-bg-elevated);
}

.ops-panel__icon--danger {
  color: var(--color-danger);
}

.ops-panel__title {
  font-size: var(--font-size-12);
  font-weight: 900;
  letter-spacing: 0.15em;
  color: var(--color-text-primary);
  margin: 0;
}

.ops-panel__content {
  flex: 1;
  overflow-y: auto;
  padding: 1.25rem;
}

.panel-note {
  font-size: var(--font-size-12);
  font-weight: 700;
  color: var(--color-text-muted);
  text-align: center;
  padding: 3rem 0;
}
</style>
