<script setup lang="ts">
import { ExternalLink } from 'lucide-vue-next'

import type {
  ContestAWDWorkspaceTargetServiceData,
  ContestAWDWorkspaceTargetTeamData,
  ID,
} from '@/api/contracts'

type AttackTargetItem = ContestAWDWorkspaceTargetTeamData & {
  active_service?: ContestAWDWorkspaceTargetServiceData
}

const props = defineProps<{
  targets: AttackTargetItem[]
  activeChallengeRuntimeKey: string
  openingTargetKey: string
  submittingKey: string
  flagInputs: Record<string, string>
  formatServiceRef: (serviceId?: string) => string
}>()

const emit = defineEmits<{
  openTarget: [teamId: ID]
  updateFlag: [payload: { stateKey: string; value: string }]
  submit: [teamId: ID]
}>()

function buildAttackStateKey(teamId: string): string {
  return `${props.activeChallengeRuntimeKey}:${teamId}`
}
</script>

<template>
  <div class="target-grid">
    <article v-for="target in targets" :key="target.team_id" class="target-card">
      <div class="target-info">
        <div class="target-team font-black">
          {{ target.team_name.toUpperCase() }}
        </div>
        <div class="target-ref">
          {{ formatServiceRef(target.active_service?.service_id) }}
        </div>
        <div class="target-url font-mono">
          {{ target.active_service?.reachable ? '代理链路已就绪' : '不可达' }}
        </div>
      </div>
      <div class="target-action">
        <button
          :data-testid="`awd-open-target-${activeChallengeRuntimeKey}-${target.team_id}`"
          :disabled="
            !target.active_service?.reachable || openingTargetKey === buildAttackStateKey(target.team_id)
          "
          class="target-open-btn"
          type="button"
          @click="emit('openTarget', target.team_id)"
        >
          <ExternalLink class="h-3 w-3" />
          <span>{{
            openingTargetKey === buildAttackStateKey(target.team_id) ? '...' : '打开'
          }}</span>
        </button>
        <input
          :value="flagInputs[buildAttackStateKey(target.team_id)] || ''"
          placeholder="输入获取到的 Flag..."
          class="flag-input"
          @input="
            emit('updateFlag', {
              stateKey: buildAttackStateKey(target.team_id),
              value: String(($event.target as HTMLInputElement).value),
            })
          "
          @keyup.enter="emit('submit', target.team_id)"
        />
        <button
          :disabled="
            !target.active_service?.reachable || submittingKey === buildAttackStateKey(target.team_id)
          "
          class="submit-btn"
          @click="emit('submit', target.team_id)"
        >
          {{ submittingKey === buildAttackStateKey(target.team_id) ? '...' : '提交' }}
        </button>
      </div>
    </article>
  </div>
</template>

<style scoped>
.target-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(20rem, 1fr));
  gap: 1.25rem;
}

.target-card {
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border-default);
  padding: 1.25rem;
  border-radius: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.target-team {
  font-size: var(--font-size-14);
  letter-spacing: 0.05em;
  color: var(--color-primary);
}

.target-url {
  font-size: var(--font-size-11);
  color: var(--color-text-muted);
}

.target-ref {
  margin-top: 0.2rem;
  font-size: var(--font-size-11);
  font-weight: 800;
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
}

.target-action {
  display: flex;
  gap: 0.5rem;
}

.target-open-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-1);
  min-width: 4.5rem;
  border: 1px solid var(--color-border-default);
  border-radius: 0.5rem;
  background: var(--color-bg-surface);
  color: var(--color-text-secondary);
  font-size: var(--font-size-12);
  font-weight: 900;
  cursor: pointer;
  transition: all 0.2s ease;
}

.target-open-btn:hover:not(:disabled) {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.target-open-btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.flag-input {
  flex: 1;
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-default);
  padding: 0.5rem 0.75rem;
  border-radius: 0.5rem;
  color: var(--color-text-primary);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-13);
  outline: none;
  transition: border-color 0.2s ease;
}

.flag-input:focus {
  border-color: var(--color-primary);
}

.submit-btn {
  background: var(--color-danger);
  color: var(--color-bg-base);
  border: none;
  padding: 0 1.25rem;
  border-radius: 0.5rem;
  font-weight: 900;
  font-size: var(--font-size-12);
  cursor: pointer;
  transition: all 0.2s ease;
}

.submit-btn:hover:not(:disabled) {
  background: color-mix(in srgb, var(--color-danger) 80%, var(--color-bg-base));
}

.submit-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@media (max-width: 720px) {
  .target-action {
    flex-direction: column;
  }
}
</style>
