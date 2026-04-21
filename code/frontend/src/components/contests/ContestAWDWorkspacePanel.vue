<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { 
  ShieldAlert, 
  Sword, 
  Target, 
  Activity, 
  Wifi, 
  Zap, 
  BarChart3, 
  History, 
  Terminal,
  ExternalLink,
  RefreshCw,
  AlertTriangle
} from 'lucide-vue-next'

import AppEmpty from '@/components/common/AppEmpty.vue'
import ScoreboardRealtimeBridge from '@/components/scoreboard/ScoreboardRealtimeBridge.vue'
import { useContestAWDWorkspace } from '@/composables/useContestAWDWorkspace'
import type {
  ContestAWDWorkspaceServiceData,
  ContestChallengeItem,
  ContestDetailData,
} from '@/api/contracts'
import { formatTime } from '@/utils/format'

const props = defineProps<{
  contest: ContestDetailData
  challenges: ContestChallengeItem[]
}>()

const activeChallengeKey = ref('')
const flagInputs = ref<Record<string, string>>({})
const targetKeyword = ref('')
const showOnlyReachableTargets = ref(false)

const {
  workspace,
  scoreboardRows,
  loading,
  error,
  hasTeam,
  submitResult,
  startingServiceKey,
  submittingKey,
  shouldAutoRefresh,
  lastSyncedAt,
  refreshAll,
  startService,
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
  props.challenges.filter(
    (item): item is ContestChallengeItem & { awd_service_id: string } => Boolean(item.awd_service_id)
  )
)

const challengeByChallengeId = computed(() => {
  const map = new Map<string, ContestChallengeItem>()
  for (const item of props.challenges) {
    map.set(item.challenge_id, item)
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
const topScore = computed(() => scoreboardRows.value[0]?.score ?? 0)
const lastSyncedLabel = computed(() =>
  lastSyncedAt.value ? formatTime(lastSyncedAt.value) : 'UNSYNCED'
)
const targetFilterKeyword = computed(() => targetKeyword.value.trim().toLowerCase())

const activeChallenge = computed(
  () =>
    runtimeChallenges.value.find((item) => getChallengeRuntimeKey(item) === activeChallengeKey.value) ||
    null
)
const activeChallengeRuntimeKey = computed(() => getChallengeRuntimeKey(activeChallenge.value))

const activeTargets = computed(() => {
  if (!workspace.value || !activeChallenge.value) {
    return []
  }
  return workspace.value.targets.map((target) => ({
    ...target,
    active_service: target.services.find((service) =>
      isTargetServiceForChallenge(service, activeChallenge.value as ContestChallengeItem)
    ),
  }))
})

const filteredTargets = computed(() =>
  activeTargets.value.filter((target) => {
    const matchesKeyword =
      targetFilterKeyword.value.length === 0 ||
      target.team_name.toLowerCase().includes(targetFilterKeyword.value)
    const matchesReachable =
      !showOnlyReachableTargets.value || Boolean(target.active_service?.access_url)
    return matchesKeyword && matchesReachable
  })
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
    let statusLabel = 'STABLE'
    let tone: 'danger' | 'warning' = 'warning'

    if (service.service_status === 'compromised') {
      issues.push('INFILTRATED')
      statusLabel = 'CRITICAL'
      tone = 'danger'
    } else if (service.service_status === 'down') {
      issues.push('OFFLINE')
      statusLabel = 'ALERT'
    }

    if ((service.attack_received ?? 0) > 0) {
      issues.push(`${service.attack_received} HITS DETECTED`)
    }

    if (issues.length === 0) continue

    items.push({
      challengeId: challenge.challenge_id,
      challengeTitle: challenge.title,
      statusLabel,
      tone,
      issues,
    })
  }
  return items
})

watch(
  () => runtimeChallenges.value.map((item) => getChallengeRuntimeKey(item)),
  (challengeKeys) => {
    if (challengeKeys.length === 0) {
      activeChallengeKey.value = ''
      return
    }
    if (!challengeKeys.includes(activeChallengeKey.value)) {
      activeChallengeKey.value = challengeKeys[0]
    }
  },
  { immediate: true }
)

function getServiceStatusLabel(status?: string): string {
  switch (status) {
    case 'up': return 'STABLE'
    case 'down': return 'OFFLINE'
    case 'compromised': return 'CRITICAL'
    default: return 'STANDBY'
  }
}

function getServiceStatusClass(status?: string): string {
  return `status-badge status-badge--${status || 'pending'}`
}

function eventDirectionLabel(direction: 'attack_in' | 'attack_out'): string {
  return direction === 'attack_out' ? 'OUTBOUND' : 'INBOUND'
}

function eventResultLabel(success: boolean): string {
  return success ? 'SUCCESS' : 'FAILED'
}

function formatServiceRef(serviceId?: string): string {
  return `SRV-${serviceId || '--'}`
}

function getChallengeRuntimeKey(challenge: ContestChallengeItem | null | undefined): string {
  return challenge?.awd_service_id || ''
}

function getServiceStartKey(challenge: ContestChallengeItem): string {
  return challenge.awd_service_id || ''
}

function getChallengeTitleForEvent(event: { service_id?: string; challenge_id: string }): string {
  if (event.service_id) {
    const matchedByService = challengeByServiceId.value.get(event.service_id)
    if (matchedByService) return matchedByService.title
  }
  return challengeByChallengeId.value.get(event.challenge_id)?.title || event.challenge_id
}

function getSubmitResultMessage(): string {
  if (!submitResult.value) return ''
  return formatAttackResultToast(submitResult.value)
}

function formatAttackResultToast(result: {
  service_id?: string
  challenge_id: string
  is_success: boolean
  score_gained: number
}): string {
  const challengeTitle = getChallengeTitleForEvent(result)
  if (result.is_success) return `${challengeTitle}: HIT SUCCESSFUL. +${result.score_gained} PTS`
  return `${challengeTitle}: NO VALID FLAG RECOVERED.`
}

function getWorkspaceService(challenge: ContestChallengeItem): ContestAWDWorkspaceServiceData | undefined {
  if (!challenge.awd_service_id) return undefined
  return servicesByServiceId.value.get(challenge.awd_service_id)
}

function isTargetServiceForChallenge(service: { service_id?: string; challenge_id: string }, challenge: ContestChallengeItem): boolean {
  return Boolean(challenge.awd_service_id) && service.service_id === challenge.awd_service_id
}

function buildAttackStateKey(serviceKey: string, teamId: string): string {
  return `${serviceKey}:${teamId}`
}

async function handleSubmit(serviceKey: string, teamId: string): Promise<void> {
  const stateKey = buildAttackStateKey(serviceKey, teamId)
  const flag = flagInputs.value[stateKey] || ''
  const result = await submitAttack(serviceKey, Number(teamId), flag)
  if (result) {
    flagInputs.value[stateKey] = ''
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

    <!-- HUD KPI Strip -->
    <header class="awd-hud-strip">
      <div class="hud-item">
        <div class="hud-label">
          OPERATIONAL ROUND
        </div>
        <div class="hud-value font-mono">
          {{ currentRound ? `#${String(currentRound.round_number).padStart(2, '0')}` : '--' }}
        </div>
        <div class="hud-helper">
          {{ currentRound?.status.toUpperCase() || 'WAITING' }}
        </div>
      </div>
      <div class="hud-item">
        <div class="hud-label">
          BATTLE SQUAD
        </div>
        <div class="hud-value">
          {{ myTeam?.team_name || 'UNASSIGNED' }}
        </div>
        <div class="hud-helper">
          RANK: #{{ scoreboardRows.find(r => r.team_id === myTeam?.team_id)?.rank || '--' }}
        </div>
      </div>
      <div class="hud-item">
        <div class="hud-label">
          SQUAD ASSETS
        </div>
        <div class="hud-value font-mono">
          {{ workspace?.services.length || 0 }}
        </div>
        <div class="hud-helper">
          ACTIVE SERVICES
        </div>
      </div>
      <div class="hud-item">
        <div class="hud-label">
          APEX SCORE
        </div>
        <div class="hud-value font-mono text-cyan-400">
          {{ topScore }}
        </div>
        <div class="hud-helper">
          BATTLEFIELD PEAK
        </div>
      </div>
      <div class="hud-actions">
        <button
          class="hud-refresh-btn"
          :disabled="loading"
          @click="refreshAll"
        >
          <RefreshCw
            class="h-4 w-4"
            :class="{ 'animate-spin': loading }"
          />
          <span>{{ lastSyncedLabel }}</span>
        </button>
      </div>
    </header>

    <div
      v-if="loading && !workspace"
      class="war-room-loading"
    >
      <div class="radar-scan" />
      <p>ESTABLISHING BATTLEFIELD LINK...</p>
    </div>

    <AppEmpty
      v-else-if="!hasTeam"
      icon="Users"
      title="JOIN A SQUAD"
      description="You must be part of a team to access the AWD battlefield."
      class="war-room-empty"
    />

    <div
      v-else
      class="war-room-grid"
    >
      <!-- 1. Defense Monitor (Left) -->
      <aside class="war-room-col column-defense">
        <section class="ops-panel">
          <header class="ops-panel__header">
            <ShieldAlert class="h-4 w-4 text-orange-500" />
            <h3 class="ops-panel__title">
              DEFENSE MONITOR
            </h3>
          </header>
          
          <div class="ops-panel__content custom-scrollbar">
            <!-- Alerts -->
            <div
              v-if="defenseAlerts.length > 0"
              class="defense-alerts"
            >
              <div
                v-for="alert in defenseAlerts"
                :key="alert.challengeId"
                class="defense-alert"
                :class="alert.tone"
              >
                <div class="flex items-center justify-between">
                  <span class="alert-title">{{ alert.challengeTitle }}</span>
                  <span class="alert-badge">{{ alert.statusLabel }}</span>
                </div>
                <div class="alert-issues">
                  <span
                    v-for="issue in alert.issues"
                    :key="issue"
                  >{{ issue }}</span>
                </div>
              </div>
            </div>

            <!-- Services -->
            <div class="asset-list mt-4">
              <div class="asset-header">
                SQUAD SERVICES
              </div>
              <div
                v-for="challenge in runtimeChallenges"
                :key="challenge.id"
                class="asset-item"
              >
                <div class="asset-main">
                  <div class="flex items-center justify-between">
                    <span class="asset-title">{{ challenge.title }}</span>
                    <span :class="getServiceStatusClass(getWorkspaceService(challenge)?.service_status)">
                      {{ getServiceStatusLabel(getWorkspaceService(challenge)?.service_status) }}
                    </span>
                  </div>
                  <div class="asset-meta font-mono text-[10px]">
                    {{ getWorkspaceService(challenge)?.access_url || 'WAITING FOR ALLOCATION' }}
                  </div>
                </div>
                <div class="asset-actions">
                  <a
                    v-if="getWorkspaceService(challenge)?.access_url"
                    :href="getWorkspaceService(challenge)?.access_url"
                    target="_blank"
                    class="asset-btn"
                  ><ExternalLink class="h-3 w-3" /></a>
                  <button
                    :disabled="startingServiceKey === getServiceStartKey(challenge)"
                    class="asset-btn asset-btn--primary"
                    @click="challenge.awd_service_id && startService(challenge.awd_service_id)"
                  >
                    {{ startingServiceKey === getServiceStartKey(challenge) ? '...' : 'REBOOT' }}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </section>
      </aside>

      <!-- 2. Attack Vector (Middle) -->
      <main class="war-room-col column-attack">
        <section class="ops-panel">
          <header class="ops-panel__header">
            <Sword class="h-4 w-4 text-red-500" />
            <h3 class="ops-panel__title">
              ATTACK VECTOR
            </h3>
          </header>

          <div class="ops-panel__toolbar">
            <div class="toolbar-field">
              <label>TARGET SECTOR</label>
              <select
                v-model="activeChallengeKey"
                class="war-room-select"
              >
                <option
                  v-for="challenge in runtimeChallenges"
                  :key="getChallengeRuntimeKey(challenge)"
                  :value="getChallengeRuntimeKey(challenge)"
                >
                  {{ challenge.title }}
                </option>
              </select>
            </div>
            <div class="toolbar-field">
              <label>SQUAD FILTER</label>
              <input
                v-model="targetKeyword"
                type="text"
                placeholder="FILTER BY NAME..."
                class="war-room-input"
              >
            </div>
          </div>

          <div class="ops-panel__content custom-scrollbar">
            <div
              v-if="!activeChallenge"
              class="panel-note"
            >
              SELECT A TARGET SECTOR TO COMMENCE ATTACK.
            </div>
            <div
              v-else-if="filteredTargets.length === 0"
              class="panel-note"
            >
              NO MATCHING HOSTILES IN CURRENT SECTOR.
            </div>
            <div
              v-else
              class="target-grid"
            >
              <article
                v-for="target in filteredTargets"
                :key="target.team_id"
                class="target-card"
              >
                <div class="target-info">
                  <div class="target-team font-black text-cyan-400">
                    {{ target.team_name.toUpperCase() }}
                  </div>
                  <div class="target-url font-mono">
                    {{ target.active_service?.access_url || 'UNREACHABLE' }}
                  </div>
                </div>
                <div class="target-action">
                  <input
                    :value="flagInputs[buildAttackStateKey(activeChallengeRuntimeKey, target.team_id)] || ''"
                    placeholder="ENTER STOLEN FLAG..."
                    class="flag-input"
                    @input="flagInputs[buildAttackStateKey(activeChallengeRuntimeKey, target.team_id)] = String(($event.target as HTMLInputElement).value)"
                    @keyup.enter="handleSubmit(activeChallengeRuntimeKey, target.team_id)"
                  >
                  <button
                    :disabled="!target.active_service?.access_url || submittingKey === buildAttackStateKey(activeChallengeRuntimeKey, target.team_id)"
                    class="submit-btn"
                    @click="handleSubmit(activeChallengeRuntimeKey, target.team_id)"
                  >
                    {{ submittingKey === buildAttackStateKey(activeChallengeRuntimeKey, target.team_id) ? '...' : 'SUBMIT' }}
                  </button>
                </div>
              </article>
            </div>
          </div>

          <footer
            v-if="submitResult"
            class="ops-panel__footer"
          >
            <div
              class="result-alert"
              :class="submitResult.is_success ? 'success' : 'danger'"
            >
              <Terminal class="h-3.5 w-3.5" />
              <span>{{ getSubmitResultMessage().toUpperCase() }}</span>
            </div>
          </footer>
        </section>
      </main>

      <!-- 3. Intelligence (Right) -->
      <aside class="war-room-col column-intel">
        <section class="ops-panel h-1/2 mb-4">
          <header class="ops-panel__header">
            <BarChart3 class="h-4 w-4 text-cyan-500" />
            <h3 class="ops-panel__title">
              FIELD INTEL
            </h3>
          </header>
          <div class="ops-panel__content custom-scrollbar">
            <div
              v-for="item in scoreboardRows.slice(0, 10)"
              :key="item.team_id"
              class="intel-row"
              :class="{ 'is-me': item.team_id === myTeam?.team_id }"
            >
              <span class="intel-rank">#{{ item.rank }}</span>
              <span class="intel-name truncate">{{ item.team_name }}</span>
              <span class="intel-score font-mono">{{ item.score }}</span>
            </div>
          </div>
        </section>

        <section class="ops-panel h-1/2">
          <header class="ops-panel__header">
            <History class="h-4 w-4 text-purple-500" />
            <h3 class="ops-panel__title">
              RECENT FEEDBACK
            </h3>
          </header>
          <div class="ops-panel__content custom-scrollbar">
            <div
              v-for="event in workspace?.recent_events"
              :key="event.id"
              class="feedback-item"
              :class="event.direction"
            >
              <div class="flex items-center justify-between text-[10px] font-black">
                <span>{{ eventDirectionLabel(event.direction) }}</span>
                <span>{{ formatTime(event.created_at) }}</span>
              </div>
              <div class="mt-1 text-xs">
                {{ event.peer_team_name }} / {{ getChallengeTitleForEvent(event) }}
              </div>
              <div class="mt-1 flex items-center justify-between font-mono text-[10px]">
                <span :class="event.is_success ? 'text-emerald-400' : 'text-slate-500'">{{ eventResultLabel(event.is_success) }}</span>
                <span class="text-cyan-400">+{{ event.score_gained }}</span>
              </div>
            </div>
            <div
              v-if="workspace?.recent_events.length === 0"
              class="panel-note"
            >
              NO RECENT OPERATIONAL DATA.
            </div>
          </div>
        </section>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.awd-war-room {
  --war-bg: #0f172a;
  --war-panel: rgba(30, 41, 59, 0.7);
  --war-border: #334155;
  --war-accent: #22d3ee;
  --war-text: #f8fafc;
  --war-muted: #94a3b8;
  
  background: var(--war-bg);
  color: var(--war-text);
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1rem;
}

/* HUD Strip */
.awd-hud-strip {
  display: grid;
  grid-template-columns: repeat(4, 1fr) auto;
  gap: 1rem;
  background: var(--war-panel);
  border: 1px solid var(--war-border);
  border-radius: 0.5rem;
  padding: 0.75rem 1.5rem;
  backdrop-filter: blur(10px);
}

.hud-item {
  display: flex;
  flex-direction: column;
}

.hud-label {
  font-size: 9px;
  font-weight: 900;
  color: var(--war-muted);
  letter-spacing: 0.1em;
}

.hud-value {
  font-size: 1.15rem;
  font-weight: 900;
  margin: 0.15rem 0;
}

.hud-helper {
  font-size: 10px;
  font-weight: 700;
  color: var(--war-accent);
}

.hud-refresh-btn {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
  padding: 0 1rem;
  border-left: 1px solid var(--war-border);
  color: var(--war-muted);
  font-size: 10px;
  font-weight: 700;
  cursor: pointer;
}

.hud-refresh-btn:hover { color: var(--war-text); }

/* Layout Grid */
.war-room-grid {
  display: grid;
  grid-template-columns: 18rem 1fr 18rem;
  gap: 1rem;
  flex: 1;
  min-height: 0;
}

.ops-panel {
  background: var(--war-panel);
  border: 1px solid var(--war-border);
  border-radius: 0.5rem;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.ops-panel__header {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--war-border);
  display: flex;
  align-items: center;
  gap: 0.75rem;
  background: rgba(0,0,0,0.1);
}

.ops-panel__title {
  font-size: 11px;
  font-weight: 900;
  letter-spacing: 0.15em;
  color: var(--war-text);
  margin: 0;
}

.ops-panel__toolbar {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--war-border);
  display: flex;
  gap: 1rem;
}

.toolbar-field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex: 1;
}

.toolbar-field label {
  font-size: 8px;
  font-weight: 900;
  color: var(--war-muted);
}

.war-room-select, .war-room-input {
  background: #0f172a;
  border: 1px solid var(--war-border);
  border-radius: 0.25rem;
  color: var(--war-text);
  font-size: 11px;
  font-weight: 700;
  padding: 0.35rem;
  outline: none;
}

.ops-panel__content {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
}

/* Defense Components */
.defense-alert {
  background: rgba(249, 115, 22, 0.05);
  border: 1px solid rgba(249, 115, 22, 0.2);
  border-left: 3px solid #f97316;
  border-radius: 0.25rem;
  padding: 0.5rem 0.75rem;
  margin-bottom: 0.5rem;
}

.defense-alert.danger {
  background: rgba(239, 68, 68, 0.05);
  border-color: rgba(239, 68, 68, 0.2);
  border-left-color: #ef4444;
}

.alert-title { font-size: 11px; font-weight: 800; }
.alert-badge { font-size: 9px; font-weight: 900; }
.alert-issues { font-size: 9px; color: var(--war-muted); margin-top: 0.25rem; }

.asset-header {
  font-size: 9px;
  font-weight: 900;
  color: var(--war-muted);
  margin-bottom: 0.5rem;
}

.asset-item {
  background: rgba(255,255,255,0.02);
  border: 1px solid var(--war-border);
  padding: 0.75rem;
  border-radius: 0.25rem;
  margin-bottom: 0.5rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.asset-title { font-size: 12px; font-weight: 800; }
.asset-meta { color: var(--war-muted); margin-top: 0.25rem; }

.status-badge {
  font-size: 9px;
  font-weight: 900;
  padding: 0.15rem 0.45rem;
  border-radius: 99px;
}

.status-badge--up { background: rgba(16, 185, 129, 0.1); color: #10b981; }
.status-badge--down { background: rgba(239, 68, 68, 0.1); color: #ef4444; }
.status-badge--compromised { background: rgba(249, 115, 22, 0.1); color: #f97316; }

.asset-btn {
  width: 2rem;
  height: 2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 0.25rem;
  background: var(--war-border);
  color: var(--war-text);
  border: none;
  cursor: pointer;
}

.asset-btn--primary {
  width: auto;
  padding: 0 0.75rem;
  font-size: 10px;
  font-weight: 900;
  background: var(--war-accent);
  color: #0f172a;
}

/* Attack Components */
.target-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(20rem, 1fr));
  gap: 1rem;
}

.target-card {
  background: rgba(0,0,0,0.2);
  border: 1px solid var(--war-border);
  padding: 1rem;
  border-radius: 0.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.target-team { font-size: 13px; letter-spacing: 0.05em; }
.target-url { font-size: 10px; color: var(--war-muted); }

.target-action {
  display: flex;
  gap: 0.5rem;
}

.flag-input {
  flex: 1;
  background: #0f172a;
  border: 1px solid var(--war-border);
  padding: 0.5rem;
  border-radius: 0.25rem;
  color: var(--war-text);
  font-family: var(--font-family-mono);
  font-size: 12px;
  outline: none;
}

.submit-btn {
  background: #ef4444;
  color: white;
  border: none;
  padding: 0 1rem;
  border-radius: 0.25rem;
  font-weight: 900;
  font-size: 11px;
  cursor: pointer;
}

/* Intel Components */
.intel-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 0;
  border-bottom: 1px solid rgba(255,255,255,0.02);
  font-size: 12px;
}

.intel-row.is-me { color: var(--war-accent); }
.intel-rank { font-weight: 900; color: var(--war-muted); width: 2rem; }
.intel-name { flex: 1; font-weight: 700; }
.intel-score { font-weight: 900; }

.feedback-item {
  background: rgba(255,255,255,0.01);
  padding: 0.65rem;
  border-radius: 0.25rem;
  border-left: 2px solid #334155;
  margin-bottom: 0.5rem;
}

.feedback-item.attack_out { border-left-color: #ef4444; }
.feedback-item.attack_in { border-left-color: #f97316; }

.panel-note { font-size: 11px; color: var(--war-muted); text-align: center; padding: 2rem 0; }

.ops-panel__footer {
  padding: 0.75rem 1rem;
  border-top: 1px solid var(--war-border);
}

.result-alert {
  padding: 0.5rem 0.75rem;
  border-radius: 0.25rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 11px;
  font-weight: 900;
}

.result-alert.success { background: rgba(16, 185, 129, 0.1); color: #10b981; }
.result-alert.danger { background: rgba(239, 68, 68, 0.1); color: #ef4444; }

.war-room-loading {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2rem;
  font-size: 12px;
  font-weight: 900;
  color: var(--war-accent);
}

.radar-scan {
  width: 4rem;
  height: 4rem;
  border: 2px solid var(--war-accent);
  border-radius: 50%;
  position: relative;
}

.radar-scan::after {
  content: '';
  position: absolute;
  inset: 0;
  background: conic-gradient(from 0deg, var(--war-accent), transparent);
  border-radius: 50%;
  animation: radar 2s linear infinite;
}

@keyframes radar { to { transform: rotate(360deg); } }

@media (max-width: 1280px) {
  .war-room-grid { grid-template-columns: 1fr; }
  .column-defense, .column-intel { grid-row: auto; }
}
</style>
