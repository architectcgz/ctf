<script setup lang="ts">
import { computed, ref } from 'vue'

import AppEmpty from '@/components/common/AppEmpty.vue'
import AWDDefenseColumn from '@/components/contests/awd/AWDDefenseColumn.vue'
import AWDAttackVectorPanel from '@/components/contests/awd/AWDAttackVectorPanel.vue'
import AWDWorkspaceHudStrip from '@/components/contests/awd/AWDWorkspaceHudStrip.vue'
import AWDWorkspaceIntelColumn from '@/components/contests/awd/AWDWorkspaceIntelColumn.vue'
import ScoreboardRealtimeBridge from '@/components/scoreboard/ScoreboardRealtimeBridge.vue'
import {
  getVSCodeSSHCommand,
  toDefenseServiceCards,
  useAwdDefenseServiceSelection,
  useAwdWorkspaceAttackVector,
  useContestAWDWorkspace,
} from '@/features/contest-awd-workspace'
import type {
  ContestAWDWorkspaceServiceData,
  ContestChallengeItem,
  ContestDetailData,
} from '@/api/contracts'
import { useToast } from '@/composables/useToast'
import { formatTime } from '@/utils/format'

const props = defineProps<{
  contest: ContestDetailData
  challenges: ContestChallengeItem[]
}>()

const toast = useToast()
const copiedSSHCommandKey = ref('')
const copiedSSHPasswordKey = ref('')

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
  props.challenges.filter((item): item is ContestChallengeItem & { awd_service_id: string } =>
    Boolean(item.awd_service_id)
  )
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

const challengeByChallengeId = computed(() => {
  const map = new Map<string, ContestChallengeItem>()
  for (const item of props.challenges) {
    map.set(getAWDChallengeId(item), item)
  }
  return map
})

const challengeByServiceId = computed(() => {
  const map = new Map<string, ContestChallengeItem>()
  for (const item of props.challenges) {
    if (item.awd_service_id) {
      map.set(item.awd_service_id, item)
    }
  }
  return map
})

const currentRound = computed(() => workspace.value?.current_round)
const myTeam = computed(() => workspace.value?.my_team ?? null)
const selectedDefenseServiceCard = computed(
  () => defenseServiceCards.value.find((card) => card.serviceId === selectedServiceId.value) || null
)
const selectedDefenseServiceTitle = computed(() => selectedDefenseServiceCard.value?.title || '')
const selectedDefenseActionPending = computed(() =>
  selectedServiceId.value ? Boolean(defenseServiceActionPendingById.value[selectedServiceId.value]) : false
)
const selectedDefenseAccess = computed(() => getSSHAccess(selectedServiceId.value))
const selectedDefenseCopiedCommand = computed(() => copiedSSHCommandKey.value === selectedServiceId.value)
const selectedDefenseCopiedPassword = computed(
  () => copiedSSHPasswordKey.value === selectedServiceId.value
)
const currentRoundLabel = computed(() =>
  currentRound.value ? `#${String(currentRound.value.round_number).padStart(2, '0')}` : '--'
)
const currentRoundStatusLabel = computed(() => formatRoundStatusLabel(currentRound.value?.status))
const myTeamRank = computed(
  () => scoreboardRows.value.find((row) => row.team_id === myTeam.value?.team_id)?.rank || '--'
)
const serviceCount = computed(() => workspace.value?.services.length || 0)
const topScore = computed(() => scoreboardRows.value[0]?.score ?? 0)
const lastSyncedLabel = computed(() =>
  lastSyncedAt.value ? formatTime(lastSyncedAt.value) : '未同步'
)

const defenseAlerts = computed(() => {
  const items: Array<{
    challengeId: string
    challengeTitle: string
    statusLabel: string
    tone: 'danger' | 'warning'
    issues: string[]
  }> = []

  for (const challenge of runtimeChallenges.value) {
    const service = getWorkspaceService(challenge)
    if (!service) continue

    const issues: string[] = []
    let statusLabel = '正常'
    let tone: 'danger' | 'warning' = 'warning'

    if (service.service_status === 'compromised') {
      issues.push('已失陷')
      statusLabel = '严重'
      tone = 'danger'
    } else if (service.service_status === 'down' && service.instance_status !== 'running') {
      issues.push('已离线')
      statusLabel = '告警'
    }

    if ((service.attack_received ?? 0) > 0) {
      issues.push(`检测到 ${service.attack_received} 次攻击`)
    }

    if (issues.length === 0) continue

    items.push({
      challengeId: getAWDChallengeId(challenge),
      challengeTitle: challenge.title,
      statusLabel,
      tone,
      issues,
    })
  }
  return items
})

function eventDirectionLabel(direction: 'attack_in' | 'attack_out'): string {
  return direction === 'attack_out' ? '对外攻击' : '受到攻击'
}

function eventResultLabel(success: boolean): string {
  return success ? '成功' : '失败'
}

function formatServiceRef(serviceId?: string): string {
  return `服务 #${serviceId || '--'}`
}

function formatRoundStatusLabel(status?: string): string {
  switch (status) {
    case 'running':
      return '进行中'
    case 'frozen':
      return '已冻结'
    case 'finished':
    case 'completed':
    case 'ended':
      return '已结束'
    default:
      return '等待中'
  }
}

function getAWDChallengeId(challenge: ContestChallengeItem): string {
  return challenge.awd_challenge_id || challenge.challenge_id
}

function getChallengeTitleForEvent(event: {
  service_id?: string
  awd_challenge_id: string
}): string {
  if (event.service_id) {
    const matchedByService = challengeByServiceId.value.get(event.service_id)
    if (matchedByService) return matchedByService.title
  }
  return challengeByChallengeId.value.get(event.awd_challenge_id)?.title || event.awd_challenge_id
}

function getSubmitResultMessage(): string {
  if (!submitResult.value) return ''
  return formatAttackResultToast(submitResult.value)
}

function formatAttackResultToast(result: {
  service_id?: string
  awd_challenge_id: string
  is_success: boolean
  score_gained: number
}): string {
  const challengeTitle = getChallengeTitleForEvent(result)
  if (result.is_success) return `${challengeTitle}: 攻击成功，+${result.score_gained} 分`
  return `${challengeTitle}: 未获取到有效 Flag。`
}

function getWorkspaceService(
  challenge: ContestChallengeItem
): ContestAWDWorkspaceServiceData | undefined {
  if (!challenge.awd_service_id) return undefined
  return servicesByServiceId.value.get(challenge.awd_service_id)
}

function getSSHAccess(serviceId?: string) {
  if (!serviceId) return undefined
  return sshAccessByServiceId.value[serviceId]
}

function getSSHCommand(serviceId?: string): string {
  return getVSCodeSSHCommand(getSSHAccess(serviceId))
}

function openDefenseService(serviceId: string): void {
  const instanceId = servicesByServiceId.value.get(serviceId)?.instance_id
  if (!instanceId) return
  void openService(instanceId)
}

async function copyTextToClipboard(text: string, successMessage: string): Promise<boolean> {
  if (!text || typeof navigator === 'undefined' || !navigator.clipboard) {
    toast.error('复制失败，请手动选择文本')
    return false
  }

  try {
    await navigator.clipboard.writeText(text)
    toast.success(successMessage)
    return true
  } catch (err) {
    console.error(err)
    toast.error('复制失败，请手动选择文本')
    return false
  }
}

async function copySSHCommand(serviceId?: string): Promise<void> {
  if (!serviceId) return
  const copied = await copyTextToClipboard(getSSHCommand(serviceId), 'SSH 命令已复制')
  if (copied) {
    copiedSSHCommandKey.value = serviceId
  }
}

async function copySSHPassword(serviceId?: string): Promise<void> {
  if (!serviceId) return
  const password = getSSHAccess(serviceId)?.password || ''
  const copied = await copyTextToClipboard(password, 'SSH 密码已复制')
  if (copied) {
    copiedSSHPasswordKey.value = serviceId
  }
}

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
          :result-message="getSubmitResultMessage()"
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
