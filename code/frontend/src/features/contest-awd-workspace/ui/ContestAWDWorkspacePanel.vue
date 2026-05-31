<script setup lang="ts">
import { computed } from 'vue'

import AppEmpty from '@/shared/ui/common/AppEmpty.vue'
import { ScoreboardRealtimeBridge } from '@/features/scoreboard'
import {
  isAwdRuntimeChallenge,
  toDefenseServiceCards,
  useAwdDefenseAccessPanel,
  useAwdDefenseServiceSelection,
  useAwdWorkspaceAttackVector,
  useAwdWorkspacePresentation,
  useAwdWorkspaceSummary,
  useContestAWDWorkspace,
} from '../model'
import type {
  ContestAWDWorkspaceServiceData,
  ContestChallengeItem,
  ContestDetailData,
} from '@/api/contracts'
import AWDAttackVectorPanel from './AWDAttackVectorPanel.vue'
import AWDDefenseColumn from './AWDDefenseColumn.vue'
import AWDWorkspaceHudStrip from './AWDWorkspaceHudStrip.vue'
import AWDWorkspaceIntelColumn from './AWDWorkspaceIntelColumn.vue'

const props = defineProps<{
  contest: ContestDetailData
  challenges: ContestChallengeItem[]
}>()

const {
  getChallengeTitleForEvent,
  formatAttackResultToast,
  eventDirectionLabel,
  eventResultLabel,
  formatServiceRef,
} = useAwdWorkspacePresentation({
  challenges: computed(() => props.challenges),
})

const {
  workspace,
  scoreboardRows,
  loading,
  hasTeam,
  submitResult,
  startingServiceKey,
  serviceActionPendingById,
  openingServiceKey,
  openingSSHKey,
  sshAccessByServiceId,
  openingTargetKey,
  submittingKey,
  lastSyncedAt,
  refreshAll,
  restartService,
  openService,
  openDefenseSSH,
  openTarget,
  submitAttack,
} = useContestAWDWorkspace({
  contestId: computed(() => props.contest.id),
  contestStatus: computed(() => props.contest.status),
  formatAttackResultToast,
})

const servicesByServiceId = computed(() => {
  const map = new Map<string, ContestAWDWorkspaceServiceData>()
  for (const item of workspace.value?.services || []) {
    if (item.service_id) {
      map.set(item.service_id, item)
    }
  }
  return map
})

const runtimeChallenges = computed(() =>
  props.challenges.filter(isAwdRuntimeChallenge)
)
const {
  activeChallengeKey,
  flagInputs,
  targetKeyword,
  attackToolbarChallengeOptions,
  activeChallenge,
  activeChallengeRuntimeKey,
  filteredTargets,
  handleSubmit,
} = useAwdWorkspaceAttackVector({
  challenges: computed(() => runtimeChallenges.value),
  targets: computed(() => workspace.value?.targets || []),
  submitAttack,
})

const defenseServiceCards = computed(() =>
  toDefenseServiceCards({
    challenges: props.challenges,
    services: workspace.value?.services || [],
  })
)

const defenseServiceActionPendingById = computed(() => {
  const pendingById: Record<string, boolean> = {}
  for (const card of defenseServiceCards.value) {
    const service = servicesByServiceId.value.get(card.serviceId)
    pendingById[card.serviceId] = Boolean(
      startingServiceKey.value === card.serviceId ||
      serviceActionPendingById.value[card.serviceId] ||
      service?.instance_status === 'pending' ||
      service?.instance_status === 'creating' ||
      service?.operation_status === 'requested' ||
      service?.operation_status === 'provisioning' ||
      service?.operation_status === 'recovering'
    )
  }
  return pendingById
})

const { selectedServiceId, selectService } = useAwdDefenseServiceSelection(defenseServiceCards)

const selectedDefenseServiceCard = computed(
  () => defenseServiceCards.value.find((card) => card.serviceId === selectedServiceId.value) || null
)
const selectedDefenseServiceTitle = computed(() => selectedDefenseServiceCard.value?.title || '')
const selectedDefenseActionPending = computed(() =>
  selectedServiceId.value ? Boolean(defenseServiceActionPendingById.value[selectedServiceId.value]) : false
)
const {
  selectedDefenseAccess,
  selectedDefenseCopiedCommand,
  selectedDefenseCopiedPassword,
  openDefenseService,
  copySSHCommand,
  copySSHPassword,
} = useAwdDefenseAccessPanel({
  selectedServiceId: computed(() => selectedServiceId.value),
  servicesByServiceId: computed(() => servicesByServiceId.value),
  sshAccessByServiceId: computed(() => sshAccessByServiceId.value),
  openService,
})
const {
  myTeam,
  currentRoundLabel,
  currentRoundStatusLabel,
  myTeamRank,
  serviceCount,
  topScore,
  lastSyncedLabel,
  defenseAlerts,
} = useAwdWorkspaceSummary({
  workspace: computed(() => workspace.value),
  scoreboardRows: computed(() => scoreboardRows.value),
  runtimeChallenges: computed(() => runtimeChallenges.value),
  servicesByServiceId: computed(() => servicesByServiceId.value),
  lastSyncedAt: computed(() => lastSyncedAt.value),
})
const submitResultMessage = computed(() =>
  submitResult.value ? formatAttackResultToast(submitResult.value) : ''
)

</script>

<template>
  <div class="awd-war-room">
    <ScoreboardRealtimeBridge
      v-if="contest.status === 'running' || contest.status === 'frozen'"
      :contest-id="contest.id"
      @updated="refreshAll"
    />

    <AWDWorkspaceHudStrip
      :current-round-label="currentRoundLabel"
      :current-round-status-label="currentRoundStatusLabel"
      :team-name="myTeam?.team_name || '未加入'"
      :team-rank="myTeamRank"
      :service-count="serviceCount"
      :top-score="topScore"
      :last-synced-label="lastSyncedLabel"
      :loading="loading"
      @refresh="refreshAll"
    />

    <div v-if="loading && !workspace" class="war-room-loading">
      <div class="radar-scan" />
      <p>正在建立战场连接...</p>
    </div>

    <AppEmpty
      v-else-if="!hasTeam"
      icon="Users"
      title="先加入队伍"
      description="需要先加入队伍后才能进入 AWD 战场。"
      class="war-room-empty"
    />

    <div v-else class="war-room-grid">
      <AWDDefenseColumn
        :alerts="defenseAlerts"
        :services="defenseServiceCards"
        :selected-service-id="selectedServiceId"
        :opening-service-key="openingServiceKey"
        :opening-ssh-key="openingSSHKey"
        :service-action-pending-by-id="defenseServiceActionPendingById"
        :service-card="selectedDefenseServiceCard"
        :service-title="selectedDefenseServiceTitle"
        :action-pending="selectedDefenseActionPending"
        :loading="loading"
        :access="selectedDefenseAccess"
        :copied-command="selectedDefenseCopiedCommand"
        :copied-password="selectedDefenseCopiedPassword"
        @select-service="selectService"
        @open-service="openDefenseService"
        @request-ssh="openDefenseSSH"
        @restart-service="restartService"
        @refresh="refreshAll"
        @copy-command="copySSHCommand"
        @copy-password="copySSHPassword"
      />

      <!-- 2. Attack Vector (Middle) -->
      <main class="war-room-col column-attack">
        <AWDAttackVectorPanel
          :challenge-options="attackToolbarChallengeOptions"
          :active-challenge-key="activeChallengeKey"
          :target-keyword="targetKeyword"
          :has-active-challenge="Boolean(activeChallenge)"
          :targets="filteredTargets"
          :active-challenge-runtime-key="activeChallengeRuntimeKey"
          :opening-target-key="openingTargetKey"
          :submitting-key="submittingKey"
          :flag-inputs="flagInputs"
          :show-result="Boolean(submitResult)"
          :result-success="submitResult?.is_success ?? false"
          :result-message="submitResultMessage"
          :format-service-ref="formatServiceRef"
          @update:active-challenge-key="activeChallengeKey = $event"
          @update:target-keyword="targetKeyword = $event"
          @open-target="openTarget(activeChallengeRuntimeKey, $event)"
          @update-flag="flagInputs[$event.stateKey] = $event.value"
          @submit="handleSubmit(activeChallengeRuntimeKey, $event)"
        />
      </main>

      <AWDWorkspaceIntelColumn
        :scoreboard-rows="scoreboardRows"
        :my-team-id="myTeam?.team_id"
        :recent-events="workspace?.recent_events || []"
        :get-challenge-title-for-event="getChallengeTitleForEvent"
        :event-direction-label="eventDirectionLabel"
        :event-result-label="eventResultLabel"
        :format-service-ref="formatServiceRef"
      />
    </div>
  </div>
</template>

<style scoped>
.awd-war-room {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
  padding-top: var(--space-4);
}

/* Layout Grid */
.war-room-grid {
  display: grid;
  grid-template-columns: minmax(20rem, 24rem) minmax(0, 1fr);
  grid-template-areas:
    'defense attack'
    'defense intel';
  gap: var(--space-5);
  flex: 1;
  min-height: 0;
  align-items: stretch;
}

.column-attack {
  grid-area: attack;
}

.war-room-loading {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2rem;
  font-size: 14px;
  font-weight: 900;
  color: var(--color-primary);
  padding: 4rem 0;
}

.radar-scan {
  width: 5rem;
  height: 5rem;
  border: 2px solid var(--color-primary);
  border-radius: 50%;
  position: relative;
}

.radar-scan::after {
  content: '';
  position: absolute;
  inset: 0;
  background: conic-gradient(from 0deg, var(--color-primary), transparent);
  border-radius: 50%;
  animation: radar 2s linear infinite;
  opacity: 0.3;
}

@keyframes radar {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 1280px) {
  .war-room-grid {
    grid-template-columns: 1fr;
    grid-template-areas:
      'defense'
      'attack'
      'intel';
  }
}

</style>
